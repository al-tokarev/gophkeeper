package recordrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type recordRepository struct {
	db     *sql.DB
	logger *zap.SugaredLogger
}

func NewRepository(db *sql.DB, logger *zap.SugaredLogger) RecordRepository {
	return &recordRepository{
		db:     db,
		logger: logger.With(zap.String("component", "record repository")),
	}
}

func (r *recordRepository) Create(ctx context.Context, rec *record.Record) (*record.Record, error) {
	r.logger.Infow("start create record")

	stmt, err := r.db.PrepareContext(ctx, "INSERT INTO records (user_id, name, type, data, metadata) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at, updated_at, deleted_at")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, rec.UserID, rec.Name, rec.Type, rec.Data, rec.MetaData).Scan(&rec.ID, &rec.CreatedAt, &rec.UpdatedAt, &rec.DeletedAt)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (r *recordRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	r.logger.Infow("start find record")

	stmt, err := r.db.PrepareContext(ctx, "SELECT name, type, data, metadata, created_at, updated_at, deleted_at FROM records WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id, userID)

	rec := record.Record{
		ID:     id,
		UserID: userID,
	}
	err = row.Scan(&rec.Name, &rec.Type, &rec.Data, &rec.MetaData, &rec.CreatedAt, &rec.UpdatedAt, &rec.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, record.ErrNotFound
		}
		return nil, fmt.Errorf("error by row scan: %w", err)
	}

	return &rec, nil
}

func (r *recordRepository) FindAll(ctx context.Context, userID uuid.UUID) ([]record.Record, error) {
	r.logger.Infow("start find records", "user_id", userID)

	stmt, err := r.db.PrepareContext(ctx, "SELECT id, type, name, metadata, created_at, updated_at FROM records WHERE user_id = $1 AND deleted_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("by prepare: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
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

func (r *recordRepository) Update(ctx context.Context, id uuid.UUID, rec *record.Record, userID uuid.UUID) (*record.Record, error) {
	r.logger.Infow("start record update")

	query := `
		UPDATE records
		SET name = $1, data = $2, metadata = $3, updated_at = now()
		WHERE id = $4 AND user_id = $5 AND deleted_at IS NULL
		RETURNING id, user_id, type, name, data, metadata, created_at, updated_at, deleted_at
	`

	var out record.Record
	err := r.db.QueryRowContext(ctx, query, rec.Name, rec.Data, rec.MetaData, id, userID).Scan(
		&out.ID, &out.UserID, &out.Type, &out.Name,
		&out.Data, &out.MetaData, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, record.ErrNotFound
		}
		return nil, fmt.Errorf("update record: %w", err)
	}

	return &out, nil
}

func (r *recordRepository) MarkDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*record.Record, error) {
	r.logger.Infow("start record mark delete")

	query := `
		UPDATE records
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		RETURNING id, name, metadata, created_at, updated_at, deleted_at
	`

	var out record.Record
	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&out.ID, &out.Name, &out.MetaData, &out.CreatedAt, &out.UpdatedAt, &out.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, record.ErrNotFound
		}
		return nil, fmt.Errorf("mark delete record: %w", err)
	}

	return &out, nil
}
