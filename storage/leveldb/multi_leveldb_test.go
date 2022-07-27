package leveldb

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const multiLeveldbPath = "./ledger"

func TestMultiLdb_New(t *testing.T) {
	// 路径下为空，创建一个新的多层leveldb
	os.RemoveAll(multiLeveldbPath)
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)
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
	os.RemoveAll(multiLeveldbPath)
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)

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
	os.RemoveAll(multiLeveldbPath)
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)

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
}

func TestMultiLdb_Delete(t *testing.T) {
	os.RemoveAll(multiLeveldbPath)
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)

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
	os.RemoveAll(multiLeveldbPath)
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)

	// 对同一个key写入两次不同的值，新值在最上层，旧值不在最上层
	mLdb.Put([]byte("key"), []byte("0123456789ABCDEF"))
	for i := 0; i < 10000; i++ {
		mLdb.Put([]byte(fmt.Sprintf("%d", i)), []byte("0123456789ABCDEF"))
	}
	mLdb.Put([]byte("key"), []byte("0123456789")) // 写入最上层

	// 迭代器中同一个key可能会出现重复的值，且新值在前
	it := mLdb.Iterator([]byte("key"), []byte("kez"))
	it.Next()
	assert.Equal(t, []byte("key"), it.Key())
	assert.Equal(t, []byte("0123456789"), it.Value())
	it.Next()
	assert.Equal(t, []byte("key"), it.Key())
	assert.Equal(t, []byte("0123456789ABCDEF"), it.Value())
}

func TestMultiLdbBatch_Commit(t *testing.T) {
	os.RemoveAll(multiLeveldbPath)
	mLdb, err := NewMultiLdb(multiLeveldbPath, &opt.Options{
		WriteBuffer: opt.KiB,
	}, 10*1024)
	require.Nil(t, err)

	// 检查put
	for i := 0; i < 10; i++ {
		batch := mLdb.NewBatch()
		for j := 0; j < 1000; j++ {
			batch.Put([]byte(fmt.Sprintf("%d-%d", i, j)), []byte("0123456789ABCDEF"))
		}
		st := time.Now()
		batch.Commit()
		fmt.Printf("commit elapse: %s\n", time.Since(st))
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
