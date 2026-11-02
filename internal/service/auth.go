package service

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/IssierHM/CraftShelf/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, email, username, passwordHash string) (*domain.User, error)
	FindByEmailForAuth(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type TokenSigner interface {
	Sign(userID int64) (string, error)
}

type AuthService struct {
	users     UserRepository
	tokens    TokenSigner
	cost      int
	dummyHash []byte
}

func NewAuthService(users UserRepository, tokens TokenSigner, bcryptcost int) *AuthService {
	dummyHash, error := bcrypt.GenerateFromPassword([]byte("craftshelf-dummy-password"), bcryptcost)
	if error != nil {
		dummyHash, _ = bcrypt.GenerateFromPassword([]byte("craftshelf-dummy-password"), bcrypt.DefaultCost)
	}
	return &AuthService{
		users:     users,
		tokens:    tokens,
		cost:      bcryptcost,
		dummyHash: dummyHash,
	}
}

func (s *AuthService) Register(ctx context.Context, email, username, password string) (*domain.User, error) {
	exists, err := s.users.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrEmailExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.Create(ctx, email, username, string(hash))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.users.FindByEmailForAuth(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(password))
		return "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Sign(user.ID)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return token, nil
}

func (s *AuthService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}
