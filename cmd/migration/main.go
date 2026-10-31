package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"openingjobs/internal/bootstrap/config"
	"openingjobs/internal/controlplane/tenant"
	schemamigration "openingjobs/internal/operations/schema_migration"

	"github.com/manifoldco/promptui"
)

func main() {
	var err error
	logger := config.GetLogger()
	tenantSlug := flag.String("tenant", config.GetEnv("TENANT_SLUG", ""), "Tenant identifier")
	flag.Parse()

	err = config.Init()
	if err != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}

	defer config.GetDB().Close()

	schemaName := ""
	if *tenantSlug != "" {
		schemaName, err = tenant.SchemaName(*tenantSlug)
		if err != nil {
			logger.Errorf("invalid tenant: %v", err)
			return
		}
	}
	err = schemamigration.InitializeSchemaMigrationDBConext(schemaName)
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
			if err := runUpForTargets(*tenantSlug, migrationsPath); err != nil {
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

func runUpForTargets(tenantSlug, migrationsPath string) error {
	if tenantSlug != "" {
		schemaName, err := tenant.SchemaName(tenantSlug)
		if err != nil {
			return err
		}
		return migrationUseCase(schemaName).RunUp(migrationsPath)
	}

	slugs, err := activeTenantSlugs()
	if err != nil {
		return err
	}
	selected, err := selectTargetSlugs(slugs)
	if err != nil {
		return err
	}
	for _, slug := range selected {
		schemaName, err := tenant.SchemaName(slug)
		if err != nil {
			return err
		}
		fmt.Printf("Aplicando migrations no schema %s\n", schemaName)
		if err := migrationUseCase(schemaName).RunUp(migrationsPath); err != nil {
			return fmt.Errorf("%s: %w", slug, err)
		}
	}
	return nil
}

func migrationUseCase(schemaName string) *schemamigration.SchemaMigrationUseCase {
	return schemamigration.NewSchemaMigrationUseCase(
		schemamigration.NewSchemaMigrationStorage(config.GetDB(), schemaName),
	)
}

func activeTenantSlugs() ([]string, error) {
	var slugs []string
	if err := config.GetDB().Select(&slugs, `SELECT slug FROM control.tenants WHERE active = TRUE ORDER BY slug`); err != nil {
		return nil, fmt.Errorf("list active tenants: %w", err)
	}
	if len(slugs) == 0 {
		return nil, errors.New("nenhum tenant ativo encontrado")
	}
	return slugs, nil
}

func selectTargetSlugs(slugs []string) ([]string, error) {
	targetPrompt := promptui.Select{
		Label: "Destino das migrations",
		Items: []string{"Todos os schemas ativos", "Selecionar schemas"},
	}
	_, target, err := targetPrompt.Run()
	if err != nil {
		return nil, err
	}
	if target == "Todos os schemas ativos" {
		return slugs, nil
	}

	for index, slug := range slugs {
		fmt.Printf("%d. %s\n", index+1, slug)
	}
	selectionPrompt := promptui.Prompt{
		Label: "Números dos schemas separados por vírgula",
		Validate: func(value string) error {
			_, err := parseTargetSelection(value, slugs)
			return err
		},
	}
	selection, err := selectionPrompt.Run()
	if err != nil {
		return nil, err
	}
	return parseTargetSelection(selection, slugs)
}

func parseTargetSelection(value string, slugs []string) ([]string, error) {
	values := strings.Split(value, ",")
	selected := make([]string, 0, len(values))
	seen := make(map[int]struct{}, len(values))
	for _, value := range values {
		index, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || index < 1 || index > len(slugs) {
			return nil, errors.New("informe números válidos separados por vírgula")
		}
		if _, found := seen[index]; found {
			continue
		}
		seen[index] = struct{}{}
		selected = append(selected, slugs[index-1])
	}
	if len(selected) == 0 {
		return nil, errors.New("selecione ao menos um schema")
	}
	return selected, nil
}
