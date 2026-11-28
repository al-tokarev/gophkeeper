package usertransport

import (
	"github.com/go-playground/validator"
	"go.uber.org/zap"
)

type UserHandler struct {
	service  UserService
	validate *validator.Validate
	logger   *zap.SugaredLogger
}

func NewHandler(service UserService, validate *validator.Validate, logger *zap.SugaredLogger) *UserHandler {
	return &UserHandler{
		service:  service,
		validate: validate,
		logger:   logger,
	}
}
