package service

import (
	"context"

	"em_tz_anvar/internal/models"
	"em_tz_anvar/internal/repository"

	"github.com/google/uuid"
)

type SubscriptionService interface {
	Create(ctx context.Context, req *models.CreateSubscriptionReq) (*models.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error)
	GetAll(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error)
	Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionReq) (*models.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetTotalCost(ctx context.Context, filter *models.CostFilter) (*models.TotalCostResponse, error)
}

type Service struct {
	Subscription SubscriptionService
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Subscription: NewSubscriptionService(repos.Subscription),
	}
}
