package main

type ListUsersResponse struct {
	Users []User `json:"users"`
}

func (s *Service) ListUsers(c *fiber.Ctx) error {
	query := c.Query("q", "")
	users, err := s.listDBUsers(c.Context(), query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	return c.JSON(ListUsersResponse{users})
}

func (s *Service) AddUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	ID, err := s.insertDBUser(c.Context(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}
	user.ID = ID

	return c.JSON(user)
}
