package http

import (
	"net/http"

	"openingjobs/internal/tenant"
	"openingjobs/pkg/response"

	"github.com/gin-gonic/gin"
)

const (
	tenantHeader = "X-Tenant"
	apiKeyHeader = "X-Tenant-Key"
)

func RequireTenant(resolver tenant.Resolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		resolved, err := resolver.Resolve(c.Request.Context(), c.GetHeader(tenantHeader), c.GetHeader(apiKeyHeader))
		if err != nil {
			response.Error(c, http.StatusUnauthorized, tenant.ErrUnauthenticated, "invalid tenant credentials")
			c.Abort()
			return
		}

		c.Request = c.Request.WithContext(tenant.WithContext(c.Request.Context(), resolved))
		c.Next()
	}
}
