package azureai

import (
	"context"
	"healthmatefood-api/models"
)

type IAzureAIRepository interface {
	ConversationWithChat(ctx context.Context, prompt string) (string, error)
	GetChatCompletion(ctx context.Context, prompt string) (*models.AI, error)
}
