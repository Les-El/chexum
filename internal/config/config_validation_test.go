package config

import (
	"os"
	"testing"
	"time"
)

func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		format  string
		wantErr bool
	}{
		{"default", false},
		{"verbose", false},
		{"json", false},
		{"plain", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			err := ValidateOutputFormat(tt.format)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateOutputFormat(%q) expected error", tt.format)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateOutputFormat(%q) unexpected error = %v", tt.format, err)
			}
		})
	}
}

func TestValidateAlgorithm(t *testing.T) {
	tests := []struct {
		algorithm string
		wantErr   bool
	}{
		{"sha256", false},
		{"md5", false},
		{"sha1", false},
		{"sha512", false},
		{"invalid", true},
		{"SHA256", true},
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			err := ValidateAlgorithm(tt.algorithm)
			if tt.wantErr && err == nil {
				t.Errorf("ValidateAlgorithm(%q) expected error", tt.algorithm)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidateAlgorithm(%q) unexpected error = %v", tt.algorithm, err)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Files = []string{"test.txt"}
	os.WriteFile("test.txt", []byte("hello"), 0644)
	defer os.Remove("test.txt")

	_, err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig() error = %v", err)
	}
}

func TestValidateConfigDates(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Files = []string{"test.txt"}
	os.WriteFile("test.txt", []byte("hello"), 0644)
	defer os.Remove("test.txt")

	now := time.Now()
	cfg.ModifiedAfter = now
	cfg.ModifiedBefore = now.Add(-time.Hour)

	_, err := ValidateConfig(cfg)
	if err == nil {
		t.Errorf("ValidateConfig() expected error for modified-after > modified-before")
	}

	cfg.ModifiedBefore = now.Add(time.Hour)
	_, err = ValidateConfig(cfg)
	if err != nil {
		t.Errorf("ValidateConfig() unexpected error for valid dates: %v", err)
	}
}
