package product

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.elastic.co/apm/v2"
)

type ProductClient struct {
	baseUrl string
}

func NewProductClient(baseUrl string) *ProductClient {
	return &ProductClient{
		baseUrl: baseUrl,
	}
}

func (p ProductClient) GetByUUID(ctx context.Context, uuid string) (*Product, error) {
	span, _ := apm.StartSpan(ctx, "GetByUUID", "ProductClient")
	defer span.End()

	url := fmt.Sprintf("%s/_private/api/v1/books/%s", p.baseUrl, uuid)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	var product Product
	err = json.NewDecoder(resp.Body).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}
