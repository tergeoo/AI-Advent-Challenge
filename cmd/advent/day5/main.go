package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/config"
	"github.com/georgijter-grigoranc/ai-advent-challenge/pkg/utils"
	openai "github.com/sashabaranov/go-openai"
)

// Информация о модели
type ModelInfo struct {
	Name        string
	DisplayName string
	Tier        string  // "weak", "medium", "strong"
	InputPrice  float64 // цена за 1M токенов (input)
	OutputPrice float64 // цена за 1M токенов (output)
}

// Результат теста модели
type ModelResult struct {
	Model            ModelInfo
	Response         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	ExecutionTime    time.Duration
	InputCost        float64
	OutputCost       float64
	TotalCost        float64
}

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Заголовок
	utils.PrintHeader("Day 5: Сравнение версий моделей")

	// Описание эксперимента
	printExperimentDescription()

	// Определяем модели для тестирования
	models := []ModelInfo{
		{
			Name:        openai.GPT4oMini,
			DisplayName: "GPT-4o-mini",
			Tier:        "weak",
			InputPrice:  0.150, // $0.150 per 1M input tokens
			OutputPrice: 0.600, // $0.600 per 1M output tokens
		},
		{
			Name:        openai.GPT4o,
			DisplayName: "GPT-4o",
			Tier:        "medium",
			InputPrice:  2.50,  // $2.50 per 1M input tokens
			OutputPrice: 10.00, // $10.00 per 1M output tokens
		},
		{
			Name:        openai.GPT4TurboPreview,
			DisplayName: "GPT-4 Turbo",
			Tier:        "strong",
			InputPrice:  10.00, // $10.00 per 1M input tokens
			OutputPrice: 30.00, // $30.00 per 1M output tokens
		},
	}

	// Тестовый промпт - сложная задача, требующая рассуждений
	prompt := `Реши следующую логическую задачу:

В комнате находятся 3 лампочки, а выключатели для них - в другой комнате.
Ты можешь включить любые выключатели, но зайти в комнату с лампочками можешь только один раз.
Как определить, какой выключатель управляет какой лампочкой?

Объясни решение пошагово и дай обоснование.`

	utils.PrintSection("📋", "ТЕСТОВЫЙ ПРОМПТ")
	fmt.Printf("%s\n\n", prompt)
	utils.PrintDivider()

	// Запуск тестов для каждой модели
	results := make([]ModelResult, 0, len(models))

	for _, model := range models {
		result := testModel(cfg.OpenAIKey, model, prompt)
		results = append(results, result)

		// Небольшая пауза между запросами
		time.Sleep(1 * time.Second)
	}

	// Сравнение результатов
	compareModels(results)

	// Рекомендации
	printRecommendations()

	utils.PrintDivider()
	utils.PrintSuccess("Задание Day 5 выполнено!")
}

func printExperimentDescription() {
	utils.PrintSection("🧪", "ОПИСАНИЕ ЭКСПЕРИМЕНТА")

	fmt.Println("Мы сравним три модели OpenAI разного уровня:")
	fmt.Println()
	fmt.Println("🔹 Слабая модель:   GPT-4o-mini")
	fmt.Println("   • Быстрая и недорогая")
	fmt.Println("   • Подходит для простых задач")
	fmt.Println("   • $0.15/$0.60 per 1M tokens (input/output)")
	fmt.Println()
	fmt.Println("🔸 Средняя модель:  GPT-4o")
	fmt.Println("   • Баланс скорости и качества")
	fmt.Println("   • Универсальная модель")
	fmt.Println("   • $2.50/$10.00 per 1M tokens (input/output)")
	fmt.Println()
	fmt.Println("🔺 Сильная модель:  GPT-4 Turbo")
	fmt.Println("   • Расширенный контекст (128K токенов)")
	fmt.Println("   • Высокое качество для сложных задач")
	fmt.Println("   • $10.00/$30.00 per 1M tokens (input/output)")
	fmt.Println()
	fmt.Println("Критерии сравнения:")
	fmt.Println("  • Качество ответа (полнота, правильность)")
	fmt.Println("  • Время выполнения")
	fmt.Println("  • Количество токенов")
	fmt.Println("  • Стоимость запроса")
	fmt.Println()

	utils.PrintDivider()
}

func testModel(apiKey string, model ModelInfo, prompt string) ModelResult {
	utils.PrintSection("🤖", fmt.Sprintf("ТЕСТИРОВАНИЕ: %s", model.DisplayName))
	fmt.Printf("Tier: %s\n", model.Tier)
	fmt.Printf("Цена: $%.3f (input) / $%.3f (output) per 1M tokens\n\n", model.InputPrice, model.OutputPrice)

	client := openai.NewClient(apiKey)
	ctx := context.Background()

	start := time.Now()

	// Создаем запрос
	req := openai.ChatCompletionRequest{
		Model: model.Name,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
	}

	resp, err := client.CreateChatCompletion(ctx, req)
	elapsed := time.Since(start)

	if err != nil {
		log.Printf("❌ Ошибка при тестировании модели %s: %v\n", model.DisplayName, err)
		utils.PrintDivider()
		return ModelResult{Model: model}
	}

	if len(resp.Choices) == 0 {
		log.Printf("❌ Пустой ответ от модели %s\n", model.DisplayName)
		utils.PrintDivider()
		return ModelResult{Model: model}
	}

	response := resp.Choices[0].Message.Content
	promptTokens := resp.Usage.PromptTokens
	completionTokens := resp.Usage.CompletionTokens
	totalTokens := resp.Usage.TotalTokens

	// Расчет стоимости
	inputCost := float64(promptTokens) / 1_000_000 * model.InputPrice
	outputCost := float64(completionTokens) / 1_000_000 * model.OutputPrice
	totalCost := inputCost + outputCost

	// Вывод ответа (первые 500 символов)
	fmt.Println("Ответ:")
	if len(response) > 500 {
		fmt.Printf("%s...\n\n", response[:500])
		fmt.Printf("(показаны первые 500 из %d символов)\n\n", len(response))
	} else {
		fmt.Printf("%s\n\n", response)
	}

	// Статистика
	utils.PrintTokenStats(totalTokens, promptTokens, completionTokens)
	utils.PrintKeyValue("Время выполнения", elapsed.Round(time.Millisecond).String())
	utils.PrintKeyValue("Стоимость (input)", fmt.Sprintf("$%.6f", inputCost))
	utils.PrintKeyValue("Стоимость (output)", fmt.Sprintf("$%.6f", outputCost))
	utils.PrintKeyValue("Стоимость (всего)", fmt.Sprintf("$%.6f", totalCost))

	utils.PrintDivider()

	return ModelResult{
		Model:            model,
		Response:         response,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		ExecutionTime:    elapsed,
		InputCost:        inputCost,
		OutputCost:       outputCost,
		TotalCost:        totalCost,
	}
}

func compareModels(results []ModelResult) {
	utils.PrintSection("📊", "СРАВНИТЕЛЬНЫЙ АНАЛИЗ")

	if len(results) == 0 {
		fmt.Println("Нет результатов для сравнения")
		return
	}

	// Таблица с основными метриками
	fmt.Println("\n┌──────────────────────┬─────────────┬────────────┬──────────────┬──────────────┐")
	fmt.Println("│ Модель               │ Токены      │ Время      │ Стоимость    │ $/1K токенов │")
	fmt.Println("├──────────────────────┼─────────────┼────────────┼──────────────┼──────────────┤")

	for _, result := range results {
		if result.TotalTokens == 0 {
			continue
		}

		costPer1K := (result.TotalCost / float64(result.TotalTokens)) * 1000

		fmt.Printf("│ %-20s │ %11d │ %10s │ $%11.6f │ $%11.6f │\n",
			truncate(result.Model.DisplayName, 20),
			result.TotalTokens,
			result.ExecutionTime.Round(time.Millisecond).String(),
			result.TotalCost,
			costPer1K,
		)
	}

	fmt.Println("└──────────────────────┴─────────────┴────────────┴──────────────┴──────────────┘")

	// Анализ качества ответов
	fmt.Println("\n📝 АНАЛИЗ КАЧЕСТВА ОТВЕТОВ:\n")

	for i, result := range results {
		if result.TotalTokens == 0 {
			continue
		}

		fmt.Printf("%d. %s (%s модель):\n", i+1, result.Model.DisplayName, result.Model.Tier)

		responseLength := len(result.Response)
		wordCount := len(splitWords(result.Response))

		utils.PrintKeyValue("  Длина ответа", fmt.Sprintf("%d символов, ~%d слов", responseLength, wordCount))

		// Проверка наличия ключевых элементов правильного решения
		hasTemperature := contains(result.Response, "температур") || contains(result.Response, "тепл") || contains(result.Response, "горяч")
		hasSteps := contains(result.Response, "шаг") || contains(result.Response, "Шаг")
		hasExplanation := contains(result.Response, "потому") || contains(result.Response, "так как") || contains(result.Response, "поэтому")

		if hasTemperature {
			utils.PrintSuccess("  ✓ Упоминает температуру (ключ к решению)")
		} else {
			utils.PrintError("  ✗ Не упоминает температуру")
		}

		if hasSteps {
			utils.PrintSuccess("  ✓ Структурированное пошаговое решение")
		}

		if hasExplanation {
			utils.PrintSuccess("  ✓ Присутствуют объяснения и обоснования")
		}

		// Оценка на основе tier
		switch result.Model.Tier {
		case "weak":
			utils.PrintInfo("  💡 Быстро и дешево, подходит для простых задач")
		case "medium":
			utils.PrintInfo("  💡 Оптимальный баланс цены и качества")
		case "strong":
			utils.PrintInfo("  💡 Максимальное качество рассуждений")
		}

		fmt.Println()
	}

	// Сравнение скорости
	fmt.Println("⚡ СРАВНЕНИЕ СКОРОСТИ:\n")

	if len(results) > 1 {
		fastest := results[0]
		slowest := results[0]

		for _, r := range results {
			if r.TotalTokens == 0 {
				continue
			}
			if r.ExecutionTime < fastest.ExecutionTime {
				fastest = r
			}
			if r.ExecutionTime > slowest.ExecutionTime {
				slowest = r
			}
		}

		utils.PrintSuccess(fmt.Sprintf("Самая быстрая: %s - %s", fastest.Model.DisplayName, fastest.ExecutionTime.Round(time.Millisecond)))
		utils.PrintError(fmt.Sprintf("Самая медленная: %s - %s", slowest.Model.DisplayName, slowest.ExecutionTime.Round(time.Millisecond)))

		if slowest.ExecutionTime > 0 {
			speedup := float64(slowest.ExecutionTime) / float64(fastest.ExecutionTime)
			fmt.Printf("\nСамая быстрая модель работает в %.1fx раз быстрее самой медленной\n", speedup)
		}
	}

	fmt.Println()

	// Сравнение стоимости
	fmt.Println("💰 СРАВНЕНИЕ СТОИМОСТИ:\n")

	if len(results) > 1 {
		cheapest := results[0]
		expensive := results[0]

		for _, r := range results {
			if r.TotalTokens == 0 {
				continue
			}
			if r.TotalCost < cheapest.TotalCost {
				cheapest = r
			}
			if r.TotalCost > expensive.TotalCost {
				expensive = r
			}
		}

		utils.PrintSuccess(fmt.Sprintf("Самая дешевая: %s - $%.6f", cheapest.Model.DisplayName, cheapest.TotalCost))
		utils.PrintError(fmt.Sprintf("Самая дорогая: %s - $%.6f", expensive.Model.DisplayName, expensive.TotalCost))

		if cheapest.TotalCost > 0 {
			priceRatio := expensive.TotalCost / cheapest.TotalCost
			fmt.Printf("\nСамая дорогая модель стоит в %.1fx раз больше самой дешевой\n", priceRatio)
		}
	}

	fmt.Println()

	// Расчет стоимости на 1000 запросов
	fmt.Println("💵 СТОИМОСТЬ НА 1000 ЗАПРОСОВ:\n")

	for _, result := range results {
		if result.TotalTokens == 0 {
			continue
		}

		costPer1000 := result.TotalCost * 1000
		fmt.Printf("  %s: $%.2f\n", result.Model.DisplayName, costPer1000)
	}

	fmt.Println()
	utils.PrintDivider()
}

func printRecommendations() {
	utils.PrintSection("🎯", "РЕКОМЕНДАЦИИ ПО ВЫБОРУ МОДЕЛИ")

	fmt.Println("\n┌──────────────────┬────────────────────────────────────────────────────────┐")
	fmt.Println("│ Модель           │ Когда использовать                                     │")
	fmt.Println("├──────────────────┼────────────────────────────────────────────────────────┤")
	fmt.Println("│ GPT-4o-mini      │ • Простые задачи (FAQ, классификация)                 │")
	fmt.Println("│ (Слабая)         │ • Высокая нагрузка (много запросов)                    │")
	fmt.Println("│                  │ • Ограниченный бюджет                                  │")
	fmt.Println("│                  │ • Быстрые ответы критичны                              │")
	fmt.Println("│                  │ ✅ Лучший выбор для: чат-боты, резюме, перевод         │")
	fmt.Println("├──────────────────┼────────────────────────────────────────────────────────┤")
	fmt.Println("│ GPT-4o           │ • Универсальные задачи                                 │")
	fmt.Println("│ (Средняя)        │ • Баланс качества и стоимости                          │")
	fmt.Println("│                  │ • Сложный анализ текста                                │")
	fmt.Println("│                  │ • Генерация контента                                   │")
	fmt.Println("│                  │ ✅ Лучший выбор для: статьи, код, анализ данных        │")
	fmt.Println("├──────────────────┼────────────────────────────────────────────────────────┤")
	fmt.Println("│ GPT-4 Turbo      │ • Сложные логические задачи                            │")
	fmt.Println("│ (Сильная)        │ • Большой контекст (128K токенов)                      │")
	fmt.Println("│                  │ • Анализ больших документов                            │")
	fmt.Println("│                  │ • Критически важные решения                            │")
	fmt.Println("│                  │ ✅ Лучший выбор для: исследования, большие тексты, код │")
	fmt.Println("└──────────────────┴────────────────────────────────────────────────────────┘")

	fmt.Println("\n📝 КЛЮЧЕВЫЕ ВЫВОДЫ:\n")

	fmt.Println("1. Закон убывающей отдачи:")
	utils.PrintInfo("   Переход от слабой к средней модели дает больший прирост качества,")
	utils.PrintInfo("   чем переход от средней к сильной (при учете стоимости)")

	fmt.Println("\n2. Правило 80/20:")
	utils.PrintSuccess("   Для 80% задач достаточно GPT-4o-mini или GPT-4o")
	utils.PrintInfo("   Сильные модели нужны только для 20% сложных задач")

	fmt.Println("\n3. Оптимизация затрат:")
	utils.PrintInfo("   Используйте слабую модель для фильтрации/предобработки")
	utils.PrintInfo("   Затем сильную модель только для сложных случаев")

	fmt.Println("\n4. Компромисс скорость/качество:")
	utils.PrintInfo("   Слабая модель: до 10x быстрее, но может пропустить детали")
	utils.PrintInfo("   Сильная модель: медленнее, но надежнее для критичных задач")

	fmt.Println("\n💡 ПРАКТИЧЕСКИЕ СОВЕТЫ:\n")

	fmt.Println("   • Начните с GPT-4o-mini, переходите к более сильным при необходимости")
	fmt.Println("   • Используйте A/B тестирование для оценки реальной разницы")
	fmt.Println("   • Кэшируйте результаты для экономии (особенно для дорогих моделей)")
	fmt.Println("   • Следите за новыми релизами - модели постоянно улучшаются")
	fmt.Println()

	fmt.Println("🔗 ПОЛЕЗНЫЕ ССЫЛКИ:\n")
	fmt.Println("   • OpenAI Pricing: https://openai.com/api/pricing/")
	fmt.Println("   • Model Documentation: https://platform.openai.com/docs/models")
	fmt.Println("   • HuggingFace Leaderboard: https://huggingface.co/spaces/lmsys/chatbot-arena-leaderboard")
	fmt.Println()
}

// Вспомогательные функции

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitWords(s string) []string {
	words := make([]string, 0)
	word := ""

	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' || r == '.' || r == ',' || r == '!' || r == '?' {
			if len(word) > 0 {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}

	if len(word) > 0 {
		words = append(words, word)
	}

	return words
}
