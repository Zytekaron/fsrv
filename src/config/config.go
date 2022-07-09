package config

import (
	"errors"
	"github.com/pelletier/go-toml"
	"os"
)

var ErrNoConfigs = errors.New("no configuration file found")

type Config struct {
	Server      *Server      `toml:"server"`
	Database    *Database    `toml:"database"`
	FileManager *FileManager `toml:"file_manager"`
}

type Server struct {
	User string `toml:"user"`
	Port int16  `toml:"port"`
}

type Database struct {
	Path    string `toml:"path"`
	Version int    `toml:"version"`
}

type FileManager struct {
	Path     string `toml:"path"`
	MaxDepth int    `toml:"max_depth"`
}

func Load(paths []string) (*Config, error) {
	for _, path := range paths {
		file, err := os.Open(path)
		if err == os.ErrNotExist {
			continue
		}
		if err != nil {
			return nil, err
		}

		cfg := &Config{}
		return cfg, toml.NewDecoder(file).Decode(cfg)
	}

	return nil, ErrNoConfigs
}
