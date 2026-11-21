package userrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type UserRepository struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewRepository(db *pgxpool.Pool, logger *zap.SugaredLogger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger.With(zap.String("component", "user repository")),
	}
}

func (r *UserRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	r.logger.Infow("Start user create")

	query := "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at"

	err := r.db.QueryRow(ctx, query, u.Email, u.PasswordHash).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, user.ErrDuplicateEmail
		}
		return nil, err
	}

	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	r.logger.Infow("Start find user by id")

	query := "SELECT id,email,created_at,updated_at FROM users WHERE id = $1"
	row := r.db.QueryRow(ctx, query, id)

	var u user.User
	err := row.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	r.logger.Infow("Start find user by email")

	query := "SELECT id, email, password_hash FROM users WHERE email = $1"
	row := r.db.QueryRow(ctx, query, email)

	var u user.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("error by row scan: %w", err)
	}

	return &u, nil
}
