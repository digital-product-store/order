package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.elastic.co/apm/v2"
)

type PaymentClient struct {
	baseUrl string
}

func NewPaymentClient(baseUrl string) *PaymentClient {
	return &PaymentClient{
		baseUrl: baseUrl,
	}
}

func (p PaymentClient) MakePayment(ctx context.Context, paymentReq PaymentRequest) (*PaymentResponse, error) {
	span, _ := apm.StartSpan(ctx, "MakePayment", "PaymentClient")
	defer span.End()

	url := fmt.Sprintf("%s/_private/api/v1/payment", p.baseUrl)

	payload, err := json.Marshal(paymentReq)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	var paymentRes PaymentResponse
	err = json.NewDecoder(resp.Body).Decode(&paymentRes)
	if err != nil {
		return nil, err
	}

	return &paymentRes, nil
}
