package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"em_tz_anvar/internal/models"
	"em_tz_anvar/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSubscriptionRepo struct {
	createFn          func(ctx context.Context, sub *models.Subscription) error
	getByIDFn         func(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	getAllFn          func(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error)
	updateFn          func(ctx context.Context, sub *models.Subscription) error
	updateAtomicallyFn func(ctx context.Context, id uuid.UUID, updateFn func(*models.Subscription) error) (*models.Subscription, error)
	deleteFn          func(ctx context.Context, id uuid.UUID) error
	getTotalCostFn    func(ctx context.Context, filter *models.CostFilter) (int, error)
}

func (m *mockSubscriptionRepo) Create(ctx context.Context, sub *models.Subscription) error {
	if m.createFn != nil {
		return m.createFn(ctx, sub)
	}
	return nil
}

func (m *mockSubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSubscriptionRepo) GetAll(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error) {
	if m.getAllFn != nil {
		return m.getAllFn(ctx, filter)
	}
	return nil, nil
}

func (m *mockSubscriptionRepo) Update(ctx context.Context, sub *models.Subscription) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, sub)
	}
	return nil
}

func (m *mockSubscriptionRepo) UpdateAtomically(ctx context.Context, id uuid.UUID, updateFn func(*models.Subscription) error) (*models.Subscription, error) {
	if m.updateAtomicallyFn != nil {
		return m.updateAtomicallyFn(ctx, id, updateFn)
	}
	return nil, nil
}

func (m *mockSubscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockSubscriptionRepo) GetTotalCost(ctx context.Context, filter *models.CostFilter) (int, error) {
	if m.getTotalCostFn != nil {
		return m.getTotalCostFn(ctx, filter)
	}
	return 0, nil
}

func TestSubscriptionService_Create(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	var capturedSub *models.Subscription
	repo := &mockSubscriptionRepo{
		createFn: func(ctx context.Context, sub *models.Subscription) error {
			capturedSub = sub
			return nil
		},
	}
	svc := NewSubscriptionService(repo)

	req := &models.CreateSubscriptionReq{
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      userID,
		StartDate:   "01-2025",
		EndDate:     "",
	}

	sub, err := svc.Create(ctx, req)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, sub.ID)
	assert.Equal(t, "Yandex Plus", sub.ServiceName)
	assert.Equal(t, 400, sub.Price)
	assert.Equal(t, userID, capturedSub.UserID.String())
	assert.Equal(t, "01-2025", capturedSub.StartDate.Format("01-2006"))
}

func TestSubscriptionService_Create_InvalidUserID(t *testing.T) {
	ctx := context.Background()
	repo := &mockSubscriptionRepo{}
	svc := NewSubscriptionService(repo)

	req := &models.CreateSubscriptionReq{
		ServiceName: "Yandex",
		Price:       100,
		UserID:      "not-a-uuid",
		StartDate:   "01-2025",
	}

	sub, err := svc.Create(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "user_id")
}

func TestSubscriptionService_Create_InvalidStartDate(t *testing.T) {
	ctx := context.Background()
	repo := &mockSubscriptionRepo{}
	svc := NewSubscriptionService(repo)

	req := &models.CreateSubscriptionReq{
		ServiceName: "Yandex",
		Price:       100,
		UserID:      uuid.New().String(),
		StartDate:   "2025-01",
	}

	sub, err := svc.Create(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, sub)
	assert.Contains(t, err.Error(), "start_date")
}

func TestSubscriptionService_GetByID(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	expected := &models.Subscription{
		ID:          id,
		ServiceName: "Netflix",
		Price:       500,
		UserID:      id,
		StartDate:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo := &mockSubscriptionRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
			return expected, nil
		},
	}
	svc := NewSubscriptionService(repo)

	sub, err := svc.GetByID(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, expected, sub)
}

func TestSubscriptionService_GetByID_NotFound(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	repo := &mockSubscriptionRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
			return nil, repository.ErrNotFound
		},
	}
	svc := NewSubscriptionService(repo)

	sub, err := svc.GetByID(ctx, id)
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, sub)
}

func TestSubscriptionService_GetAll(t *testing.T) {
	ctx := context.Background()
	list := []models.Subscription{
		{ID: uuid.New(), ServiceName: "A", Price: 100, UserID: uuid.New(), StartDate: time.Now(), UpdatedAt: time.Now()},
	}
	repo := &mockSubscriptionRepo{
		getAllFn: func(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error) {
			return list, nil
		},
	}
	svc := NewSubscriptionService(repo)

	result, err := svc.GetAll(ctx, &models.SubscriptionFilter{Limit: 10})
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "A", result[0].ServiceName)
}

func TestSubscriptionService_Update(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	repo := &mockSubscriptionRepo{
		updateAtomicallyFn: func(ctx context.Context, id uuid.UUID, fn func(*models.Subscription) error) (*models.Subscription, error) {
			sub := &models.Subscription{ID: id, ServiceName: "Old", Price: 100, UserID: id, StartDate: time.Now(), UpdatedAt: time.Now()}
			if err := fn(sub); err != nil {
				return nil, err
			}
			return sub, nil
		},
	}
	svc := NewSubscriptionService(repo)

	req := &models.UpdateSubscriptionReq{
		ServiceName: "Updated",
		Price:       600,
		StartDate:   "02-2025",
	}
	sub, err := svc.Update(ctx, id, req)
	require.NoError(t, err)
	assert.Equal(t, "Updated", sub.ServiceName)
	assert.Equal(t, 600, sub.Price)
}

func TestSubscriptionService_Update_InvalidDate(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	repo := &mockSubscriptionRepo{
		updateAtomicallyFn: func(ctx context.Context, id uuid.UUID, fn func(*models.Subscription) error) (*models.Subscription, error) {
			sub := &models.Subscription{ID: id, UpdatedAt: time.Now()}
			return nil, fn(sub)
		},
	}
	svc := NewSubscriptionService(repo)

	req := &models.UpdateSubscriptionReq{StartDate: "invalid"}
	sub, err := svc.Update(ctx, id, req)
	assert.Error(t, err)
	assert.Nil(t, sub)
}

func TestSubscriptionService_Update_NotFound(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	repo := &mockSubscriptionRepo{
		updateAtomicallyFn: func(ctx context.Context, id uuid.UUID, fn func(*models.Subscription) error) (*models.Subscription, error) {
			return nil, repository.ErrNotFound
		},
	}
	svc := NewSubscriptionService(repo)

	sub, err := svc.Update(ctx, id, &models.UpdateSubscriptionReq{Price: 100})
	assert.ErrorIs(t, err, repository.ErrNotFound)
	assert.Nil(t, sub)
}

func TestSubscriptionService_Delete(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	repo := &mockSubscriptionRepo{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return nil
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.Delete(ctx, id)
	require.NoError(t, err)
}

func TestSubscriptionService_Delete_NotFound(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	repo := &mockSubscriptionRepo{
		deleteFn: func(ctx context.Context, id uuid.UUID) error {
			return repository.ErrNotFound
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.Delete(ctx, id)
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestSubscriptionService_GetTotalCost(t *testing.T) {
	ctx := context.Background()
	repo := &mockSubscriptionRepo{
		getTotalCostFn: func(ctx context.Context, filter *models.CostFilter) (int, error) {
			return 1200, nil
		},
	}
	svc := NewSubscriptionService(repo)

	filter := &models.CostFilter{
		StartDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	resp, err := svc.GetTotalCost(ctx, filter)
	require.NoError(t, err)
	assert.Equal(t, 1200, resp.TotalCost)
	assert.Equal(t, "RUB", resp.Currency)
}

func TestSubscriptionService_GetTotalCost_RepoError(t *testing.T) {
	ctx := context.Background()
	repoErr := errors.New("db error")
	repo := &mockSubscriptionRepo{
		getTotalCostFn: func(ctx context.Context, filter *models.CostFilter) (int, error) {
			return 0, repoErr
		},
	}
	svc := NewSubscriptionService(repo)

	resp, err := svc.GetTotalCost(ctx, &models.CostFilter{})
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, resp)
}
