//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"em_tz_anvar/internal/handler"
	"em_tz_anvar/internal/models"
	"em_tz_anvar/internal/repository"
	"em_tz_anvar/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testRouter *gin.Engine

func TestMain(m *testing.M) {
	ctx := context.Background()
	postgresContainer, connStr, err := startPostgres(ctx)
	if err != nil {
		panic("failed to start postgres: " + err.Error())
	}
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	// PostgreSQL в контейнере может быть ещё не готов принимать соединения — повторяем попытки
	var db *sqlx.DB
	for i := 0; i < 15; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err == nil {
			break
		}
		if i < 14 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	if err != nil {
		panic("failed to connect: " + err.Error())
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)
	testRouter = setupRouter(handlers)

	os.Exit(m.Run())
}

func startPostgres(ctx context.Context) (*postgres.PostgresContainer, string, error) {
	// RunContainer — API до v0.32; в v0.32+ можно использовать postgres.Run(ctx, "postgres:16-alpine", ...)
	container, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("subscriptions_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		return nil, "", err
	}

	connStr, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, "", err
	}
	// Контейнер PostgreSQL без SSL — явно отключаем
	if !strings.Contains(connStr, "sslmode=") {
		if strings.Contains(connStr, "?") {
			connStr += "&sslmode=disable"
		} else if strings.HasPrefix(connStr, "postgres://") {
			connStr += "?sslmode=disable"
		} else {
			connStr += " sslmode=disable"
		}
	}
	return container, connStr, nil
}

func runMigrations(db *sqlx.DB) error {
	// Находим migrations относительно корня модуля
	workDir, _ := os.Getwd()
	for i := 0; i < 5; i++ {
		upPath := filepath.Join(workDir, "migrations", "000001_init.up.sql")
		if _, err := os.Stat(upPath); err == nil {
			data, err := os.ReadFile(upPath)
			if err != nil {
				return err
			}
			_, err = db.Exec(string(data))
			return err
		}
		workDir = filepath.Dir(workDir)
	}
	return nil
}

// setupRouter создаёт роутер без swagger (для интеграционных тестов)
func setupRouter(h *handler.Handler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	api := router.Group("/api/v1")
	{
		subs := api.Group("/subscriptions")
		{
			subs.POST("", h.CreateSubscription)
			subs.GET("", h.GetAllSubscriptions)
			subs.GET("/cost", h.GetTotalCost)
			subs.GET("/:id", h.GetSubscription)
			subs.PUT("/:id", h.UpdateSubscription)
			subs.DELETE("/:id", h.DeleteSubscription)
		}
	}
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return router
}

func TestIntegration_Health(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestIntegration_CRUD_Flow(t *testing.T) {
	userID := uuid.New().String()
	createBody := `{
		"service_name": "Yandex Plus",
		"price": 400,
		"user_id": "` + userID + `",
		"start_date": "01-2025",
		"end_date": "12-2025"
	}`

	// Create
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, "create: %s", rec.Body.String())

	var created models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
	assert.NotEqual(t, uuid.Nil, created.ID)
	assert.Equal(t, "Yandex Plus", created.ServiceName)
	assert.Equal(t, 400, created.Price)
	subscriptionID := created.ID.String()

	// GetByID
	req = httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+subscriptionID, nil)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var got models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Yandex Plus", got.ServiceName)

	// GetAll
	req = httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions?user_id="+userID, nil)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var list []models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	assert.GreaterOrEqual(t, len(list), 1)

	// Update
	updateBody := `{"service_name": "Yandex Plus Updated", "price": 500}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/subscriptions/"+subscriptionID, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var updated models.Subscription
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &updated))
	assert.Equal(t, "Yandex Plus Updated", updated.ServiceName)
	assert.Equal(t, 500, updated.Price)

	// GetTotalCost
	req = httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/cost?start_date=01-2025&end_date=12-2025&user_id="+userID, nil)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var costResp models.TotalCostResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &costResp))
	assert.Equal(t, "RUB", costResp.Currency)
	assert.GreaterOrEqual(t, costResp.TotalCost, 0)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/"+subscriptionID, nil)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// GetByID after delete -> 404
	req = httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+subscriptionID, nil)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestIntegration_GetSubscription_NotFound(t *testing.T) {
	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+id, nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestIntegration_Create_Validation(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", strings.NewReader(`{"service_name":"","price":0}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestIntegration_GetCost_MissingParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/cost", nil)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
