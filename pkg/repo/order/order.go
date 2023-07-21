package order

type Item struct {
	Id    string  `json:"id"`
	Name  string  `json:"name"`
	Price float32 `json:"price"`
}

type Order struct {
	Id        string  `json:"id"`
	Status    string  `json:"status"`
	PaymentId *string `json:"payment_id,omitempty"`
	UserId    string  `json:"user_id"`
	Total     float32 `json:"total"`
	Items     []Item  `json:"items"`
}
