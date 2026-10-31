package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSuccessAndError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name   string
		write  func(*gin.Context)
		status int
		ok     bool
	}{
		{name: "success", write: func(c *gin.Context) { Success(c, http.StatusCreated, map[string]string{"id": "1"}) }, status: http.StatusCreated, ok: true},
		{name: "error", write: func(c *gin.Context) { Error(c, http.StatusBadRequest, errors.New("bad input"), "invalid request") }, status: http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			test.write(context)
			if recorder.Code != test.status {
				t.Fatalf("status = %d", recorder.Code)
			}
			var body APIResponse[json.RawMessage]
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.Success != test.ok {
				t.Fatalf("body = %s, error = %v", recorder.Body.String(), err)
			}
		})
	}
}
