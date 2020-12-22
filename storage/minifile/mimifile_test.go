package minifile

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBatchFile(t *testing.T) {
	path, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)

	b, err := New(path)
	assert.Nil(t, err)
	assert.Equal(t, path, b.path)

	err = b.Close()
	assert.Nil(t, err)

	b, err = New("")
	assert.Nil(t, err)

	err = b.Close()
	assert.Nil(t, err)

	b, err = New(".")
	assert.Nil(t, err)
	assert.NotEqual(t, ".", b.path)

	b, err = New(".")
	assert.NotNil(t, err)
}

func TestBatchFile_Put(t *testing.T) {
	path, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)

	b, err := New(path)
	assert.Nil(t, err)

	key := "abc"
	val := []byte{1, 2, 3}

	err = b.Put(key, val)
	assert.Nil(t, err)

	v, e := b.Get(key)
	assert.Nil(t, e)
	assert.Equal(t, val, v)

	e = b.Delete(key)
	assert.Nil(t, e)

	v, e = b.Get(key)
	assert.Nil(t, e)
	assert.Nil(t, v)
}

func TestBatchFile_Prefix(t *testing.T) {
	path, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)

	b, err := New(path)
	assert.Nil(t, err)

	prefix := "abc"
	val := []byte{1, 2, 3}

	err = b.Put(prefix, val)
	assert.Nil(t, err)

	err = b.Put(prefix+"1", val)
	assert.Nil(t, err)

	err = b.Put(prefix+"2", val)
	assert.Nil(t, err)

	err = b.Put("2", val)
	assert.Nil(t, err)

	files, err := b.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(files))

	err = b.DeleteAll()
	assert.Nil(t, err)

	m, err := b.GetAll()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(m))

	err = b.Close()
	assert.Nil(t, err)
}
