package http

import (
	"healthmatefood-api/models"
	"healthmatefood-api/service/agent-ai"
	"healthmatefood-api/service/user"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type agentAIHandler struct {
	agentUs agent.IAgentAIUsecase
	userUs  user.IUserUsecase
}

func NewAgentAIHandler(agentUs agent.IAgentAIUsecase, userUs user.IUserUsecase) agent.IAgentAIHandler {
	return &agentAIHandler{
		agentUs: agentUs,
		userUs:  userUs,
	}
}

func (h *agentAIHandler) GenerateMealsPlan(c *fiber.Ctx) error {
	ctx := c.UserContext()
	params := c.Locals("params").(map[string]interface{})
	userInfo := models.NewUserInfoWithParams(params, nil)
	userInfo.GetBMR()
	userInfo.GetCaloriesLimit()
	user := new(models.User)
	user.UserInfo = userInfo

	plan, err := h.agentUs.GenerateMealsPlan(ctx, user)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}

	resp := map[string]interface{}{
		"plan": plan,
	}

	return c.Status(http.StatusOK).JSON(resp)
}
