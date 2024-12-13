package handler

import (
	"errors"
	"fmt"
	"healthmatefood-api/models"
	user_mocks "healthmatefood-api/service/user/mocks"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/Pheethy/psql/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchAllUsers(t *testing.T) {
	queryparams := url.Values{}
	queryparams.Add("page", "1")
	queryparams.Add("per_page", "10")
	queryparams.Add("search_word", "test")
	id := uuid.FromStringOrNil("48a2ad72-9133-4358-b905-b20621ed8297")
	ti := helper.NewTimestampFromTime(time.Now())
	mockUsers := []*models.User{
		{
			Id:        &id,
			Username:  "pheet",
			Password:  "pheet1234",
			Email:     "e7xKm@example.com",
			RoleId:    1,
			Role:      "customer",
			CreatedAt: &ti,
			UpdatedAt: &ti,
		},
	}
	t.Run("success", func(t *testing.T) {
		app := fiber.New()
		userUs := new(user_mocks.IUserUsecase)
		userUs.On("FetchAllUsers", mock.Anything, mock.AnythingOfType("*sync.Map")).Return(mockUsers, nil).Run(func(args mock.Arguments) {
			epCtx := args.Get(0)
			epArg := args.Get(1).(*sync.Map)

			epPage, _ := epArg.Load("page")
			epPerPage, _ := epArg.Load("per_page")
			epSearchWord, _ := epArg.Load("search_word")

			assert.NotNil(t, epCtx)
			assert.Equal(t, epPage, 1)
			assert.Equal(t, epPerPage, 10)
			assert.Equal(t, epSearchWord, "test")
		})
		userHandler := NewUserHandler(userUs)
		app.Get("/v1/user/list", func(c *fiber.Ctx) error {
			return userHandler.FetchAllUsers(c)
		})
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/user/list?%s", queryparams.Encode()), nil)
		req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
	t.Run("error_internal_server", func(t *testing.T) {
		app := fiber.New()
		userUs := new(user_mocks.IUserUsecase)
		userUs.On("FetchAllUsers", mock.Anything, mock.AnythingOfType("*sync.Map")).Return(nil, errors.New("unexpected")).Run(func(args mock.Arguments) {
			epCtx := args.Get(0)
			epArg := args.Get(1).(*sync.Map)

			epPage, _ := epArg.Load("page")
			epPerPage, _ := epArg.Load("per_page")
			epSearchWord, _ := epArg.Load("search_word")

			assert.NotNil(t, epCtx)
			assert.Equal(t, epPage, 1)
			assert.Equal(t, epPerPage, 10)
			assert.Equal(t, epSearchWord, "test")
		})
		userHandler := NewUserHandler(userUs)
		app.Get("/v1/user/list", func(c *fiber.Ctx) error {
			return userHandler.FetchAllUsers(c)
		})
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/user/list?%s", queryparams.Encode()), nil)
		req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
