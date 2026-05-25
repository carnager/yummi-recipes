package main

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sessionExpiry() time.Time {
	return time.Now().Add(30 * 24 * time.Hour)
}
