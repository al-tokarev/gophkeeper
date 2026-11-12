package userservice

import (
	"context"
	"fmt"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	userrepository "github.com/al-tokarev/gophkeeper/internal/modules/user/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type userService struct {
	repo       userrepository.UserRepository
	jwtManager *auth.JWTManager
	logger     *zap.SugaredLogger
}

func NewService(repo userrepository.UserRepository, jwtManager *auth.JWTManager, logger *zap.SugaredLogger) UserService {
	return &userService{
		repo:       repo,
		jwtManager: jwtManager,
		logger:     logger.With(zap.String("component", "user service")),
	}
}

func (s *userService) Create(ctx context.Context, u *user.User) (*user.User, error) {
	created, err := s.repo.Create(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	s.logger.Infow("User is created", "id", created.ID)
	return created, nil
}

func (s *userService) Login(ctx context.Context, email string, password string) (string, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("error find by email: %w", err)
	}

	if !auth.VerifyPassword(password, u.PasswordHash) {
		return "", user.ErrIncorrectLogin
	} else {
		token, err := s.jwtManager.GenerateToken(u.ID)
		if err != nil {
			return "", fmt.Errorf("error by generate token: %w", err)
		}

		return token, nil
	}
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	return s.repo.FindByID(ctx, id)
}
