package main

import (
	"context"
	"time"

	"openingjobs/internal/tenant"
	"openingjobs/pkg/config"
)

func main() {
	logger := config.GetLogger()
	if err := config.Init(); err != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}
	defer config.GetDB().Close()

	slug := config.GetEnv("TENANT_SLUG", "")
	apiKey := config.GetEnv("TENANT_API_KEY", "")
	migrationsPath := config.GetEnv("MIGRATIONS_PATH", "")
	if slug == "" || apiKey == "" || migrationsPath == "" {
		logger.Error("TENANT_SLUG, TENANT_API_KEY, and MIGRATIONS_PATH are required")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := tenant.NewProvisioner(config.GetDB()).Provision(ctx, slug, apiKey, migrationsPath); err != nil {
		logger.Errorf("tenant provisioning failed: %v", err)
		return
	}
	logger.Infof("tenant %q provisioned successfully", slug)
}
