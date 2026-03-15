package user_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/edgekit/edgekit/internal/core/auth"
	"github.com/edgekit/edgekit/internal/core/user"
	"github.com/edgekit/edgekit/pkg/apperror"
	"github.com/edgekit/edgekit/pkg/jwt"
)

type mockUserRepo struct {
	users   map[string]*user.User
	byEmail map[string]*user.User
	createFn func(ctx context.Context, u *user.User) error
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[string]*user.User),
		byEmail: make(map[string]*user.User),
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, u)
	}
	u.ID = "user-1"
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	m.users[u.ID] = u
	m.byEmail[u.Email] = u
	return nil
}

func (m *mockUserRepo) FindByID(_ context.Context, id string) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}

func (m *mockUserRepo) Update(_ context.Context, u *user.User) error {
	m.users[u.ID] = u
	m.byEmail[u.Email] = u
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, id string) error {
	if u, ok := m.users[id]; ok {
		delete(m.byEmail, u.Email)
		delete(m.users, id)
	}
	return nil
}

func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]*user.User, int64, error) {
	var all []*user.User
	for _, u := range m.users {
		all = append(all, u)
	}
	return all, int64(len(all)), nil
}

type mockUserCache struct {
	store        map[string]*user.User
	getCalled    bool
	setCalled    bool
	invalidated  map[string]bool
}

func newMockUserCache() *mockUserCache {
	return &mockUserCache{
		store:       make(map[string]*user.User),
		invalidated: make(map[string]bool),
	}
}

func (m *mockUserCache) Get(_ context.Context, id string) (*user.User, error) {
	m.getCalled = true
	u, ok := m.store[id]
	if !ok {
		return nil, errors.New("cache miss")
	}
	return u, nil
}

func (m *mockUserCache) Set(_ context.Context, u *user.User) error {
	m.setCalled = true
	m.store[u.ID] = u
	return nil
}

func (m *mockUserCache) Invalidate(_ context.Context, id string) error {
	m.invalidated[id] = true
	delete(m.store, id)
	return nil
}

type mockHasher struct {
	hashFn   func(password string) (string, error)
	verifyFn func(password, hash string) error
}

func (m *mockHasher) Hash(password string) (string, error) {
	if m.hashFn != nil {
		return m.hashFn(password)
	}
	return "hashed-" + password, nil
}

func (m *mockHasher) Verify(password, hash string) error {
	if m.verifyFn != nil {
		return m.verifyFn(password, hash)
	}
	if hash != "hashed-"+password {
		return errors.New("password mismatch")
	}
	return nil
}

func setupTokenService(t *testing.T) *auth.TokenService {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519 key: %v", err)
	}

	dir := t.TempDir()

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal private key: %v", err)
	}
	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	privPath := filepath.Join(dir, "private.pem")
	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		t.Fatalf("write private key: %v", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	pubPath := filepath.Join(dir, "public.pem")
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	mgr, err := jwt.NewManager(jwt.Config{
		PrivateKeyPath: privPath,
		PublicKeyPath:  pubPath,
		AccessTTL:      15 * time.Minute,
		RefreshTTL:     24 * time.Hour,
	})
	if err != nil {
		t.Fatalf("create jwt manager: %v", err)
	}

	return auth.NewTokenService(mgr)
}

func setupService(t *testing.T) (*user.Service, *mockUserRepo, *mockUserCache, *mockHasher) {
	t.Helper()
	repo := newMockUserRepo()
	cache := newMockUserCache()
	hasher := &mockHasher{}
	tokenSvc := setupTokenService(t)
	svc := user.NewService(repo, cache, hasher, tokenSvc)
	return svc, repo, cache, hasher
}

func TestRegister(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, _, _, _ := setupService(t)
		ctx := context.Background()

		u, err := svc.Register(ctx, user.CreateUserInput{
			Username: "alice",
			Email:    "alice@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Username != "alice" {
			t.Errorf("username = %q, want %q", u.Username, "alice")
		}
		if u.Email != "alice@example.com" {
			t.Errorf("email = %q, want %q", u.Email, "alice@example.com")
		}
		if u.PasswordHash != "hashed-password123" {
			t.Errorf("password hash = %q, want %q", u.PasswordHash, "hashed-password123")
		}
		if u.Role != user.RoleUser {
			t.Errorf("role = %q, want %q", u.Role, user.RoleUser)
		}
	})

	t.Run("EmailAlreadyExists", func(t *testing.T) {
		svc, repo, _, _ := setupService(t)
		ctx := context.Background()

		repo.byEmail["alice@example.com"] = &user.User{
			ID:    "existing-1",
			Email: "alice@example.com",
		}

		_, err := svc.Register(ctx, user.CreateUserInput{
			Username: "alice",
			Email:    "alice@example.com",
			Password: "password123",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeConflict) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeConflict)
		}
	})

	t.Run("HashFailure", func(t *testing.T) {
		svc, _, _, hasher := setupService(t)
		ctx := context.Background()

		hasher.hashFn = func(string) (string, error) {
			return "", errors.New("hash exploded")
		}

		_, err := svc.Register(ctx, user.CreateUserInput{
			Username: "alice",
			Email:    "alice@example.com",
			Password: "password123",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if apperror.Is(err, apperror.CodeConflict) {
			t.Error("should not be a conflict error")
		}
	})
}

func TestLogin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, repo, _, _ := setupService(t)
		ctx := context.Background()

		repo.byEmail["alice@example.com"] = &user.User{
			ID:           "user-1",
			Username:     "alice",
			Email:        "alice@example.com",
			PasswordHash: "hashed-password123",
			Role:         user.RoleUser,
			Scopes:       []string{"read", "write"},
		}

		result, err := svc.Login(ctx, user.LoginInput{
			Email:    "alice@example.com",
			Password: "password123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken == "" {
			t.Error("access token is empty")
		}
		if result.RefreshToken == "" {
			t.Error("refresh token is empty")
		}
		if result.ExpiresIn <= 0 {
			t.Errorf("expires_in = %d, want > 0", result.ExpiresIn)
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		svc, _, _, _ := setupService(t)
		ctx := context.Background()

		_, err := svc.Login(ctx, user.LoginInput{
			Email:    "nobody@example.com",
			Password: "password123",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeUnauthorized) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeUnauthorized)
		}
	})

	t.Run("WrongPassword", func(t *testing.T) {
		svc, repo, _, _ := setupService(t)
		ctx := context.Background()

		repo.byEmail["alice@example.com"] = &user.User{
			ID:           "user-1",
			Email:        "alice@example.com",
			PasswordHash: "hashed-correctpassword",
			Role:         user.RoleUser,
		}

		_, err := svc.Login(ctx, user.LoginInput{
			Email:    "alice@example.com",
			Password: "wrongpassword",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeUnauthorized) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeUnauthorized)
		}
	})
}

func TestGetProfile(t *testing.T) {
	t.Run("CacheHit", func(t *testing.T) {
		svc, repo, cache, _ := setupService(t)
		ctx := context.Background()

		cache.store["user-1"] = &user.User{
			ID:       "user-1",
			Username: "alice",
			Email:    "alice@example.com",
			Role:     user.RoleUser,
		}

		profile, err := svc.GetProfile(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.ID != "user-1" {
			t.Errorf("id = %q, want %q", profile.ID, "user-1")
		}
		if profile.Username != "alice" {
			t.Errorf("username = %q, want %q", profile.Username, "alice")
		}
		if cache.getCalled != true {
			t.Error("cache.Get was not called")
		}
		if _, exists := repo.users["user-1"]; exists {
			t.Error("repo should not have been populated — cache should have been used")
		}
	})

	t.Run("CacheMiss", func(t *testing.T) {
		svc, repo, cache, _ := setupService(t)
		ctx := context.Background()

		repo.users["user-1"] = &user.User{
			ID:       "user-1",
			Username: "alice",
			Email:    "alice@example.com",
			Role:     user.RoleUser,
		}

		profile, err := svc.GetProfile(ctx, "user-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.ID != "user-1" {
			t.Errorf("id = %q, want %q", profile.ID, "user-1")
		}
		if !cache.setCalled {
			t.Error("cache.Set was not called after cache miss")
		}
		if _, exists := cache.store["user-1"]; !exists {
			t.Error("user should have been cached after repo fetch")
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		svc, _, _, _ := setupService(t)
		ctx := context.Background()

		_, err := svc.GetProfile(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeNotFound) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeNotFound)
		}
	})
}

func TestUpdateProfile(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		svc, repo, _, _ := setupService(t)
		ctx := context.Background()

		repo.users["user-1"] = &user.User{
			ID:       "user-1",
			Username: "alice",
			Email:    "alice@example.com",
			Role:     user.RoleUser,
		}
		repo.byEmail["alice@example.com"] = repo.users["user-1"]

		profile, err := svc.UpdateProfile(ctx, "user-1", user.UpdateProfileInput{
			Username: "alice-updated",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.Username != "alice-updated" {
			t.Errorf("username = %q, want %q", profile.Username, "alice-updated")
		}
		if repo.users["user-1"].Username != "alice-updated" {
			t.Error("repo was not updated")
		}
	})

	t.Run("UserNotFound", func(t *testing.T) {
		svc, _, _, _ := setupService(t)
		ctx := context.Background()

		_, err := svc.UpdateProfile(ctx, "nonexistent", user.UpdateProfileInput{
			Username: "bob",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !apperror.Is(err, apperror.CodeNotFound) {
			t.Errorf("error code = %v, want %v", err, apperror.CodeNotFound)
		}
	})

	t.Run("CacheInvalidated", func(t *testing.T) {
		svc, repo, cache, _ := setupService(t)
		ctx := context.Background()

		repo.users["user-1"] = &user.User{
			ID:       "user-1",
			Username: "alice",
			Email:    "alice@example.com",
			Role:     user.RoleUser,
		}
		repo.byEmail["alice@example.com"] = repo.users["user-1"]

		cache.store["user-1"] = repo.users["user-1"]

		_, err := svc.UpdateProfile(ctx, "user-1", user.UpdateProfileInput{
			Username: "alice-new",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cache.invalidated["user-1"] {
			t.Error("cache.Invalidate was not called for user-1")
		}
		if _, exists := cache.store["user-1"]; exists {
			t.Error("cached entry should have been removed after invalidation")
		}
	})
}
