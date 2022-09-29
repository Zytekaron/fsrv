package main

import (
	"fsrv/src/config"
	"fsrv/src/database/dbutil"
	"fsrv/src/database/impl/cache"
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
	db, err := dbutil.Create(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	cachedDB := cache.NewCache(db)

	serv := server.New(cachedDB, cfg)

	addr := ":" + strconv.Itoa(int(cfg.Server.Port))
	err = serv.Start(addr)
	if err != nil {
		log.Fatal(err)
	}
}
