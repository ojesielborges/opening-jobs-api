package tenant

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
)

var (
	ErrUnauthenticated = errors.New("tenant authentication failed")
	ErrMissing         = errors.New("tenant is missing from context")
	schemaNamePattern  = regexp.MustCompile(`^tenant_[a-z0-9_]{2,50}$`)
	slugPattern        = regexp.MustCompile(`^[a-z0-9_]{2,50}$`)
)

type Tenant struct {
	ID         uuid.UUID
	Slug       string
	SchemaName string
}

func (t Tenant) Valid() bool {
	return t.ID != uuid.Nil && schemaNamePattern.MatchString(t.SchemaName)
}

func SchemaName(slug string) (string, error) {
	if !slugPattern.MatchString(slug) {
		return "", errors.New("tenant slug must contain 2-50 lowercase letters, numbers, or underscores")
	}
	return "tenant_" + slug, nil
}

type Resolver interface {
	Resolve(ctx context.Context, slug, apiKey string) (Tenant, error)
}

type contextKey struct{}

func WithContext(ctx context.Context, value Tenant) context.Context {
	return context.WithValue(ctx, contextKey{}, value)
}

func FromContext(ctx context.Context) (Tenant, error) {
	value, ok := ctx.Value(contextKey{}).(Tenant)
	if !ok || !value.Valid() {
		return Tenant{}, ErrMissing
	}
	return value, nil
}
