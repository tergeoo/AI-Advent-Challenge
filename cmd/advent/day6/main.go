package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

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
	utils.PrintHeader("Day 6: Первый AI Агент")

	// Описание
	printWelcome()

	// Создаем агента с конфигурацией
	agentConfig := agent.AgentConfig{
		APIKey:      cfg.OpenAIKey,
		Model:       openai.GPT4oMini,
		Temperature: 0.7,
		MaxTokens:   500,
		SystemPrompt: `Ты - полезный AI ассистент. Отвечай кратко и по делу.
Если не знаешь ответа, так и скажи. Будь дружелюбным и профессиональным.`,
	}

	aiAgent := agent.NewAgent(agentConfig)

	utils.PrintSuccess("✓ Агент инициализирован и готов к работе!")
	utils.PrintInfo(fmt.Sprintf("Модель: %s", agentConfig.Model))
	utils.PrintInfo(fmt.Sprintf("Temperature: %.1f", agentConfig.Temperature))
	fmt.Println()

	// Запускаем интерактивный режим
	runInteractiveMode(aiAgent)
}

func printWelcome() {
	utils.PrintSection("🤖", "О АГЕНТЕ")

	fmt.Println("Это простой AI агент с памятью диалога.")
	fmt.Println()
	fmt.Println("Возможности агента:")
	fmt.Println("  • Запоминает контекст разговора")
	fmt.Println("  • Отвечает на вопросы")
	fmt.Println("  • Помогает с задачами")
	fmt.Println("  • Поддерживает команды для управления")
	fmt.Println()
	fmt.Println("Доступные команды:")
	fmt.Println("  /help     - показать справку")
	fmt.Println("  /history  - показать историю диалога")
	fmt.Println("  /clear    - очистить историю")
	fmt.Println("  /stats    - показать статистику")
	fmt.Println("  /exit     - выйти из программы")
	fmt.Println()

	utils.PrintDivider()
}

func runInteractiveMode(aiAgent *agent.Agent) {
	reader := bufio.NewReader(os.Stdin)
	totalTokens := 0
	requestCount := 0

	for {
		// Приглашение для ввода
		fmt.Print("\n💬 Вы: ")

		// Читаем ввод пользователя
		input, err := reader.ReadString('\n')
		if err != nil {
			utils.PrintError(fmt.Sprintf("Ошибка чтения ввода: %v", err))
			continue
		}

		// Очищаем ввод
		input = strings.TrimSpace(input)

		// Пропускаем пустые строки
		if input == "" {
			continue
		}

		// Обрабатываем команды
		if strings.HasPrefix(input, "/") {
			handleCommand(input, aiAgent, totalTokens, requestCount)
			continue
		}

		// Отправляем запрос агенту
		fmt.Print("\n🤖 Агент: ")

		response, err := aiAgent.Ask(input)
		if err != nil {
			utils.PrintError(fmt.Sprintf("\nОшибка: %v", err))
			continue
		}

		// Выводим ответ
		fmt.Printf("%s\n", response.Content)

		// Обновляем статистику
		totalTokens += response.TokensUsed
		requestCount++

		// Показываем метаданные (опционально, можно закомментировать)
		fmt.Printf("\n")
		utils.PrintKeyValue("├─ Токены", fmt.Sprintf("%d", response.TokensUsed))
		utils.PrintKeyValue("├─ Время", response.ExecutionTime.String())
		utils.PrintKeyValue("└─ Сообщений в истории", fmt.Sprintf("%d", aiAgent.GetHistorySize()))
	}
}

func handleCommand(cmd string, aiAgent *agent.Agent, totalTokens, requestCount int) {
	cmd = strings.ToLower(cmd)

	switch cmd {
	case "/help":
		printHelp()

	case "/history":
		printHistory(aiAgent)

	case "/clear":
		aiAgent.ClearHistory()
		utils.PrintSuccess("\n✓ История диалога очищена")

	case "/stats":
		printStats(aiAgent, totalTokens, requestCount)

	case "/exit", "/quit":
		fmt.Println()
		utils.PrintSuccess("До свидания! 👋")
		os.Exit(0)

	default:
		utils.PrintError(fmt.Sprintf("\n❌ Неизвестная команда: %s", cmd))
		fmt.Println("Используйте /help для списка доступных команд")
	}
}

func printHelp() {
	fmt.Println()
	utils.PrintSection("📖", "СПРАВКА")

	fmt.Println("Доступные команды:\n")

	commands := []struct {
		cmd  string
		desc string
	}{
		{"/help", "Показать эту справку"},
		{"/history", "Показать историю диалога"},
		{"/clear", "Очистить историю диалога"},
		{"/stats", "Показать статистику использования"},
		{"/exit", "Выйти из программы"},
	}

	for _, c := range commands {
		fmt.Printf("  %-12s - %s\n", c.cmd, c.desc)
	}

	fmt.Println()
	fmt.Println("Просто введите свой вопрос для общения с агентом.")
	utils.PrintDivider()
}

func printHistory(aiAgent *agent.Agent) {
	history := aiAgent.GetHistory()

	if len(history) == 0 {
		utils.PrintInfo("\n📭 История диалога пуста")
		return
	}

	fmt.Println()
	utils.PrintSection("📜", "ИСТОРИЯ ДИАЛОГА")

	for i, msg := range history {
		var prefix string
		var color string

		if msg.Role == "user" {
			prefix = "💬 Вы"
			color = "\033[36m" // Cyan
		} else {
			prefix = "🤖 Агент"
			color = "\033[32m" // Green
		}

		fmt.Printf("\n%s%s [%s]:\033[0m\n", color, prefix, msg.Timestamp.Format("15:04:05"))

		// Обрезаем длинные сообщения
		content := msg.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}

		fmt.Printf("%s\n", content)

		// Разделитель между сообщениями (кроме последнего)
		if i < len(history)-1 {
			fmt.Println(strings.Repeat("─", 60))
		}
	}

	utils.PrintDivider()
}

func printStats(aiAgent *agent.Agent, totalTokens, requestCount int) {
	fmt.Println()
	utils.PrintSection("📊", "СТАТИСТИКА")

	historySize := aiAgent.GetHistorySize()
	estimatedTokens := aiAgent.GetTotalTokens()

	fmt.Println()
	utils.PrintKeyValue("Запросов выполнено", fmt.Sprintf("%d", requestCount))
	utils.PrintKeyValue("Сообщений в истории", fmt.Sprintf("%d", historySize))
	utils.PrintKeyValue("Токенов использовано", fmt.Sprintf("%d", totalTokens))
	utils.PrintKeyValue("Токенов в памяти (оценка)", fmt.Sprintf("%d", estimatedTokens))

	if requestCount > 0 {
		avgTokens := float64(totalTokens) / float64(requestCount)
		utils.PrintKeyValue("Среднее токенов/запрос", fmt.Sprintf("%.1f", avgTokens))
	}

	// Примерная стоимость для GPT-4o-mini
	cost := float64(totalTokens) / 1_000_000 * 0.40 // Примерная средняя цена
	utils.PrintKeyValue("Примерная стоимость", fmt.Sprintf("$%.6f", cost))

	fmt.Println()
	utils.PrintDivider()
}
