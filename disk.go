package datum

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type DiskStore struct {
	Root string
}

func NewDiskStore(root string) *DiskStore {
	return &DiskStore{root}
}

func (d *DiskStore) Set(token, space string, val []byte) error {
	dir := filepath.Join(d.Root, token)
	os.MkdirAll(dir, 0755)

	return ioutil.WriteFile(filepath.Join(d.Root, token, space), val, 0644)
}

func (d *DiskStore) Get(token, space string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(d.Root, token, space))
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return nil, nil
		}

		return nil, err
	}

	return data, nil
}
