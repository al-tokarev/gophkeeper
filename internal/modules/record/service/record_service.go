package recordservice

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/al-tokarev/gophkeeper/internal/cryptoutil"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	recordrepository "github.com/al-tokarev/gophkeeper/internal/modules/record/repository"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type recordService struct {
	repo     recordrepository.RecordRepository
	ciph     *cryptoutil.Cipher
	validate *validator.Validate
	logger   *zap.SugaredLogger
}

func NewService(repo recordrepository.RecordRepository, ciph *cryptoutil.Cipher, validate *validator.Validate, logger *zap.SugaredLogger) RecordService {
	return &recordService{
		repo:     repo,
		ciph:     ciph,
		logger:   logger.With(zap.String("component", "record service")),
		validate: validate,
	}
}

func (s *recordService) Create(ctx context.Context, rec *record.Record) (*record.Record, error) {
	if !rec.Type.Valid() {
		return nil, record.ErrInvalidType
	}
	if err := s.validateData(rec.Data, rec.Type); err != nil {
		return nil, err
	}

	ciphData, err := s.ciph.Encrypt(rec.Data)
	if err != nil {
		return nil, err
	}

	rec.Data = json.RawMessage(`"` + ciphData + `"`)

	return s.repo.Create(ctx, rec)
}

func (s *recordService) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	rec, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	decrypt, err := s.ciph.Decrypt(string(rec.Data[1 : len(rec.Data)-1]))
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	rec.Data = decrypt

	return rec, nil
}

func (s *recordService) List(ctx context.Context, userID uuid.UUID) ([]record.Record, error) {
	return s.repo.FindAll(ctx, userID)
}

func (s *recordService) Update(ctx context.Context, id uuid.UUID, rec *record.Record, userID uuid.UUID) (*record.Record, error) {
	r, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if err := s.validateData(rec.Data, r.Type); err != nil {
		s.logger.Warnw("validate for update: %w", err)
		return nil, record.ErrInvalidData
	}

	ciphData, err := s.ciph.Encrypt(rec.Data)
	if err != nil {
		return nil, err
	}

	rec.Data = json.RawMessage(`"` + ciphData + `"`)

	updated, err := s.repo.Update(ctx, id, rec, userID)
	if err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}

	decrypt, err := s.ciph.Decrypt(string(rec.Data[1 : len(rec.Data)-1]))
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	updated.Data = decrypt

	return updated, nil
}

func (s *recordService) MarkDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	return s.repo.MarkDelete(ctx, id, userID)
}

func (s *recordService) validateData(data json.RawMessage, t record.RecordType) error {
	var d any

	switch t {
	case record.TypeLogin:
		d = &record.LoginData{}
	case record.TypeText:
		d = &record.TextData{}
	case record.TypeBinary:
		d = &record.BinaryData{}
	case record.TypeCard:
		d = &record.CardData{}
	default:
		return record.ErrInvalidType
	}

	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("%w: %v", record.ErrInvalidData, err)
	}
	if err := s.validate.Struct(d); err != nil {
		return fmt.Errorf("%w: %v", record.ErrInvalidData, err)
	}

	return nil
}
