package management

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	coreauth "github.com/router-for-me/CLIProxyAPI/v6/sdk/cliproxy/auth"
)

type memoryAuthStore struct {
	mu    sync.Mutex
	items map[string]*coreauth.Auth
}

func (s *memoryAuthStore) List(_ context.Context) ([]*coreauth.Auth, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]*coreauth.Auth, 0, len(s.items))
	for _, item := range s.items {
		out = append(out, item)
	}
	return out, nil
}

func (s *memoryAuthStore) Save(_ context.Context, auth *coreauth.Auth) (string, error) {
	if auth == nil {
		return "", nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.items == nil {
		s.items = make(map[string]*coreauth.Auth)
	}
	s.items[auth.ID] = auth
	return auth.ID, nil
}

func (s *memoryAuthStore) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, id)
	return nil
}

func (s *memoryAuthStore) SetBaseDir(string) {}

type failingAuthStore struct {
	saveErr error
	baseDir string
}

func (s *failingAuthStore) List(context.Context) ([]*coreauth.Auth, error) { return nil, nil }

func (s *failingAuthStore) Save(_ context.Context, auth *coreauth.Auth) (string, error) {
	path := pathFromAuth(auth)
	if path != "" && !filepath.IsAbs(path) {
		path = filepath.Join(s.baseDir, path)
	}
	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			return "", err
		}
		if err := os.WriteFile(path, []byte(`{"type":"codex","email":"alpha@example.com"}`), 0o600); err != nil {
			return "", err
		}
	}
	if s != nil && s.saveErr != nil {
		return "", s.saveErr
	}
	return "", errors.New("save failed")
}

func (s *failingAuthStore) Delete(context.Context, string) error { return nil }

func (s *failingAuthStore) SetBaseDir(dir string) { s.baseDir = dir }

func (s *failingAuthStore) PersistAuthFiles(context.Context, string, ...string) error {
	if s != nil && s.saveErr != nil {
		return s.saveErr
	}
	return errors.New("persist auth failed")
}

func pathFromAuth(auth *coreauth.Auth) string {
	if auth == nil {
		return ""
	}
	if auth.Attributes != nil {
		if path := strings.TrimSpace(auth.Attributes["path"]); path != "" {
			return path
		}
	}
	if fileName := strings.TrimSpace(auth.FileName); fileName != "" {
		return fileName
	}
	return strings.TrimSpace(auth.ID)
}
