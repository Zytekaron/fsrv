package main

import (
	"fsrv/src/config"
	"fsrv/src/database"
	"fsrv/src/server"
	"log"
	"strconv"
)

var cfg *config.Config
var configPaths = []string{"/etc/fsrv/config.toml", "config.toml"}

func init() {
	var err error
	cfg, err = config.Load(configPaths)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	db, err := database.Create(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	serv := server.New(db)

	addr := ":" + strconv.Itoa(int(cfg.Server.Port))
	err = serv.Start(addr)
	if err != nil {
		log.Fatal(err)
	}
}
