package config

import (
	"testing"
	"time"
)

func TestGetEnvAndDuration(t *testing.T) {
	t.Setenv("TEST_VALUE", "configured")
	if got := GetEnv("TEST_VALUE", "default"); got != "configured" {
		t.Fatalf("GetEnv() = %q", got)
	}
	if got := GetEnv("MISSING_VALUE", "default"); got != "default" {
		t.Fatalf("GetEnv() = %q", got)
	}

	t.Setenv("TEST_DURATION", "2s")
	got, err := GetDuration("TEST_DURATION", time.Minute)
	if err != nil || got != 2*time.Second {
		t.Fatalf("GetDuration() = %v, %v", got, err)
	}
	t.Setenv("TEST_DURATION", "0s")
	if _, err := GetDuration("TEST_DURATION", time.Minute); err == nil {
		t.Fatal("expected duration validation error")
	}
}

func TestInitializeQueryBuilderUsesDollarPlaceholders(t *testing.T) {
	query, args, err := InitializeQueryBuilder().Select("id").From("opening").Where("id = ?", 10).ToSql()
	if err != nil || query != "SELECT id FROM opening WHERE id = $1" || len(args) != 1 || args[0] != 10 {
		t.Fatalf("ToSql() = %q, %v, %v", query, args, err)
	}
}
