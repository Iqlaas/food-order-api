package models

import "time"

type MenuItem struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	PriceCents   int       `json:"price_cents"`
	Availability string    `json:"availability"`
	CreatedAt    time.Time `json:"created_at"`
}

type UpdateAvailabilityPayload struct {
	Availability string `json:"availability"`
}

type OrderItemPayload struct {
	MenuItemID int `json:"menu_item_id"`
	Quantity   int `json:"quantity"`
}

type PlaceOrderPayload struct {
	Items []OrderItemPayload `json:"items"`
}

type OrderItemResponse struct {
	MenuItemID int    `json:"menu_item_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PriceCents int    `json:"price_cents"`
}

type OrderResponse struct {
	ID         int                 `json:"id"`
	TotalCents int                 `json:"total_cents"`
	Status     string              `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
	Items      []OrderItemResponse `json:"items"`
}

type UpdateStatusPayload struct {
	Status string `json:"status"`
}

type OrderPlacedEvent struct {
	OrderID    int       `json:"order_id"`
	TotalCents int       `json:"total_cents"`
	Timestamp  time.Time `json:"timestamp"`
}