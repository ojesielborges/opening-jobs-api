package errors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"openingjobs/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

func TestWriteMapsTimeoutAndCustomErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name   string
		err    error
		mapper Mapper
		status int
	}{
		{name: "timeout", err: context.DeadlineExceeded, status: http.StatusGatewayTimeout},
		{name: "custom", err: errors.New("custom"), status: http.StatusTeapot, mapper: func(c *gin.Context, err error) bool { response.Error(c, http.StatusTeapot, err, "mapped"); return true }},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			Write(context, test.err, test.mapper)
			if recorder.Code != test.status {
				t.Fatalf("status = %d", recorder.Code)
			}
			var body response.APIResponse[json.RawMessage]
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil || body.Success {
				t.Fatalf("body = %s, error = %v", recorder.Body.String(), err)
			}
		})
	}
}
