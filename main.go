package main

import (
	"fsrv/src/config"
	"fsrv/src/database/dbutil"
	"fsrv/src/database/impl/cache"
	"fsrv/src/filemanager"
	"fsrv/src/server"
	"log"
	"strconv"
)

var cfg *config.Config
var configPaths = []string{
	"/etc/fsrv/config.toml",
	"config.toml",
}

func init() {
	var err error
	cfg, err = config.Load(configPaths)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// setup database
	db, err := dbutil.Create(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	db = cache.NewCache(cfg.Cache, db)

	// setup file manager
	fm := filemanager.New(cfg.FileManager)

	// setup server
	serv := server.New(cfg, db, fm)

	// begin server
	addr := ":" + strconv.Itoa(int(cfg.Server.Port))
	err = serv.Start(addr)
	if err != nil {
		log.Fatal(err)
	}
}
