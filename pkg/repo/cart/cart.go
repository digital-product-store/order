package cart

type CartItem struct {
	Id    string  `json:"id"`
	Name  string  `json:"name"`
	Price float32 `json:"price"`
}

type Cart struct {
	Items []CartItem `json:"items"`
}
