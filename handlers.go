package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

func (s *Server) dashboardHandler(c *fiber.Ctx) error {
	fromdate := time.Now().AddDate(0, 0, -10).Format("2006-01-02")
	return c.Render("dashboard/main", fiber.Map{
		"fromdate": fromdate,
	}, "layout")
}
