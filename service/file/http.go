package file

import "github.com/gofiber/fiber/v2"

type IFileHandler interface {
	UploadFile(c *fiber.Ctx) error
	DeleteFile(c *fiber.Ctx) error
}
