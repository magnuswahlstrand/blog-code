package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/magnuswahlstrand/blog-code/idempotency/middleware"
	"log"
)

type Service struct {
	Process func()
}

func setup() (*fiber.App, *Service) {
	mongo := middleware.NewMongoDB("mongodb://localhost:27017", "idempotency", "keys")
	idempotencyMiddleware := middleware.New(mongo).IdempotencyMiddleware
	service := &Service{
		Process: func() {}, // Can be overridden in tests
	}

	app := fiber.New()
	app.Use(idempotencyMiddleware)
	app.Post("/order", service.HandlerOrder)
	return app, service
}

func main() {
	app, _ := setup()
	log.Fatal(app.Listen(":8080"))
}
