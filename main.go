package main

import (
	"fsrv/src/config"
	"log"
)

var cfg *config.Config

func init() {
	var err error
	cfg, err = config.Load([]string{"config.example.toml"})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
}
