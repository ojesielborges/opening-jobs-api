package opening

import (
	openingapp "openingjobs/internal/business/opening/application"
	openingpostgres "openingjobs/internal/business/opening/infrastructure/postgres"
	openinghttpv1 "openingjobs/internal/business/opening/transport/http/v1"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func MountHTTPV1(routes *gin.RouterGroup, db *sqlx.DB) {
	repository := openingpostgres.NewRepository(db)
	service := openingapp.NewService(repository)
	handler := openinghttpv1.NewHandler(service)

	openinghttpv1.RegisterRoutes(routes, handler)
}
