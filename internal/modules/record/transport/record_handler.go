package recordtransport

import (
	"github.com/go-playground/validator"
	"go.uber.org/zap"
)

type RecordHandler struct {
	service  RecordService
	validate *validator.Validate
	logger   *zap.SugaredLogger
}

func NewHandler(service RecordService, validate *validator.Validate, logger *zap.SugaredLogger) *RecordHandler {
	return &RecordHandler{
		service:  service,
		validate: validate,
		logger:   logger,
	}
}
