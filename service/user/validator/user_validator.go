package validator

import (
	"fmt"
	"net/http"

	"github.com/Pheethy/psql/helper"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

type Validation struct{}

func (v Validation) ValidateSignUp() fiber.Handler {
	return func(c *fiber.Ctx) error {
		params := c.Locals("params").(map[string]interface{})
		var key string

		/* key params */
		key = "email"
		email, emailOK := params[key]
		if !emailOK {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: was missing on body", key))
		}
		if err := validation.Validate(email, validation.By(helper.ValidateTypeString)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		key = "username"
		username, usernameOK := params[key]
		if !usernameOK {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: was missing on body", key))
		}
		if err := validation.Validate(username, validation.By(helper.ValidateTypeString)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		key = "password"
		password, passwordOK := params[key]
		if !passwordOK {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: was missing on body", key))
		}
		if err := validation.Validate(password, validation.By(helper.ValidateTypeString)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}
		return c.Next()
	}
}

func (v Validation) ValidateCreateUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		params := c.Locals("params").(map[string]interface{})
		var key string

		/* key params */
		key = "username"
		username, usernameOK := params[key]
		if !usernameOK {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: was missing on body", key))
		}
		if err := validation.Validate(username, validation.By(helper.ValidateTypeString)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		key = "password"
		password, passwordOK := params[key]
		if !passwordOK {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: was missing on body", key))
		}
		if err := validation.Validate(password, validation.By(helper.ValidateTypeString)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}
		return c.Next()
	}
}

func (v Validation) ValidateSignIn() fiber.Handler {
	return func(c *fiber.Ctx) error {
		params := c.Locals("params").(map[string]interface{})
		var key string

		/* key params */
		key = "email"
		email, emailOK := params[key]
		if !emailOK {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: was missing on body", key))
		}
		if err := validation.Validate(email, validation.By(helper.ValidateTypeString)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}

		key = "password"
		password, passwordOK := params[key]
		if !passwordOK {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: was missing on body", key))
		}
		if err := validation.Validate(password, validation.By(helper.ValidateTypeString)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}
		return c.Next()
	}
}

func (v Validation) ValidateParams(key string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		params := c.Params(key)
		if err := validation.Validate(params, validation.By(helper.ValidateTypeUUID)); err != nil {
			return fiber.NewError(http.StatusBadRequest, fmt.Sprintf("%s: %s", key, err.Error()))
		}
		return c.Next()
	}
}
