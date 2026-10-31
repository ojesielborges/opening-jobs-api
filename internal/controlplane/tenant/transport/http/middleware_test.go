package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"openingjobs/internal/controlplane/tenant"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type resolverStub struct {
	resolved tenant.Tenant
	err      error
}

func (s resolverStub) Resolve(context.Context, string, string) (tenant.Tenant, error) {
	return s.resolved, s.err
}

func TestRequireTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	valid := tenant.Tenant{ID: uuid.New(), Slug: "acme", SchemaName: "tenant_acme"}
	for _, test := range []struct {
		name     string
		resolver resolverStub
		status   int
	}{
		{name: "authorized", resolver: resolverStub{resolved: valid}, status: http.StatusNoContent},
		{name: "unauthorized", resolver: resolverStub{err: errors.New("invalid")}, status: http.StatusUnauthorized},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequireTenant(test.resolver))
			router.GET("/", func(c *gin.Context) {
				if _, err := tenant.FromContext(c.Request.Context()); err != nil {
					t.Fatalf("tenant missing from context: %v", err)
				}
				c.Status(http.StatusNoContent)
			})
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set(tenantHeader, "acme")
			request.Header.Set(apiKeyHeader, "key")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Fatalf("status = %d", recorder.Code)
			}
		})
	}
}
