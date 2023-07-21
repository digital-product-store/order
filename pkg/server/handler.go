package server

import (
	"net/http"
	"orderservice/pkg/client/exchange"
	"orderservice/pkg/client/payment"
	"orderservice/pkg/client/product"
	"orderservice/pkg/repo/cart"
	"orderservice/pkg/repo/order"
	"orderservice/pkg/server/gen"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.elastic.co/apm/v2"
	"go.uber.org/zap"
)

type Handler struct {
	logger         *zap.Logger
	cartStorage    cart.CartStorage
	orderStorage   order.OrderStorage
	productClient  *product.ProductClient
	exchangeClient *exchange.ExchangeClient
	paymentClient  *payment.PaymentClient
}

func NewHandler(logger *zap.Logger, cartStorage cart.CartStorage, orderStorage order.OrderStorage, productClient *product.ProductClient, exchangeClient *exchange.ExchangeClient, paymentClient *payment.PaymentClient) Handler {
	return Handler{
		logger:         logger,
		cartStorage:    cartStorage,
		orderStorage:   orderStorage,
		productClient:  productClient,
		exchangeClient: exchangeClient,
		paymentClient:  paymentClient,
	}
}

func (h Handler) Health(ctx echo.Context) error {
	return ctx.NoContent(http.StatusOK)
}

func (h Handler) ClearCart(ctx echo.Context, params gen.ClearCartParams) error {
	span, apmCtx := apm.StartSpan(ctx.Request().Context(), "ClearCart", "request")
	defer span.End()

	err := h.cartStorage.Delete(apmCtx, params.XUserId)
	if err != nil {
		h.logger.Error("error on deleting key", zap.String("key", params.XUserId), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h Handler) GetCart(ctx echo.Context, params gen.GetCartParams) error {
	span, apmCtx := apm.StartSpan(ctx.Request().Context(), "GetCart", "request")
	defer span.End()

	cart, err := h.cartStorage.Get(apmCtx, params.XUserId)
	if err != nil {
		h.logger.Error("error on getting key", zap.String("key", params.XUserId), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if cart == nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	items := []gen.CartItem{}
	for _, item := range cart.Items {
		cartItem := gen.CartItem{
			Id:    item.Id,
			Name:  item.Name,
			Price: item.Price,
		}

		items = append(items, cartItem)
	}

	return ctx.JSON(http.StatusOK, items)
}

func (h Handler) UpdateCart(ctx echo.Context, params gen.UpdateCartParams) error {
	span, apmCtx := apm.StartSpan(ctx.Request().Context(), "UpdateCart", "request")
	defer span.End()

	ids := &[]string{}
	if err := ctx.Bind(ids); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	items := []cart.CartItem{}
	for _, id := range *ids {
		product, err := h.productClient.GetByUUID(apmCtx, id)
		if err != nil {
			h.logger.Error("error getting product detail", zap.String("user_id", params.XUserId), zap.String("product_id", id), zap.Error(err))
			return ctx.NoContent(http.StatusInternalServerError)
		}

		price, err := strconv.ParseFloat(product.Price, 32)
		if err != nil {
			h.logger.Error("error casting product price to float32", zap.String("user_id", params.XUserId), zap.String("product_id", id), zap.String("price", product.Price), zap.Error(err))
			return ctx.NoContent(http.StatusInternalServerError)
		}
		item := cart.CartItem{
			Id:    id,
			Name:  product.BookName,
			Price: float32(price),
		}
		items = append(items, item)
	}

	cart := cart.Cart{
		Items: items,
	}

	err := h.cartStorage.Set(apmCtx, params.XUserId, &cart)
	if err != nil {
		h.logger.Error("error on setting key", zap.String("key", params.XUserId), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	return ctx.NoContent(http.StatusOK)
}

func (h Handler) CheckoutCart(ctx echo.Context, params gen.CheckoutCartParams) error {
	span, apmCtx := apm.StartSpan(ctx.Request().Context(), "CheckoutCart", "request")
	defer span.End()

	cardInfo := new(gen.CardInfo)
	if err := ctx.Bind(cardInfo); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	// get cart
	cart, err := h.cartStorage.Get(apmCtx, params.XUserId)
	if err != nil {
		h.logger.Error("error on getting key", zap.String("key", params.XUserId), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if cart == nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	items := []order.Item{}
	var total float32 = 0
	for _, item := range cart.Items {
		orderItem := order.Item{
			Id:    item.Id,
			Name:  item.Name,
			Price: item.Price,
		}

		items = append(items, orderItem)
		total += item.Price
	}

	// create order with status ready
	order, err := h.orderStorage.Create(apmCtx, params.XUserId, total, items)
	if err != nil {
		h.logger.Error("error during creating new order", zap.String("user_id", params.XUserId), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// exchange rate
	exchangeResult, err := h.exchangeClient.GetTotal(apmCtx, "EUR", "USD", total)
	if err != nil {
		h.logger.Error("error exchange result", zap.String("user_id", params.XUserId), zap.Float32("total", total), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// payment
	paymentRequest := payment.PaymentRequest{
		Amount:     float64(exchangeResult.Total),
		Currency:   "USD",
		CardNumber: cardInfo.Number,
		ExpDate:    cardInfo.ExpDate,
		CVV:        cardInfo.Cvv,
	}
	paymentResult, err := h.paymentClient.MakePayment(apmCtx, paymentRequest)
	if err != nil {
		h.logger.Error("error on payment result", zap.String("user_id", params.XUserId), zap.String("cvv", paymentRequest.CVV), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// update order with status complete
	if err := h.orderStorage.Complete(apmCtx, order.Id, paymentResult.Id); err != nil {
		h.logger.Error("error during completing the order", zap.String("user_id", params.XUserId), zap.String("order_id", order.Id), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// get updated order
	order, err = h.orderStorage.Get(apmCtx, order.Id)
	if err != nil {
		h.logger.Error("error getting created order", zap.String("user_id", params.XUserId), zap.String("order_id", order.Id), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// clear cart
	err = h.cartStorage.Delete(apmCtx, params.XUserId)
	if err != nil {
		h.logger.Error("error on clearing user's cart", zap.String("user_id", params.XUserId), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	// representation
	output := gen.Order{
		Id:        order.Id,
		PaymentId: order.PaymentId,
		Status:    order.Status,
		Total:     order.Total,
		Items:     []gen.CartItem{},
	}
	for _, item := range order.Items {
		genItem := gen.CartItem{
			Id:    item.Id,
			Name:  item.Name,
			Price: item.Price,
		}
		output.Items = append(output.Items, genItem)
	}

	return ctx.JSON(http.StatusOK, output)
}

func (h Handler) ListOrders(ctx echo.Context, params gen.ListOrdersParams) error {
	span, apmCtx := apm.StartSpan(ctx.Request().Context(), "ListOrders", "request")
	defer span.End()

	orders, err := h.orderStorage.List(apmCtx, params.XUserId)
	if err != nil {
		h.logger.Error("error on listing orders for user", zap.String("key", params.XUserId), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	outputOrders := []gen.Order{}
	for _, order := range orders {
		outputOrder := gen.Order{
			Id:        order.Id,
			PaymentId: order.PaymentId,
			Status:    order.Status,
			Total:     order.Total,
			Items:     []gen.CartItem{},
		}

		for _, item := range order.Items {
			genItem := gen.CartItem{
				Id:    item.Id,
				Name:  item.Name,
				Price: item.Price,
			}
			outputOrder.Items = append(outputOrder.Items, genItem)
		}

		outputOrders = append(outputOrders, outputOrder)
	}
	return ctx.JSON(http.StatusOK, outputOrders)
}

func (h Handler) GetOrderDetail(ctx echo.Context, uuid string) error {
	span, apmCtx := apm.StartSpan(ctx.Request().Context(), "GetOrderDetail", "request")
	defer span.End()

	order, err := h.orderStorage.Get(apmCtx, uuid)
	if err != nil {
		h.logger.Error("error on getting order by id", zap.String("order_id", uuid), zap.Error(err))
		return ctx.NoContent(http.StatusInternalServerError)
	}

	if order == nil {
		return ctx.NoContent(http.StatusNotFound)
	}

	output := gen.Order{
		Id:        order.Id,
		PaymentId: order.PaymentId,
		Status:    order.Status,
		Total:     order.Total,
		Items:     []gen.CartItem{},
	}
	for _, item := range order.Items {
		genItem := gen.CartItem{
			Id:    item.Id,
			Name:  item.Name,
			Price: item.Price,
		}
		output.Items = append(output.Items, genItem)
	}

	return ctx.JSON(http.StatusOK, output)
}
