package usertransport

import (
	userservice "github.com/al-tokarev/gophkeeper/internal/modules/user/service"
	"github.com/go-playground/validator"
	"go.uber.org/zap"
)

type UserHandler struct {
	service  userservice.UserService
	validate *validator.Validate
	logger   *zap.SugaredLogger
}

func NewHandler(service userservice.UserService, validate *validator.Validate, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{
		service:  service,
		validate: validate,
		logger:   logger,
	}
}
