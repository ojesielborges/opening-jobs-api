package tenant

import (
	"context"
	"testing"
)

func TestProvisionRejectsInvalidInputBeforeDatabaseAccess(t *testing.T) {
	provisioner := NewProvisioner(nil)
	if err := provisioner.Provision(context.Background(), "acme", "short"); err == nil {
		t.Fatal("expected API key validation error")
	}
	if err := provisioner.Provision(context.Background(), "invalid slug", "12345678901234567890123456789012"); err == nil {
		t.Fatal("expected slug validation error")
	}
}
