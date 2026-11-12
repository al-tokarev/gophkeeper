package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/al-tokarev/gophkeeper/internal/transport"
	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

type Auth struct {
	jwtManager *JWTManager
}

func NewAuth(jwtManager *JWTManager) *Auth {
	return &Auth{
		jwtManager: jwtManager,
	}
}

func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			transport.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing token")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		userID, err := a.jwtManager.ParseToken(tokenString)
		if err != nil {
			transport.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}

// WithUserID кладёт userID в контекст
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}
