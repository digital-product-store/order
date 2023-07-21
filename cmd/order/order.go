//go:generate oapi-codegen -package gen -generate types,server,spec -o ../../pkg/server/gen/specs.gen.go ../../specs/openapi3.yaml

package main

import (
	"orderservice/pkg/client/exchange"
	"orderservice/pkg/client/payment"
	"orderservice/pkg/client/product"
	"orderservice/pkg/config"
	"orderservice/pkg/repo/cart"
	"orderservice/pkg/repo/order"
	"orderservice/pkg/server"
	"os"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	conf, err := config.LoadConfig()
	if err != nil {
		logger.Error("error on during configuration", zap.Error(err))
		os.Exit(-1)
	}

	redisCartStorage, err := cart.NewRedisCartStorage(conf.RedisUrl)
	if err != nil {
		logger.Error("error on creating redis", zap.Error(err))
		os.Exit(-1)
	}

	pgOrderStorage, err := order.NewPGOrderStorage(conf.PostgresqlUrl)
	if err != nil {
		logger.Error("error on creating pg store", zap.Error(err))
		os.Exit(-1)
	}

	productClient := product.NewProductClient(conf.ProductServiceUrl)
	exchangeClient := exchange.NewExchangeClient(conf.ExchangeServiceUrl)
	paymentClient := payment.NewPaymentClient(conf.PaymentServiceUrl)

	handler := server.NewHandler(logger, redisCartStorage, pgOrderStorage, productClient, exchangeClient, paymentClient)
	srvr := server.NewServer(&handler, conf)

	srvr.Listen()
}
