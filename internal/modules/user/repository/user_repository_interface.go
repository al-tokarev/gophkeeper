package userrepository

import (
	"context"

	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/mock_user_repository.go -package=mocks github.com/al-tokarev/gophkeeper/internal/modules/user/repository UserRepository

type UserRepository interface {
	Create(ctx context.Context, u *user.User) (*user.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*user.User, error)
	FindByEmail(ctx context.Context, email string) (*user.User, error)
}
