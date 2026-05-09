package auth

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

type countingStore struct {
	saveCount atomic.Int32
}

type failingStore struct {
	saveCount atomic.Int32
	err       error
}

func (s *countingStore) List(context.Context) ([]*Auth, error) { return nil, nil }

func (s *countingStore) Save(context.Context, *Auth) (string, error) {
	s.saveCount.Add(1)
	return "", nil
}

func (s *countingStore) Delete(context.Context, string) error { return nil }

func (s *failingStore) List(context.Context) ([]*Auth, error) { return nil, nil }

func (s *failingStore) Save(context.Context, *Auth) (string, error) {
	s.saveCount.Add(1)
	if s.err != nil {
		return "", s.err
	}
	return "", errors.New("save failed")
}

func (s *failingStore) Delete(context.Context, string) error { return nil }

func TestWithSkipPersist_DisablesUpdatePersistence(t *testing.T) {
	store := &countingStore{}
	mgr := NewManager(store, nil, nil)
	auth := &Auth{
		ID:       "auth-1",
		Provider: "antigravity",
		Metadata: map[string]any{"type": "antigravity"},
	}

	if _, err := mgr.Update(context.Background(), auth); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if got := store.saveCount.Load(); got != 1 {
		t.Fatalf("expected 1 Save call, got %d", got)
	}

	ctxSkip := WithSkipPersist(context.Background())
	if _, err := mgr.Update(ctxSkip, auth); err != nil {
		t.Fatalf("Update(skipPersist) returned error: %v", err)
	}
	if got := store.saveCount.Load(); got != 1 {
		t.Fatalf("expected Save call count to remain 1, got %d", got)
	}
}

func TestWithSkipPersist_DisablesRegisterPersistence(t *testing.T) {
	store := &countingStore{}
	mgr := NewManager(store, nil, nil)
	auth := &Auth{
		ID:       "auth-1",
		Provider: "antigravity",
		Metadata: map[string]any{"type": "antigravity"},
	}

	if _, err := mgr.Register(WithSkipPersist(context.Background()), auth); err != nil {
		t.Fatalf("Register(skipPersist) returned error: %v", err)
	}
	if got := store.saveCount.Load(); got != 0 {
		t.Fatalf("expected 0 Save calls, got %d", got)
	}
}

func TestRegister_PersistenceFailureReturnsErrorAndDoesNotMutateManager(t *testing.T) {
	store := &failingStore{err: errors.New("persist failed")}
	mgr := NewManager(store, nil, nil)
	auth := &Auth{
		ID:       "auth-1",
		Provider: "antigravity",
		Metadata: map[string]any{"type": "antigravity"},
	}

	if _, err := mgr.Register(context.Background(), auth); err == nil {
		t.Fatal("expected Register to return persistence error")
	}
	if got := store.saveCount.Load(); got != 1 {
		t.Fatalf("expected 1 Save call, got %d", got)
	}
	if auths := mgr.List(); len(auths) != 0 {
		t.Fatalf("expected manager state to remain empty, got %d auths", len(auths))
	}
}

func TestUpdate_PersistenceFailureReturnsErrorAndPreservesExistingState(t *testing.T) {
	store := &failingStore{err: errors.New("persist failed")}
	mgr := NewManager(nil, nil, nil)
	if _, err := mgr.Register(WithSkipPersist(context.Background()), &Auth{
		ID:       "auth-1",
		Provider: "antigravity",
		Label:    "before",
		Metadata: map[string]any{"type": "antigravity"},
	}); err != nil {
		t.Fatalf("seed Register returned error: %v", err)
	}
	mgr.SetStore(store)

	if _, err := mgr.Update(context.Background(), &Auth{
		ID:       "auth-1",
		Provider: "antigravity",
		Label:    "after",
		Metadata: map[string]any{"type": "antigravity"},
	}); err == nil {
		t.Fatal("expected Update to return persistence error")
	}
	if got := store.saveCount.Load(); got != 1 {
		t.Fatalf("expected 1 Save call, got %d", got)
	}
	auth, ok := mgr.GetByID("auth-1")
	if !ok || auth == nil {
		t.Fatal("expected existing auth to remain present")
	}
	if auth.Label != "before" {
		t.Fatalf("expected existing auth label to remain %q, got %q", "before", auth.Label)
	}
}
