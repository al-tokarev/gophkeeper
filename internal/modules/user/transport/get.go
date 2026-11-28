package usertransport

import (
	"errors"
	"net/http"

	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	"github.com/al-tokarev/gophkeeper/internal/transport"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		transport.RespondError(w, http.StatusBadRequest, "INVALID_ID", "invalid id")
		return
	}

	u, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			transport.RespondError(w, http.StatusNotFound, "NOT_FOUND", "user not found")
		default:
			h.logger.Errorw("get user failed", "error", err)
			transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		}
		return
	}

	transport.RespondJSON(w, http.StatusOK, u)
}
