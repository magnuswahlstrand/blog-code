package main

import (
	"github.com/gofiber/fiber/v2"
	errorhandler "github.com/magnuswahlstrand/blog-code/idempotency/error_handler"
	"net/http"
)

type Order struct {
	ProductType string `json:"product_type"`
}

func (s *Service) HandlerOrder(c *fiber.Ctx) error {
	var order Order
	if err := c.BodyParser(&order); err != nil {
		return errorhandler.Handle(c, http.StatusInternalServerError)
	}

	// Do work here
	s.Process()

	// Store result
	return c.Status(http.StatusCreated).JSON(order)
}
