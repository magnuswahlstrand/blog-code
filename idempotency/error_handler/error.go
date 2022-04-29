package errorhandler

import "github.com/gofiber/fiber/v2"

func Handle(c *fiber.Ctx, request int, message ...string) error {
	return c.Status(request).JSON(fiber.NewError(request, message...))
}
