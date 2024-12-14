package route

import (
	"healthmatefood-api/service/user"
	user_validator "healthmatefood-api/service/user/validator"

	"github.com/gofiber/fiber/v2"
)

type Route struct {
	e fiber.Router
}

func NewRoute(e fiber.Router) *Route {
	return &Route{
		e: e,
	}
}

func (r *Route) RegisterUser(handler user.IUserHandler, validator user_validator.Validation) {
	r.e.Get("/user/list", handler.FetchAllUsers)
	r.e.Get("/user/:user_id", handler.FetchOneUserById)
	r.e.Get("/user/info/:user_id", handler.FetchOneUserInfoByUserId)
	r.e.Post("/user/sign-in", validator.ValidateSignIn(), handler.SignIn)
	r.e.Post("/user/sign-up", validator.ValidateSignUp(), handler.SignUp)
	r.e.Post("/user/admin", validator.ValidateSignUp(), handler.SignUpAdmin)
	r.e.Post("/user/refresh", handler.RefreshUserPassport)
	r.e.Post("/user/info", handler.CreateUserInfo)
	r.e.Put("/user/info/:user_id", handler.UpdateUserInfo)
}
