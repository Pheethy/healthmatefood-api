package usecase

import (
	"context"
	"healthmatefood-api/models"
	"healthmatefood-api/service/agent-ai"
)

type agentAIUsecase struct {
	agentRepo agent.IAgentAIRepository
}

func NewAgentAIUsecase(agentRepo agent.IAgentAIRepository) agent.IAgentAIUsecase {
	return &agentAIUsecase{
		agentRepo: agentRepo,
	}
}

func (u *agentAIUsecase) GenerateMealsPlan(ctx context.Context, user *models.User) (string, error) {
	return u.agentRepo.GenerateMealsPlan(ctx, user)
}
