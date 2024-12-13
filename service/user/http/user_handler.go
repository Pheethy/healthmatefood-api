package handler

import (
	"healthmatefood-api/constants"
	"healthmatefood-api/models"
	"healthmatefood-api/service/user"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
)

type userHandler struct {
	userUs user.IUserUsecase
}

func NewUserHandler(userUs user.IUserUsecase) user.IUserHandler {
	return &userHandler{
		userUs: userUs,
	}
}

// @Summary     FetchAllUsers
// @Description Get list users
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       search_word query string false "example: john doe"
// @Param       page        query int    false "example: 1"
// @Param       per_page    query int    false "example: 10"
// @Success     200         {object}     map[string]interface{}
// @Failure     500         {object}     constants.ErrorResponse
// @Router      /v1/user/list [get]
func (u *userHandler) FetchAllUsers(c *fiber.Ctx) error {
	ctx := c.UserContext()
	args := new(sync.Map)
	searchWord := c.Query("search_word")
	page, pageErr := strconv.Atoi(c.Query("page", "1"))
	perPage, perPageErr := strconv.Atoi(c.Query("per_page", "10"))
	if pageErr == nil && perPageErr == nil {
		args.Store("page", page)
		args.Store("per_page", perPage)
	}
	if searchWord != "" {
		args.Store("search_word", searchWord)
	}

	users, err := u.userUs.FetchAllUsers(ctx, args)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	resp := map[string]interface{}{
		"users": users,
	}

	return c.Status(http.StatusOK).JSON(resp)
}

// @Summary     FetchOneUserById
// @Description Get One users
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user_id path string true "example:257d3552-c186-4c23-aa5d-1ea53f453e2a"
// @Success     200         {object}     map[string]interface{}
// @Failure     500         {object}     constants.ErrorResponse
// @Router      /v1/user/{user_id} [get]
func (u *userHandler) FetchOneUserById(c *fiber.Ctx) error {
	ctx := c.UserContext()
	id := uuid.FromStringOrNil(c.Params("user_id"))

	user, err := u.userUs.FetchOneUserById(ctx, &id)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	resp := map[string]interface{}{
		"user": user,
	}
	return c.Status(http.StatusOK).JSON(resp)
}

// @Summary     SignUp
// @Description Sign-up to system with email and password
// @Tags        users
// @Accept      multipart/form-data
// @Produce     json
// @Param       username formData string true "username user" default:"john_doe"
// @Param       email    formData string true "email user" example:"customer001@odor.com"
// @Param       password formData string true "password user" example:"strongpassword123"
// @Param       files    formData file   false "user profile image"
// @Success     200 {object} map[string]interface{} "Successful response" example({"message":"successful","user_id":"uuid-123","username":"john_doe"})
// @Failure     400 {object} constants.ErrorResponse "Invalid email format, duplicate username, or duplicate email"
// @Failure     422 {object} constants.ErrorResponse "Password hashing error"
// @Failure     500 {object} constants.ErrorResponse "Internal server error"
// @Router      /v1/user/sign-up [post]
func (u *userHandler) SignUp(c *fiber.Ctx) error {
	ctx := c.UserContext()
	params := c.Locals("params").(map[string]interface{})
	files := c.Locals("files").([]*multipart.FileHeader)
	user := models.NewUserWithParams(params, nil)
	user.NewID()
	user.SetCreatedAt()
	user.SetUpdatedAt()

	if ok := user.IsEmail(); !ok {
		return fiber.NewError(http.StatusBadRequest, constants.ERROR_EMAIL_PATTERN_IS_INVALID)
	}

	if err := user.BcryptHashing(); err != nil {
		return fiber.NewError(http.StatusUnprocessableEntity, err.Error())
	}

	if err := u.userUs.UpsertUser(ctx, user, false, files); err != nil {
		if ok := strings.Contains(err.Error(), constants.ERROR_USERNAME_WAS_DUPLICATED); ok {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if ok := strings.Contains(err.Error(), constants.ERROR_EMAIL_WAS_DUPLICATED); ok {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"message":  "successful",
		"user_id":  user.Id,
		"username": user.Username,
	}
	return c.Status(http.StatusOK).JSON(resp)
}

// @Summary     SignIn
// @Description Sign-in to system with email and password
// @Tags        users
// @Accept      multipart/form-data
// @Produce     json
// @Param       email formData string true "Email user"
// @Param       password formData string true "Password user"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} constants.ErrorResponse "email pattern is invalid"
// @Failure     404 {object} constants.ErrorResponse "user not found"
// @Failure     400 {object} constants.ErrorResponse "password is invalid"
// @Failure     500 {object} constants.ErrorResponse  "Internal server error"
// @Router      /v1/user/sign-in [post]
func (u *userHandler) SignIn(c *fiber.Ctx) error {
	ctx := c.UserContext()
	params := c.Locals("params").(map[string]interface{})
	user := models.NewUserWithParams(params, nil)

	if ok := user.IsEmail(); !ok {
		return fiber.NewError(http.StatusBadRequest, constants.ERROR_EMAIL_PATTERN_IS_INVALID)
	}

	userPassport, err := u.userUs.FetchUserPassport(ctx, user)
	if err != nil {
		if ok := strings.Contains(err.Error(), constants.ERROR_USER_NOT_FOUND); ok {
			return fiber.NewError(http.StatusNotFound, err.Error())
		}
		if ok := strings.Contains(err.Error(), constants.ERROR_PASSWORD_IS_INVALID); ok {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"passport": userPassport,
	}

	return c.Status(http.StatusOK).JSON(resp)
}

// @Summary     SignUpAdmin
// @Description Sign-up admin to system with email and password
// @Tags        users
// @Accept      multipart/form-data
// @Produce     json
// @Param       username formData string true "Username user"
// @Param       email    formData string true "Email user" example:"example@odor.com"
// @Param       password formData string true "Password user" example:"strongpassword123"
// @Param       files    formData file   false "User profile image"
// @Success     200 {object} map[string]interface{} "Successful response" example({"message":"successful","user_id":"uuid-123","username":"john_doe"})
// @Failure     400 {object} constants.ErrorResponse "Invalid email format, duplicate username, or duplicate email"
// @Failure     422 {object} constants.ErrorResponse "Password hashing error"
// @Failure     500 {object} constants.ErrorResponse "Internal server error"
// @Router      /v1/user/admin [post]
func (u *userHandler) SignUpAdmin(c *fiber.Ctx) error {
	ctx := c.UserContext()
	params := c.Locals("params").(map[string]interface{})
	files := c.Locals("files").([]*multipart.FileHeader)
	user := models.NewUserWithParams(params, nil)
	user.NewID()
	user.SetCreatedAt()
	user.SetUpdatedAt()

	if ok := user.IsEmail(); !ok {
		return fiber.NewError(http.StatusBadRequest, constants.ERROR_EMAIL_PATTERN_IS_INVALID)
	}

	if err := user.BcryptHashing(); err != nil {
		return fiber.NewError(http.StatusUnprocessableEntity, err.Error())
	}

	if err := u.userUs.UpsertUser(ctx, user, true, files); err != nil {
		if ok := strings.Contains(err.Error(), constants.ERROR_USERNAME_WAS_DUPLICATED); ok {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if ok := strings.Contains(err.Error(), constants.ERROR_EMAIL_WAS_DUPLICATED); ok {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"message":  "successful",
		"user_id":  user.Id,
		"username": user.Username,
	}
	return c.Status(http.StatusOK).JSON(resp)
}

// @Summary     RefreshUserPassport
// @Description Refresh user passport
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       refresh_token query string true "refresh_token"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} constants.ErrorResponse
// @Router      /v1/user/refresh-passport [get]
func (u *userHandler) RefreshUserPassport(c *fiber.Ctx) error {
	ctx := c.UserContext()
	refreshToken := c.Query("refresh_token")

	passport, err := u.userUs.RefreshUserPassport(ctx, refreshToken)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	resp := map[string]interface{}{
		"passport": passport,
	}

	return c.Status(http.StatusOK).JSON(resp)
}
