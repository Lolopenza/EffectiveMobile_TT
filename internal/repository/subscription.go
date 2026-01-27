package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"em_tz_anvar/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

var ErrNotFound = errors.New("subscription not found")

type subscriptionRepository struct {
	db *sqlx.DB
}

func NewSubscriptionRepository(db *sqlx.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// Create
func (r *subscriptionRepository) Create(ctx context.Context, subscription *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	log.Debug().
		Str("subscription_id", subscription.ID.String()).
		Str("service_name", subscription.ServiceName).
		Msg("Creating subscription")

	_, err := r.db.ExecContext(ctx, query,
		subscription.ID,
		subscription.ServiceName,
		subscription.Price,
		subscription.UserID,
		subscription.StartDate,
		subscription.EndDate,
		subscription.CreatedAt,
		subscription.UpdatedAt,
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create subscription")
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	return nil
}

// GetByID
func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`

	log.Debug().Str("subscription_id", id.String()).Msg("Getting subscription by ID")

	var subscription models.Subscription
	err := r.db.GetContext(ctx, &subscription, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		log.Error().Err(err).Str("subscription_id", id.String()).Msg("Failed to get subscription")
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return &subscription, nil
}

// GetAll with filters
func (r *subscriptionRepository) GetAll(ctx context.Context, filter *models.SubscriptionFilter) ([]models.Subscription, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
	`

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, *filter.UserID)
		argNum++
	}

	if filter.ServiceName != "" {
		conditions = append(conditions, fmt.Sprintf("service_name ILIKE $%d", argNum))
		args = append(args, "%"+filter.ServiceName+"%")
		argNum++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	log.Debug().
		Interface("filter", filter).
		Str("query", query).
		Msg("Getting all subscriptions")

	var subscriptions []models.Subscription
	err := r.db.SelectContext(ctx, &subscriptions, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subscriptions")
		return nil, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	return subscriptions, nil
}

// Update
func (r *subscriptionRepository) Update(ctx context.Context, subscription *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = $5
		WHERE id = $6
	`

	log.Debug().
		Str("subscription_id", subscription.ID.String()).
		Msg("Updating subscription")

	result, err := r.db.ExecContext(ctx, query,
		subscription.ServiceName,
		subscription.Price,
		subscription.StartDate,
		subscription.EndDate,
		subscription.UpdatedAt,
		subscription.ID,
	)

	if err != nil {
		log.Error().Err(err).Str("subscription_id", subscription.ID.String()).Msg("Failed to update subscription")
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// Delete
func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	log.Debug().Str("subscription_id", id.String()).Msg("Deleting subscription")

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error().Err(err).Str("subscription_id", id.String()).Msg("Failed to delete subscription")
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

// GetTotalCost
func (r *subscriptionRepository) GetTotalCost(ctx context.Context, filter *models.CostFilter) (int, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	// Base query
	query := `
		SELECT COALESCE(SUM(price * 
			(EXTRACT(YEAR FROM LEAST(COALESCE(end_date, $1::timestamp), $1::timestamp)) * 12 + 
			 EXTRACT(MONTH FROM LEAST(COALESCE(end_date, $1::timestamp), $1::timestamp)) -
			 EXTRACT(YEAR FROM GREATEST(start_date, $2::timestamp)) * 12 - 
			 EXTRACT(MONTH FROM GREATEST(start_date, $2::timestamp)) + 1)
		), 0)::integer as total_cost
		FROM subscriptions
		WHERE start_date <= $1 AND (end_date IS NULL OR end_date >= $2)
	`

	args = append(args, filter.EndDate, filter.StartDate)
	argNum = 3

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argNum))
		args = append(args, *filter.UserID)
		argNum++
	}

	if filter.ServiceName != "" {
		conditions = append(conditions, fmt.Sprintf("service_name ILIKE $%d", argNum))
		args = append(args, "%"+filter.ServiceName+"%")
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	log.Debug().
		Interface("filter", filter).
		Msg("Calculating total cost")

	var totalCost int
	err := r.db.GetContext(ctx, &totalCost, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to calculate total cost")
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	return totalCost, nil
}
