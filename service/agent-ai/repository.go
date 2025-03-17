package agent

import (
	"context"
	"healthmatefood-api/models"
)

type IAgentAIRepository interface {
	GenerateMealsPlan(ctx context.Context, user *models.User) (string, error)
	ConversationWithChat(ctx context.Context, prompt string) (string, error)
}
