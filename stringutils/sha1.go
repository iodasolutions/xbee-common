package stringutils

import (
	"crypto/sha1"
	"fmt"
)

func Sha1String(message string) string {
	return Sha1Bytes([]byte(message))
}

func Sha1StringTruncated(message string) string {
	sha1 := Sha1String(message)
	return sha1[:10]
}

func Sha1Bytes(message []byte) string {
	hash := sha1.New()
	if _, err := hash.Write(message); err != nil {
		panic(err)
	}
	hashBytes := hash.Sum(nil)

	return fmt.Sprintf("%x", hashBytes)
}
