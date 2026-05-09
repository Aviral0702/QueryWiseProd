package hash_test

import (
	"testing"

	"querywise/internal/hash"
)

func TestSHA256Hex(t *testing.T) {
	want := "2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae"
	if got := hash.SHA256Hex("foo"); got != want {
		t.Fatalf("SHA256Hex(foo) = %q, want %q", got, want)
	}
}
