package postgres

import (
	"context"
	"errors"
	"testing"

	"openingjobs/internal/controlplane/tenant"

	"github.com/jmoiron/sqlx"
)

func TestWithinTransactionRequiresTenantContext(t *testing.T) {
	db := NewTenantDB(nil)
	err := db.WithinTransaction(context.Background(), func(*sqlx.Tx) error { return nil })
	if !errors.Is(err, tenant.ErrMissing) {
		t.Fatalf("WithinTransaction() error = %v", err)
	}
}
