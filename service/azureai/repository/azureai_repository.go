package repository

import (
	"context"
	"encoding/json"
	"healthmatefood-api/config"
	"healthmatefood-api/models"

	"github.com/sashabaranov/go-openai"
)

type azureAIRepository struct {
	client *openai.Client
	cfg    config.IAzureAIConfig
}

func NewAzureAIRepository(client *openai.Client, cfg config.IAzureAIConfig) *azureAIRepository {
	return &azureAIRepository{
		client: client,
		cfg:    cfg,
	}
}

func (r *azureAIRepository) ConversationWithChat(ctx context.Context, prompt string) (string, error) {
	request := openai.ChatCompletionRequest{
		Model: r.cfg.DeploymentName(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := r.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", nil
	}

	var content string
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return content, nil
}

func (r *azureAIRepository) GetChatCompletion(ctx context.Context, prompt string) (*models.AI, error) {
	resp, err := r.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: r.cfg.DeploymentName(),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role: openai.ChatMessageRoleSystem,
					Content: `***Return only json format,this is format
				          {
				            "user": "pheet",
				            "say": "Why is the sky blue?",
				            "age": 25
				          }
			          `,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
			MaxTokens:   800,
		},
	)
	if err != nil {
		return nil, err
	}
	response := new(models.AI)
	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		if err := json.Unmarshal([]byte(content), response); err != nil {
			return nil, err
		}
	}

	return response, nil
}
