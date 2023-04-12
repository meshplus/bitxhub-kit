package leveldb

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"sync"

	"github.com/meshplus/bitxhub-kit/storage"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/comparer"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	layerNamePrefix = "leveldb" // the prefix of leveldb name at each layer
)

type multiLdb struct {
	dbList        []*leveldb.DB // the i-th db is i-th layer, the last db is top layer, the first db (0-th db) is top layer
	path          string        // the path of multi-leveldb
	sizeThreshold int64         // the threshold of size for each layer leveldb (Byte)
	opt           *opt.Options  // option of each layer leveldb
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
		path:          dirPath,
		sizeThreshold: sizeThreshold,
		opt:           opt,
	}

	// if it's empty under path, create 0-th layer leveldb
	if len(fList) == 0 {
		db, err := leveldb.OpenFile(mLdb.getLayerPath(0), mLdb.opt)
		if err != nil {
			return nil, err
		}
		mLdb.dbList = append(mLdb.dbList, db)
		return mLdb, nil
	}

	// if it's not empty, connect exist each layer leveldb sequentially
	sort.Slice(fList, func(i, j int) bool {
		// sort by length firstly
		return len(fList[i]) < len(fList[j]) || (len(fList[i]) == len(fList[j]) && fList[i] < fList[j])
	})
	for i := 0; i < len(fList); i++ {
		// if dir name is incorrect, return error
		// e.g. '/leveldb0' <-> 0-th layer leveldb,  '/leveldb1' <-> 1-th layer leveldb
		if fList[i] != mLdb.getLayerPath(i) {
			return nil, fmt.Errorf("missing file or filename error under %s: expect %s, get %s",
				mLdb.path, mLdb.getLayerPath(i), fList[i])
		}
		// connect i-th layer leveldb
		db, err := leveldb.OpenFile(mLdb.getLayerPath(i), opt)
		if err != nil {
			return nil, err
		}
		mLdb.dbList = append(mLdb.dbList, db)
	}

	return mLdb, nil
}

// getLayerPath get path of i-th layer leveldb
func (l *multiLdb) getLayerPath(i int) string {
	return path.Join(l.path, fmt.Sprintf("/%s%d", layerNamePrefix, i))
}

// getLayers get layers in order from top to bottom
func (l *multiLdb) getLayers() []*leveldb.DB {
	layers := make([]*leveldb.DB, 0)
	// the last db in l.dbList is top layer leveldb, the first db (0-th db) is bottom layer
	for i := len(l.dbList) - 1; i >= 0; i-- {
		layers = append(layers, l.dbList[i])
	}
	return layers
}

// getTopLayer get top layer leveldb
func (l *multiLdb) getTopLayer() (*leveldb.DB, error) {
	if len(l.dbList) == 0 {
		return nil, fmt.Errorf("dbList length is 0")
	}
	return l.dbList[len(l.dbList)-1], nil
}

// addTopLayer add new leveldb as top layer
func (l *multiLdb) addTopLayer(curLayerCnt int) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// when several goroutine call addTopLayer, only one goroutine can success
	if len(l.dbList) > curLayerCnt {
		return
	}

	// create new leveldb, len(l.dbList) is the index of new leveldb
	db, err := leveldb.OpenFile(l.getLayerPath(len(l.dbList)), l.opt)
	if err != nil {
		return
	}

	// append new leveldb to l.dbList, then it becomes the top layer
	l.dbList = append(l.dbList, db)
}

// checkTopLayerSize check the size of top layer leveldb
func (l *multiLdb) checkTopLayerSize() error {
	db, err := l.getTopLayer()
	if err != nil {
		return err
	}
	stats := leveldb.DBStats{}
	if err := db.Stats(&stats); err != nil {
		return err
	}

	// if size of top layer bigger than l.sizeThreshold, call addTopLayer
	if stats.LevelSizes.Sum() > l.sizeThreshold {
		go l.addTopLayer(len(l.dbList))
	}

	return nil
}

// GetStats get stats of each layer from top to bottom, return []*leveldb.DBStats
func (l *multiLdb) GetStats() (interface{}, error) {
	statesList := make([]*leveldb.DBStats, 0)
	for _, db := range l.getLayers() {
		stats := &leveldb.DBStats{}
		if err := db.Stats(stats); err != nil {
			return nil, err
		}
		statesList = append(statesList, stats)
	}
	return statesList, nil
}

type KeyValueEntry struct {
	key, value []byte
}

// myArray use for iterator
type myArray struct {
	entries []KeyValueEntry
	bytes   int
	cmp     comparer.Comparer
}

func (a *myArray) Len() int {
	return len(a.entries)
}

func (a *myArray) Search(key []byte) int {
	return sort.Search(a.Len(), func(i int) bool {
		k, _ := a.Index(i)
		return a.cmp.Compare(k, key) >= 0
	})
}

func (a *myArray) Index(i int) (key, value []byte) {
	if i < 0 || i >= len(a.entries) {
		return nil, nil
	}
	return a.entries[i].key, a.entries[i].value
}

func (a *myArray) Put(key, value []byte) {
	a.entries = append(a.entries, KeyValueEntry{key, value})
	a.bytes += len(key) + len(value)
}

// iterator merge iterator in each layer. For the same key, only the latest value is returned
func (l *multiLdb) iterator(rg *util.Range) storage.Iterator {
	arr := &myArray{cmp: l.opt.GetComparer()}
	m := make(map[string]bool)
	// iterate from top to bottom. for the same key, the latest value is in higher layer
	for _, db := range l.getLayers() {
		it := db.NewIterator(rg, nil)
		for it.Next() {
			// if key appears for the first time, append key-value to arr
			if _, ok := m[string(it.Key())]; !ok {
				// it.key() and it.value() return reference. copy firstly, then put to arr
				key := make([]byte, len(it.Key()))
				val := make([]byte, len(it.Value()))
				copy(key, it.Key())
				copy(val, it.Value())
				m[string(key)] = true
				arr.Put(key, val)
			}
		}
	}

	return &iter{iter: iterator.NewArrayIterator(arr)}
}

// Put only put to top layer
func (l *multiLdb) Put(key, value []byte) {
	db, err := l.getTopLayer()
	if err != nil {
		panic(err)
	}

	if err := db.Put(key, value, nil); err != nil {
		panic(err)
	}

	if err := l.checkTopLayerSize(); err != nil {
		panic(err)
	}
}

// Delete delete in each layer
func (l *multiLdb) Delete(key []byte) {
	for _, db := range l.getLayers() {
		if err := db.Delete(key, nil); err != nil {
			panic(err)
		}
	}
}

// Get get from top to bottom
func (l *multiLdb) Get(key []byte) []byte {
	for _, db := range l.getLayers() {
		val, err := db.Get(key, nil)
		if err == nil {
			// if current layer get key, return
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

func (l *multiLdb) Iterator(start, end []byte) storage.Iterator {
	return l.iterator(&util.Range{
		Start: start,
		Limit: end,
	})
}

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
	// putBatch write to top layer
	db, err := b.mLdb.getTopLayer()
	if err != nil {
		panic(err)
	}
	if err := db.Write(b.putBatch, nil); err != nil {
		panic(err)
	}

	// delBatch write to each layer
	for _, db := range b.mLdb.getLayers() {
		if err := db.Write(b.delBatch, nil); err != nil {
			panic(err)
		}
	}

	if err := b.mLdb.checkTopLayerSize(); err != nil {
		panic(err)
	}
}
