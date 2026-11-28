package recordtransport

import (
	"errors"
	"net/http"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/al-tokarev/gophkeeper/internal/transport"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

type deleteResponse struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	MetaData  string     `json:"metadata"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (h *RecordHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)

	if err != nil {
		h.logger.Warnw("uuid parse", "error", err)
		transport.RespondError(w, http.StatusBadRequest, "BAD_REQUEST", "incorrect id")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		transport.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "empty user ID")
		return
	}

	deletedRec, err := h.service.MarkDelete(r.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, record.ErrNotFound):
			transport.RespondError(w, http.StatusNotFound, "NOT_FOUND", "record not found")
		default:
			h.logger.Warnw("get record failed", "error", err)
			transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		}
		return
	}

	response := deleteResponse{
		ID:        deletedRec.ID,
		Name:      deletedRec.Name,
		MetaData:  deletedRec.MetaData,
		CreatedAt: deletedRec.CreatedAt,
		UpdatedAt: deletedRec.UpdatedAt,
		DeletedAt: deletedRec.DeletedAt,
	}

	transport.RespondJSON(w, http.StatusOK, response)
}
