package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"em_tz_anvar/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	db := sqlx.NewDb(sqlDB, "postgres")
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestSubscriptionRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	sub := &models.Subscription{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		ServiceName: "Test",
		Price:       100,
		UserID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec("INSERT INTO subscriptions").
		WithArgs(sub.ID, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate, sub.CreatedAt, sub.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, sub)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_GetByID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	rows := sqlmock.NewRows([]string{"id", "service_name", "price", "user_id", "start_date", "end_date", "created_at", "updated_at"}).
		AddRow(id, "Yandex", 300, uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), nil, time.Now(), time.Now())

	mock.ExpectQuery("SELECT .+ FROM subscriptions WHERE id").
		WithArgs(id).
		WillReturnRows(rows)

	sub, err := repo.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, sub.ID)
	assert.Equal(t, "Yandex", sub.ServiceName)
	assert.Equal(t, 300, sub.Price)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_GetByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	mock.ExpectQuery("SELECT .+ FROM subscriptions WHERE id").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	sub, err := repo.GetByID(ctx, id)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, sub)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_GetAll(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	rows := sqlmock.NewRows([]string{"id", "service_name", "price", "user_id", "start_date", "end_date", "created_at", "updated_at"}).
		AddRow(id, "Yandex", 300, id, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), nil, time.Now(), time.Now())

	mock.ExpectQuery("SELECT .+ FROM subscriptions").
		WillReturnRows(rows)

	filter := &models.SubscriptionFilter{Limit: 10}
	list, err := repo.GetAll(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Yandex", list[0].ServiceName)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_Update(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	sub := &models.Subscription{
		ID:          id,
		ServiceName: "Updated",
		Price:       500,
		UserID:      id,
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:     nil,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectExec("UPDATE subscriptions").
		WithArgs(sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.UpdatedAt, sub.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, sub)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_Update_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	sub := &models.Subscription{ID: id, ServiceName: "X", Price: 1, UserID: id, StartDate: time.Now(), UpdatedAt: time.Now()}

	mock.ExpectExec("UPDATE subscriptions").
		WithArgs(sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.UpdatedAt, sub.ID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(ctx, sub)
	assert.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_UpdateAtomically(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "service_name", "price", "user_id", "start_date", "end_date", "created_at", "updated_at"}).
		AddRow(id, "Old", 100, id, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), nil, time.Now(), time.Now())
	mock.ExpectQuery("SELECT .+ FOR UPDATE").
		WithArgs(id).
		WillReturnRows(rows)
	mock.ExpectExec("UPDATE subscriptions").
		WithArgs("New", 200, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), nil, sqlmock.AnyArg(), id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	updated, err := repo.UpdateAtomically(ctx, id, func(s *models.Subscription) error {
		s.ServiceName = "New"
		s.Price = 200
		s.UpdatedAt = time.Now()
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.ServiceName)
	assert.Equal(t, 200, updated.Price)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_UpdateAtomically_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .+ FOR UPDATE").
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	updated, err := repo.UpdateAtomically(ctx, id, func(s *models.Subscription) error { return nil })
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Nil(t, updated)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	mock.ExpectExec("DELETE FROM subscriptions WHERE id").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctx, id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_Delete_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	mock.ExpectExec("DELETE FROM subscriptions WHERE id").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(ctx, id)
	assert.ErrorIs(t, err, ErrNotFound)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_GetTotalCost(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewSubscriptionRepository(db)
	ctx := context.Background()
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"total_cost"}).AddRow(3600)
	mock.ExpectQuery("SELECT COALESCE").
		WillReturnRows(rows)

	filter := &models.CostFilter{StartDate: start, EndDate: end}
	cost, err := repo.GetTotalCost(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 3600, cost)
	require.NoError(t, mock.ExpectationsWereMet())
}
