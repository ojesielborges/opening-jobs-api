package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"openingjobs/internal/bootstrap/config"
	"openingjobs/internal/business/opening"
	"openingjobs/internal/controlplane/tenant"
	tenantpostgres "openingjobs/internal/controlplane/tenant/infrastructure/postgres"
	tenanthttp "openingjobs/internal/controlplane/tenant/transport/http"
	"openingjobs/internal/platform/http/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := config.GetLogger()

	if err := config.Init(); err != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}

	defer config.GetDB().Close()
	requestTimeout, err := config.GetDuration("HTTP_REQUEST_TIMEOUT", 10*time.Second)
	if err != nil {
		logger.Errorf("invalid HTTP request timeout: %v", err)
		return
	}
	tenantCacheTTL, err := config.GetDuration("TENANT_CACHE_TTL", 5*time.Minute)
	if err != nil {
		logger.Errorf("invalid tenant cache TTL: %v", err)
		return
	}
	tenantNegativeCacheTTL, err := config.GetDuration("TENANT_NEGATIVE_CACHE_TTL", 15*time.Second)
	if err != nil {
		logger.Errorf("invalid tenant negative cache TTL: %v", err)
		return
	}

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Errorf("configure trusted proxies: %v", err)
		return
	}

	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.RequestTimeout(requestTimeout))
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})

	tenants := tenant.NewCachedResolver(
		tenantpostgres.NewRepository(config.GetDB()),
		tenantCacheTTL,
		tenantNegativeCacheTTL,
	)

	api := router.Group("/api/v1")
	api.Use(tenanthttp.RequireTenant(tenants))
	opening.MountHTTPV1(api, config.GetDB())

	server := &http.Server{
		Addr:              config.GetEnv("HTTP_ADDR", ":8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Infof("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("HTTP server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("graceful shutdown failed: %v", err)
	}
}
