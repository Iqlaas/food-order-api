package repository

import (
	"context"
	"fmt"
	"food-order-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetMenu(ctx context.Context) ([]models.MenuItem, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, category, price_cents, availability, created_at FROM menu_items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.MenuItem
	for rows.Next() {
		var item models.MenuItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.PriceCents, &item.Availability, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) GetMenuItem(ctx context.Context, id int) (*models.MenuItem, error) {
	var item models.MenuItem
	err := r.db.QueryRow(ctx, "SELECT id, name, category, price_cents, availability, created_at FROM menu_items WHERE id = $1", id).
		Scan(&item.ID, &item.Name, &item.Category, &item.PriceCents, &item.Availability, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) UpdateItemAvailability(ctx context.Context, id int, availability string) error {
	res, err := r.db.Exec(ctx, "UPDATE menu_items SET availability = $1 WHERE id = $2", availability, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}

func (r *Repository) CreateOrder(ctx context.Context, payload models.PlaceOrderPayload) (*models.OrderResponse, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var totalCents int
	type verifiedItem struct {
		id       int
		name     string
		price    int
		quantity int
	}
	var verifiedItems []verifiedItem

	for _, reqItem := range payload.Items {
		var item models.MenuItem
		err := tx.QueryRow(ctx, "SELECT id, name, price_cents, availability FROM menu_items WHERE id = $1", reqItem.MenuItemID).
			Scan(&item.ID, &item.Name, &item.PriceCents, &item.Availability)
		if err != nil {
			return nil, fmt.Errorf("item %d does not exist", reqItem.MenuItemID)
		}
		if item.Availability != "in_stock" {
			return nil, fmt.Errorf("item %s is out of stock", item.Name)
		}

		totalCents += item.PriceCents * reqItem.Quantity
		verifiedItems = append(verifiedItems, verifiedItem{
			id:       item.ID,
			name:     item.Name,
			price:    item.PriceCents,
			quantity: reqItem.Quantity,
		})
	}

	var orderID int
	err = tx.QueryRow(ctx, "INSERT INTO orders (total_cents, status) VALUES ($1, 'received') RETURNING id", totalCents).Scan(&orderID)
	if err != nil {
		return nil, err
	}

	var itemResponses []models.OrderItemResponse
	for _, vi := range verifiedItems {
		_, err = tx.Exec(ctx, "INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_purchase_cents) VALUES ($1, $2, $3, $4)",
			orderID, vi.id, vi.quantity, vi.price)
		if err != nil {
			return nil, err
		}
		itemResponses = append(itemResponses, models.OrderItemResponse{
			MenuItemID: vi.id,
			Name:       vi.name,
			Quantity:   vi.quantity,
			PriceCents: vi.price,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &models.OrderResponse{
		ID:         orderID,
		TotalCents: totalCents,
		Status:     "received",
		Items:      itemResponses,
	}, nil
}

func (r *Repository) GetOrder(ctx context.Context, id int) (*models.OrderResponse, error) {
	var order models.OrderResponse
	err := r.db.QueryRow(ctx, "SELECT id, total_cents, status, created_at FROM orders WHERE id = $1", id).
		Scan(&order.ID, &order.TotalCents, &order.Status, &order.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT oi.menu_item_id, m.name, oi.quantity, oi.price_at_purchase_cents 
		FROM order_items oi
		JOIN menu_items m ON oi.menu_item_id = m.id
		WHERE oi.order_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItemResponse
		if err := rows.Scan(&item.MenuItemID, &item.Name, &item.Quantity, &item.PriceCents); err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, id int, status string) error {
	res, err := r.db.Exec(ctx, "UPDATE orders SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}