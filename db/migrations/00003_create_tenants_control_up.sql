CREATE SCHEMA IF NOT EXISTS control;

CREATE TABLE IF NOT EXISTS control.tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    schema_name VARCHAR(63) UNIQUE NOT NULL,
    api_key_hash CHAR(64) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (slug ~ '^[a-z0-9_]{2,50}$'),
    CHECK (schema_name ~ '^tenant_[a-z0-9_]{2,50}$')
);
