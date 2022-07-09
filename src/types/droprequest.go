package types

// DropRequest is a request for a file or set of files to be dropped into a private directory.
type DropRequest struct {
	// Directory is the directory that dropped files will go into.
	Directory string `json:"directory"`
	// MaxFiles represents the number of files allowed to be dropped.
	MaxFiles int `json:"max_files"`
	// MaxSize represents the maximum size for the complete drop.
	MaxSize int `json:"max_size"`
}
