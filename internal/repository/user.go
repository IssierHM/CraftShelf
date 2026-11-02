package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/IssierHM/CraftShelf/internal/domain"
	"github.com/IssierHM/CraftShelf/internal/repository/sqlc"
)

type UserRepo struct {
	q *sqlc.Queries
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: sqlc.New(pool)}
}

func (r *UserRepo) FindByEmailForAuth(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmailForAuth(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepo) Create(ctx context.Context, email, username, passwordHash string) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if mapped := mapUniqueViolation(err); mapped != nil {
			return nil, mapped
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.q.ExistsUserByEmail(ctx, email)
	if err != nil {
		return false, fmt.Errorf("exists user by email: %w", err)
	}
	return exists, nil
}

func toDomainUser(row sqlc.User) *domain.User {
	return &domain.User{
		ID:           row.ID,
		PublicID:     row.PublicID.String(),
		Username:     row.Username,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Bio:          derefStr(row.Bio),
		AvatarURL:    derefStr(row.AvatarUrl),
		Role:         domain.Role(row.Role),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code != "23505" {
		return nil
	}
	switch pgErr.ConstraintName {
	case "user_email_key":
		return domain.ErrEmailExists
	case "user_username_key":
		return domain.ErrUsernameExists
	default:
		return nil
	}
}

func derefStr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
