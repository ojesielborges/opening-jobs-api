package main

import "testing"

func TestGenerateAPIKey(t *testing.T) {
	key, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generateAPIKey() error = %v", err)
	}
	if len(key) < 32 {
		t.Fatalf("generateAPIKey() length = %d, want at least 32", len(key))
	}
}
