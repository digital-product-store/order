package order

import "context"

type OrderStorage interface {
	Create(ctx context.Context, userId string, total float32, items []Item) (*Order, error)
	Complete(ctx context.Context, orderId string, paymentId string) error
	List(ctx context.Context, userId string) ([]Order, error)
	Get(ctx context.Context, orderId string) (*Order, error)
}
