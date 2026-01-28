package repository

import (
	"context"
	"em_tz_anvar/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// SubscriptionRepository interface to work with subs
type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	GetAll(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error)
	Update(ctx context.Context, subscription *models.Subscription) error
	// UpdateAtomically выполняет атомарное обновление подписки с SELECT FOR UPDATE
	UpdateAtomically(ctx context.Context, id uuid.UUID, updateFn func(*models.Subscription) error) (*models.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetTotalCost(ctx context.Context, filter *models.CostFilter) (int, error)
}

// All repositories
type Repository struct {
	Subscription SubscriptionRepository
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Subscription: NewSubscriptionRepository(db),
	}
}
