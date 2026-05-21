package internal

import (
	"crypto/rand"
	"math/big"
)

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digitChars  = "0123456789"
	symbolChars = "!@#$%^&*()-_=+[]{}|;:,.<>?"
)

func generatePassword(length int, lower, upper, digits, symbols bool) string {
	var pool string
	var mandatory []byte

	if lower {
		pool += lowerChars
		mandatory = append(mandatory, lowerChars[randInt(len(lowerChars))])
	}
	if upper {
		pool += upperChars
		mandatory = append(mandatory, upperChars[randInt(len(upperChars))])
	}
	if digits {
		pool += digitChars
		mandatory = append(mandatory, digitChars[randInt(len(digitChars))])
	}
	if symbols {
		pool += symbolChars
		mandatory = append(mandatory, symbolChars[randInt(len(symbolChars))])
	}
	if pool == "" {
		pool = lowerChars + upperChars + digitChars + symbolChars
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = pool[randInt(len(pool))]
	}
	
	for i, b := range mandatory {
		if i < length {
			result[i] = b
		}
	}
	
	for i := range result {
		j := randInt(len(result))
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

func randInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		panic(err)
	}
	return int(n.Int64())
}
