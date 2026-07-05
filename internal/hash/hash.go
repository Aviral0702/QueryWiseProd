package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// SHA256Hex returns lowercase hex-encoded SHA-256.
func SHA256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// HMACHex returns lowercase hex-encoded HMAC-SHA256(key, s).
func HMACHex(key, s string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(s))
	return hex.EncodeToString(mac.Sum(nil))
}

// Fingerprint returns a keyed HMAC-SHA256 fingerprint when key is non-empty.
// With an empty key it falls back to plain SHA-256, which is obfuscation only:
// normalized query text is low-entropy and reversible for a known schema.
func Fingerprint(key, s string) string {
	if key != "" {
		return HMACHex(key, s)
	}
	return SHA256Hex(s)
}
