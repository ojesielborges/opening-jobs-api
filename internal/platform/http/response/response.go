package response

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Success[T any](c *gin.Context, status int, data T) {
	c.JSON(status, APIResponse[T]{
		Success: true,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, err error, message string) {
	c.JSON(status, APIResponse[any]{
		Success: false,
		Message: message,
		Error:   err.Error(),
	})
}

func NotFoundError() error {
	return errors.New("not found")
}
