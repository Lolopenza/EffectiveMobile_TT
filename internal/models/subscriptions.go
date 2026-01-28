package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	ServiceName string     `json:"service_name" db:"service_name"`
	Price       int        `json:"price" db:"price"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	StartDate   time.Time  `json:"start_date" db:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type CreateSubscriptionReq struct {
	ServiceName string `json:"service_name" binding:"required"`
	Price       int    `json:"price" binding:"required,min=1"`
	UserID      string `json:"user_id" binding:"required,uuid"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date,omitempty"`
}

type UpdateSubscriptionReq struct {
	ServiceName string `json:"service_name,omitempty"`
	Price       int    `json:"price,omitempty" binding:"omitempty,min=1"`
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName string
	Limit       int
	Offset      int
}

type CostFilter struct {
	UserID      *uuid.UUID
	ServiceName string
	StartDate   time.Time
	EndDate     time.Time
}

type TotalCostResponse struct {
	TotalCost int    `json:"total_cost"`
	Currency  string `json:"currency"`
}
