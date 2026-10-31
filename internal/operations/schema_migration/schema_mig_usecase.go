package schemamigration

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SchemaMigrationUseCase struct {
	storage *SchemaMigrationStorage
}

func NewSchemaMigrationUseCase(storage *SchemaMigrationStorage) *SchemaMigrationUseCase {
	return &SchemaMigrationUseCase{
		storage: storage,
	}
}

func (useCase *SchemaMigrationUseCase) RunUp(migrationsPath string) error {
	pathFile := filepath.Join(migrationsPath, "/*_up.sql")
	files, _ := filepath.Glob(pathFile)
	sort.Strings(files)

	applied, err := useCase.AppliedMigrations()
	if err != nil {
		return err
	}

	for _, f := range files {
		if applied[f] {
			continue
		}
		fmt.Printf("🚀 Aplicando migração: %s\n", f)
		if err := useCase.runSQL(f); err != nil {
			return fmt.Errorf("erro ao aplicar %s: %v", f, err)
		}
		if err := useCase.MarkMigrationApplied(f); err != nil {
			return err
		}
	}
	fmt.Println("✅ Todas as migrações foram aplicadas.")
	return nil
}

func (useCase *SchemaMigrationUseCase) RunDown(migrationsPath string) error {
	pathFile := filepath.Join(migrationsPath, "/*_down.sql")
	files, _ := filepath.Glob(pathFile)
	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	applied, err := useCase.AppliedMigrations()
	if err != nil {
		return err
	}

	for _, f := range files {
		upFile := strings.Replace(f, "_down.sql", "_up.sql", 1)
		if applied[upFile] {
			fmt.Printf("🧩 Revertendo migração: %s\n", f)
			if err := useCase.runSQL(f); err != nil {
				return fmt.Errorf("erro ao reverter %s: %v", f, err)
			}
			if err := useCase.UnmarkMigration(upFile); err != nil {
				return err
			}
			fmt.Printf("↩️  Migração revertida: %s\n", upFile)
			return nil
		}
	}
	fmt.Println("Nenhuma migração para reverter.")
	return nil
}

func (useCase *SchemaMigrationUseCase) RunStatus(migrationsPath string) error {
	pathFile := filepath.Join(migrationsPath, "/*_up.sql")
	files, _ := filepath.Glob(pathFile)
	sort.Strings(files)

	applied, err := useCase.AppliedMigrations()
	if err != nil {
		return err
	}

	for _, f := range files {
		if applied[f] {
			fmt.Printf("✅ %s\n", f)
		} else {
			fmt.Printf("❌ %s\n", f)
		}
	}
	return nil
}

func (useCase *SchemaMigrationUseCase) runSQL(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = useCase.storage.execSql(string(content))
	return err
}

func (useCase *SchemaMigrationUseCase) AppliedMigrations() (map[string]bool, error) {
	return useCase.storage.appliedMigrations()
}

func (useCase *SchemaMigrationUseCase) MarkMigrationApplied(filename string) error {
	return useCase.storage.markMigrationApplied(filename)
}

func (useCase *SchemaMigrationUseCase) UnmarkMigration(filename string) error {
	return useCase.storage.unmarkMigration(filename)
}
