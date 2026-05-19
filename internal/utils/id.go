package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func GenerateSessionID() string {

	timestamp := time.Now().UnixNano()

	randomBytes := make([]byte, 3)

	_, err := rand.Read(randomBytes)

	if err != nil {
		return fmt.Sprintf(
			"session_%d",
			timestamp,
		)
	}

	randomPart := hex.EncodeToString(randomBytes)

	return fmt.Sprintf(
		"session_%d_%s",
		timestamp,
		randomPart,
	)
}
