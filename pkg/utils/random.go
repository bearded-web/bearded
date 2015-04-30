package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

func RandomString(n int) string {
	id := make([]byte, n)

	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		panic(err) // This shouldn't happen
	}
	return hex.EncodeToString(id)
}
