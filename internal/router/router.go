package router

import (
	"net/http"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/handler"
	"github.com/al-tokarev/gophkeeper/internal/logger"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// NewRouter возвращает необходимый handler по полученному маршруту
func NewRouter(h *handler.Handler, a *auth.Auth, l *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()

	r.Use(logger.WithLogging(l))

	r.Post("/users", h.User.Create)
	r.Post("/users/login", h.User.Login)

	r.Group(func(r chi.Router) {
		r.Use(a.AuthMiddleware)

		r.Get("/users/{id}", h.User.GetByID)
		r.Post("/records", h.Record.Create)
		r.Get("/records", h.Record.List)
		r.Get("/records/{id}", h.Record.Get)
		r.Put("/records/{id}", h.Record.Update)
		r.Delete("/records/{id}", h.Record.Delete)
	})

	return r
}
