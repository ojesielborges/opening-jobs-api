package main

import (
	"reflect"
	"testing"
)

func TestParseTargetSelection(t *testing.T) {
	slugs := []string{"acme", "globex", "initech"}
	got, err := parseTargetSelection("1, 3, 1", slugs)
	if err != nil {
		t.Fatalf("parseTargetSelection() error = %v", err)
	}
	want := []string{"acme", "initech"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseTargetSelection() = %v, want %v", got, want)
	}
}

func TestParseTargetSelectionRejectsInvalidIndex(t *testing.T) {
	if _, err := parseTargetSelection("0,2", []string{"acme"}); err == nil {
		t.Fatal("expected invalid selection error")
	}
}
