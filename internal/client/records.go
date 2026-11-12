package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Record повторяет модель записи на сервере.
type Record struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data"`
	MetaData  string          `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at"`
}

// CreateRecord вызывает POST /records.
func (c *Client) CreateRecord(ctx context.Context, rec Record) (*Record, error) {
	respBody, err := c.do(ctx, "POST", "/records", rec, true)
	if err != nil {
		return nil, err
	}
	var out Record
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("client: decode create: %w", err)
	}
	return &out, nil
}

// ListRecords вызывает GET /records.
func (c *Client) ListRecords(ctx context.Context) ([]Record, error) {
	respBody, err := c.do(ctx, "GET", "/records", nil, true)
	if err != nil {
		return nil, err
	}
	var out []Record
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("client: decode list: %w", err)
	}
	return out, nil
}

// GetRecord вызывает GET /records/{id}.
func (c *Client) GetRecord(ctx context.Context, id string) (*Record, error) {
	respBody, err := c.do(ctx, "GET", "/records/"+id, nil, true)
	if err != nil {
		return nil, err
	}
	var out Record
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("client: decode get: %w", err)
	}
	return &out, nil
}

// UpdateRecord вызывает PUT /records/{id}.
func (c *Client) UpdateRecord(ctx context.Context, id string, rec Record) (*Record, error) {
	respBody, err := c.do(ctx, "PUT", "/records/"+id, rec, true)
	if err != nil {
		return nil, err
	}
	var out Record
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("client: decode update: %w", err)
	}
	return &out, nil
}

// DeleteRecord вызывает DELETE /records/{id}.
func (c *Client) DeleteRecord(ctx context.Context, id string) (*Record, error) {
	respBody, err := c.do(ctx, "DELETE", "/records/"+id, nil, true)
	if err != nil {
		return nil, err
	}
	var out Record
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("client: decode delete: %w", err)
	}
	return &out, nil
}
