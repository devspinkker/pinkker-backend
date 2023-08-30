package helpers

import (
	"math/rand"
	"time"
)

func KeyTransmission(length int) string {
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	bytes := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length; i++ {
		bytes[i] = possible[rand.Intn(len(possible))]
	}
	return string(bytes)
}
