package token

import "crypto/rand"

// chars represents the set of characters that can be included in a generated token.
const chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Generate creates a randomized string with the provided length to be used as a token The token
// consists of a random combination of numbers and mixed-case letters.
func Generate(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes), nil
}
