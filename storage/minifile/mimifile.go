package minifile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type MiniFile struct {
	path string
}

func New(path string) (*MiniFile, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(abs, 0755)
	if err != nil {
		return nil, err
	}
	return &MiniFile{path: abs}, nil
}

func (b *MiniFile) Put(key string, value []byte) error {
	if key == "" {
		return fmt.Errorf("store file with empty key")
	}

	name := filepath.Join(b.path, key)
	if err := ioutil.WriteFile(name, value, 0644); err != nil {
		return fmt.Errorf("fail to write file %s: %w", name, err)
	}

	return nil
}

func (b *MiniFile) Delete(key string) error {
	err := os.Remove(filepath.Join(b.path, key))
	if err != nil && isNoFileError(err) {
		return nil
	}

	return err
}

func (b *MiniFile) Get(key string) ([]byte, error) {
	val, err := ioutil.ReadFile(filepath.Join(b.path, key))
	if err != nil && isNoFileError(err) {
		return nil, nil
	}

	return val, err
}

func (b *MiniFile) Has(key string) (bool, error) {
	val, err := b.Get(key)
	if err != nil {
		return false, err
	}

	return val != nil, nil
}

func (b *MiniFile) Close() error {
	return nil
}

func (b *MiniFile) DeleteAll() error {
	if err := os.RemoveAll(b.path); err != nil {
		return err
	}

	_ = os.MkdirAll(b.path, 0755)
	return nil
}

func (b *MiniFile) Prefix(prefix string) ([]string, error) {
	var files []string

	if err := filepath.Walk(b.path, func(path string, info os.FileInfo, err error) error {
		if path != b.path {
			name := filepath.Base(path)
			if strings.HasPrefix(name, prefix) {
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
