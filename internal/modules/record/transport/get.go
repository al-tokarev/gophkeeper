package recordtransport

import (
	"errors"
	"net/http"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/al-tokarev/gophkeeper/internal/transport"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (h *RecordHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	rec, err := h.service.Get(r.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, record.ErrNotFound):
			transport.RespondError(w, http.StatusNotFound, "NOT_FOUND", "record not found")
		default:
			h.logger.Errorw("get record failed", "error", err)
			transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		}
		return
	}

	transport.RespondJSON(w, http.StatusOK, rec)
}
