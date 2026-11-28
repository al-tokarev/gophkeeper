package recordrepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type RecordRepository struct {
	db     *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewRepository(db *pgxpool.Pool, logger *zap.SugaredLogger) *RecordRepository {
	return &RecordRepository{
		db:     db,
		logger: logger.With(zap.String("component", "record repository")),
	}
}

func (r *RecordRepository) Create(ctx context.Context, rec *record.Record) (*record.Record, error) {
	r.logger.Infow("start create record")

	query := "INSERT INTO records (user_id, name, type, data, metadata) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at, deleted_at"
	err := r.db.QueryRow(ctx, query, rec.UserID, rec.Name, rec.Type, rec.Data, rec.MetaData).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt, &rec.DeletedAt)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (r *RecordRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	r.logger.Infow("start find record")

	query := "SELECT name, type, data, metadata, created_at, updated_at, deleted_at FROM records WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL"
	row := r.db.QueryRow(ctx, query, id, userID)

	rec := record.Record{
		ID:     id,
		UserID: userID,
	}
	err := row.Scan(&rec.Name, &rec.Type, &rec.Data, &rec.MetaData, &rec.CreatedAt, &rec.UpdatedAt, &rec.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, record.ErrNotFound
		}
		return nil, fmt.Errorf("error by row scan: %w", err)
	}

	return &rec, nil
}

func (r *RecordRepository) FindAll(ctx context.Context, userID uuid.UUID, searchFilter string) ([]record.Record, error) {
	r.logger.Infow("start find records", "user_id", userID)

	query := "SELECT id, type, name, metadata, created_at, updated_at FROM records WHERE user_id = $1 AND deleted_at IS NULL AND ($2 = '' OR metadata ILIKE '%' || $2 || '%')"
	rows, err := r.db.Query(ctx, query, userID, searchFilter)
	if err != nil {
		return nil, fmt.Errorf("by query: %w", err)
	}
	defer rows.Close()

	var result []record.Record
	for rows.Next() {
		var rec record.Record
		err := rows.Scan(&rec.ID, &rec.Type, &rec.Name, &rec.MetaData, &rec.CreatedAt, &rec.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("by scan: %w", err)
		}

		result = append(result, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return result, nil
}

func (r *RecordRepository) Update(ctx context.Context, id uuid.UUID, rec *record.Record, userID uuid.UUID) (*record.Record, error) {
	r.logger.Infow("start record update")

	query := `
		UPDATE records
		SET name = $1, data = $2, metadata = $3, version = version + 1, updated_at = now()
		WHERE id = $4 AND user_id = $5 AND version = $6 AND deleted_at IS NULL
		RETURNING id, user_id, type, name, data, metadata, version, created_at, updated_at, deleted_at
	`

	var out record.Record
	err := r.db.QueryRow(ctx, query, rec.Name, rec.Data, rec.MetaData, id, userID, rec.Version).Scan(
		&out.ID, &out.UserID, &out.Type, &out.Name,
		&out.Data, &out.MetaData, &out.Version, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, record.ErrConflict
		}
		return nil, fmt.Errorf("update record: %w", err)
	}

	return &out, nil
}

func (r *RecordRepository) MarkDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	r.logger.Infow("start record mark delete")

	query := `
		UPDATE records
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, name, metadata, created_at, updated_at, deleted_at
	`

	var out record.Record
	err := r.db.QueryRow(ctx, query, id, userID).Scan(
		&out.ID, &out.Name, &out.MetaData, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, record.ErrNotFound
		}
		return nil, fmt.Errorf("mark delete record: %w", err)
	}

	return &out, nil
}
