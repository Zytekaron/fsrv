package keygen

import (
	"crypto/rand"
	"encoding/base64"
	"fsrv/utils"
	"log"
	"math"
)

func GetKeySize(randomBytes, checksumBytes int) (randSize, checkSize int) {
	const b64repMlt float64 = 1 / (6.0 / 8) //base64 representation multiplier
	randSize = int(math.Ceil(b64repMlt * float64(randomBytes)))
	checkSize = int(math.Ceil(b64repMlt * float64(checksumBytes)))
	return
}

func MintKey(key, salt []byte, checksumBytes int) string {
	sum := utils.Sha512Sum(key, salt)[:checksumBytes]
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
