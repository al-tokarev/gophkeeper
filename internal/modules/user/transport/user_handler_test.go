package usertransport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/modules/user"
	"github.com/al-tokarev/gophkeeper/internal/modules/user/repository/mocks"
	userservice "github.com/al-tokarev/gophkeeper/internal/modules/user/service"
	"github.com/al-tokarev/gophkeeper/internal/validation"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newUserHandler(t *testing.T) (*UserHandler, *mocks.MockUserRepository) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	jwt, err := auth.NewJWTManager([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	require.NoError(t, err)

	svc := userservice.NewService(mockRepo, jwt, zap.NewNop().Sugar())
	return NewHandler(svc, validation.New(), zap.NewNop().Sugar()), mockRepo
}

func serveUser(route string, handler http.HandlerFunc, method, path, body string) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Method(method, route, handler)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUserCreate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setup      func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name: "ok", body: `{"email":"a@b.c","password":"secret"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, u *user.User) (*user.User, error) {
						u.ID = uuid.New()
						u.CreatedAt = time.Now()
						u.UpdatedAt = time.Now()
						return u, nil
					})
			},
			wantStatus: http.StatusCreated,
		},
		{name: "bad json", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "empty email", body: `{"email":"","password":"x"}`, wantStatus: http.StatusBadRequest},
		{name: "empty password", body: `{"email":"a@b.c","password":""}`, wantStatus: http.StatusBadRequest},
		{
			name: "duplicate email", body: `{"email":"a@b.c","password":"secret"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, user.ErrDuplicateEmail)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "repo error", body: `{"email":"a@b.c","password":"secret"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo := newUserHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			w := serveUser("/users", h.Create, http.MethodPost, "/users", tt.body)
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestUserLogin(t *testing.T) {
	userID := uuid.New()

	hash, err := auth.GetPasswordHash("secret")
	require.NoError(t, err)

	userWithHash := &user.User{ID: userID, Email: "a@b.c", PasswordHash: hash}

	tests := []struct {
		name       string
		body       string
		setup      func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name: "ok", body: `{"email":"a@b.c","password":"secret"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().FindByEmail(gomock.Any(), "a@b.c").Return(userWithHash, nil)
			},
			wantStatus: http.StatusOK,
		},
		{name: "bad json", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "bad email", body: `{"email":"not-email","password":"secret"}`, wantStatus: http.StatusBadRequest},
		{name: "empty password", body: `{"email":"a@b.c","password":""}`, wantStatus: http.StatusBadRequest},
		{
			name: "wrong password", body: `{"email":"a@b.c","password":"wrong"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().FindByEmail(gomock.Any(), "a@b.c").Return(userWithHash, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "user not found", body: `{"email":"a@b.c","password":"secret"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().FindByEmail(gomock.Any(), "a@b.c").Return(nil, user.ErrNotFound)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "repo error", body: `{"email":"a@b.c","password":"secret"}`,
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().FindByEmail(gomock.Any(), "a@b.c").Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo := newUserHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			w := serveUser("/users/login", h.Login, http.MethodPost, "/users/login", tt.body)
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestUserGetByID(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		pathID     string
		setup      func(*mocks.MockUserRepository)
		wantStatus int
	}{
		{
			name: "ok", pathID: userID.String(),
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().FindByID(gomock.Any(), userID).
					Return(&user.User{ID: userID, Email: "a@b.c"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{name: "bad id", pathID: "not-a-uuid", wantStatus: http.StatusBadRequest},
		{
			name: "not found", pathID: userID.String(),
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().FindByID(gomock.Any(), userID).Return(nil, user.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "repo error", pathID: userID.String(),
			setup: func(m *mocks.MockUserRepository) {
				m.EXPECT().FindByID(gomock.Any(), userID).Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo := newUserHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo)
			}
			w := serveUser("/users/{id}", h.GetByID, http.MethodGet, "/users/"+tt.pathID, "")
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}
