package tenant

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type resolverStub struct {
	mu       sync.Mutex
	calls    int
	resolved Tenant
	err      error
}

func (s *resolverStub) Resolve(context.Context, string, string) (Tenant, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.resolved, s.err
}

func (s *resolverStub) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.calls
}

func TestCachedResolverCachesSuccessfulResolution(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	next := &resolverStub{resolved: Tenant{ID: uuid.New(), Slug: "acme", SchemaName: "tenant_acme"}}
	resolver := NewCachedResolver(next, time.Minute, time.Second)
	resolver.now = func() time.Time { return now }

	for range 2 {
		got, err := resolver.Resolve(context.Background(), "acme", "a-valid-key")
		if err != nil || got != next.resolved {
			t.Fatalf("Resolve() = %#v, %v", got, err)
		}
	}
	if next.count() != 1 {
		t.Fatalf("resolver called %d times, want 1", next.count())
	}
}

func TestCachedResolverExpiresEntries(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	next := &resolverStub{resolved: Tenant{ID: uuid.New(), Slug: "acme", SchemaName: "tenant_acme"}}
	resolver := NewCachedResolver(next, time.Minute, time.Second)
	resolver.now = func() time.Time { return now }

	_, _ = resolver.Resolve(context.Background(), "acme", "a-valid-key")
	now = now.Add(time.Minute)
	_, _ = resolver.Resolve(context.Background(), "acme", "a-valid-key")
	if next.count() != 2 {
		t.Fatalf("resolver called %d times, want 2", next.count())
	}
}

func TestCachedResolverCachesOnlyAuthenticationFailures(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	next := &resolverStub{err: ErrUnauthenticated}
	resolver := NewCachedResolver(next, time.Minute, time.Second)
	resolver.now = func() time.Time { return now }

	for range 2 {
		_, err := resolver.Resolve(context.Background(), "acme", "wrong-key")
		if !errors.Is(err, ErrUnauthenticated) {
			t.Fatalf("Resolve() error = %v", err)
		}
	}
	if next.count() != 1 {
		t.Fatalf("resolver called %d times, want 1", next.count())
	}

	next.err = errors.New("database unavailable")
	_, err := resolver.Resolve(context.Background(), "other", "key")
	if err == nil {
		t.Fatal("expected database error")
	}
	_, _ = resolver.Resolve(context.Background(), "other", "key")
	if next.count() != 3 {
		t.Fatalf("resolver called %d times, want 3", next.count())
	}
}
