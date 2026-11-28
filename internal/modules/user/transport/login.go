package usertransport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	"github.com/al-tokarev/gophkeeper/internal/transport"
)

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Debugw("decode request body failed", "error", err)
		transport.RespondError(w, http.StatusBadRequest, "DECODE_ERROR", "invalid request body")
		return
	}

	if err := h.validate.Struct(request); err != nil {
		h.logger.Debugw("validation failed", "error", err)
		transport.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request data")
		return
	}

	token, err := h.service.Login(r.Context(), request.Email, request.Password)
	if err != nil {
		if errors.Is(err, user.ErrIncorrectLogin) || errors.Is(err, user.ErrNotFound) {
			h.logger.Debugw("invalid credentials", "email", request.Email)
			transport.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "incorrect login/password")
			return
		}

		h.logger.Errorw("login failed", "error", err, "email", request.Email)
		transport.RespondError(w, http.StatusInternalServerError, "INTERNAL", "internal error")
		return
	}

	response := loginResponse{
		Token: token,
	}

	transport.RespondJSON(w, http.StatusOK, response)
}
