package user

import "github.com/gofiber/fiber/v2"

type IUserHandler interface {
	FetchAllUsers(c *fiber.Ctx) error
	FetchOneUserById(c *fiber.Ctx) error
	SignIn(c *fiber.Ctx) error
	SignUp(c *fiber.Ctx) error
	SignUpAdmin(c *fiber.Ctx) error
	RefreshUserPassport(c *fiber.Ctx) error
}
