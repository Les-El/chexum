package hash

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"testing"
	"testing/quick"
)

func TestProperty_HashDeterminism(t *testing.T) {
	f := func(input []byte) bool {
		if len(input) == 0 {
			return true
		}
		
		h1, _ := NewComputer("sha256")
		hash1, _ := h1.ComputeBytes(input)
		
		h2, _ := NewComputer("sha256")
		hash2, _ := h2.ComputeBytes(input)
		
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
	}

	for _, tt := range tests {
		f := func(input []byte) bool {
			h, _ := NewComputer(tt.algo)
			hash, _ := h.ComputeBytes(input)
			return len(hash) == tt.length
		}
		if err := quick.Check(f, nil); err != nil {
			t.Errorf("%s: %v", tt.algo, err)
		}
	}
}
