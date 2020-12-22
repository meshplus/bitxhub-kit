package minifile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/prometheus/tsdb/fileutil"
)

type MiniFile struct {
	path         string
	instanceLock fileutil.Releaser // File-system lock to prevent double opens
	namedLock    *NamedLock
	lock         *sync.RWMutex
	closed       int64
}

type NamedLock struct {
	lock  *sync.RWMutex
	locks map[string]*sync.RWMutex
}

func NewNamedLock() *NamedLock {
	return &NamedLock{
		lock:  &sync.RWMutex{},
		locks: make(map[string]*sync.RWMutex),
	}
}

func New(path string) (*MiniFile, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	flock, _, err := fileutil.Flock(filepath.Join(abs, "FLOCK"))
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(abs, 0755)
	if err != nil {
		return nil, err
	}
	return &MiniFile{
		path:         abs,
		instanceLock: flock,
		lock:         &sync.RWMutex{},
		namedLock:    NewNamedLock(),
	}, nil
}

func (mf *MiniFile) Put(key string, value []byte) error {
	if mf.isClosed() {
		return fmt.Errorf("the miniFile storage is closed")
	}

	if key == "" {
		return fmt.Errorf("store file with empty key")
	}

	mf.lock.RLock()
	defer mf.lock.RUnlock()

	mf.namedLock.Lock(key)
	defer mf.namedLock.UnLock(key, false)

	name := filepath.Join(mf.path, key)

	if err := ioutil.WriteFile(name, value, 0644); err != nil {
		return fmt.Errorf("fail to write file %s: %w", name, err)
	}

	return nil
}

func (mf *MiniFile) Delete(key string) error {
	if mf.isClosed() {
		return fmt.Errorf("the miniFile storage is closed")
	}

	mf.lock.RLock()
	defer mf.lock.RUnlock()

	mf.namedLock.Lock(key)
	defer mf.namedLock.UnLock(key, true)

	err := os.Remove(filepath.Join(mf.path, key))
	if err != nil && isNoFileError(err) {
		return nil
	}

	return err
}
func (mf *MiniFile) Get(key string) ([]byte, error) {
	if mf.isClosed() {
		return nil, fmt.Errorf("the miniFile storage is closed")
	}

	mf.lock.RLock()
	defer mf.lock.RUnlock()

	return mf.get(key)
}

func (mf *MiniFile) get(key string) ([]byte, error) {
	mf.namedLock.RLock(key)

	val, err := ioutil.ReadFile(filepath.Join(mf.path, key))
	if err != nil && isNoFileError(err) {
		mf.namedLock.RUnLock(key, true)
		return nil, nil
	}

	mf.namedLock.RUnLock(key, false)

	return val, err
}

func (mf *MiniFile) Has(key string) (bool, error) {
	val, err := mf.Get(key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (mf *MiniFile) Close() error {
	if mf.isClosed() {
		return nil
	}
	atomic.StoreInt64(&mf.closed, 1)
	return mf.instanceLock.Release()
}

func (mf *MiniFile) GetAll() (map[string][]byte, error) {
	if mf.isClosed() {
		return nil, fmt.Errorf("the miniFile storage is closed")
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	all := make(map[string][]byte)

	files, err := mf.prefix("")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		val, err := mf.get(file)
		if err != nil {
			return nil, err
		}

		all[file] = val
	}

	return all, nil
}

func (mf *MiniFile) DeleteAll() error {
	if mf.isClosed() {
		return fmt.Errorf("the miniFile storage is closed")
	}

	mf.lock.Lock()
	defer mf.lock.Unlock()

	if err := os.RemoveAll(mf.path); err != nil {
		return err
	}

	_ = os.MkdirAll(mf.path, 0755)
	return nil
}

func (mf *MiniFile) prefix(prefix string) ([]string, error) {
	if mf.isClosed() {
		return nil, fmt.Errorf("the miniFile storage is closed")
	}

	var files []string

	if err := filepath.Walk(mf.path, func(path string, info os.FileInfo, err error) error {
		if path != mf.path {
			name := filepath.Base(path)
			if strings.HasPrefix(name, prefix) && name != "FLOCK" {
				files = append(files, name)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}

func isNoFileError(err error) bool {
	return strings.Contains(err.Error(), "no such file or directory")
}

func (mf *MiniFile) isClosed() bool {
	return atomic.LoadInt64(&mf.closed) == 1
}

func (nl *NamedLock) RLock(key string) {
	nl.lock.Lock()
	defer nl.lock.Unlock()

	lock, ok := nl.locks[key]
	if !ok {
		lock = &sync.RWMutex{}
		nl.locks[key] = lock
	}

	lock.RLock()
}

func (nl *NamedLock) RUnLock(key string, delLock bool) error {
	nl.lock.Lock()
	defer nl.lock.Unlock()

	lock, ok := nl.locks[key]
	if !ok {
		return fmt.Errorf("cannot get lock for %s", key)
	}

	lock.RUnlock()

	if delLock {
		delete(nl.locks, key)
	}

	return nil
}

func (nl *NamedLock) Lock(key string) {
	nl.lock.Lock()
	defer nl.lock.Unlock()

	lock, ok := nl.locks[key]
	if !ok {
		lock = &sync.RWMutex{}
		nl.locks[key] = lock
	}

	lock.Lock()
}

func (nl *NamedLock) UnLock(key string, delLock bool) error {
	nl.lock.Lock()
	defer nl.lock.Unlock()

	lock, ok := nl.locks[key]
	if !ok {
		return fmt.Errorf("cannot get lock for %s", key)
	}

	lock.Unlock()

	if delLock {
		delete(nl.locks, key)
	}

	return nil
}

func (nl *NamedLock) LockAll() {
	nl.lock.RLock()
	defer nl.lock.RUnlock()

	for _, lock := range nl.locks {
		lock.Lock()
	}
}

func (nl *NamedLock) UnlockAll() {
	for _, lock := range nl.locks {
		lock.Unlock()
	}
}
