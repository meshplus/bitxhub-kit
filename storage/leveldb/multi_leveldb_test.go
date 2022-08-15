package leveldb

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const multiLeveldbPath = "./test_leveldb"

func TestMultiLdb_New(t *testing.T) {
	// 路径下为空，创建一个新的多层leveldb
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	defer os.RemoveAll(multiLeveldbPath)
	mLdb.Put([]byte("key"), []byte("0123456789ABCDEF"))
	for i := 0; i < 10000; i++ {
		mLdb.Put([]byte(fmt.Sprintf("%d", i)), []byte("0123456789ABCDEF"))
	}
	mLdb.Put([]byte("key"), []byte("0123456789")) // 写入最上层

	mLdb.Close()

	// 路径下不为空，连接这个已存在的多层leveldb
	mLdb, err = NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	assert.Equal(t, []byte("0123456789"), mLdb.Get([]byte("key"))) // 获取的数据应为最上层的新数据
}

func TestMultiLdb_Put(t *testing.T) {
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	defer os.RemoveAll(multiLeveldbPath)

	// 写
	mLdb.Put([]byte("key"), []byte("0123456789ABCDEF"))
	assert.Equal(t, []byte("0123456789ABCDEF"), mLdb.Get([]byte("key")))

	// 多次写，检查是否触发分层
	for i := 0; i < 10000; i++ {
		// st := time.Now()
		mLdb.Put([]byte(fmt.Sprintf("%d", i)), []byte("0123456789ABCDEF"))
		// fmt.Printf("put elapse: %s\n", time.Since(st))
	}

	// 修改
	mLdb.Put([]byte("key"), []byte("0123"))
	assert.Equal(t, []byte("0123"), mLdb.Get([]byte("key")))
}

func TestMultiLdb_Get(t *testing.T) {
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	defer os.RemoveAll(multiLeveldbPath)

	// 读不存在
	assert.Nil(t, mLdb.Get([]byte("key1")))

	// 读刚写入的数据（数据在最上层）
	mLdb.Put([]byte("key1"), []byte("0123456789ABCDEF"))
	assert.Equal(t, []byte("0123456789ABCDEF"), mLdb.Get([]byte("key1")))

	// 读此前写入的数据（数据不在最上层）
	for i := 0; i < 10000; i++ {
		mLdb.Put([]byte(fmt.Sprintf("%d", i)), []byte("0123456789ABCDEF"))
	}
	assert.Equal(t, []byte("0123456789ABCDEF"), mLdb.Get([]byte("key1")))

	// 读修改后的新值
	mLdb.Put([]byte("key1"), []byte("0123456789"))
	assert.Equal(t, []byte("0123456789"), mLdb.Get([]byte("key1")))
}

func TestMultiLdb_Delete(t *testing.T) {
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	defer os.RemoveAll(multiLeveldbPath)

	// 删除刚写入的数据（数据在最上层）
	mLdb.Put([]byte("key1"), []byte("0123456789ABCDEF"))
	assert.Equal(t, []byte("0123456789ABCDEF"), mLdb.Get([]byte("key1")))
	mLdb.Delete([]byte("key1"))
	assert.Nil(t, mLdb.Get([]byte("key1")))

	// 删除此前写入的数据（数据不在最上层）
	mLdb.Put([]byte("key2"), []byte("0123456789ABCDEF"))
	assert.Equal(t, []byte("0123456789ABCDEF"), mLdb.Get([]byte("key2")))
	for i := 0; i < 10000; i++ {
		mLdb.Put([]byte(fmt.Sprintf("%d", i)), []byte("0123456789ABCDEF"))
	}
	mLdb.Delete([]byte("key2"))
	assert.Nil(t, mLdb.Get([]byte("key2")))
}

func TestMultiLdb_Iterator(t *testing.T) {
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	defer os.RemoveAll(multiLeveldbPath)

	// 对同一个key写入两次不同的值，新值在最上层，旧值不在最上层
	mLdb.Put([]byte("key"), []byte("0123456789ABCDEF"))
	for i := 0; i < 10000; i++ {
		mLdb.Put([]byte(fmt.Sprintf("%d", i)), []byte("0123456789ABCDEF"))
	}
	mLdb.Put([]byte("key"), []byte("0123456789")) // 写入最上层

	// 对同一个key，迭代器中只存在最新值
	it := mLdb.Iterator([]byte("key"), []byte("kez"))
	it.Next()
	assert.Equal(t, []byte("key"), it.Key())
	assert.Equal(t, []byte("0123456789"), it.Value())
	assert.Equal(t, false, it.Next())

	// 范围匹配
	for i := 0; i < 3; i++ {
		mLdb.Put([]byte(fmt.Sprintf("key%d", i)), []byte("0123456789ABCDEF"))
	}
	it = mLdb.Iterator([]byte("key0"), []byte("key3"))
	for i := 0; i < 3; i++ {
		it.Next()
		assert.Equal(t, []byte(fmt.Sprintf("key%d", i)), it.Key())
		assert.Equal(t, []byte("0123456789ABCDEF"), it.Value())
	}
}

func TestMultiLdbBatch_Commit(t *testing.T) {
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	defer os.RemoveAll(multiLeveldbPath)

	// 检查put
	for i := 0; i < 10; i++ {
		batch := mLdb.NewBatch()
		for j := 0; j < 1000; j++ {
			batch.Put([]byte(fmt.Sprintf("%d-%d", i, j)), []byte("0123456789ABCDEF"))
		}
		batch.Commit()
	}
	assert.Equal(t, []byte("0123456789ABCDEF"), mLdb.Get([]byte("0-0")))
	assert.Equal(t, []byte("0123456789ABCDEF"), mLdb.Get([]byte("9-999")))

	// 检查delete
	batch := mLdb.NewBatch()
	for i := 0; i < 10; i++ {
		batch.Delete([]byte(fmt.Sprintf("%d-0", i)))
	}
	batch.Commit()
	assert.Nil(t, mLdb.Get([]byte("0-0")))
}

func TestMultiLdb_GetStats(t *testing.T) {
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
	defer os.RemoveAll(multiLeveldbPath)

	for i := 0; i < 10000; i++ {
		mLdb.Put([]byte(fmt.Sprintf("%d", i)), []byte("0123456789ABCDEF"))
	}
	statsList, err := mLdb.GetStats()
	require.Nil(t, err)
	for i, stats := range statsList.([]*leveldb.DBStats) {
		fmt.Printf("layer %d size: %d\n", i, stats.LevelSizes.Sum())
	}
}
