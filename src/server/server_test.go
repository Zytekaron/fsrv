package server

import (
	"encoding/base64"
	"fmt"
	"fsrv/src/config"
	"fsrv/src/database/dbimpl/cache"
	"fsrv/src/database/dbutil"
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
		time.Sleep(1 * time.Second)
		t.Log("second: ", i)
		i++
	}()
	time.Sleep(15 * time.Second)
}

func TestRequestServer(t *testing.T) {
	setup()

	randSize, checkSize := cfg.Server.KeyRandomBytes, cfg.Server.KeyCheckBytes
	keys := makeKeys(t, 20000, randSize, checkSize, []byte(cfg.Server.KeyValidationSecret))

	t.Log(">MAKING REQUESTS")

	totalRequests := int64(0)

	go func() {
		size, _ := keygen.GetKeySize(randSize, checkSize)
		for {
			d := keygen.GetRand(size)
			key := base64.RawURLEncoding.EncodeToString(d)
			url := fmt.Sprintf("http://127.0.0.1:1337/bad/?key=%s", key)
			req, _ := http.NewRequest("GET", url, nil)
			res, _ := http.DefaultClient.Do(req)
			//if err != nil {
			//	t.Log(err)
			//	t.Fail()
			//}
			if res != nil {
				//body, _ := io.ReadAll(res.Body)
				//fmt.Println(res)
				//fmt.Println(string(body))
				//err = res.Body.Close()
				//if err != nil {
				//	t.Log(err)
				//	t.Fail()
				//}
			}

			time.Sleep(20 * time.Microsecond)
			atomic.AddInt64(&totalRequests, 1)
		}
	}()

	go func() {
		client := http.Client{}
		for i := 0; i < 20000000000; i++ {
			url := fmt.Sprintf("http://127.0.0.1:1337/cool/path/%d", i)
			_, _ = client.Get(url)
			//if err != nil {
			//	t.Log(err)
			//	t.Fail()
			//}
			//if res != nil {
			//body, _ := io.ReadAll(res.Body)
			//fmt.Println(res)
			//fmt.Println(string(body))
			//err = res.Body.Close()
			//if err != nil {
			//	t.Log(err)
			//	t.Fail()
			//}
			//}
			time.Sleep(20 * time.Microsecond)
			atomic.AddInt64(&totalRequests, 1)
		}
	}()

	go func() {
		client := http.Client{}
		for _, key := range keys {
			key := key
			go func() {
				for i := 0; i < 10; i++ {
					url := fmt.Sprintf("http://127.0.0.1:1337/?key=%s", key)
					_, _ = client.Get(url)
					//if err != nil {
					//	t.Log(err)
					//	t.Fail()
					//}
					//if res != nil {
					//body, _ := io.ReadAll(res.Body)
					//fmt.Println(res)
					//fmt.Println(string(body))
					//err = res.Body.Close()
					//if err != nil {
					//	t.Log(err)
					//	t.Fail()
					//}
					//}
					atomic.AddInt64(&totalRequests, 1)
					time.Sleep(20 * time.Microsecond)
				}
			}()
			time.Sleep(100 * time.Microsecond)
		}
	}()

	time.Sleep(10 * time.Second)

	fmt.Println(totalRequests)
}
