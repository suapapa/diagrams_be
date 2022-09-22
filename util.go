package main

import (
	"crypto/rand"
	"encoding/hex"
)

func randHex(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		// TODO: proper error check
		return ""
	}
	return hex.EncodeToString(bytes)
}
