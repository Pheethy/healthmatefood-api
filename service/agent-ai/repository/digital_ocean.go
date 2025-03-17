package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tmc/langchaingo/llms"
)

type digitalOceanLLM struct {
	endpoint  string
	accessKey string
	client    *http.Client
}

// NewDigitalOceanLLM สร้าง instance ของ custom LLM
func NewDigitalOceanLLM(endpoint, accessKey string) *digitalOceanLLM {
	return &digitalOceanLLM{
		endpoint:  endpoint,
		accessKey: accessKey,
		client:    &http.Client{},
	}
}

func (d *digitalOceanLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (*llms.ContentResponse, error) {
	var requestMessages []map[string]interface{}
	for _, msg := range messages {
		role := "user"
		if msg.Role == llms.ChatMessageTypeSystem {
			role = "system"
		}
		content := ""
		for _, part := range msg.Parts {
			if text, ok := part.(llms.TextContent); ok {
				content = text.Text
			}
		}
		requestMessages = append(requestMessages, map[string]interface{}{
			"role":    role,
			"content": content,
		})
	}

	// สร้าง request body
	body, err := json.Marshal(map[string]interface{}{
		"messages": requestMessages,
	})
	if err != nil {
		return nil, err
	}

	// สร้าง HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", d.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// ตั้ง header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.accessKey))

	// ส่ง request
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// อ่าน response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// ดึง content จาก response
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}
	content, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid content format")
	}

	// สร้าง response สำหรับ LangChainGo
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{Content: content},
		},
	}, nil
}

// Call ฟังก์ชันที่ LangChainGo ต้องการ (สำหรับ prompt เดี่ยว)
func (d *digitalOceanLLM) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
		},
	}
	resp, err := d.GenerateContent(ctx, messages, opts...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}
	return resp.Choices[0].Content, nil
}
