package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)

func generateRandomCode() (string, error) {
	const codeLength = 6
	const charset = "0123456789"

	randomCode := make([]byte, codeLength)
	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < codeLength; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", err
		}
		randomCode[i] = charset[randomIndex.Int64()]
	}

	return string(randomCode), nil
}

func GenerateOTP(c context.Context) (string, error) {
	code, err := generateRandomCode()
	if err != nil {
		log.Printf("failed to generate code: %v\n", err)
		return "", fmt.Errorf("failed to generate code")
	}
	return code, nil
}
