package pubsubhubbub

import (
	"crypto/rand"
	"encoding/base64"
)

func randomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateChallenge() string {
	challenge, err := randomString(32)
	if err != nil {
		panic(err)
	}
	return challenge
}
