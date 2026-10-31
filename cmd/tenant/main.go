package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"time"

	"openingjobs/internal/bootstrap/config"
	"openingjobs/internal/controlplane/tenant"
)

func main() {
	logger := config.GetLogger()
	if err := config.Init(); err != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}
	defer config.GetDB().Close()

	slug := flag.String("slug", config.GetEnv("TENANT_SLUG", ""), "Tenant identifier")
	apiKey := flag.String("api-key", config.GetEnv("TENANT_API_KEY", ""), "Tenant API key; generated when omitted")
	flag.Parse()
	if *slug == "" {
		logger.Error("slug is required")
		return
	}
	generatedKey := false
	if *apiKey == "" {
		var err error
		*apiKey, err = generateAPIKey()
		if err != nil {
			logger.Errorf("generate tenant API key: %v", err)
			return
		}
		generatedKey = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := tenant.NewProvisioner(config.GetDB()).Provision(ctx, *slug, *apiKey); err != nil {
		logger.Errorf("tenant provisioning failed: %v", err)
		return
	}
	logger.Infof("tenant %q provisioned successfully", *slug)
	if generatedKey {
		fmt.Printf("Tenant API key (save it now; it will not be shown again): %s\n", *apiKey)
	}
}

func generateAPIKey() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
