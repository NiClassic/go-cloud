package service

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

func GenerateToken(len int) string {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return hex.EncodeToString(b)
}
