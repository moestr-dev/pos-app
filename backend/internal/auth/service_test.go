package auth

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockRepo struct {
	users map[string]*User
	perms map[string][]string
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[string]*User), perms: make(map[string][]string)}
}

func (m *mockRepo) CreateUser(ctx context.Context, username, email, passwordHash string) (*User, error) {
	u := &User{ID: uuid.New(), Username: username, Email: email, PasswordHash: passwordHash, IsActive: true}
	m.users[email] = u
	return u, nil
}

func (m *mockRepo) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepo) ListUsers(ctx context.Context) ([]User, error) {
	out := make([]User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, *u)
	}
	return out, nil
}

func (m *mockRepo) UserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return m.perms[userID.String()], nil
}

func (m *mockRepo) WriteAuditLog(ctx context.Context, userID *uuid.UUID, action, detail string) error {
	return nil
}

func TestRegister_HashesPassword(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	resp, err := svc.Register(context.Background(), "budi", "budi@pos.com", "rahasia123")
	require.NoError(t, err)
	assert.Equal(t, "budi", resp.Username)

	stored := repo.users["budi@pos.com"]
	require.NotNil(t, stored)
	assert.NotEqual(t, "rahasia123", stored.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("rahasia123")))
}

func TestLogin(t *testing.T) {
	seedActive := func(t *testing.T, svc *Service, repo *mockRepo) {
		t.Helper()
		_, err := svc.Register(context.Background(), "budi", "budi@pos.com", "rahasia123")
		require.NoError(t, err)
		u := repo.users["budi@pos.com"]
		repo.perms[u.ID.String()] = []string{"inventory.read", "inventory.transfer"}
	}
	seedInactive := func(t *testing.T, svc *Service, repo *mockRepo) {
		t.Helper()
		_, err := svc.Register(context.Background(), "budi", "budi@pos.com", "rahasia123")
		require.NoError(t, err)
		repo.users["budi@pos.com"].IsActive = false
	}
	seedNone := func(t *testing.T, svc *Service, repo *mockRepo) {}

	tests := []struct {
		name     string
		setup    func(t *testing.T, svc *Service, repo *mockRepo)
		email    string
		password string
		wantErr  error
	}{
		{"success", seedActive, "budi@pos.com", "rahasia123", nil},
		{"wrong password", seedActive, "budi@pos.com", "wrong-password", ErrInvalidCredentials},
		{"user not found", seedNone, "hantu@pos.com", "any-password", ErrInvalidCredentials},
		{"inactive user", seedInactive, "budi@pos.com", "rahasia123", ErrUserInactive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockRepo()
			svc := NewService(repo, "test-secret")
			tt.setup(t, svc, repo)

			token, resp, err := svc.Login(context.Background(), tt.email, tt.password)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.email, resp.Email)

			parsed, err := jwt.Parse(token, func(tok *jwt.Token) (any, error) {
				return []byte("test-secret"), nil
			})
			require.NoError(t, err)
			require.True(t, parsed.Valid)
			claims, ok := parsed.Claims.(jwt.MapClaims)
			require.True(t, ok)
			stored := repo.users[tt.email]
			assert.Equal(t, stored.ID.String(), claims["sub"])
			assert.Equal(t, tt.email, claims["email"])
			assert.Contains(t, claims["perms"], "inventory.read")
		})
	}
}

func TestList_ReturnsDTOs(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	_, err := svc.Register(context.Background(), "budi", "budi@pos.com", "rahasia123")
	require.NoError(t, err)
	_, err = svc.Register(context.Background(), "siti", "siti@pos.com", "rahasia123")
	require.NoError(t, err)

	out, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, out, 2)
	assert.NotNil(t, out)
}

func TestPermissions(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo, "test-secret")

	id := uuid.New()
	repo.perms[id.String()] = []string{"user.read", "inventory.transfer"}

	got, err := svc.Permissions(context.Background(), id)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"user.read", "inventory.transfer"}, got)
}
