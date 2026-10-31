package schemamigration

import "testing"

func TestRunSQLReturnsReadError(t *testing.T) {
	useCase := NewSchemaMigrationUseCase(nil)
	if err := useCase.runSQL("missing.sql"); err == nil {
		t.Fatal("expected file read error")
	}
}
