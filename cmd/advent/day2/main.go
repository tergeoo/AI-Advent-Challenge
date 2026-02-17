package main

import (
	"fmt"
	"log"

	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/client"
	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/config"
	"github.com/georgijter-grigoranc/ai-advent-challenge/pkg/utils"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создание клиента
	aiClient := client.NewOpenAIClient(cfg.OpenAIKey)

	// Заголовок
	utils.PrintHeader("Day 2: Сравнение запросов с разным уровнем контроля")

	// Базовый запрос
	basePrompt := "Расскажи про искусственный интеллект"

	// 1. Запрос без ограничений
	runRequestWithoutConstraints(aiClient, basePrompt)

	// 2. Запрос с ограничениями
	runRequestWithConstraints(aiClient)

	// 3. Запрос с жесткими ограничениями (JSON)
	runRequestWithStrictConstraints(aiClient)

	// Сравнение результатов
	printComparison()
}

func runRequestWithoutConstraints(aiClient *client.OpenAIClient, prompt string) {
	utils.PrintSection("📝", "ЗАПРОС 1: БЕЗ ОГРАНИЧЕНИЙ")
	fmt.Printf("Промпт: %s\n\n", prompt)

	resp, err := aiClient.CreateCompletion(client.CompletionRequest{
		Prompt:      prompt,
		Temperature: 0.7,
	})

	if err != nil {
		log.Printf("Ошибка в запросе 1: %v\n", err)
		return
	}

	fmt.Printf("Ответ:\n%s\n\n", resp.Content)
	utils.PrintTokenStats(resp.TotalTokens, resp.PromptTokens, resp.CompletionTokens)
	utils.PrintKeyValue("Модель", resp.Model)
	utils.PrintDivider()
}

func runRequestWithConstraints(aiClient *client.OpenAIClient) {
	utils.PrintSection("📝", "ЗАПРОС 2: С ОГРАНИЧЕНИЯМИ")

	controlledPrompt := `Расскажи про искусственный интеллект.

ФОРМАТ ОТВЕТА:
1. Определение (1 предложение)
2. Основные направления (список из 3 пунктов)
3. Практическое применение (2-3 примера)

ОГРАНИЧЕНИЯ:
- Максимум 150 слов
- Структурированный формат
- Завершить ответ фразой "[КОНЕЦ ОТВЕТА]"`

	fmt.Printf("Промпт:\n%s\n\n", controlledPrompt)

	resp, err := aiClient.CreateCompletion(client.CompletionRequest{
		Prompt:      controlledPrompt,
		MaxTokens:   300,
		Temperature: 0.7,
		Stop:        []string{"[КОНЕЦ ОТВЕТА]"},
	})

	if err != nil {
		log.Printf("Ошибка в запросе 2: %v\n", err)
		return
	}

	fmt.Printf("Ответ:\n%s\n\n", resp.Content)
	utils.PrintTokenStats(resp.TotalTokens, resp.PromptTokens, resp.CompletionTokens)
	utils.PrintKeyValue("Модель", resp.Model)
	utils.PrintKeyValue("Finish reason", resp.FinishReason)
	utils.PrintDivider()
}

func runRequestWithStrictConstraints(aiClient *client.OpenAIClient) {
	utils.PrintSection("📝", "ЗАПРОС 3: С ЖЕСТКИМИ ОГРАНИЧЕНИЯМИ (JSON)")

	strictPrompt := `Расскажи про искусственный интеллект.

СТРОГИЙ ФОРМАТ ОТВЕТА (JSON):
{
  "definition": "краткое определение (1 предложение)",
  "types": ["тип1", "тип2", "тип3"],
  "applications": ["применение1", "применение2"]
}

ТРЕБОВАНИЯ:
- Только валидный JSON
- Без дополнительных пояснений
- Максимум 50 токенов`

	fmt.Printf("Промпт:\n%s\n\n", strictPrompt)

	resp, err := aiClient.CreateCompletion(client.CompletionRequest{
		Prompt:      strictPrompt,
		MaxTokens:   150,
		Temperature: 0.3,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})

	if err != nil {
		log.Printf("Ошибка в запросе 3: %v\n", err)
		return
	}

	fmt.Printf("Ответ:\n%s\n\n", resp.Content)
	utils.PrintTokenStats(resp.TotalTokens, resp.PromptTokens, resp.CompletionTokens)
	utils.PrintKeyValue("Модель", resp.Model)
	utils.PrintKeyValue("Finish reason", resp.FinishReason)
	utils.PrintDivider()
}

func printComparison() {
	utils.PrintSection("📊", "СРАВНЕНИЕ РЕЗУЛЬТАТОВ")

	fmt.Println("1. Без ограничений:")
	utils.PrintSuccess("Получен развернутый, подробный ответ")
	utils.PrintSuccess("Больше токенов использовано")
	utils.PrintError("Формат ответа непредсказуем")
	fmt.Println()

	fmt.Println("2. С ограничениями:")
	utils.PrintSuccess("Структурированный ответ")
	utils.PrintSuccess("Ограниченная длина (MaxTokens)")
	utils.PrintSuccess("Контролируемое завершение (Stop sequence)")
	utils.PrintSuccess("Меньше токенов = экономия $$$")
	fmt.Println()

	fmt.Println("3. С жесткими ограничениями (JSON):")
	utils.PrintSuccess("Валидный JSON формат")
	utils.PrintSuccess("Низкая температура = более предсказуемый результат")
	utils.PrintSuccess("Минимальное количество токенов")
	utils.PrintSuccess("Легко парсится программно")
	fmt.Println()

	utils.PrintInfo("ВЫВОД:")
	fmt.Println("   Уровень контроля влияет на:")
	fmt.Println("   - Формат и структуру ответа")
	fmt.Println("   - Стоимость запроса (токены)")
	fmt.Println("   - Предсказуемость результата")
	fmt.Println("   - Удобство программной обработки")

	utils.PrintDivider()
	utils.PrintSuccess("Задание Day 2 выполнено!")
}
