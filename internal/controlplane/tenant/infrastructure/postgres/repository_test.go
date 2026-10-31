package postgres

import (
	"context"
	"errors"
	"testing"

	"openingjobs/internal/controlplane/tenant"
)

func TestResolveRejectsMissingCredentialsWithoutDatabaseAccess(t *testing.T) {
	repository := NewRepository(nil)
	for _, credentials := range [][2]string{{"", "key"}, {"acme", ""}} {
		_, err := repository.Resolve(context.Background(), credentials[0], credentials[1])
		if !errors.Is(err, tenant.ErrUnauthenticated) {
			t.Fatalf("Resolve() error = %v", err)
		}
	}
}
