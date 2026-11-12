package recordrepository

import (
	"context"

	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=mocks/mock_record_repository.go -package=mocks github.com/al-tokarev/gophkeeper/internal/modules/record/repository RecordRepository

type RecordRepository interface {
	Create(ctx context.Context, rec *record.Record) (*record.Record, error)
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error)
	FindAll(ctx context.Context, userID uuid.UUID) ([]record.Record, error)
	Update(ctx context.Context, id uuid.UUID, rec *record.Record, userID uuid.UUID) (*record.Record, error)
	MarkDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error)
}
