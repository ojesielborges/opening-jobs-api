package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSchemaNameAcceptsOnlySafeSlug(t *testing.T) {
	schema, err := SchemaName("acme_1")
	if err != nil || schema != "tenant_acme_1" {
		t.Fatalf("expected safe schema name, got %q, %v", schema, err)
	}
	if _, err := SchemaName("acme; DROP SCHEMA public"); err == nil {
		t.Fatal("expected unsafe slug to be rejected")
	}
}

func TestTenantContextRejectsInvalidTenant(t *testing.T) {
	invalid := Tenant{ID: uuid.New(), SchemaName: "public"}
	if _, err := FromContext(WithContext(context.Background(), invalid)); err == nil {
		t.Fatal("expected invalid tenant context to be rejected")
	}
}
