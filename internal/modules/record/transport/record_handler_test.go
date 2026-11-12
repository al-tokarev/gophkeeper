package recordtransport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/al-tokarev/gophkeeper/internal/auth"
	"github.com/al-tokarev/gophkeeper/internal/cryptoutil"
	"github.com/al-tokarev/gophkeeper/internal/modules/record"
	"github.com/al-tokarev/gophkeeper/internal/modules/record/repository/mocks"
	recordservice "github.com/al-tokarev/gophkeeper/internal/modules/record/service"
	"github.com/al-tokarev/gophkeeper/internal/validation"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func newHandler(t *testing.T) (*RecordHandler, *mocks.MockRecordRepository, *cryptoutil.Cipher) {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockRecordRepository(ctrl)
	ciph, err := cryptoutil.NewCipher([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	svc := recordservice.NewService(mockRepo, ciph, validation.New(), zap.NewNop().Sugar())
	return NewHandler(svc, validation.New(), zap.NewNop().Sugar()), mockRepo, ciph
}

func serve(route string, handler http.HandlerFunc, method, path, body string, userID uuid.UUID) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	r.Method(method, route, handler)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != uuid.Nil {
		req = req.WithContext(auth.WithUserID(req.Context(), userID))
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func encrypted(t *testing.T, c *cryptoutil.Cipher, plaintext string) json.RawMessage {
	t.Helper()
	enc, err := c.Encrypt([]byte(plaintext))
	require.NoError(t, err)
	return json.RawMessage(`"` + enc + `"`)
}

func TestCreate(t *testing.T) {
	userID := uuid.New()
	validBody := `{"type":"login","name":"gmail","data":{"login":"a","password":"b"},"metadata":"mail"}`

	tests := []struct {
		name       string
		body       string
		userID     uuid.UUID
		setup      func(*mocks.MockRecordRepository, *cryptoutil.Cipher)
		wantStatus int
	}{
		{
			name: "ok", body: validBody, userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, rec *record.Record) (*record.Record, error) {
						rec.ID = uuid.New()
						rec.CreatedAt = time.Now()
						return rec, nil
					})
			},
			wantStatus: http.StatusCreated,
		},
		{name: "bad json", body: `{`, userID: userID, wantStatus: http.StatusBadRequest},
		{name: "bad type", body: `{"type":"nope","name":"x","data":{}}`, userID: userID, wantStatus: http.StatusBadRequest},
		{name: "no user", body: validBody, userID: uuid.Nil, wantStatus: http.StatusUnauthorized},
		{
			name: "invalid payload", body: `{"type":"login","name":"x","data":{"login":""}}`, userID: userID,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "repo error", body: validBody, userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo, ciph := newHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo, ciph)
			}
			w := serve("/records", h.Create, http.MethodPost, "/records", tt.body, tt.userID)
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestGet(t *testing.T) {
	userID := uuid.New()
	recordID := uuid.New()

	tests := []struct {
		name       string
		pathID     string
		userID     uuid.UUID
		setup      func(*mocks.MockRecordRepository, *cryptoutil.Cipher)
		wantStatus int
	}{
		{
			name: "ok", pathID: recordID.String(), userID: userID,
			setup: func(m *mocks.MockRecordRepository, c *cryptoutil.Cipher) {
				m.EXPECT().FindByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&record.Record{
						ID: recordID, UserID: userID, Type: record.TypeLogin,
						Name: "gmail",
						Data: encrypted(t, c, `{"login":"a","password":"b"}`),
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{name: "bad id", pathID: "not-a-uuid", userID: userID, wantStatus: http.StatusBadRequest},
		{name: "no user", pathID: recordID.String(), userID: uuid.Nil, wantStatus: http.StatusUnauthorized},
		{
			name: "not found", pathID: recordID.String(), userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				m.EXPECT().FindByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, record.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo, ciph := newHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo, ciph)
			}
			w := serve("/records/{id}", h.Get, http.MethodGet, "/records/"+tt.pathID, "", tt.userID)
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestList(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		userID     uuid.UUID
		setup      func(*mocks.MockRecordRepository, *cryptoutil.Cipher)
		wantStatus int
	}{
		{
			name: "ok", userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				m.EXPECT().FindAll(gomock.Any(), gomock.Any()).
					Return([]record.Record{{ID: uuid.New(), Name: "gmail"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{name: "no user", userID: uuid.Nil, wantStatus: http.StatusUnauthorized},
		{
			name: "repo error", userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				m.EXPECT().FindAll(gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo, ciph := newHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo, ciph)
			}
			w := serve("/records", h.List, http.MethodGet, "/records", "", tt.userID)
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestUpdate(t *testing.T) {
	userID := uuid.New()
	recordID := uuid.New()
	validBody := `{"name":"gmail","data":{"login":"a","password":"b"},"metadata":"mail"}`

	tests := []struct {
		name       string
		pathID     string
		body       string
		userID     uuid.UUID
		setup      func(*mocks.MockRecordRepository, *cryptoutil.Cipher)
		wantStatus int
	}{
		{
			name: "ok", pathID: recordID.String(), body: validBody, userID: userID,
			setup: func(m *mocks.MockRecordRepository, c *cryptoutil.Cipher) {
				m.EXPECT().FindByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&record.Record{ID: recordID, Type: record.TypeLogin, Data: encrypted(t, c, `{}`)}, nil)
				m.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ uuid.UUID, rec *record.Record, _ uuid.UUID) (*record.Record, error) {
						rec.ID = recordID
						return rec, nil
					})
			},
			wantStatus: http.StatusOK,
		},
		{name: "bad id", pathID: "x", body: validBody, userID: userID, wantStatus: http.StatusBadRequest},
		{name: "bad json", pathID: recordID.String(), body: `{`, userID: userID, wantStatus: http.StatusBadRequest},
		{name: "no user", pathID: recordID.String(), body: validBody, userID: uuid.Nil, wantStatus: http.StatusUnauthorized},
		{
			name: "not found", pathID: recordID.String(), body: validBody, userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				m.EXPECT().FindByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, record.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid data", pathID: recordID.String(), body: `{"name":"x","data":{"login":""}}`, userID: userID,
			setup: func(m *mocks.MockRecordRepository, c *cryptoutil.Cipher) {
				m.EXPECT().FindByID(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&record.Record{ID: recordID, Type: record.TypeLogin, Data: encrypted(t, c, `{}`)}, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo, ciph := newHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo, ciph)
			}
			w := serve("/records/{id}", h.Update, http.MethodPut, "/records/"+tt.pathID, tt.body, tt.userID)
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestDelete(t *testing.T) {
	userID := uuid.New()
	recordID := uuid.New()

	tests := []struct {
		name       string
		pathID     string
		userID     uuid.UUID
		setup      func(*mocks.MockRecordRepository, *cryptoutil.Cipher)
		wantStatus int
	}{
		{
			name: "ok", pathID: recordID.String(), userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				now := time.Now()
				m.EXPECT().MarkDelete(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&record.Record{ID: recordID, Name: "gmail", DeletedAt: &now}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{name: "bad id", pathID: "x", userID: userID, wantStatus: http.StatusBadRequest},
		{name: "no user", pathID: recordID.String(), userID: uuid.Nil, wantStatus: http.StatusUnauthorized},
		{
			name: "not found", pathID: recordID.String(), userID: userID,
			setup: func(m *mocks.MockRecordRepository, _ *cryptoutil.Cipher) {
				m.EXPECT().MarkDelete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, record.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, mockRepo, ciph := newHandler(t)
			if tt.setup != nil {
				tt.setup(mockRepo, ciph)
			}
			w := serve("/records/{id}", h.Delete, http.MethodDelete, "/records/"+tt.pathID, "", tt.userID)
			require.Equal(t, tt.wantStatus, w.Code, "body: %s", w.Body.String())
		})
	}
}
