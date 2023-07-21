package server

import (
	"net/http"
	"orderservice/pkg/config"
	"orderservice/pkg/server/gen"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.elastic.co/apm/module/apmechov4/v2"
)

type Server struct {
	handler *Handler
	config  *config.Config
}

func NewServer(handler *Handler, conf *config.Config) *Server {
	return &Server{
		handler: handler,
		config:  conf,
	}
}

func (s *Server) Listen() error {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(apmechov4.Middleware())

	swagger, err := gen.GetSwagger()
	if err != nil {
		e.Logger.Fatal("openapi3 error", err)
		return err
	}

	jsonByte, err := swagger.MarshalJSON()
	if err != nil {
		e.Logger.Fatal("openapi3 serialization error", err)
		return err
	}

	e.GET("/_openapi3.json", func(c echo.Context) error {
		return c.JSONBlob(http.StatusOK, jsonByte)
	})

	gen.RegisterHandlers(e, s.handler)

	if err := e.Start(s.config.ListenAddr); err != nil {
		e.Logger.Fatal(err)
		return err
	}
	return nil
}
