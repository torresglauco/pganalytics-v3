package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	path string
	data map[string]string
}

func NewFileStore(path string) *FileStore {
	store := &FileStore{
		path: path,
		data: make(map[string]string),
	}
	store.load()
	return store
}

func (fs *FileStore) Set(key, value string) error {
	fs.data[key] = value
	return fs.save()
}

func (fs *FileStore) Get(key string) (string, error) {
	val, exists := fs.data[key]
	if !exists {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, nil
}

func (fs *FileStore) GetAll() map[string]string {
	return fs.data
}

func (fs *FileStore) Delete(key string) error {
	delete(fs.data, key)
	return fs.save()
}

func (fs *FileStore) load() error {
	data, err := os.ReadFile(fs.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(data) > 0 {
		return json.Unmarshal(data, &fs.data)
	}

	fs.data = make(map[string]string)
	return nil
}

func (fs *FileStore) save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(fs.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.path, data, 0600)
}
