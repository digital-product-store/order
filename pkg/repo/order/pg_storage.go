package order

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.elastic.co/apm/v2"
)

type PGOrderStorage struct {
	db *sql.DB
}

func NewPGOrderStorage(url string) (*PGOrderStorage, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	s := &PGOrderStorage{
		db: db,
	}
	return s, nil
}

func (s PGOrderStorage) Create(ctx context.Context, userId string, total float32, items []Item) (*Order, error) {
	span, ctx := apm.StartSpan(ctx, "Create", "PGOrderStorage")
	defer span.End()

	itemsJson, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}

	id := uuid.NewString()
	_, err = s.db.Exec("INSERT INTO orders (id, status, user_id, total, items) VALUES ($1, $2, $3, $4, $5)", id, "ready", userId, total, itemsJson)
	if err != nil {
		return nil, err
	}

	order := new(Order)
	order.Id = id
	order.Status = "ready"
	order.UserId = userId
	order.Total = total
	order.Items = items

	return order, nil
}

func (s PGOrderStorage) Complete(ctx context.Context, orderId string, paymentId string) error {
	span, ctx := apm.StartSpan(ctx, "Complete", "PGOrderStorage")
	defer span.End()

	_, err := s.db.Exec("UPDATE orders SET status=$1, payment_id=$2 WHERE id=$3", "completed", paymentId, orderId)
	return err
}

func (s PGOrderStorage) List(ctx context.Context, userId string) ([]Order, error) {
	span, ctx := apm.StartSpan(ctx, "List", "PGOrderStorage")
	defer span.End()

	query := "SELECT id, status, payment_id, user_id, total, items FROM orders WHERE user_id = $1"
	rows, err := s.db.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var id, status, paymentID, userID string
		var total float64
		var itemsJSON []byte

		err := rows.Scan(&id, &status, &paymentID, &userID, &total, &itemsJSON)
		if err != nil {
			return nil, err
		}

		var items []Item
		err = json.Unmarshal(itemsJSON, &items)
		if err != nil {
			return nil, err
		}

		order := Order{
			Id:        id,
			Status:    status,
			PaymentId: &paymentID,
			UserId:    userID,
			Total:     float32(total),
			Items:     items,
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s PGOrderStorage) Get(ctx context.Context, orderId string) (*Order, error) {
	span, ctx := apm.StartSpan(ctx, "Get", "PGOrderStorage")
	defer span.End()

	query := "SELECT id, status, payment_id, user_id, total, items FROM orders WHERE id = $1"
	row := s.db.QueryRow(query, orderId)

	var id, status, paymentID, userID string
	var total float64
	var itemsJSON []byte

	err := row.Scan(&id, &status, &paymentID, &userID, &total, &itemsJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return nil and nil error if no order found
			return nil, nil
		}
		return nil, err
	}

	var items []Item
	err = json.Unmarshal(itemsJSON, &items)
	if err != nil {
		return nil, err
	}

	order := &Order{
		Id:        id,
		Status:    status,
		PaymentId: &paymentID,
		UserId:    userID,
		Total:     float32(total),
		Items:     items,
	}

	return order, nil
}
