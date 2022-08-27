package keygen

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"log"
	"math"
)

func GetKeySize(randomBytes, checksumBytes int) (randSize, checkSize int) {
	const b64repMlt float64 = 1 / (6.0 / 8) //base64 representation multiplier
	randSize = int(math.Ceil(b64repMlt * float64(randomBytes)))
	checkSize = int(math.Ceil(b64repMlt * float64(checksumBytes)))
	return
}

func getSum(data ...[]byte) []byte {
	sha := sha512.New()
	for _, d := range data {
		sha.Write(d)
	}
	return sha.Sum(nil)
}

func MintKey(key, salt []byte, checksumBytes int) string {
	sum := getSum(key, salt)[:checksumBytes]
	return base64.RawURLEncoding.EncodeToString(append(key, sum...))
}

func GetRand(size int) []byte {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
