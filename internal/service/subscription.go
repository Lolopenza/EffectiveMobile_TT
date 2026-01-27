package service

import (
	"context"
	"fmt"
	"time"

	"em_tz_anvar/internal/models"
	"em_tz_anvar/internal/repository"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type subscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

// Create
func (s *subscriptionService) Create(ctx context.Context, req *models.CreateSubscriptionReq) (*models.Subscription, error) {
	log.Info().
		Str("service_name", req.ServiceName).
		Str("user_id", req.UserID).
		Msg("Creating new subscription")

	//User parsing
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	//Parsing
	startDate, err := parseMonthYear(req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}

	//Parsing
	var endDate *time.Time
	if req.EndDate != "" {
		ed, err := parseMonthYear(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		endDate = &ed
	}

	now := time.Now()
	subscription := &models.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, subscription); err != nil {
		return nil, err
	}

	log.Info().
		Str("subscription_id", subscription.ID.String()).
		Msg("Subscription created successfully")

	return subscription, nil
}

// GetByID
func (s *subscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	log.Info().Str("subscription_id", id.String()).Msg("Getting subscription")
	return s.repo.GetByID(ctx, id)
}

// GetAll
func (s *subscriptionService) GetAll(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error) {
	log.Info().Interface("filter", filter).Msg("Getting all subscriptions")
	return s.repo.GetAll(ctx, filter)
}

// Update
func (s *subscriptionService) Update(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionReq) (*models.Subscription, error) {
	log.Info().Str("subscription_id", id.String()).Msg("Updating subscription")

	subscription, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ServiceName != "" {
		subscription.ServiceName = req.ServiceName
	}

	if req.Price > 0 {
		subscription.Price = req.Price
	}

	if req.StartDate != "" {
		startDate, err := parseMonthYear(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		subscription.StartDate = startDate
	}

	if req.EndDate != "" {
		endDate, err := parseMonthYear(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		subscription.EndDate = &endDate
	}

	subscription.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, subscription); err != nil {
		return nil, err
	}

	log.Info().
		Str("subscription_id", subscription.ID.String()).
		Msg("Subscription updated successfully")

	return subscription, nil
}

// Delete
func (s *subscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	log.Info().Str("subscription_id", id.String()).Msg("Deleting subscription")
	return s.repo.Delete(ctx, id)
}

// GetTotalCost
func (s *subscriptionService) GetTotalCost(ctx context.Context, filter *models.CostFilter) (*models.TotalCostResponse, error) {
	log.Info().
		Interface("filter", filter).
		Msg("Calculating total cost")

	totalCost, err := s.repo.GetTotalCost(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &models.TotalCostResponse{
		TotalCost: totalCost,
		Currency:  "RUB",
	}, nil
}

// parseMonthYear
func parseMonthYear(s string) (time.Time, error) {
	return time.Parse("01-2006", s)
}
