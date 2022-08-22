package keygen

import (
	"crypto/sha512"
	"encoding/base64"
	"github.com/zytekaron/gotil/v2/random"
	"math"
)

func GetKeySize(randomBytes, checksumBytes int) (randSize, checkSize int) {
	const b64repMlt float64 = 1 / (6.0 / 8) //base64 representation multiplier
	randSize = int(math.Ceil(b64repMlt * float64(randomBytes)))
	checkSize = int(math.Ceil(b64repMlt * float64(checksumBytes)))
	return
}

func getSum(returnedBytes int, data ...[]byte) []byte {
	sha := sha512.New()
	for _, d := range data {
		sha.Write(d)
	}
	return sha.Sum(nil)[:returnedBytes]
}

// todo: fix
func MintKey(b64KeyString, salt string, checksumBytes int) (string, error) {
	key, err := base64.URLEncoding.DecodeString(b64KeyString)
	if err != nil {
		return "", err
	}

	data := key[:len(b64KeyString)]
	sum := getSum(checksumBytes, data, []byte(salt))
	return base64.URLEncoding.EncodeToString(sum), nil
}

func GetRand(size int) string {
	return random.MustSecureString(size, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890+/")
}
