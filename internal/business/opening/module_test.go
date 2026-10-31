package opening

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMountHTTPV1RegistersOpeningRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	MountHTTPV1(router.Group("/api/v1"), nil)
	routes := make(map[string]bool)
	for _, route := range router.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{
		http.MethodPost + " /api/v1/openings",
		http.MethodGet + " /api/v1/openings",
		http.MethodGet + " /api/v1/openings/:id",
		http.MethodPut + " /api/v1/openings/:id",
		http.MethodDelete + " /api/v1/openings/:id",
	} {
		if !routes[route] {
			t.Fatalf("route %s was not registered", route)
		}
	}
}
