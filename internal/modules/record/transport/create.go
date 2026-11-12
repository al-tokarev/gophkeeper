package recordtransport

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/al-tokarev/gophkeeper/internal/transport"
	"github.com/google/uuid"
)

type createRequest struct {
	Type     record.RecordType `json:"type"`
	Name     string            `json:"name"`
	Data     json.RawMessage   `json:"data"`
	MetaData string            `json:"metadata"`
}

type createResponse struct {
	ID        uuid.UUID         `json:"id"`
	Type      record.RecordType `json:"type"`
	Name      string            `json:"name"`
	CreatedAt time.Time         `json:"created_at"`
}

func (h *RecordHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warnw("Decode failed", "error", err.Error())
		transport.RespondError(w, http.StatusBadRequest, "DECODE_ERROR", err.Error())
		return
	}

	if err := h.validate.Struct(request); err != nil {
		transport.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	} else if !request.Type.Valid() {
		transport.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "incorrect type")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		transport.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "empty user ID")
		return
	}

	rec := record.Record{
		UserID:   userID,
		Type:     request.Type,
		Name:     request.Name,
		Data:     request.Data,
		MetaData: request.MetaData,
	}

	createdRec, err := h.service.Create(r.Context(), &rec)
	if err != nil {
		transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "error by create")
		return
	}

	response := createResponse{
		ID:        createdRec.ID,
		Type:      createdRec.Type,
		Name:      createdRec.Name,
		CreatedAt: createdRec.CreatedAt,
	}

	transport.RespondJSON(w, http.StatusCreated, response)
}
