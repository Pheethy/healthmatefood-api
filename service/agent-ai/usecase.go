package agent

import (
	"context"
	"healthmatefood-api/models"
)

type IAgentAIUsecase interface {
	GenerateMealsPlan(ctx context.Context, user *models.User) (string, error)
}
