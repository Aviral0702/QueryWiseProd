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

func TestHMACHex(t *testing.T) {
	tests := []struct {
		name string
		key  string
		in   string
		want string
	}{
		{
			name: "known key and input is stable",
			key:  "secret",
			in:   "foo",
			want: "773ba44693c7553d6ee20f61ea5d2757a9a4f4a44d2841ae4e95b52e4cd62db4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hash.HMACHex(tt.key, tt.in); got != tt.want {
				t.Fatalf("HMACHex(%q, %q) = %q, want %q", tt.key, tt.in, got, tt.want)
			}
		})
	}

	if a, b := hash.HMACHex("key1", "foo"), hash.HMACHex("key2", "foo"); a == b {
		t.Fatalf("HMACHex with different keys produced identical output %q", a)
	}
}

func TestFingerprint(t *testing.T) {
	tests := []struct {
		name string
		key  string
		in   string
		want string
	}{
		{
			name: "empty key delegates to SHA256Hex",
			key:  "",
			in:   "foo",
			want: hash.SHA256Hex("foo"),
		},
		{
			name: "non-empty key delegates to HMACHex",
			key:  "secret",
			in:   "foo",
			want: hash.HMACHex("secret", "foo"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hash.Fingerprint(tt.key, tt.in); got != tt.want {
				t.Fatalf("Fingerprint(%q, %q) = %q, want %q", tt.key, tt.in, got, tt.want)
			}
		})
	}
}
