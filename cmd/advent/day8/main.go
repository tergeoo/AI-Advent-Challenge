package main

import (
	"fmt"
	"log"

	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/agent"
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

	// Заголовок
	utils.PrintHeader("Day 8: Работа с токенами")

	// Описание
	printIntro()

	// Демонстрация различных сценариев
	fmt.Println("Выберите сценарий для демонстрации:\n")
	fmt.Println("1. Короткий диалог (отслеживание токенов)")
	fmt.Println("2. Длинный диалог (рост стоимости)")
	fmt.Println("3. Переполнение контекста (демонстрация проблемы)")
	fmt.Println("4. Все сценарии подряд")
	fmt.Println()

	fmt.Print("Выбор (1-4): ")
	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		runShortDialogScenario(cfg.OpenAIKey)
	case 2:
		runLongDialogScenario(cfg.OpenAIKey)
	case 3:
		runOverflowScenario(cfg.OpenAIKey)
	case 4:
		runShortDialogScenario(cfg.OpenAIKey)
		fmt.Println("\n" + utils.Repeat("=", 80) + "\n")
		runLongDialogScenario(cfg.OpenAIKey)
		fmt.Println("\n" + utils.Repeat("=", 80) + "\n")
		runOverflowScenario(cfg.OpenAIKey)
	default:
		fmt.Println("Неверный выбор. Запуск всех сценариев...")
		runShortDialogScenario(cfg.OpenAIKey)
		fmt.Println("\n" + utils.Repeat("=", 80) + "\n")
		runLongDialogScenario(cfg.OpenAIKey)
		fmt.Println("\n" + utils.Repeat("=", 80) + "\n")
		runOverflowScenario(cfg.OpenAIKey)
	}

	// Итоговые выводы
	printConclusions()

	utils.PrintDivider()
	utils.PrintSuccess("Задание Day 8 выполнено!")
}

func printIntro() {
	utils.PrintSection("📊", "О ТОКЕНАХ")

	fmt.Println("Токены - это базовые единицы текста для LLM:")
	fmt.Println("  • 1 токен ≈ 4 символа (английский)")
	fmt.Println("  • 1 токен ≈ 2-3 символа (русский)")
	fmt.Println("  • Стоимость = (input токены * цена_input) + (output токены * цена_output)")
	fmt.Println()
	fmt.Println("Лимиты контекста (примеры):")
	fmt.Println("  • GPT-4o-mini:   128,000 токенов")
	fmt.Println("  • GPT-4o:        128,000 токенов")
	fmt.Println("  • GPT-4:         8,192 токенов")
	fmt.Println("  • GPT-3.5-turbo: 16,385 токенов")
	fmt.Println()
	fmt.Println("Что произойдет при превышении лимита:")
	fmt.Println("  ❌ API вернет ошибку")
	fmt.Println("  ❌ Запрос не будет обработан")
	fmt.Println("  ❌ Старые сообщения нужно удалять вручную")
	fmt.Println()

	utils.PrintDivider()
}

func runShortDialogScenario(apiKey string) {
	utils.PrintSection("1️⃣", "СЦЕНАРИЙ 1: Короткий диалог")

	fmt.Println("Демонстрация: отслеживание токенов в коротком диалоге\n")

	// Создаем агента
	agentConfig := agent.AgentConfig{
		APIKey:       apiKey,
		Model:        openai.GPT4oMini,
		Temperature:  0.7,
		MaxTokens:    100,
		SystemPrompt: "Ты - краткий помощник. Отвечай максимально кратко.",
	}

	aiAgent := agent.NewAgent(agentConfig)

	// Получаем информацию о модели
	modelLimit := agent.GetModelLimit(agentConfig.Model)
	inputPrice, outputPrice := agent.GetModelPricing(agentConfig.Model)

	// Создаем трекер токенов
	tokenStats := agent.NewTokenStats(modelLimit, inputPrice, outputPrice)

	utils.PrintInfo(fmt.Sprintf("Модель: %s", agentConfig.Model))
	utils.PrintInfo(fmt.Sprintf("Лимит контекста: %d токенов", modelLimit))
	utils.PrintInfo(fmt.Sprintf("Цена: $%.3f/$%.3f per 1M tokens", inputPrice, outputPrice))
	fmt.Println()

	// Короткий диалог
	messages := []string{
		"Привет!",
		"Сколько будет 2+2?",
		"Спасибо!",
	}

	for i, msg := range messages {
		fmt.Printf("\n💬 Вы: %s\n", msg)

		resp, err := aiAgent.Ask(msg)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Ошибка: %v", err))
			continue
		}

		fmt.Printf("🤖 Агент: %s\n", resp.Content)

		// Обновляем статистику
		tokenStats.AddRequest(resp.PromptTokens, resp.CompletionTokens)
		tokenStats.UpdateContextSize(aiAgent.GetTotalTokens())

		// Показываем детальную статистику
		fmt.Println()
		printDetailedStats(tokenStats, resp, i+1)
	}

	// Итоговая статистика
	fmt.Println()
	printFinalStats(tokenStats)
}

func runLongDialogScenario(apiKey string) {
	utils.PrintSection("2️⃣", "СЦЕНАРИЙ 2: Длинный диалог (рост стоимости)")

	fmt.Println("Демонстрация: как растут токены и стоимость по мере диалога\n")

	// Создаем агента
	agentConfig := agent.AgentConfig{
		APIKey:       apiKey,
		Model:        openai.GPT4oMini,
		Temperature:  0.7,
		MaxTokens:    200,
		SystemPrompt: "Ты - подробный помощник. Давай развернутые ответы.",
	}

	aiAgent := agent.NewAgent(agentConfig)

	// Получаем информацию о модели
	modelLimit := agent.GetModelLimit(agentConfig.Model)
	inputPrice, outputPrice := agent.GetModelPricing(agentConfig.Model)

	// Создаем трекер токенов
	tokenStats := agent.NewTokenStats(modelLimit, inputPrice, outputPrice)

	utils.PrintInfo(fmt.Sprintf("Модель: %s", agentConfig.Model))
	utils.PrintInfo(fmt.Sprintf("Лимит контекста: %d токенов", modelLimit))
	fmt.Println()

	// Длинный диалог - 10 сообщений
	messages := []string{
		"Расскажи про язык программирования Go",
		"Какие у него преимущества?",
		"А какие недостатки?",
		"Для каких проектов он подходит?",
		"Сравни Go с Python",
		"А с Rust?",
		"Какие крупные компании используют Go?",
		"Какие фреймворки популярны в Go?",
		"Как начать изучать Go?",
		"Посоветуй хорошие книги по Go",
	}

	// Показываем прогресс каждые 2 сообщения
	for i, msg := range messages {
		fmt.Printf("\n💬 Вы (#%d): %s\n", i+1, msg)

		resp, err := aiAgent.Ask(msg)
		if err != nil {
			utils.PrintError(fmt.Sprintf("Ошибка: %v", err))
			continue
		}

		// Обрезаем длинный ответ для отображения
		shortResp := resp.Content
		if len(shortResp) > 150 {
			shortResp = shortResp[:150] + "..."
		}
		fmt.Printf("🤖 Агент: %s\n", shortResp)

		// Обновляем статистику
		tokenStats.AddRequest(resp.PromptTokens, resp.CompletionTokens)
		tokenStats.UpdateContextSize(aiAgent.GetTotalTokens())

		// Показываем прогресс
		if (i+1)%2 == 0 || i == len(messages)-1 {
			fmt.Println()
			printProgressStats(tokenStats, i+1)
		}
	}

	// Итоговая статистика
	fmt.Println()
	printFinalStats(tokenStats)

	// Показываем динамику роста
	printGrowthAnalysis(tokenStats)
}

func runOverflowScenario(apiKey string) {
	utils.PrintSection("3️⃣", "СЦЕНАРИЙ 3: Переполнение контекста")

	fmt.Println("Демонстрация: что происходит при превышении лимита\n")
	fmt.Println("⚠️  Для демонстрации используем GPT-4 с маленьким контекстом (8K токенов)\n")

	// Используем GPT-4 с маленьким контекстом для демонстрации
	agentConfig := agent.AgentConfig{
		APIKey:      apiKey,
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   500,
		SystemPrompt: `Ты - очень подробный помощник.
Давай максимально развернутые и детальные ответы с примерами.`,
	}

	aiAgent := agent.NewAgent(agentConfig)

	// Получаем информацию о модели
	modelLimit := agent.GetModelLimit(agentConfig.Model)
	inputPrice, outputPrice := agent.GetModelPricing(agentConfig.Model)

	// Создаем трекер токенов
	tokenStats := agent.NewTokenStats(modelLimit, inputPrice, outputPrice)

	utils.PrintInfo(fmt.Sprintf("Модель: %s", agentConfig.Model))
	utils.PrintInfo(fmt.Sprintf("Лимит контекста: %d токенов (МАЛЕНЬКИЙ!)", modelLimit))
	utils.PrintInfo(fmt.Sprintf("Цена: $%.2f/$%.2f per 1M tokens", inputPrice, outputPrice))
	fmt.Println()

	// Генерируем много длинных сообщений
	fmt.Println("Начинаем отправлять много длинных сообщений...\n")

	for i := 1; i <= 20; i++ {
		msg := fmt.Sprintf(`Расскажи подробно о теме номер %d:
История развития языков программирования.
Включи примеры кода, исторический контекст,
влияние на индустрию и будущие перспективы.`, i)

		fmt.Printf("💬 Запрос #%d (длинный запрос про программирование)\n", i)

		resp, err := aiAgent.Ask(msg)

		// Обновляем статистику перед проверкой ошибки
		if resp != nil {
			tokenStats.AddRequest(resp.PromptTokens, resp.CompletionTokens)
			tokenStats.UpdateContextSize(aiAgent.GetTotalTokens())
		}

		if err != nil {
			fmt.Println()
			utils.PrintError(fmt.Sprintf("❌ ОШИБКА: %v", err))
			fmt.Println()
			utils.PrintInfo("🔍 Анализ ошибки:")
			utils.PrintInfo(fmt.Sprintf("  • Контекст: %d токенов", tokenStats.CurrentContextTokens))
			utils.PrintInfo(fmt.Sprintf("  • Лимит: %d токенов", tokenStats.MaxContextTokens))
			utils.PrintInfo(fmt.Sprintf("  • Превышение: %d токенов",
				tokenStats.CurrentContextTokens-tokenStats.MaxContextTokens))
			fmt.Println()
			utils.PrintError("💥 Контекст переполнен! API отклонил запрос.")
			fmt.Println()
			break
		}

		// Короткий ответ для отображения
		shortResp := resp.Content
		if len(shortResp) > 100 {
			shortResp = shortResp[:100] + "..."
		}
		fmt.Printf("🤖 Ответ: %s\n", shortResp)

		// Показываем прогресс
		fmt.Println()
		contextBar := tokenStats.FormatContextBar(50)
		fmt.Printf("Контекст: %s\n", contextBar)
		fmt.Printf("Токенов: %d / %d\n", tokenStats.CurrentContextTokens, tokenStats.MaxContextTokens)

		warning := tokenStats.GetWarningMessage()
		if warning != "" {
			fmt.Println(warning)
		}

		fmt.Println()

		// Проверяем, не близки ли мы к лимиту
		if tokenStats.IsOverLimit() {
			utils.PrintError("Достигнут лимит контекста!")
			break
		}
	}

	// Итоговая статистика
	fmt.Println()
	printFinalStats(tokenStats)

	// Рекомендации
	fmt.Println()
	utils.PrintSection("💡", "РЕШЕНИЯ ПРОБЛЕМЫ")
	fmt.Println()
	fmt.Println("1. Обрезка истории:")
	utils.PrintInfo("   Удалять старые сообщения, сохраняя только последние N")
	fmt.Println()
	fmt.Println("2. Суммаризация:")
	utils.PrintInfo("   Сжимать старые сообщения в краткое резюме")
	fmt.Println()
	fmt.Println("3. Выбор модели с большим контекстом:")
	utils.PrintInfo("   GPT-4o-mini: 128K токенов (vs GPT-4: 8K)")
	fmt.Println()
	fmt.Println("4. Разделение диалога:")
	utils.PrintInfo("   Создавать новую сессию при достижении лимита")
	fmt.Println()
}

func printDetailedStats(stats *agent.TokenStats, resp *agent.Response, requestNum int) {
	fmt.Printf("├─ Запрос #%d:\n", requestNum)
	fmt.Printf("│  ├─ Токены запроса: %d\n", resp.PromptTokens)
	fmt.Printf("│  ├─ Токены ответа: %d\n", resp.CompletionTokens)
	fmt.Printf("│  ├─ Всего токенов: %d\n", resp.TokensUsed)
	fmt.Printf("│  ├─ Время: %s\n", resp.ExecutionTime)
	fmt.Printf("│  └─ Стоимость: $%.6f\n",
		float64(resp.PromptTokens)/1_000_000*stats.InputPrice+
			float64(resp.CompletionTokens)/1_000_000*stats.OutputPrice)
	fmt.Printf("│\n")
	fmt.Printf("└─ Накопительно:\n")
	fmt.Printf("   ├─ Всего запросов: %d\n", stats.TotalRequests)
	fmt.Printf("   ├─ Всего токенов: %d\n", stats.TotalTokens)
	fmt.Printf("   ├─ Контекст: %d токенов\n", stats.CurrentContextTokens)
	fmt.Printf("   └─ Общая стоимость: $%.6f\n", stats.TotalCost)
}

func printProgressStats(stats *agent.TokenStats, requestNum int) {
	fmt.Printf("📊 Прогресс после %d запросов:\n", requestNum)
	fmt.Printf("├─ Токенов использовано: %d\n", stats.TotalTokens)
	fmt.Printf("├─ Контекст: %d / %d токенов\n",
		stats.CurrentContextTokens, stats.MaxContextTokens)

	contextBar := stats.FormatContextBar(40)
	fmt.Printf("├─ %s\n", contextBar)
	fmt.Printf("├─ Стоимость: $%.6f\n", stats.TotalCost)
	fmt.Printf("└─ Средняя стоимость/запрос: $%.6f\n", stats.GetAverageCostPerRequest())

	warning := stats.GetWarningMessage()
	if warning != "" {
		fmt.Printf("\n%s\n", warning)
	}
}

func printFinalStats(stats *agent.TokenStats) {
	utils.PrintSection("📈", "ИТОГОВАЯ СТАТИСТИКА")
	fmt.Println()

	fmt.Printf("Всего запросов: %d\n", stats.TotalRequests)
	fmt.Printf("Всего токенов: %d\n", stats.TotalTokens)
	fmt.Printf("  ├─ Input:  %d токенов\n", stats.TotalPromptTokens)
	fmt.Printf("  └─ Output: %d токенов\n", stats.TotalCompletionTokens)
	fmt.Println()

	fmt.Printf("Контекст: %d / %d токенов\n",
		stats.CurrentContextTokens, stats.MaxContextTokens)
	contextBar := stats.FormatContextBar(50)
	fmt.Printf("%s\n", contextBar)
	fmt.Println()

	fmt.Printf("Стоимость:\n")
	fmt.Printf("  ├─ Всего: $%.6f\n", stats.TotalCost)
	fmt.Printf("  ├─ Средняя на запрос: $%.6f\n", stats.GetAverageCostPerRequest())
	fmt.Printf("  └─ На 1000 запросов: $%.2f\n", stats.GetAverageCostPerRequest()*1000)
	fmt.Println()

	fmt.Printf("Средние значения:\n")
	fmt.Printf("  ├─ Токенов на запрос: %.1f\n", stats.GetAverageTokensPerRequest())
	fmt.Printf("  └─ Осталось токенов: %d\n", stats.GetRemainingTokens())
}

func printGrowthAnalysis(stats *agent.TokenStats) {
	utils.PrintSection("📊", "АНАЛИЗ РОСТА")
	fmt.Println()

	avgPerRequest := stats.GetAverageTokensPerRequest()

	fmt.Println("Динамика роста токенов:")
	fmt.Printf("  • Средний рост: %.1f токенов/запрос\n", avgPerRequest)
	fmt.Println()

	// Прогноз
	remaining := stats.GetRemainingTokens()
	estimatedRequests := int(float64(remaining) / avgPerRequest)

	fmt.Println("Прогноз:")
	if estimatedRequests > 0 {
		utils.PrintInfo(fmt.Sprintf("  • Можно сделать еще ~%d запросов", estimatedRequests))
		utils.PrintInfo(fmt.Sprintf("  • До переполнения: %d токенов", remaining))
	} else {
		utils.PrintError("  • Контекст близок к переполнению!")
		utils.PrintInfo("  • Требуется очистка истории")
	}
	fmt.Println()

	// Экстраполяция стоимости
	fmt.Println("Экстраполяция стоимости:")
	fmt.Printf("  • При 100 запросах: $%.4f\n", stats.GetAverageCostPerRequest()*100)
	fmt.Printf("  • При 1,000 запросах: $%.2f\n", stats.GetAverageCostPerRequest()*1000)
	fmt.Printf("  • При 10,000 запросах: $%.2f\n", stats.GetAverageCostPerRequest()*10000)
}

func printConclusions() {
	fmt.Println()
	utils.PrintSection("🎓", "ВЫВОДЫ")
	fmt.Println()

	fmt.Println("1. Токены растут линейно с количеством сообщений:")
	utils.PrintInfo("   Каждое новое сообщение добавляет в контекст")
	fmt.Println()

	fmt.Println("2. Стоимость растет пропорционально токенам:")
	utils.PrintInfo("   Длинные диалоги = больше токенов = выше стоимость")
	fmt.Println()

	fmt.Println("3. Переполнение контекста - реальная проблема:")
	utils.PrintError("   API отклоняет запросы при превышении лимита")
	fmt.Println()

	fmt.Println("4. Необходимо управление контекстом:")
	utils.PrintSuccess("   Обрезка, суммаризация, или разделение диалогов")
	fmt.Println()

	fmt.Println("5. Выбор модели влияет на лимиты:")
	utils.PrintInfo("   GPT-4o-mini (128K) vs GPT-4 (8K) - разница в 16x!")
	fmt.Println()
}
