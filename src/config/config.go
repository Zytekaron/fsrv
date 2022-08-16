package config

import (
	"errors"
	"fsrv/src/database/entities"
	"github.com/pelletier/go-toml"
	"os"
)

var ErrNotFound = errors.New("no configuration file found")

type DatabaseImpl string

const (
	DatabaseSQLite DatabaseImpl = "sqlite"
)

type Config struct {
	Server      *Server      `toml:"server"`
	FileManager *FileManager `toml:"file_manager"`
	Database    *Database    `toml:"database"`
	Cache       *Cache       `toml:"cache"`
	Logging     *Logging     `toml:"logging"`
}

type Server struct {
	User             string             `toml:"user"`
	Port             int16              `toml:"port"`
	KeyAuthDefaultRL entities.RateLimit `toml:"key_auth_default_rl"`
	KeyBadAuthRL     entities.RateLimit `toml:"key_bad_rl"`
	IPAnonymousRL    entities.RateLimit `toml:"ip_anonymous_rl"`
}

type FileManager struct {
	Path     string `toml:"path"`
	MaxDepth int    `toml:"max_depth"`
}

type Database struct {
	Path    string       `toml:"path"`
	Version int          `toml:"version"`
	Type    DatabaseImpl `toml:"type"`
}

type Cache struct {
	Keys             int    `toml:"keys"`
	PermissionIDs    int    `toml:"permission_ids"`
	PermissionIDHash string `toml:"permission_id_hash"`
}

type Logging struct {
	StdoutLevel  string `toml:"stdout_level"`
	StdoutFormat string `toml:"stdout_format"`
	File         string `toml:"file"`
	FileLevel    string `toml:"file_level"`
	FileFormat   string `toml:"file_format"`
	MaxFileSize  string `toml:"max_file_size"`
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
