package leveldb

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type multiLdb struct {
	dbList        []*leveldb.DB // 第i个代表第i层(i>=0)，最后一个代表最上层
	dirPath       string        // 数据库文件所在目录的路径
	sizeThreshold int64         // 每层leveldb的大小阈值（字节为单位）
	opt           *opt.Options  // leveldb初始化参数
	mu            sync.Mutex
}

// NewMultiLdb New a multi layer leveldb.
// When size of top layer leveldb exceeds sizeThreshold(Byte), it will add a new layer leveldb above the top layer.
func NewMultiLdb(dirPath string, opt *opt.Options, sizeThreshold int64) (storage.Storage, error) {
	fList, err := filepath.Glob(path.Join(dirPath, "*"))
	if err != nil {
		return nil, err
	}

	mLdb := &multiLdb{
		dbList:        make([]*leveldb.DB, 0),
		dirPath:       dirPath,
		sizeThreshold: sizeThreshold,
		opt:           opt,
	}

	// 如果路径下没有文件，新增第0层leveldb
	if len(fList) == 0 {
		mLdb.addTopLayer(0)
		return mLdb, nil
	}

	// 如果路径下有文件，依次连接各层leveldb
	sort.Slice(fList, func(i, j int) bool {
		// 先按照长度排序，再按字典序排序
		return len(fList[i]) < len(fList[j]) || (len(fList[i]) == len(fList[j]) && fList[i] < fList[j])
	})
	// 从 0 到 len(fList)-1 依次启动各层leveldb。若文件顺序不匹配，返回错误
	for i := 0; i < len(fList); i++ {
		if fList[i] != path.Join(dirPath, fmt.Sprintf("/%d", i)) {
			return nil, fmt.Errorf("missing file or filename error under %s: expect %s, get %s",
				dirPath, path.Join(dirPath, fmt.Sprintf("/%d", i)), fList[i])
		}
		db, err := leveldb.OpenFile(path.Join(dirPath, fmt.Sprintf("/%d", i)), opt)
		if err != nil {
			return nil, err
		}
		mLdb.dbList = append(mLdb.dbList, db)
	}

	return mLdb, nil
}

// getLayers 获取各层leveldb，从最上层往下层排序（dbList最后一个元素为最上层，第一个元素为最下层）
func (l *multiLdb) getLayers() []*leveldb.DB {
	layers := make([]*leveldb.DB, 0)
	for i := len(l.dbList) - 1; i >= 0; i-- {
		layers = append(layers, l.dbList[i])
	}
	return layers
}

// getTopLayer 获取最上层leveldb
func (l *multiLdb) getTopLayer() *leveldb.DB {
	if len(l.dbList) == 0 {
		panic(fmt.Errorf("dbList length is 0"))
	}
	return l.dbList[len(l.dbList)-1]
}

// addTopLayer 新增一层leveldb作为最上层，需传入当前层数
func (l *multiLdb) addTopLayer(curLayerCnt int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 多个线程同时调用 addTopLayer 时，只允许一个成功
	if len(l.dbList) > curLayerCnt {
		return
	}

	st := time.Now()
	// 新增一层leveldb
	index := len(l.dbList) // 新增leveldb的index
	db, err := leveldb.OpenFile(path.Join(l.dirPath, fmt.Sprintf("/%d", index)), l.opt)
	if err != nil {
		panic(err)
	}
	fmt.Printf("new leveldb elapse: %s\n", time.Since(st))

	// 新增的leveldb添加到dbList最后
	l.dbList = append(l.dbList, db)
}

// checkTopLayerSize 检查最上层leveldb的大小
func (l *multiLdb) checkTopLayerSize() {
	stats := leveldb.DBStats{}
	if err := l.getTopLayer().Stats(&stats); err != nil {
		panic(err)
	}

	// 如果最上层leveldb大小超过阈值，则新增一层
	if stats.LevelSizes.Sum() > l.sizeThreshold {
		fmt.Println("exceed size threshold")
		go l.addTopLayer(len(l.dbList))
	}
}

// iterator 将各层iterator拼接，可能会存在重复的key，新值在前，旧值在后
func (l *multiLdb) iterator(rg *util.Range) storage.Iterator {
	its := make([]iterator.Iterator, 0)
	for _, db := range l.getLayers() {
		it := db.NewIterator(rg, nil)
		its = append(its, it)
	}

	return &iter{iter: iterator.NewMergedIterator(its, l.opt.GetComparer(), true)}
}

// Put 只写入最上层
func (l *multiLdb) Put(key, value []byte) {
	db := l.getTopLayer()
	if err := db.Put(key, value, nil); err != nil {
		panic(err)
	}

	l.checkTopLayerSize()
}

// Delete 各层都需执行删除
func (l *multiLdb) Delete(key []byte) {
	for _, db := range l.getLayers() {
		if err := db.Delete(key, nil); err != nil {
			panic(err)
		}
	}
}

// Get 从最上层往下查询，查到直接返回
func (l *multiLdb) Get(key []byte) []byte {
	for _, db := range l.getLayers() {
		val, err := db.Get(key, nil)
		if err == nil {
			// 查到了直接返回
			return val
		} else if err != errors.ErrNotFound {
			panic(err)
		}
	}
	return nil
}

func (l *multiLdb) Has(key []byte) bool {
	return l.Get(key) != nil
}

// Iterator 可能会存在重复的key，新值在前，旧值在后
func (l *multiLdb) Iterator(start, end []byte) storage.Iterator {
	return l.iterator(&util.Range{
		Start: start,
		Limit: end,
	})
}

// Prefix 可能会存在重复的key，新值在前，旧值在后
func (l *multiLdb) Prefix(prefix []byte) storage.Iterator {
	return l.iterator(util.BytesPrefix(prefix))
}

func (l *multiLdb) Close() error {
	for _, db := range l.getLayers() {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLdb) NewBatch() storage.Batch {
	return &multiLdbBatch{
		mLdb:     l,
		putBatch: &leveldb.Batch{},
		delBatch: &leveldb.Batch{},
	}
}

type multiLdbBatch struct {
	mLdb     *multiLdb
	putBatch *leveldb.Batch
	delBatch *leveldb.Batch
}

func (b *multiLdbBatch) Put(key, value []byte) {
	b.putBatch.Put(key, value)
}

func (b *multiLdbBatch) Delete(key []byte) {
	b.delBatch.Delete(key)
}

func (b *multiLdbBatch) Commit() {
	// putBatch写入最上层
	if err := b.mLdb.getTopLayer().Write(b.putBatch, nil); err != nil {
		panic(err)
	}

	// delBatch写入每一层
	for _, db := range b.mLdb.getLayers() {
		if err := db.Write(b.delBatch, nil); err != nil {
			panic(err)
		}
	}

	b.mLdb.checkTopLayerSize()
}
