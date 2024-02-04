package helpers

import (
	"math/rand"
	"strconv"
	"time"
)

func GenerateRandomCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000
	return strconv.Itoa(code)
}
