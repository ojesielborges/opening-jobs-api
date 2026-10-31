package postgres

import "testing"

func TestNewRepository(t *testing.T) {
	if repository := NewRepository(nil); repository == nil {
		t.Fatal("NewRepository() returned nil")
	}
}
