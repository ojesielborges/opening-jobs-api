package domain

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrNotFound = errors.New("opening not found")
	ErrInvalid  = errors.New("invalid opening")
)

type Opening struct {
	ID       uuid.UUID
	Role     string
	Company  string
	Location string
	Remote   bool
	Link     string
	Salary   int64
}

func (o Opening) Validate() error {
	if strings.TrimSpace(o.Role) == "" || strings.TrimSpace(o.Company) == "" {
		return fmt.Errorf("%w: role and company are required", ErrInvalid)
	}
	if o.Salary < 0 {
		return fmt.Errorf("%w: salary cannot be negative", ErrInvalid)
	}
	if o.Link != "" {
		u, err := url.ParseRequestURI(o.Link)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("%w: link must be an absolute HTTP(S) URL", ErrInvalid)
		}
	}
	return nil
}

type Repository interface {
	Create(context.Context, Opening) (Opening, error)
	GetByID(context.Context, uuid.UUID) (Opening, error)
	List(context.Context) ([]Opening, error)
	Update(context.Context, Opening) (Opening, error)
	Delete(context.Context, uuid.UUID) error
}
