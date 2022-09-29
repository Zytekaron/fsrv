package filemanager

import (
	"fsrv/src/config"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestNew(t *testing.T) {
	New(&config.FileManager{
		Path:     "/home/fsrv",
		MaxDepth: 5,
	})
}

func TestFileSystem_Clean(t *testing.T) {
	f := New(&config.FileManager{
		Path:     "/home/fsrv",
		MaxDepth: 5,
	})

	assert.Equal(t, f.CleanPath("/dir"), "/home/fsrv/dir")
	assert.Equal(t, f.CleanPath("/dir/file"), "/home/fsrv/dir/file")
	assert.Equal(t, f.CleanPath("/dir/ignored/../file"), "/home/fsrv/dir/file")
	assert.Equal(t, f.CleanPath("/./dir/./file"), "/home/fsrv/dir/file")
	assert.Equal(t, f.CleanPath("/../../dir/../../../dir/file"), "/home/fsrv/dir/file")
}

func TestFileSystem_CheckDepth(t *testing.T) {
	f := New(&config.FileManager{
		Path:     "/home/fsrv",
		MaxDepth: 5,
	})

	assert.Equal(t, f.CheckDepth("/home/fsrv/one/two"), true)
	assert.Equal(t, f.CheckDepth("/home/fsrv/one/two/three/four/five"), true)
	assert.Equal(t, f.CheckDepth("/home/fsrv/one/two/ignored/ignored/../../three"), true)

	assert.Equal(t, f.CheckDepth("/home/fsrv/one/two/three/four/five/six"), false)
	assert.Equal(t, f.CheckDepth("/home/fsrv/one/two/three/ignored/../four/five/ignored/../six"), false)
}
