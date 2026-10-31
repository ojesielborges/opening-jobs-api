package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"openingjobs/internal/business/opening/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type serviceStub struct {
	created domain.Opening
	err     error
}

func (s *serviceStub) Create(context.Context, domain.Opening) (domain.Opening, error) {
	return s.created, s.err
}
func (s *serviceStub) Get(context.Context, string) (domain.Opening, error) {
	return domain.Opening{}, s.err
}
func (s *serviceStub) List(context.Context) ([]domain.Opening, error) { return nil, s.err }
func (s *serviceStub) Update(context.Context, string, domain.Opening) (domain.Opening, error) {
	return domain.Opening{}, s.err
}
func (s *serviceStub) Delete(context.Context, string) error { return s.err }

func TestCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	opening := domain.Opening{ID: uuid.New(), Role: "Engineer", Company: "Acme"}
	for _, test := range []struct {
		name    string
		body    string
		service serviceStub
		status  int
	}{
		{name: "created", body: `{"role":"Engineer","company":"Acme"}`, service: serviceStub{created: opening}, status: http.StatusCreated},
		{name: "invalid JSON", body: `{`, status: http.StatusBadRequest},
		{name: "invalid opening", body: `{"role":"Engineer","company":"Acme"}`, service: serviceStub{err: errors.Join(domain.ErrInvalid, errors.New("bad"))}, status: http.StatusUnprocessableEntity},
	} {
		t.Run(test.name, func(t *testing.T) {
			router := gin.New()
			RegisterRoutes(router.Group(""), NewHandler(&test.service))
			request := httptest.NewRequest(http.MethodPost, "/openings", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestRequestAndResponseConversion(t *testing.T) {
	request := openingRequest{Role: "Engineer", Company: "Acme", Remote: true, Salary: 100}
	opening := request.toDomain()
	if opening.Role != request.Role || opening.Company != request.Company || !opening.Remote || opening.Salary != request.Salary {
		t.Fatalf("toDomain() = %#v", opening)
	}
	opening.ID = uuid.New()
	response := toResponse(opening)
	if response.ID != opening.ID.String() || response.Role != opening.Role {
		t.Fatalf("toResponse() = %#v", response)
	}
}
