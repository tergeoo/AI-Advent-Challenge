package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Message представляет одно сообщение в диалоге
type Message struct {
	Role      string    // "user" или "assistant"
	Content   string    // Содержание сообщения
	Timestamp time.Time // Время сообщения
}

// AgentConfig конфигурация агента
type AgentConfig struct {
	APIKey       string
	Model        string
	Temperature  float32
	MaxTokens    int
	SystemPrompt string
}

// Agent представляет AI агента с памятью диалога
type Agent struct {
	config    AgentConfig
	client    *openai.Client
	ctx       context.Context
	history   []Message // История диалога
	systemMsg *Message  // Системное сообщение (опционально)
}

// Response ответ агента
type Response struct {
	Content          string
	TokensUsed       int
	PromptTokens     int
	CompletionTokens int
	ExecutionTime    time.Duration
	Model            string
}

// NewAgent создает нового агента
func NewAgent(config AgentConfig) *Agent {
	agent := &Agent{
		config:  config,
		client:  openai.NewClient(config.APIKey),
		ctx:     context.Background(),
		history: make([]Message, 0),
	}

	// Добавляем системное сообщение, если оно указано
	if config.SystemPrompt != "" {
		agent.systemMsg = &Message{
			Role:      "system",
			Content:   config.SystemPrompt,
			Timestamp: time.Now(),
		}
	}

	return agent
}

// Ask отправляет запрос агенту и получает ответ
func (a *Agent) Ask(userMessage string) (*Response, error) {
	// Добавляем сообщение пользователя в историю
	a.history = append(a.history, Message{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now(),
	})

	// Формируем сообщения для API
	messages := a.buildMessages()

	// Создаем запрос
	req := openai.ChatCompletionRequest{
		Model:       a.config.Model,
		Messages:    messages,
		Temperature: a.config.Temperature,
	}

	if a.config.MaxTokens > 0 {
		req.MaxTokens = a.config.MaxTokens
	}

	// Отправляем запрос
	start := time.Now()
	resp, err := a.client.CreateChatCompletion(a.ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе к API: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("получен пустой ответ от API")
	}

	assistantMessage := resp.Choices[0].Message.Content

	// Добавляем ответ ассистента в историю
	a.history = append(a.history, Message{
		Role:      "assistant",
		Content:   assistantMessage,
		Timestamp: time.Now(),
	})

	return &Response{
		Content:          assistantMessage,
		TokensUsed:       resp.Usage.TotalTokens,
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		ExecutionTime:    elapsed,
		Model:            resp.Model,
	}, nil
}

// buildMessages формирует список сообщений для API из истории
func (a *Agent) buildMessages() []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0)

	// Добавляем системное сообщение
	if a.systemMsg != nil {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.systemMsg.Content,
		})
	}

	// Добавляем историю диалога
	for _, msg := range a.history {
		var role string
		if msg.Role == "user" {
			role = openai.ChatMessageRoleUser
		} else {
			role = openai.ChatMessageRoleAssistant
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	return messages
}

// GetHistory возвращает историю диалога
func (a *Agent) GetHistory() []Message {
	return a.history
}

// ClearHistory очищает историю диалога
func (a *Agent) ClearHistory() {
	a.history = make([]Message, 0)
}

// GetHistorySize возвращает количество сообщений в истории
func (a *Agent) GetHistorySize() int {
	return len(a.history)
}

// GetTotalTokens подсчитывает примерное количество токенов в истории
// (упрощенная оценка: ~4 символа на токен для английского, ~2 для русского)
func (a *Agent) GetTotalTokens() int {
	total := 0

	if a.systemMsg != nil {
		total += len(a.systemMsg.Content) / 3 // Примерная оценка
	}

	for _, msg := range a.history {
		total += len(msg.Content) / 3 // Примерная оценка
	}

	return total
}

// SetSystemPrompt устанавливает системный промпт
func (a *Agent) SetSystemPrompt(prompt string) {
	if prompt == "" {
		a.systemMsg = nil
	} else {
		a.systemMsg = &Message{
			Role:      "system",
			Content:   prompt,
			Timestamp: time.Now(),
		}
	}
}

// GetLastMessage возвращает последнее сообщение ассистента
func (a *Agent) GetLastMessage() *Message {
	for i := len(a.history) - 1; i >= 0; i-- {
		if a.history[i].Role == "assistant" {
			return &a.history[i]
		}
	}
	return nil
}

// SaveHistory сохраняет историю диалога в JSON файл
func (a *Agent) SaveHistory(filename string) error {
	// Создаем структуру для сохранения
	data := struct {
		SystemPrompt string    `json:"system_prompt,omitempty"`
		History      []Message `json:"history"`
		SavedAt      time.Time `json:"saved_at"`
	}{
		History: a.history,
		SavedAt: time.Now(),
	}

	// Добавляем системный промпт, если есть
	if a.systemMsg != nil {
		data.SystemPrompt = a.systemMsg.Content
	}

	// Сериализуем в JSON с отступами для читаемости
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	// Записываем в файл
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("ошибка записи в файл: %w", err)
	}

	return nil
}

// LoadHistory загружает историю диалога из JSON файла
func (a *Agent) LoadHistory(filename string) error {
	// Читаем файл
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Файл не существует - это нормально для первого запуска
			return nil
		}
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	// Десериализуем JSON
	var data struct {
		SystemPrompt string    `json:"system_prompt,omitempty"`
		History      []Message `json:"history"`
		SavedAt      time.Time `json:"saved_at"`
	}

	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return fmt.Errorf("ошибка десериализации: %w", err)
	}

	// Загружаем историю
	a.history = data.History

	// Загружаем системный промпт, если он был сохранен
	// (но не перезаписываем, если новый уже установлен)
	if data.SystemPrompt != "" && a.systemMsg == nil {
		a.systemMsg = &Message{
			Role:      "system",
			Content:   data.SystemPrompt,
			Timestamp: time.Now(),
		}
	}

	return nil
}

// AutoSave автоматически сохраняет историю после каждого сообщения
func (a *Agent) AutoSave(filename string) error {
	return a.SaveHistory(filename)
}
