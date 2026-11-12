package recordservice

import (
	"context"

	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/google/uuid"
)

type RecordService interface {
	Create(ctx context.Context, rec *record.Record) (*record.Record, error)
	Get(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error)
	List(ctx context.Context, userID uuid.UUID) ([]record.Record, error)
	Update(ctx context.Context, id uuid.UUID, rec *record.Record, userID uuid.UUID) (*record.Record, error)
	MarkDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error)
}
