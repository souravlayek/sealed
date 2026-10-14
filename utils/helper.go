package utils

import (
	"math/rand/v2"
	"strings"
)

const base32ReadableCharset = "23456789ABCDEFGHJKMNPQRSTUVWXYZ"

// GenerateDeviceUserCode produces a GitHub-like string (e.g., "WDJB-MJHT")
func GenerateDeviceUserCode() string {
	const totalChars = 8
	const halfway = totalChars / 2

	var sb strings.Builder
	// 8 characters + 1 hyphen = 9 bytes total allocated upfront
	sb.Grow(totalChars + 1)

	for i := 0; i < totalChars; i++ {
		// Inject the hyphen right in the middle
		if i == halfway {
			sb.WriteByte('-')
		}

		// Securely pick a random index from our optimized character set
		randomIndex := rand.IntN(len(base32ReadableCharset))
		sb.WriteByte(base32ReadableCharset[randomIndex])
	}

	return sb.String()
}
