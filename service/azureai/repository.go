package azureai

import "healthmatefood-api/models"

type IAzureAIRepository interface {
	GetChatCompletion(prompt string) (*models.AI, error)
}
