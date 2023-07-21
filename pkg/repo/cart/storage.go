package cart

import "context"

type CartStorage interface {
	Get(ctx context.Context, key string) (*Cart, error)
	Set(ctx context.Context, key string, cart *Cart) error
	Delete(ctx context.Context, key string) error
}
