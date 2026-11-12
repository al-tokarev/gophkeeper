package recordtransport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/al-tokarev/gophkeeper/internal/transport"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type updateRequest struct {
	Name     string          `json:"name"`
	Data     json.RawMessage `json:"data"`
	MetaData string          `json:"metadata"`
}

func (h *RecordHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)

	if err != nil {
		h.logger.Warnw("uuid parse", "error", err)
		transport.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "incorrect id")
		return
	}

	var request updateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warnw("Decode failed", "error", err.Error())
		transport.RespondError(w, http.StatusBadRequest, "DECODE_ERROR", err.Error())
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		transport.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "empty user ID")
		return
	}

	rec := record.Record{
		UserID:   userID,
		Name:     request.Name,
		Data:     request.Data,
		MetaData: request.MetaData,
	}

	updatedRec, err := h.service.Update(r.Context(), id, &rec, userID)
	if err != nil {
		switch {
		case errors.Is(err, record.ErrNotFound):
			transport.RespondError(w, http.StatusNotFound, "NOT_FOUND", "record not found")
		case errors.Is(err, record.ErrInvalidData):
			transport.RespondError(w, http.StatusBadRequest, "INVALID", err.Error())
		default:
			h.logger.Warnw("get record failed", "error", err)
			transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		}
		return
	}

	transport.RespondJSON(w, http.StatusOK, updatedRec)
}
