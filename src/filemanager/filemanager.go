package filemanager

import (
	"fmt"
	"fsrv/src/config"
	"os"
	"path/filepath"
	"strings"
)

type FileManager struct {
	baseDir  string
	maxDepth int
}

// New creates a new FileManager.
func New(cfg *config.FileManager) *FileManager {
	return &FileManager{
		baseDir:  cfg.Path,
		maxDepth: cfg.MaxDepth,
	}
}

// CleanPath returns a path which is guaranteed to be at the level
// of or deeper than the base directory of the file manager. The
// path is first cleaned, then joined with the base directory.
func (f *FileManager) CleanPath(name string) string {
	cleaned := filepath.Join(f.baseDir, filepath.Clean(name))
	if !strings.HasPrefix(cleaned, f.baseDir) {
		// cleaning failed. this should not be possible.
		panic(fmt.Sprintf("input was not properly cleaned: '%s' does not have prefix '%s'", cleaned, f.baseDir))
	}
	return cleaned
}

// CheckDepth takes a path and returns whether the depth of
// the path, not including the base directory, is less than
// the maximum depth allowed by the file manager. The input
// must begin with the base path of the file manager.
func (f *FileManager) CheckDepth(name string) bool {
	raw := filepath.Clean(name)
	raw = strings.TrimPrefix(raw, f.baseDir)
	raw = strings.TrimPrefix(raw, "/")
	if len(name) == len(raw) { // did not strip the base dir
		panic("path must begin with base dir")
	}

	// fixme: potentially insecure splitting
	parts := strings.Split(raw, string(os.PathSeparator))
	return len(parts) <= f.maxDepth
}
