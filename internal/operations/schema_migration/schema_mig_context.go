package schemamigration

import "openingjobs/internal/bootstrap/config"

var (
	UseCase *SchemaMigrationUseCase
)

func InitializeSchemaMigrationDBConext(schemaName string) error {
	db := config.GetDB()

	storage := NewSchemaMigrationStorage(db, schemaName)
	UseCase = NewSchemaMigrationUseCase(storage)

	return nil
}
