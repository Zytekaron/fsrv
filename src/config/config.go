package config

import (
	"errors"
	"github.com/pelletier/go-toml"
	"os"
)

var ErrNotFound = errors.New("no configuration file found")

type Config struct {
	Server      *Server      `toml:"server"`
	FileManager *FileManager `toml:"file_manager"`
	Database    *Database    `toml:"database"`
	Cache       *Cache       `toml:"cache"`
}

type Server struct {
	User string `toml:"user"`
	Port int16  `toml:"port"`
}

type FileManager struct {
	Path     string `toml:"path"`
	MaxDepth int    `toml:"max_depth"`
}

type Database struct {
	Path    string `toml:"path"`
	Version int    `toml:"version"`
}

type Cache struct {
	Keys            int    `toml:"keys"`
	Permissions     int    `toml:"permissions"`
	PermissionsHash string `toml:"permissions_hash"`
}

func Load(paths []string) (*Config, error) {
	for _, path := range paths {
		file, err := os.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		cfg := &Config{}
		return cfg, toml.NewDecoder(file).Decode(cfg)
	}

	return nil, ErrNotFound
}
