package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"em_tz_anvar/internal/models"
	"em_tz_anvar/internal/repository"
	"em_tz_anvar/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSubscriptionService реализует service.SubscriptionService для тестов
type mockSubscriptionService struct {
	createFn       func(ctx context.Context, req *models.CreateSubscriptionReq) (*models.Subscription, error)
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	getAllFn       func(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error)
	updateFn       func(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionReq) (*models.Subscription, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
	getTotalCostFn func(ctx context.Context, filter *models.CostFilter) (*models.TotalCostResponse, error)
}

func (m *mockSubscriptionService) Create(ctx context.Context, req *models.CreateSubscriptionReq) (*models.Subscription, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}

func (m *mockSubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSubscriptionService) GetAll(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockSubscriptionService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionReq) (*models.Subscription, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return nil, nil
}

func (m *mockSubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockSubscriptionService) GetTotalCost(ctx context.Context, filter *models.CostFilter) (*models.TotalCostResponse, error) {
	if m.getTotalCostFn != nil {
		return m.getTotalCostFn(ctx, filter)
	}
	return nil, nil
}

func handlerWithMock(mock *mockSubscriptionService) *Handler {
	svc := &service.Service{
		Subscription: mock,
	}
	return NewHandler(svc)
}

func TestHandler_CreateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	mock := &mockSubscriptionService{
		createFn: func(ctx context.Context, req *models.CreateSubscriptionReq) (*models.Subscription, error) {
			return &models.Subscription{
				ID:          id,
				ServiceName: req.ServiceName,
				Price:       req.Price,
				UserID:      uuid.MustParse(req.UserID),
				StartDate:   time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.POST("/api/v1/subscriptions", h.CreateSubscription)

	body := `{"service_name":"Yandex","price":400,"user_id":"` + uuid.New().String() + `","start_date":"01-2025"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	var sub models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &sub))
	assert.Equal(t, "Yandex", sub.ServiceName)
	assert.Equal(t, 400, sub.Price)
}

func TestHandler_CreateSubscription_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlerWithMock(&mockSubscriptionService{})
	router := gin.New()
	router.POST("/api/v1/subscriptions", h.CreateSubscription)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_CreateSubscription_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockSubscriptionService{
		createFn: func(ctx context.Context, req *models.CreateSubscriptionReq) (*models.Subscription, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.POST("/api/v1/subscriptions", h.CreateSubscription)

	body := `{"service_name":"Yandex","price":400,"user_id":"` + uuid.New().String() + `","start_date":"01-2025"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandler_GetSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mock := &mockSubscriptionService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
			return &models.Subscription{
				ID:          id,
				ServiceName: "Netflix",
				Price:       500,
				UserID:      id,
				StartDate:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.GET("/api/v1/subscriptions/:id", h.GetSubscription)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var sub models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &sub))
	assert.Equal(t, id, sub.ID)
	assert.Equal(t, "Netflix", sub.ServiceName)
}

func TestHandler_GetSubscription_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlerWithMock(&mockSubscriptionService{})
	router := gin.New()
	router.GET("/api/v1/subscriptions/:id", h.GetSubscription)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/not-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetSubscription_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockSubscriptionService{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
			return nil, repository.ErrNotFound
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.GET("/api/v1/subscriptions/:id", h.GetSubscription)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_GetAllSubscriptions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockSubscriptionService{
		getAllFn: func(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error) {
			return []models.Subscription{
				{ID: uuid.New(), ServiceName: "A", Price: 100, UserID: uuid.New(), StartDate: time.Now(), UpdatedAt: time.Now()},
			}, nil
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.GET("/api/v1/subscriptions", h.GetAllSubscriptions)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var list []models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	assert.Len(t, list, 1)
	assert.Equal(t, "A", list[0].ServiceName)
}

func TestHandler_UpdateSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)
	subID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mock := &mockSubscriptionService{
		updateFn: func(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionReq) (*models.Subscription, error) {
			return &models.Subscription{
				ID:          id,
				ServiceName: "Updated",
				Price:       600,
				UserID:      id,
				StartDate:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.PUT("/api/v1/subscriptions/:id", h.UpdateSubscription)

	body := `{"service_name":"Updated","price":600}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscriptions/"+subID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var sub models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &sub))
	assert.Equal(t, "Updated", sub.ServiceName)
	assert.Equal(t, 600, sub.Price)
}

func TestHandler_UpdateSubscription_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockSubscriptionService{
		updateFn: func(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionReq) (*models.Subscription, error) {
			return nil, repository.ErrNotFound
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.PUT("/api/v1/subscriptions/:id", h.UpdateSubscription)

	body := `{"price":100}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_DeleteSubscription(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockSubscriptionService{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.DELETE("/api/v1/subscriptions/:id", h.DeleteSubscription)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandler_DeleteSubscription_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockSubscriptionService{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return repository.ErrNotFound
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.DELETE("/api/v1/subscriptions/:id", h.DeleteSubscription)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/11111111-1111-1111-1111-111111111111", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandler_GetTotalCost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mock := &mockSubscriptionService{
		getTotalCostFn: func(ctx context.Context, filter *models.CostFilter) (*models.TotalCostResponse, error) {
			return &models.TotalCostResponse{TotalCost: 3600, Currency: "RUB"}, nil
		},
	}
	h := handlerWithMock(mock)
	router := gin.New()
	router.GET("/api/v1/subscriptions/cost", h.GetTotalCost)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/cost?start_date=01-2025&end_date=12-2025", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var resp models.TotalCostResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 3600, resp.TotalCost)
	assert.Equal(t, "RUB", resp.Currency)
}

func TestHandler_GetTotalCost_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlerWithMock(&mockSubscriptionService{})
	router := gin.New()
	router.GET("/api/v1/subscriptions/cost", h.GetTotalCost)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/cost", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandler_GetTotalCost_InvalidDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handlerWithMock(&mockSubscriptionService{})
	router := gin.New()
	router.GET("/api/v1/subscriptions/cost", h.GetTotalCost)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/cost?start_date=01-2025&end_date=invalid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
