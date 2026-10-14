package utils

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func GetNonce() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}
