package agent

import (
	"fmt"
)

// TokenStats статистика использования токенов
type TokenStats struct {
	TotalRequests         int     // Всего запросов
	TotalTokens           int     // Всего токенов использовано
	TotalPromptTokens     int     // Токенов в промптах
	TotalCompletionTokens int     // Токенов в ответах
	TotalCost             float64 // Общая стоимость

	CurrentContextTokens int // Токенов в текущем контексте
	MaxContextTokens     int // Максимальный лимит контекста

	// Стоимость за 1M токенов
	InputPrice  float64
	OutputPrice float64
}

// NewTokenStats создает новую статистику токенов
func NewTokenStats(maxContext int, inputPrice, outputPrice float64) *TokenStats {
	return &TokenStats{
		MaxContextTokens: maxContext,
		InputPrice:       inputPrice,
		OutputPrice:      outputPrice,
	}
}

// AddRequest добавляет информацию о запросе
func (ts *TokenStats) AddRequest(promptTokens, completionTokens int) {
	ts.TotalRequests++
	ts.TotalPromptTokens += promptTokens
	ts.TotalCompletionTokens += completionTokens
	ts.TotalTokens = ts.TotalPromptTokens + ts.TotalCompletionTokens

	// Обновляем стоимость
	inputCost := float64(promptTokens) / 1_000_000 * ts.InputPrice
	outputCost := float64(completionTokens) / 1_000_000 * ts.OutputPrice
	ts.TotalCost += inputCost + outputCost
}

// UpdateContextSize обновляет размер текущего контекста
func (ts *TokenStats) UpdateContextSize(tokens int) {
	ts.CurrentContextTokens = tokens
}

// GetContextUsagePercent возвращает процент использования контекста
func (ts *TokenStats) GetContextUsagePercent() float64 {
	if ts.MaxContextTokens == 0 {
		return 0
	}
	return float64(ts.CurrentContextTokens) / float64(ts.MaxContextTokens) * 100
}

// IsNearLimit проверяет, близки ли мы к лимиту (>80%)
func (ts *TokenStats) IsNearLimit() bool {
	return ts.GetContextUsagePercent() > 80
}

// IsOverLimit проверяет, превышен ли лимит
func (ts *TokenStats) IsOverLimit() bool {
	return ts.CurrentContextTokens > ts.MaxContextTokens
}

// GetRemainingTokens возвращает сколько токенов осталось
func (ts *TokenStats) GetRemainingTokens() int {
	remaining := ts.MaxContextTokens - ts.CurrentContextTokens
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetAverageCostPerRequest возвращает среднюю стоимость запроса
func (ts *TokenStats) GetAverageCostPerRequest() float64 {
	if ts.TotalRequests == 0 {
		return 0
	}
	return ts.TotalCost / float64(ts.TotalRequests)
}

// GetAverageTokensPerRequest возвращает среднее количество токенов на запрос
func (ts *TokenStats) GetAverageTokensPerRequest() float64 {
	if ts.TotalRequests == 0 {
		return 0
	}
	return float64(ts.TotalTokens) / float64(ts.TotalRequests)
}

// FormatContextBar создает визуальную полосу использования контекста
func (ts *TokenStats) FormatContextBar(width int) string {
	if width <= 0 {
		width = 50
	}

	percent := ts.GetContextUsagePercent()
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			if percent > 90 {
				bar += "█" // Красный (критично)
			} else if percent > 80 {
				bar += "▓" // Желтый (предупреждение)
			} else {
				bar += "▒" // Зеленый (норма)
			}
		} else {
			bar += "░"
		}
	}

	return fmt.Sprintf("[%s] %.1f%%", bar, percent)
}

// GetWarningMessage возвращает предупреждение, если есть проблемы
func (ts *TokenStats) GetWarningMessage() string {
	if ts.IsOverLimit() {
		return fmt.Sprintf("⚠️  КРИТИЧНО: Превышен лимит контекста! (%d / %d токенов)",
			ts.CurrentContextTokens, ts.MaxContextTokens)
	}

	if ts.IsNearLimit() {
		return fmt.Sprintf("⚠️  ВНИМАНИЕ: Близко к лимиту контекста (%d / %d токенов)",
			ts.CurrentContextTokens, ts.MaxContextTokens)
	}

	return ""
}

// ModelLimits информация о лимитах моделей
var ModelLimits = map[string]int{
	"gpt-4o-mini":         128000,
	"gpt-4o":              128000,
	"gpt-4-turbo-preview": 128000,
	"gpt-4":               8192,
	"gpt-3.5-turbo":       16385,
}

// ModelPricing информация о ценах моделей (input/output per 1M tokens)
var ModelPricing = map[string][2]float64{
	"gpt-4o-mini":         {0.150, 0.600},
	"gpt-4o":              {2.50, 10.00},
	"gpt-4-turbo-preview": {10.00, 30.00},
	"gpt-4":               {30.00, 60.00},
	"gpt-3.5-turbo":       {0.50, 1.50},
}

// GetModelLimit возвращает лимит контекста для модели
func GetModelLimit(model string) int {
	if limit, ok := ModelLimits[model]; ok {
		return limit
	}
	return 4096 // Дефолтное значение для неизвестных моделей
}

// GetModelPricing возвращает цены для модели
func GetModelPricing(model string) (inputPrice, outputPrice float64) {
	if pricing, ok := ModelPricing[model]; ok {
		return pricing[0], pricing[1]
	}
	return 0.50, 1.50 // Дефолтные значения
}
