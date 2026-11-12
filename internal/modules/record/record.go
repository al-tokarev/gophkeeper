package record

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Type      RecordType      `json:"type"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data"`
	MetaData  string          `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt *time.Time      `json:"deleted_at"`
}

func (t RecordType) Valid() bool {
	switch t {
	case TypeLogin, TypeText, TypeBinary, TypeCard:
		return true
	default:
		return false
	}
}
