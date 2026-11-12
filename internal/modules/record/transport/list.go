package recordtransport

import (
	"net/http"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/al-tokarev/gophkeeper/internal/transport"
	"github.com/google/uuid"
)

type responseListItem struct {
	ID        uuid.UUID         `json:"id"`
	Type      record.RecordType `json:"type"`
	Name      string            `json:"name"`
	MetaData  string            `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func (h *RecordHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		transport.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "empty user ID")
		return
	}

	records, err := h.service.List(r.Context(), userID)
	if err != nil {
		transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "error by create")
		return
	}

	var response []responseListItem
	for _, r := range records {
		response = append(response, responseListItem{
			ID:        r.ID,
			Type:      r.Type,
			Name:      r.Name,
			MetaData:  r.MetaData,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}

	transport.RespondJSON(w, http.StatusOK, response)
}
