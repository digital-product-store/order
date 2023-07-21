package exchange

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.elastic.co/apm/v2"
)

type ExchangeClient struct {
	baseUrl string
}

func NewExchangeClient(baseUrl string) *ExchangeClient {
	return &ExchangeClient{
		baseUrl: baseUrl,
	}
}

func (p ExchangeClient) GetTotal(ctx context.Context, from, to string, amount float32) (*ExchangeResult, error) {
	span, _ := apm.StartSpan(ctx, "GetTotal", "ExchangeClient")
	defer span.End()

	url := fmt.Sprintf("%s/_private/api/v1/%s/%s/%.2f", p.baseUrl, from, to, amount)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	var exchangeResult ExchangeResult
	err = json.NewDecoder(resp.Body).Decode(&exchangeResult)
	if err != nil {
		return nil, err
	}

	return &exchangeResult, nil
}
