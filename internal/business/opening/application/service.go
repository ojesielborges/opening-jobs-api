package application

import (
	"context"

	"openingjobs/internal/business/opening/domain"

	"github.com/google/uuid"
)

type Service struct{ repository domain.Repository }

func NewService(repository domain.Repository) *Service { return &Service{repository: repository} }

func (s *Service) Create(ctx context.Context, opening domain.Opening) (domain.Opening, error) {
	if err := opening.Validate(); err != nil {
		return domain.Opening{}, err
	}
	return s.repository.Create(ctx, opening)
}

func (s *Service) Get(ctx context.Context, id string) (domain.Opening, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.Opening{}, domain.ErrNotFound
	}
	return s.repository.GetByID(ctx, parsedID)
}

func (s *Service) List(ctx context.Context) ([]domain.Opening, error) {
	return s.repository.List(ctx)
}

func (s *Service) Update(ctx context.Context, id string, opening domain.Opening) (domain.Opening, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.Opening{}, domain.ErrNotFound
	}
	opening.ID = parsedID
	if err := opening.Validate(); err != nil {
		return domain.Opening{}, err
	}
	return s.repository.Update(ctx, opening)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrNotFound
	}
	return s.repository.Delete(ctx, parsedID)
}
