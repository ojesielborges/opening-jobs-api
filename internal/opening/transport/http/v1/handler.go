package v1

import (
	"context"
	"errors"
	"net/http"

	"openingjobs/internal/opening/domain"
	httperrors "openingjobs/internal/platform/http/errors"
	"openingjobs/pkg/response"

	"github.com/gin-gonic/gin"
)

type openingService interface {
	Create(context.Context, domain.Opening) (domain.Opening, error)
	Get(context.Context, string) (domain.Opening, error)
	List(context.Context) ([]domain.Opening, error)
	Update(context.Context, string, domain.Opening) (domain.Opening, error)
	Delete(context.Context, string) error
}

type Handler struct{ service openingService }

func NewHandler(service openingService) *Handler { return &Handler{service: service} }

func RegisterRoutes(routes *gin.RouterGroup, handler *Handler) {
	routes.POST("/openings", handler.create)
	routes.GET("/openings", handler.list)
	routes.GET("/openings/:id", handler.get)
	routes.PUT("/openings/:id", handler.update)
	routes.DELETE("/openings/:id", handler.delete)
}

func (h *Handler) create(c *gin.Context) {
	var request openingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperrors.BadRequest(c, err)
		return
	}
	opening, err := h.service.Create(c.Request.Context(), request.toDomain())
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, toResponse(opening))
}

func (h *Handler) get(c *gin.Context) {
	opening, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, toResponse(opening))
}

func (h *Handler) list(c *gin.Context) {
	openings, err := h.service.List(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	result := make([]openingResponse, 0, len(openings))
	for _, opening := range openings {
		result = append(result, toResponse(opening))
	}
	response.Success(c, http.StatusOK, result)
}

func (h *Handler) update(c *gin.Context) {
	var request openingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperrors.BadRequest(c, err)
		return
	}
	opening, err := h.service.Update(c.Request.Context(), c.Param("id"), request.toDomain())
	if err != nil {
		writeError(c, err)
		return
	}
	response.Success(c, http.StatusOK, toResponse(opening))
}

func (h *Handler) delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeError(c *gin.Context, err error) {
	httperrors.Write(c, err, writeDomainError)
}

func writeDomainError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		response.Error(c, http.StatusNotFound, err, "opening not found")
		return true
	case errors.Is(err, domain.ErrInvalid):
		response.Error(c, http.StatusUnprocessableEntity, err, "invalid opening")
		return true
	}
	return false
}
