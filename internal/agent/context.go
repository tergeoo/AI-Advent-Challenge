package agent

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// ContextManager управляет историей сообщений с поддержкой сжатия
type ContextManager struct {
	// Полная история всех сообщений
	fullHistory []Message

	// Сжатые блоки истории (summary)
	summaries []string

	// Количество сообщений в каждом сжатом блоке
	compressionWindow int

	// Количество последних сообщений, хранимых "как есть"
	recentWindow int

	// OpenAI клиент для создания summary
	client *openai.Client
	ctx    context.Context
}

// ContextStats содержит статистику по контексту
type ContextStats struct {
	TotalMessages      int     // Всего сообщений
	CompressedBlocks   int     // Сжатых блоков
	RecentMessages     int     // Последних сообщений
	OriginalTokens     int     // Токенов в оригинальной истории
	CompressedTokens   int     // Токенов после сжатия
	CompressionRatio   float64 // Коэффициент сжатия
	TokensSaved        int     // Сэкономлено токенов
	CompressionPercent float64 // Процент сжатия
}

// NewContextManager создает новый менеджер контекста
func NewContextManager(client *openai.Client, compressionWindow, recentWindow int) *ContextManager {
	return &ContextManager{
		fullHistory:       make([]Message, 0),
		summaries:         make([]string, 0),
		compressionWindow: compressionWindow,
		recentWindow:      recentWindow,
		client:            client,
		ctx:               context.Background(),
	}
}

// AddMessage добавляет новое сообщение в историю
func (cm *ContextManager) AddMessage(role, content string) {
	cm.fullHistory = append(cm.fullHistory, Message{
		Role:    role,
		Content: content,
	})
}

// shouldCompress проверяет, нужно ли сжимать историю
func (cm *ContextManager) shouldCompress() bool {
	// Количество сообщений, которые можно сжать
	compressibleCount := len(cm.fullHistory) - cm.recentWindow

	// Количество уже сжатых сообщений
	alreadyCompressed := len(cm.summaries) * cm.compressionWindow

	// Количество несжатых сообщений (кроме recent)
	uncompressed := compressibleCount - alreadyCompressed

	// Сжимаем, если накопилось >= compressionWindow несжатых сообщений
	return uncompressed >= cm.compressionWindow
}

// CompressIfNeeded проверяет и сжимает историю при необходимости
func (cm *ContextManager) CompressIfNeeded() error {
	if !cm.shouldCompress() {
		return nil
	}

	// Находим диапазон для сжатия
	compressedCount := len(cm.summaries) * cm.compressionWindow
	startIdx := compressedCount
	endIdx := startIdx + cm.compressionWindow

	if endIdx > len(cm.fullHistory)-cm.recentWindow {
		endIdx = len(cm.fullHistory) - cm.recentWindow
	}

	// Извлекаем блок для сжатия
	blockToCompress := cm.fullHistory[startIdx:endIdx]

	// Создаем summary
	summary, err := cm.createSummary(blockToCompress)
	if err != nil {
		return fmt.Errorf("failed to create summary: %w", err)
	}

	// Сохраняем summary
	cm.summaries = append(cm.summaries, summary)

	return nil
}

// createSummary создает краткое содержание блока сообщений
func (cm *ContextManager) createSummary(messages []Message) (string, error) {
	// Формируем текст для суммаризации
	var dialogText string
	for _, msg := range messages {
		dialogText += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	// Запрос на создание summary
	prompt := fmt.Sprintf(`Создай краткое содержание следующего диалога, сохранив ключевые факты, решения и выводы:

%s

Краткое содержание (2-3 предложения):`, dialogText)

	resp, err := cm.client.CreateChatCompletion(cm.ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.3, // Низкая температура для точности
		MaxTokens:   150,
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	return resp.Choices[0].Message.Content, nil
}

// GetContextForRequest возвращает контекст для запроса (summaries + recent messages)
func (cm *ContextManager) GetContextForRequest() []Message {
	messages := make([]Message, 0)

	// Добавляем все summaries как одно системное сообщение
	if len(cm.summaries) > 0 {
		var combinedSummary string
		for i, summary := range cm.summaries {
			combinedSummary += fmt.Sprintf("[Блок %d]: %s\n", i+1, summary)
		}
		messages = append(messages, Message{
			Role:    "system",
			Content: fmt.Sprintf("Краткое содержание предыдущего диалога:\n%s", combinedSummary),
		})
	}

	// Добавляем последние N сообщений
	recentStart := len(cm.fullHistory) - cm.recentWindow
	if recentStart < 0 {
		recentStart = 0
	}

	recentMessages := cm.fullHistory[recentStart:]
	messages = append(messages, recentMessages...)

	return messages
}

// GetFullHistory возвращает полную историю (для сравнения)
func (cm *ContextManager) GetFullHistory() []Message {
	return cm.fullHistory
}

// GetStats возвращает статистику по контексту
func (cm *ContextManager) GetStats() ContextStats {
	stats := ContextStats{
		TotalMessages:    len(cm.fullHistory),
		CompressedBlocks: len(cm.summaries),
	}

	// Количество последних сообщений
	recentStart := len(cm.fullHistory) - cm.recentWindow
	if recentStart < 0 {
		stats.RecentMessages = len(cm.fullHistory)
	} else {
		stats.RecentMessages = cm.recentWindow
	}

	// Оценка токенов (приблизительно: 1 токен ≈ 3 символа для русского)
	// Полная история
	for _, msg := range cm.fullHistory {
		stats.OriginalTokens += len(msg.Content) / 3
	}

	// Сжатая история (summaries + recent)
	for _, summary := range cm.summaries {
		stats.CompressedTokens += len(summary) / 3
	}
	for i := len(cm.fullHistory) - stats.RecentMessages; i < len(cm.fullHistory); i++ {
		stats.CompressedTokens += len(cm.fullHistory[i].Content) / 3
	}

	// Расчет экономии
	if stats.OriginalTokens > 0 {
		stats.TokensSaved = stats.OriginalTokens - stats.CompressedTokens
		stats.CompressionRatio = float64(stats.CompressedTokens) / float64(stats.OriginalTokens)
		stats.CompressionPercent = (1.0 - stats.CompressionRatio) * 100
	}

	return stats
}

// Reset очищает всю историю и summaries
func (cm *ContextManager) Reset() {
	cm.fullHistory = make([]Message, 0)
	cm.summaries = make([]string, 0)
}
