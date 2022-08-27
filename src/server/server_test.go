package server

import (
	"fmt"
	"fsrv/src/config"
	"fsrv/src/database/dbimpl/cache"
	"fsrv/src/database/dbutil"
	"fsrv/utils/keygen"
	"io"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)

var cfg *config.Config
var configPaths = []string{"/etc/fsrv/config.toml", "../../config.toml"}

func setup() {
	var err error
	cfg, err = config.Load(configPaths)
	if err != nil {
		log.Fatal(err)
	}
}

func run() {
	db, err := dbutil.Create(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	cachedDB := cache.NewCache(db)

	serv := New(cachedDB, cfg)

	addr := ":" + strconv.Itoa(int(cfg.Server.Port))
	err = serv.Start(addr)
	if err != nil {
		log.Fatal(err)
	}
}

func makeKeys(t *testing.T, numKeys, keySize, checksumBytes int, salt []byte) (keys []string) {
	for i := 0; i < numKeys; i++ {
		kStr := keygen.GetRand(keySize)
		key := keygen.MintKey(kStr, salt, checksumBytes)
		t.Log(key)
		keys = append(keys, key)
	}
	return
}

func TestServer(t *testing.T) {
	setup()
	t.Log(">STARTING SERVER")
	go run() //run server
	time.Sleep(500 * time.Millisecond)
	t.Log(">GENERATING KEYS")
	randSize, checkSize := cfg.Server.KeyRandomBytes, cfg.Server.KeyCheckBytes
	keys := makeKeys(t, 20, randSize, checkSize, []byte(cfg.Server.KeyValidationSecret))

	t.Log(">MAKING REQUESTS")
	for _, key := range keys {
		url := fmt.Sprintf("http://127.0.0.1:1337/?key=%s", key)
		req, _ := http.NewRequest("GET", url, nil)
		res, _ := http.DefaultClient.Do(req)
		body, _ := io.ReadAll(res.Body)
		fmt.Println(res)
		fmt.Println(string(body))
		err := res.Body.Close()
		if err != nil {
			t.Log(err)
			t.Fail()
		}
	}

	time.Sleep(500 * time.Millisecond)
}
