package helpers

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string, passwordHash chan string) {
	passwordHashGenerate, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		passwordHash <- "error"
		return
	}
	passwordHash <- string(passwordHashGenerate)
}

func DecodePassword(PasswordHash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(PasswordHash), []byte(password))
	return err
}
