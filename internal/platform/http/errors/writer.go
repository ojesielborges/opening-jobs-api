package errors

import (
	"context"
	"errors"
	"net/http"

	"openingjobs/pkg/response"

	"github.com/gin-gonic/gin"
)

type Mapper func(*gin.Context, error) bool

func BadRequest(c *gin.Context, err error) {
	response.Error(c, http.StatusBadRequest, err, "invalid request")
}

func Write(c *gin.Context, err error, mappers ...Mapper) {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		response.Error(c, http.StatusGatewayTimeout, errors.New("request timeout"), "request timeout")
		return
	case errors.Is(err, context.Canceled):
		c.Abort()
		return
	}

	for _, mapper := range mappers {
		if mapper(c, err) {
			return
		}
	}

	response.Error(c, http.StatusInternalServerError, errors.New("internal server error"), "internal server error")
}
