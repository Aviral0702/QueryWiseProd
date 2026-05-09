package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hex returns lowercase hex-encoded SHA-256.
func SHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
