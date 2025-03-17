package agent

import "github.com/gofiber/fiber/v2"

type IAgentAIHandler interface {
	GenerateMealsPlan(c *fiber.Ctx) error
}
