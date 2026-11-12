package usertransport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	"github.com/al-tokarev/gophkeeper/internal/transport"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *registerRequest) validate() error {
	if len(strings.TrimSpace(r.Email)) == 0 {
		return errors.New("email is required")
	}
	if len(strings.TrimSpace(r.Password)) == 0 {
		return errors.New("password is required")
	}
	return nil
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warnw("Decode failed", "error", err.Error())
		transport.RespondError(w, http.StatusBadRequest, "DECODE_ERROR", err.Error())
		return
	}

	if err := request.validate(); err != nil {
		transport.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	passwordHash, err := auth.GetPasswordHash(request.Password)
	if err != nil {
		transport.RespondError(w, http.StatusBadRequest, "GEN_PASSWORD_ERROR", err.Error())
	}

	u := user.User{
		Email:        request.Email,
		PasswordHash: passwordHash,
	}

	createdUser, err := h.service.Create(r.Context(), &u)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrDuplicateEmail):
			transport.RespondError(w, http.StatusConflict, "EMAIL_EXISTS", "email already exists")
		default:
			h.logger.Errorw("create user failed", "error", err)
			transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		}
		return
	}

	transport.RespondJSON(w, http.StatusCreated, createdUser)
}
