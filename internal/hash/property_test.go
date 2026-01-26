package hash

import (
	"testing"
	"testing/quick"
)

// Note: Removed unused bytes and crypto/hash imports to resolve build failures.
// These should be re-added when their respective hashing functions are
// actively used in the tests ("bytes", "crypto/md5", "crypto/sha1", "crypto/sha256", "crypto/sha512")

func TestProperty_HashDeterminism(t *testing.T) {
	f := func(input []byte) bool {
		if len(input) == 0 {
			return true
		}

		h1, _ := NewComputer("sha256")
		hash1 := h1.ComputeBytes(input)

		h2, _ := NewComputer("sha256")
		hash2 := h2.ComputeBytes(input)

		return hash1 == hash2
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestProperty_HashLength(t *testing.T) {
	tests := []struct {
		algo   string
		length int
	}{
		{"sha256", 64},
		{"md5", 32},
		{"sha1", 40},
		{"sha512", 128},
		{"blake2b", 128},
	}

	for _, tt := range tests {
		f := func(input []byte) bool {
			h, _ := NewComputer(tt.algo)
			hash := h.ComputeBytes(input)
			return len(hash) == tt.length
		}
		if err := quick.Check(f, nil); err != nil {
			t.Errorf("%s: %v", tt.algo, err)
		}
	}
}
