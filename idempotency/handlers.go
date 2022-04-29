package main

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

const HeaderIdempotencyKey = "Idempotency-Key"

type RecordedRequestResponse struct {
	RequestBody []byte
	Response    Response
	Completed   bool
}

type Response struct {
	Status int
	Data   []byte
}

func responseError(c *fiber.Ctx, request int, message ...string) error {
	return c.Status(request).JSON(fiber.NewError(request, message...))
}

func (s *Service) HandlerOrder(c *fiber.Ctx) error {
	var order Order
	if err := c.BodyParser(&order); err != nil {
		return responseError(c, http.StatusInternalServerError)
	}

	// Do work here
	s.Process()

	// Store result
	return c.Status(http.StatusCreated).JSON(order)
}
