package userservice

import (
	"context"

	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, user *user.User) (*user.User, error)
	Login(ctx context.Context, email string, password string) (string, error)
	GetByID(ctx context.Context, id uuid.UUID) (*user.User, error)
}
