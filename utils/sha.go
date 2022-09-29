package utils

import "crypto/sha512"

func Sha512Sum(data ...[]byte) []byte {
	sha := sha512.New()
	for _, d := range data {
		sha.Write(d)
	}
	return sha.Sum(nil)
}
