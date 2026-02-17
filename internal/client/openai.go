package client

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// OpenAIClient обертка над OpenAI клиентом
type OpenAIClient struct {
	client *openai.Client
	ctx    context.Context
}

// NewOpenAIClient создает новый OpenAI клиент
func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		ctx:    context.Background(),
	}
}

// CompletionRequest представляет запрос к API
type CompletionRequest struct {
	Prompt         string
	MaxTokens      int
	Temperature    float32
	Stop           []string
	ResponseFormat *openai.ChatCompletionResponseFormat
}

// CompletionResponse представляет ответ от API
type CompletionResponse struct {
	Content          string
	TotalTokens      int
	PromptTokens     int
	CompletionTokens int
	Model            string
	FinishReason     string
}

// CreateCompletion выполняет запрос к OpenAI API
func (c *OpenAIClient) CreateCompletion(req CompletionRequest) (*CompletionResponse, error) {
	chatReq := openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: req.Prompt,
			},
		},
	}

	// Опциональные параметры
	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		chatReq.Temperature = req.Temperature
	}
	if len(req.Stop) > 0 {
		chatReq.Stop = req.Stop
	}
	if req.ResponseFormat != nil {
		chatReq.ResponseFormat = req.ResponseFormat
	}

	resp, err := c.client.CreateChatCompletion(c.ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе к OpenAI API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("получен пустой ответ от API")
	}

	return &CompletionResponse{
		Content:          resp.Choices[0].Message.Content,
		TotalTokens:      resp.Usage.TotalTokens,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		Model:            resp.Model,
		FinishReason:     string(resp.Choices[0].FinishReason),
	}, nil
}
