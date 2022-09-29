package files

import (
	"encoding/base64"
	"fmt"
	"fsrv/src/config"
	"fsrv/src/database/dbutil"
	"fsrv/src/database/impl/cache"
	"fsrv/utils/keygen"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
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
		log.Fatal("err with db:", err)
	}
	db = cache.NewCache(cfg.Cache, db)

	serv := New(cfg, db, nil)

	addr := ":" + strconv.Itoa(int(cfg.Server.Port))
	err = serv.Start(addr)
	if err != nil {
		log.Fatal(err)
	}
}

func makeKeys( /*t *testing.V,*/ numKeys, keySize, checksumBytes int, salt []byte) (keys []string) {
	for i := 0; i < numKeys; i++ {
		kStr := keygen.GetRand(keySize)
		key := keygen.MintKey(kStr, salt, checksumBytes)
		//t.Log(key)
		keys = append(keys, key)
	}
	return
}

func TestServer(t *testing.T) {
	setup()
	t.Log(">STARTING SERVER")
	go run() //run server

	i := 1
	go func() {
		for {
			time.Sleep(1 * time.Second)
			t.Log("second: ", i)
			i++
		}
	}()
	time.Sleep(15 * time.Second)
}

func TestRequestServer(t *testing.T) {
	setup()

	randSize, checkSize := cfg.Server.KeyRandomBytes, cfg.Server.KeyCheckBytes
	keys := makeKeys( /*t,*/ 20000, randSize, checkSize, []byte(cfg.Server.KeyValidationSecret))

	t.Log(">MAKING REQUESTS")

	totalRequests := int64(0)
	totalTime := int64(0)
	for i := 0; i < 3; i++ {
		go func() {
			size, _ := keygen.GetKeySize(randSize, checkSize)
			for {
				start := time.Now()
				d := keygen.GetRand(size)
				key := base64.RawURLEncoding.EncodeToString(d)
				url := fmt.Sprintf("http://127.0.0.1:1337/bad/?key=%s", key)
				req, _ := http.NewRequest("GET", url, nil)
				_, _ = http.DefaultClient.Do(req)
				atomic.AddInt64(&totalTime, int64(time.Since(start)))

				time.Sleep(20 * time.Microsecond)
				atomic.AddInt64(&totalRequests, 1)
			}
		}()

		go func() {
			client := http.Client{}
			for i := 0; true; i++ {
				start := time.Now()
				url := fmt.Sprintf("http://127.0.0.1:1337/cool/path/%d", i)
				_, _ = client.Get(url)
				atomic.AddInt64(&totalTime, int64(time.Since(start)))
				time.Sleep(20 * time.Microsecond)
				atomic.AddInt64(&totalRequests, 1)
			}
		}()

		go func() {
			client := http.Client{}
			for _, key := range keys {
				key := key
				for i := 0; i < 4; i++ {
					start := time.Now()
					url := fmt.Sprintf("http://127.0.0.1:1337/?key=%s", key)
					_, _ = client.Get(url)
					atomic.AddInt64(&totalTime, int64(time.Since(start)))
					atomic.AddInt64(&totalRequests, 1)
					time.Sleep(20 * time.Microsecond)
				}
				time.Sleep(100 * time.Microsecond)
			}
		}()
	}

	time.Sleep(10 * time.Second)

	dursum := time.Duration(totalTime / totalRequests)
	fmt.Printf("Average time: %s\n", dursum.String())

	fmt.Println("total requests:", totalRequests)
}
