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
	"go.uber.org/zap"
)

type userRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewRepository(db *sql.DB, logger *zap.SugaredLogger) UserRepository {
	return &userRepository{
		db:     db,
		logger: logger.With(zap.String("component", "user repository")),
	}
}

func (r *userRepository) Create(ctx context.Context, u *user.User) (*user.User, error) {
	r.logger.Infow("Start user create")

	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at, updated_at")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, u.Email, u.PasswordHash).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, user.ErrDuplicateEmail
		}
		return nil, err
	}

	return u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	r.logger.Infow("Start find user by id")

	stmt, err := r.db.PrepareContext(ctx, "SELECT id,email,created_at,updated_at FROM users WHERE id = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)
	var u user.User
	err = row.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	r.logger.Infow("Start find user by email")

	stmt, err := r.db.PrepareContext(ctx, "SELECT id, email, password_hash FROM users WHERE email = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)

	var u user.User
	err = row.Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("error by row scan: %w", err)
	}

	return &u, nil
}
