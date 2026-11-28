package recordservice

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/al-tokarev/gophkeeper/internal/cryptoutil"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type RecordService struct {
	repo     RecordRepository
	ciph     *cryptoutil.Cipher
	validate *validator.Validate
	logger   *zap.SugaredLogger
}

func NewService(repo RecordRepository, ciph *cryptoutil.Cipher, validate *validator.Validate, logger *zap.SugaredLogger) *RecordService {
	return &RecordService{
		repo:     repo,
		ciph:     ciph,
		logger:   logger.With(zap.String("component", "record service")),
		validate: validate,
	}
}

func (s *RecordService) Create(ctx context.Context, rec *record.Record) (*record.Record, error) {
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

	encoded, err := json.Marshal(ciphData)
	if err != nil {
		return nil, err
	}

	rec.Data = encoded

	return s.repo.Create(ctx, rec)
}

func (s *RecordService) Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	rec, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	plain, err := s.decryptData(rec.Data)
	if err != nil {
		return nil, err
	}
	rec.Data = plain

	return rec, nil
}

func (s *RecordService) List(ctx context.Context, userID uuid.UUID, searchFilter string) ([]record.Record, error) {
	return s.repo.FindAll(ctx, userID, searchFilter)
}

func (s *RecordService) Update(ctx context.Context, id uuid.UUID, rec *record.Record, userID uuid.UUID) (*record.Record, error) {
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
	encoded, err := json.Marshal(ciphData)
	if err != nil {
		return nil, err
	}
	rec.Data = encoded

	updated, err := s.repo.Update(ctx, id, rec, userID)
	if err != nil {
		return nil, fmt.Errorf("update: %w", err)
	}

	plain, err := s.decryptData(updated.Data)
	if err != nil {
		return nil, err
	}
	updated.Data = plain

	return updated, nil
}

func (s *RecordService) MarkDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	return s.repo.MarkDelete(ctx, id, userID)
}

func (s *RecordService) validateData(data json.RawMessage, t record.RecordType) error {
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

// decryptData извлекает зашифрованную строку из JSON и расшифровывает её.
func (s *RecordService) decryptData(raw json.RawMessage) (json.RawMessage, error) {
	var enc string
	if err := json.Unmarshal(raw, &enc); err != nil {
		return nil, fmt.Errorf("decode encrypted payload: %w", err)
	}
	if enc == "" {
		return nil, fmt.Errorf("decode encrypted payload: empty")
	}

	plain, err := s.ciph.Decrypt(enc)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plain, nil
}
