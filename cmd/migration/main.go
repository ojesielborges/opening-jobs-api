package main

import (
	"fmt"
	"log"
	schemamigration "openingjobs/internal/schema_migration"
	"openingjobs/pkg/config"

	"github.com/manifoldco/promptui"
)

func main() {
	var err error
	logger := config.GetLogger()

	err = config.Init()
	if err != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}

	defer config.GetDB().Close()

	err = schemamigration.InitializeSchemaMigrationDBConext()
	if err != nil {
		logger.Errorf("InitializeSchemaMigrationDBConext error: %v", err)
		return
	}

	migrationsPath := config.GetEnv("MIGRATIONS_PATH", "")
	if migrationsPath == "" {
		logger.Errorf("MIGRATIONS_PATH not set")
		return
	}

	for {
		fmt.Println("======================================")
		fmt.Println("     🛠️  CLI de Migrações - Go")
		fmt.Println("======================================")

		prompt := promptui.Select{
			Label: "Selecione uma opção",
			Items: []string{
				"Aplicar novas migrações (UP)",
				"Reverter última migração (DOWN)",
				"Mostrar status das migrações",
				"Sair",
			},
			Size: 5,
		}

		_, result, err := prompt.Run()
		if err != nil {
			log.Fatalf("Erro ao executar prompt: %v\n", err)
		}

		switch result {
		case "Aplicar novas migrações (UP)":
			fmt.Println("➡️  Aplicando migrações...")
			if err := schemamigration.UseCase.RunUp(migrationsPath); err != nil {
				log.Printf("❌ Erro: %v\n", err)
			}

		case "Reverter última migração (DOWN)":
			fmt.Println("↩️  Revertendo última migração...")
			if err := schemamigration.UseCase.RunDown(migrationsPath); err != nil {
				log.Printf("❌ Erro: %v\n", err)
			}

		case "Mostrar status das migrações":
			fmt.Println("📋 Status das migrações:")
			if err := schemamigration.UseCase.RunStatus(migrationsPath); err != nil {
				log.Printf("❌ Erro: %v\n", err)
			}

		case "Sair":
			fmt.Println("👋 Encerrando CLI de migrações.")
			return
		}

		fmt.Println("\nPressione ENTER para continuar...")
		fmt.Scanln()
	}

}
