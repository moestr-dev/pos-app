package auth

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserInactive       = errors.New("user is inactive")
)

type Service struct {
	repo      UserRepository
	jwtSecret string
}

func NewService(repo UserRepository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

func (s *Service) Register(ctx context.Context, username, email, password string) (*UserResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u, err := s.repo.CreateUser(ctx, username, email, string(hash))
	if err != nil {
		return nil, err
	}
	detail, _ := json.Marshal(map[string]string{"email": email})
	_ = s.repo.WriteAuditLog(ctx, &u.ID, "user.register", string(detail))
	resp := u.ToResponse()
	return &resp, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, *UserResponse, error) {
	u, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}
	if !u.IsActive {
		return "", nil, ErrUserInactive
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", nil, ErrInvalidCredentials
	}

	perms, err := s.repo.UserPermissions(ctx, u.ID)
	if err != nil {
		return "", nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.ID,
		"email": u.Email,
		"perms": perms,
		"exp":   jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
		"iat":   jwt.NewNumericDate(time.Now()),
	})
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", nil, err
	}
	_ = s.repo.WriteAuditLog(ctx, &u.ID, "user.login", `{}`)
	resp := u.ToResponse()
	return signed, &resp, nil
}

func (s *Service) List(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]UserResponse, 0, len(users))
	for i := range users {
		out = append(out, users[i].ToResponse())
	}
	return out, nil
}

func (s *Service) Permissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.repo.UserPermissions(ctx, userID)
}
