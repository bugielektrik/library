package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomToken generates a random base64 encoded token
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateRandomInt generates a random integer between min and max
func GenerateRandomInt(min, max int64) (int64, error) {
	if min >= max {
		return 0, fmt.Errorf("min must be less than max")
	}

	diff := max - min
	n, err := rand.Int(rand.Reader, big.NewInt(diff))
	if err != nil {
		return 0, err
	}

	return n.Int64() + min, nil
}
