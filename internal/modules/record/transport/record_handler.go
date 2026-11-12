package recordtransport

import (
	recordservice "github.com/al-tokarev/gophkeeper/internal/modules/record/service"
	"github.com/go-playground/validator"
	"go.uber.org/zap"
)

type RecordHandler struct {
	service  recordservice.RecordService
	validate *validator.Validate
	logger   *zap.SugaredLogger
}

func NewHandler(service recordservice.RecordService, validate *validator.Validate, logger *zap.SugaredLogger) *RecordHandler {
	return &RecordHandler{
		service:  service,
		validate: validate,
		logger:   logger,
	}
}
