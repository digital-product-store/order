package payment

type PaymentRequest struct {
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	CardNumber string  `json:"card_number"`
	ExpDate    string  `json:"exp_date"`
	CVV        string  `json:"cvv"`
}

type PaymentResponse struct {
	Id string `json:"id"`
}
