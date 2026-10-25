package application

import (
	"context"
	"errors"
	"testing"

	"openingjobs/internal/opening/domain"

	"github.com/google/uuid"
)

type repositorySpy struct {
	created domain.Opening
}

func (r *repositorySpy) Create(_ context.Context, opening domain.Opening) (domain.Opening, error) {
	r.created = opening
	return opening, nil
}

func (r *repositorySpy) GetByID(context.Context, uuid.UUID) (domain.Opening, error) {
	return domain.Opening{}, errors.New("not implemented")
}

func (r *repositorySpy) List(context.Context) ([]domain.Opening, error) {
	return nil, errors.New("not implemented")
}

func (r *repositorySpy) Update(context.Context, domain.Opening) (domain.Opening, error) {
	return domain.Opening{}, errors.New("not implemented")
}

func (r *repositorySpy) Delete(context.Context, uuid.UUID) error {
	return errors.New("not implemented")
}

func TestCreateRejectsInvalidOpeningBeforeRepository(t *testing.T) {
	repository := &repositorySpy{}
	service := NewService(repository)

	_, err := service.Create(context.Background(), domain.Opening{
		Role:    "Developer",
		Company: "ACME",
		Salary:  -1,
	})
	if !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected invalid-opening error, got %v", err)
	}
	if repository.created.Role != "" {
		t.Fatal("repository must not be called for invalid input")
	}
}
