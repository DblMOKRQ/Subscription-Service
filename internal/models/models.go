package models

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ID          uuid.UUID `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
}

type SubReq struct {
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      uuid.UUID `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
}

type GetSummaryReq struct {
	ServiceName string     `json:"service_name,omitempty"`
	From        time.Time  `json:"from,omitempty"`
	To          time.Time  `json:"to,omitempty"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
}

type GetSummary struct {
	From        string     `json:"from"`
	To          string     `json:"to"`
	UserID      *uuid.UUID `json:"user_id"`
	ServiceName string     `json:"service_name"`
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID `json:"user_id"`
	ServiceName *string    `json:"service_name"`
}

type Response struct {
	Status int         `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
}
