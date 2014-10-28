package datum

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestDiskStore(t *testing.T) {
	n := neko.Start(t)

	var disk *DiskStore

	tmpdir, err := ioutil.TempDir("", "disk")
	require.NoError(t, err)

	defer os.RemoveAll(tmpdir)

	dir := filepath.Join(tmpdir, "test")

	n.Setup(func() {
		disk = NewDiskStore(dir)
	})

	n.Cleanup(func() {
		os.RemoveAll(dir)
	})

	n.It("stores blobs in directories", func() {
		err := disk.Set("aabbcc", "default", []byte("foo"))
		require.NoError(t, err)

		data, err := ioutil.ReadFile(filepath.Join(dir, "aabbcc", "default"))
		require.NoError(t, err)

		assert.Equal(t, []byte("foo"), data)
	})

	n.It("retrieves blobs in directories", func() {
		dir := filepath.Join(dir, "aabbcc")
		os.MkdirAll(dir, 0755)

		err := ioutil.WriteFile(filepath.Join(dir, "default"), []byte("foo"), 0644)
		require.NoError(t, err)

		data, err := disk.Get("aabbcc", "default")
		require.NoError(t, err)

		assert.Equal(t, []byte("foo"), data)
	})

	n.Meow()
}
