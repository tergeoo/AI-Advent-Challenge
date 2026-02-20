package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/client"
	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/config"
	"github.com/georgijter-grigoranc/ai-advent-challenge/pkg/utils"
)

// Тип задачи
type TaskType string

const (
	FactualTask    TaskType = "factual"    // Фактическая задача
	CreativeTask   TaskType = "creative"   // Креативная задача
	AnalyticalTask TaskType = "analytical" // Аналитическая задача
)

// Результат теста с определенной температурой
type TemperatureResult struct {
	Temperature float32
	Response    string
	TokensUsed  int
	TimeTaken   time.Duration
}

// Набор результатов для одной задачи
type TaskResults struct {
	TaskType    TaskType
	Prompt      string
	Description string
	Results     []TemperatureResult
}

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создание клиента
	aiClient := client.NewOpenAIClient(cfg.OpenAIKey)

	// Заголовок
	utils.PrintHeader("Day 4: Эксперимент с температурой")

	// Описание эксперимента
	printExperimentDescription()

	// Температуры для тестирования
	temperatures := []float32{0.0, 0.7, 1.2}

	// Хранилище всех результатов
	allResults := make([]TaskResults, 0, 3)

	// 1. Фактическая задача (математика/факты)
	allResults = append(allResults, runFactualTask(aiClient, temperatures))

	// 2. Креативная задача (написание текста)
	allResults = append(allResults, runCreativeTask(aiClient, temperatures))

	// 3. Аналитическая задача
	allResults = append(allResults, runAnalyticalTask(aiClient, temperatures))

	// Сравнение и анализ
	compareResults(allResults)

	// Рекомендации
	printRecommendations()

	utils.PrintDivider()
	utils.PrintSuccess("Задание Day 4 выполнено!")
}

func printExperimentDescription() {
	utils.PrintSection("🧪", "ОПИСАНИЕ ЭКСПЕРИМЕНТА")

	fmt.Println("Temperature (температура) - параметр, контролирующий случайность ответов:")
	fmt.Println()
	fmt.Println("📊 Диапазон значений: 0.0 - 2.0")
	fmt.Println()
	fmt.Println("Что означают значения:")
	fmt.Println("  • 0.0   → Детерминированный, предсказуемый ответ")
	fmt.Println("  • 0.7   → Баланс между точностью и креативностью (по умолчанию)")
	fmt.Println("  • 1.2   → Высокая креативность и разнообразие")
	fmt.Println()
	fmt.Println("Мы протестируем три типа задач:")
	fmt.Println("  1. Фактическая   (математика, факты)")
	fmt.Println("  2. Креативная    (написание историй)")
	fmt.Println("  3. Аналитическая (анализ данных)")
	fmt.Println()

	utils.PrintDivider()
}

// Задача 1: Фактическая (математика)
func runFactualTask(aiClient *client.OpenAIClient, temperatures []float32) TaskResults {
	utils.PrintSection("1️⃣", "ФАКТИЧЕСКАЯ ЗАДАЧА: Математика")

	prompt := `Реши математическую задачу:

У Маши было 15 яблок. Она отдала 1/3 своих яблок Пете,
а затем купила еще 7 яблок. Сколько яблок стало у Маши?

Ответь кратко: только решение и ответ.`

	fmt.Printf("Промпт:\n%s\n\n", prompt)

	results := TaskResults{
		TaskType:    FactualTask,
		Prompt:      prompt,
		Description: "Математическая задача с точным ответом",
		Results:     make([]TemperatureResult, 0, len(temperatures)),
	}

	for _, temp := range temperatures {
		fmt.Printf("🌡️  Temperature = %.1f\n", temp)
		fmt.Println(strings.Repeat("─", 80))

		start := time.Now()
		resp, err := aiClient.CreateCompletion(client.CompletionRequest{
			Prompt:      prompt,
			Temperature: temp,
			MaxTokens:   150,
		})
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("Ошибка: %v\n", err)
			continue
		}

		fmt.Printf("Ответ:\n%s\n\n", resp.Content)
		utils.PrintTokenStats(resp.TotalTokens, resp.PromptTokens, resp.CompletionTokens)
		utils.PrintKeyValue("Время", elapsed.Round(time.Millisecond).String())
		fmt.Println()

		results.Results = append(results.Results, TemperatureResult{
			Temperature: temp,
			Response:    resp.Content,
			TokensUsed:  resp.TotalTokens,
			TimeTaken:   elapsed,
		})
	}

	utils.PrintDivider()
	return results
}

// Задача 2: Креативная (написание текста)
func runCreativeTask(aiClient *client.OpenAIClient, temperatures []float32) TaskResults {
	utils.PrintSection("2️⃣", "КРЕАТИВНАЯ ЗАДАЧА: Написание истории")

	prompt := `Напиши короткую историю (3-4 предложения) о роботе,
который впервые увидел закат.

Используй яркие образы и эмоции.`

	fmt.Printf("Промпт:\n%s\n\n", prompt)

	results := TaskResults{
		TaskType:    CreativeTask,
		Prompt:      prompt,
		Description: "Креативное написание текста",
		Results:     make([]TemperatureResult, 0, len(temperatures)),
	}

	for _, temp := range temperatures {
		fmt.Printf("🌡️  Temperature = %.1f\n", temp)
		fmt.Println(strings.Repeat("─", 80))

		start := time.Now()
		resp, err := aiClient.CreateCompletion(client.CompletionRequest{
			Prompt:      prompt,
			Temperature: temp,
			MaxTokens:   200,
		})
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("Ошибка: %v\n", err)
			continue
		}

		fmt.Printf("Ответ:\n%s\n\n", resp.Content)
		utils.PrintTokenStats(resp.TotalTokens, resp.PromptTokens, resp.CompletionTokens)
		utils.PrintKeyValue("Время", elapsed.Round(time.Millisecond).String())
		fmt.Println()

		results.Results = append(results.Results, TemperatureResult{
			Temperature: temp,
			Response:    resp.Content,
			TokensUsed:  resp.TotalTokens,
			TimeTaken:   elapsed,
		})
	}

	utils.PrintDivider()
	return results
}

// Задача 3: Аналитическая
func runAnalyticalTask(aiClient *client.OpenAIClient, temperatures []float32) TaskResults {
	utils.PrintSection("3️⃣", "АНАЛИТИЧЕСКАЯ ЗАДАЧА: Анализ данных")

	prompt := `Проанализируй следующие данные продаж:
- Январь: 100 единиц
- Февраль: 150 единиц
- Март: 120 единиц

Какой тренд наблюдается? Дай краткую рекомендацию (2-3 предложения).`

	fmt.Printf("Промпт:\n%s\n\n", prompt)

	results := TaskResults{
		TaskType:    AnalyticalTask,
		Prompt:      prompt,
		Description: "Анализ данных и выводы",
		Results:     make([]TemperatureResult, 0, len(temperatures)),
	}

	for _, temp := range temperatures {
		fmt.Printf("🌡️  Temperature = %.1f\n", temp)
		fmt.Println(strings.Repeat("─", 80))

		start := time.Now()
		resp, err := aiClient.CreateCompletion(client.CompletionRequest{
			Prompt:      prompt,
			Temperature: temp,
			MaxTokens:   150,
		})
		elapsed := time.Since(start)

		if err != nil {
			log.Printf("Ошибка: %v\n", err)
			continue
		}

		fmt.Printf("Ответ:\n%s\n\n", resp.Content)
		utils.PrintTokenStats(resp.TotalTokens, resp.PromptTokens, resp.CompletionTokens)
		utils.PrintKeyValue("Время", elapsed.Round(time.Millisecond).String())
		fmt.Println()

		results.Results = append(results.Results, TemperatureResult{
			Temperature: temp,
			Response:    resp.Content,
			TokensUsed:  resp.TotalTokens,
			TimeTaken:   elapsed,
		})
	}

	utils.PrintDivider()
	return results
}

func compareResults(allResults []TaskResults) {
	utils.PrintSection("📊", "СРАВНИТЕЛЬНЫЙ АНАЛИЗ")

	for _, taskResult := range allResults {
		taskName := ""
		emoji := ""

		switch taskResult.TaskType {
		case FactualTask:
			taskName = "Фактическая задача"
			emoji = "1️⃣"
		case CreativeTask:
			taskName = "Креативная задача"
			emoji = "2️⃣"
		case AnalyticalTask:
			taskName = "Аналитическая задача"
			emoji = "3️⃣"
		}

		fmt.Printf("\n%s %s\n", emoji, taskName)
		fmt.Println(strings.Repeat("─", 80))

		// Анализ для каждой температуры
		for _, result := range taskResult.Results {
			fmt.Printf("\n🌡️  Temperature = %.1f:\n", result.Temperature)

			switch taskResult.TaskType {
			case FactualTask:
				analyzeFactualResponse(result)
			case CreativeTask:
				analyzeCreativeResponse(result)
			case AnalyticalTask:
				analyzeAnalyticalResponse(result)
			}
		}

		fmt.Println()
	}

	utils.PrintDivider()
}

func analyzeFactualResponse(result TemperatureResult) {
	// Проверяем наличие правильного ответа (17 яблок)
	hasCorrectAnswer := strings.Contains(result.Response, "17")
	responseLength := len(strings.Split(result.Response, " "))

	if hasCorrectAnswer {
		utils.PrintSuccess("Правильный ответ присутствует")
	} else {
		utils.PrintError("Правильный ответ не найден или неточен")
	}

	utils.PrintKeyValue("  Длина ответа", fmt.Sprintf("%d слов", responseLength))

	// Оценка точности
	if result.Temperature == 0.0 {
		utils.PrintInfo("  Точность: максимальная, ответ предсказуемый")
	} else if result.Temperature == 0.7 {
		utils.PrintInfo("  Точность: высокая, возможны вариации в объяснении")
	} else {
		utils.PrintInfo("  Точность: может варьироваться, возможны отклонения")
	}
}

func analyzeCreativeResponse(result TemperatureResult) {
	responseLength := len(strings.Split(result.Response, " "))
	sentenceCount := strings.Count(result.Response, ".") + strings.Count(result.Response, "!") + strings.Count(result.Response, "?")

	utils.PrintKeyValue("  Длина", fmt.Sprintf("%d слов", responseLength))
	utils.PrintKeyValue("  Предложений", fmt.Sprintf("~%d", sentenceCount))

	// Оценка креативности
	if result.Temperature == 0.0 {
		utils.PrintInfo("  Креативность: низкая, стандартные фразы")
	} else if result.Temperature == 0.7 {
		utils.PrintInfo("  Креативность: средняя, баланс оригинальности и связности")
	} else {
		utils.PrintInfo("  Креативность: высокая, неожиданные образы и метафоры")
	}
}

func analyzeAnalyticalResponse(result TemperatureResult) {
	responseLength := len(strings.Split(result.Response, " "))
	hasTrend := strings.Contains(strings.ToLower(result.Response), "тренд") ||
		strings.Contains(strings.ToLower(result.Response), "рост") ||
		strings.Contains(strings.ToLower(result.Response), "снижение")

	if hasTrend {
		utils.PrintSuccess("Присутствует анализ тренда")
	} else {
		utils.PrintError("Анализ тренда не явный")
	}

	utils.PrintKeyValue("  Длина ответа", fmt.Sprintf("%d слов", responseLength))

	// Оценка аналитичности
	if result.Temperature == 0.0 {
		utils.PrintInfo("  Аналитика: прямолинейная, факты без вариаций")
	} else if result.Temperature == 0.7 {
		utils.PrintInfo("  Аналитика: сбалансированная с разными точками зрения")
	} else {
		utils.PrintInfo("  Аналитика: разнообразные интерпретации")
	}
}

func printRecommendations() {
	utils.PrintSection("🎯", "РЕКОМЕНДАЦИИ ПО ИСПОЛЬЗОВАНИЮ")

	fmt.Println("\n┌─────────────────┬──────────────────────────────────────────────────────┐")
	fmt.Println("│ Temperature     │ Рекомендуемые задачи                                 │")
	fmt.Println("├─────────────────┼──────────────────────────────────────────────────────┤")
	fmt.Println("│ 0.0 - 0.3       │ • Математические вычисления                          │")
	fmt.Println("│ (Низкая)        │ • Извлечение фактов из текста                        │")
	fmt.Println("│                 │ • Классификация по четким правилам                   │")
	fmt.Println("│                 │ • Генерация кода с точным синтаксисом                │")
	fmt.Println("│                 │ • Перевод текста (точность важнее вариативности)     │")
	fmt.Println("├─────────────────┼──────────────────────────────────────────────────────┤")
	fmt.Println("│ 0.4 - 0.8       │ • Написание статей и документации                    │")
	fmt.Println("│ (Средняя)       │ • Ответы на вопросы клиентов                         │")
	fmt.Println("│                 │ • Анализ данных и рекомендации                       │")
	fmt.Println("│                 │ • Резюмирование текста                               │")
	fmt.Println("│                 │ • Общий чат-бот (баланс точности и естественности)   │")
	fmt.Println("├─────────────────┼──────────────────────────────────────────────────────┤")
	fmt.Println("│ 0.9 - 1.5       │ • Креативное написание (истории, стихи)              │")
	fmt.Println("│ (Высокая)       │ • Брейнсторминг идей                                 │")
	fmt.Println("│                 │ • Генерация названий и слоганов                      │")
	fmt.Println("│                 │ • Создание разнообразного контента                   │")
	fmt.Println("│                 │ • Экспериментальные проекты                          │")
	fmt.Println("├─────────────────┼──────────────────────────────────────────────────────┤")
	fmt.Println("│ 1.6 - 2.0       │ • Экстремальная креативность                         │")
	fmt.Println("│ (Очень высокая) │ • Абстрактное искусство                              │")
	fmt.Println("│                 │ ⚠️  Внимание: может быть непредсказуемо и нелогично   │")
	fmt.Println("└─────────────────┴──────────────────────────────────────────────────────┘")

	fmt.Println("\n📝 КЛЮЧЕВЫЕ ВЫВОДЫ:\n")

	fmt.Println("1. Для фактических задач:")
	utils.PrintSuccess("   Используйте низкую температуру (0.0-0.3)")
	utils.PrintInfo("   Гарантирует точность и повторяемость результатов")

	fmt.Println("\n2. Для креативных задач:")
	utils.PrintSuccess("   Используйте высокую температуру (0.9-1.5)")
	utils.PrintInfo("   Обеспечивает разнообразие и оригинальность")

	fmt.Println("\n3. Для большинства задач:")
	utils.PrintSuccess("   Используйте среднюю температуру (0.5-0.8)")
	utils.PrintInfo("   Оптимальный баланс между точностью и естественностью")

	fmt.Println("\n4. Температура и повторяемость:")
	utils.PrintError("   При temperature > 0: каждый запрос даст разный ответ")
	utils.PrintSuccess("   При temperature = 0: ответы будут одинаковыми")

	fmt.Println("\n5. Влияние на стоимость:")
	utils.PrintInfo("   Температура НЕ влияет на количество токенов напрямую")
	utils.PrintInfo("   Но высокая температура может генерировать более длинные ответы")

	fmt.Println("\n💡 ПРАКТИЧЕСКИЙ СОВЕТ:\n")
	fmt.Println("   Начните с temperature = 0.7 (значение по умолчанию)")
	fmt.Println("   Затем:")
	fmt.Println("   • Уменьшите, если нужна большая точность")
	fmt.Println("   • Увеличьте, если нужно больше креативности")
	fmt.Println()
}
