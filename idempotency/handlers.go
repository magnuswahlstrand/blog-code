package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"net/http"
	time "time"
)

const HeaderIdempotencyKey = "Idempotency-Key"

type RecordedRequestResponse struct {
	Param     string
	Response  Response
	Completed bool
}

type Response struct {
	Status int
	Data   map[string]interface{}
}

// TODO: Find a better name
func RespondTo(c *fiber.Ctx, r Response) error {
	return c.Status(r.Status).JSON(r.Data)
}

func responseError(c *fiber.Ctx, request int, message ...string) error {
	return c.Status(request).JSON(fiber.NewError(request, message...))
}

func (s *Service) HandlerOrder(c *fiber.Ctx) error {
	idempotencyKey := c.Get(HeaderIdempotencyKey)
	if idempotencyKey == "" {
		return responseError(c, http.StatusBadRequest, "idempotency key missing")
	}

	if _, err := uuid.Parse(idempotencyKey); err != nil {
		return responseError(c, http.StatusBadRequest, "idempotency must be a UUID V4")
	}

	var order Order
	if err := c.BodyParser(&order); err != nil {
		return responseError(c, http.StatusInternalServerError)
	}

	recorded, exists := s.db2.get(idempotencyKey)
	if exists {
		if order.ProductType != recorded.Param {
			return responseError(c, http.StatusUnprocessableEntity, "idempotency key previously used with other payload")
		}

		if recorded.Completed == false {
			return responseError(c, http.StatusConflict, "request already in process")
		}

		// Re-send previous response
		return RespondTo(c, recorded.Response)
	}

	// Request not in flight, record it
	record := RecordedRequestResponse{
		Param:     order.ProductType,
		Completed: false,
	}
	s.db2.update(idempotencyKey, record)

	// Do something here
	if s.db2.shouldSleep {
		time.Sleep(1 * time.Second)
	}

	// Store result
	record.Completed = true
	record.Response = Response{
		Status: http.StatusCreated,
		Data: map[string]interface{}{
			"hej": "da",
			"bye": "bye",
		},
	}
	s.db2.update(idempotencyKey, record)
	return RespondTo(c, record.Response)
}
