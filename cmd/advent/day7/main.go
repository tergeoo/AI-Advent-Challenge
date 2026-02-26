package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/agent"
	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/config"
	"github.com/georgijter-grigoranc/ai-advent-challenge/pkg/utils"
	openai "github.com/sashabaranov/go-openai"
)

const (
	// Путь к файлу сохранения
	defaultSaveFile = ".agent_history.json"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Заголовок
	utils.PrintHeader("Day 7: Агент с сохранением контекста")

	// Описание
	printWelcome()

	// Определяем путь к файлу сохранения
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Предупреждение: не удалось определить домашнюю директорию: %v", err)
		homeDir = "."
	}
	saveFilePath := filepath.Join(homeDir, defaultSaveFile)

	// Создаем агента
	agentConfig := agent.AgentConfig{
		APIKey:      cfg.OpenAIKey,
		Model:       openai.GPT4oMini,
		Temperature: 0.7,
		MaxTokens:   500,
		SystemPrompt: `Ты - полезный AI ассистент с долговременной памятью.
Ты помнишь все предыдущие разговоры даже после перезапуска.
Отвечай кратко и по делу. Будь дружелюбным и профессиональным.`,
	}

	aiAgent := agent.NewAgent(agentConfig)

	// Пытаемся загрузить сохраненную историю
	err = aiAgent.LoadHistory(saveFilePath)
	if err != nil {
		utils.PrintError(fmt.Sprintf("Ошибка загрузки истории: %v", err))
	} else {
		historySize := aiAgent.GetHistorySize()
		if historySize > 0 {
			utils.PrintSuccess(fmt.Sprintf("✓ Загружена история: %d сообщений", historySize))
			utils.PrintInfo("Агент помнит предыдущие разговоры!")
		} else {
			utils.PrintInfo("Это первый запуск. История пуста.")
		}
	}

	utils.PrintInfo(fmt.Sprintf("Файл сохранения: %s", saveFilePath))
	utils.PrintInfo(fmt.Sprintf("Модель: %s", agentConfig.Model))
	fmt.Println()

	// Запускаем интерактивный режим
	runInteractiveMode(aiAgent, saveFilePath)
}

func printWelcome() {
	utils.PrintSection("🤖", "АГЕНТ С ПАМЯТЬЮ")

	fmt.Println("Это AI агент с долговременной памятью диалога.")
	fmt.Println()
	fmt.Println("Ключевые особенности:")
	fmt.Println("  • Автоматическое сохранение после каждого сообщения")
	fmt.Println("  • Загрузка истории при запуске")
	fmt.Println("  • Продолжение диалога после перезапуска")
	fmt.Println("  • Контекст сохраняется между сессиями")
	fmt.Println()
	fmt.Println("Доступные команды:")
	fmt.Println("  /help     - показать справку")
	fmt.Println("  /history  - показать историю диалога")
	fmt.Println("  /save     - принудительно сохранить историю")
	fmt.Println("  /clear    - очистить историю (с подтверждением)")
	fmt.Println("  /stats    - показать статистику")
	fmt.Println("  /exit     - выйти из программы")
	fmt.Println()
	utils.PrintInfo("💡 Совет: Попробуйте начать диалог, затем перезапустите программу")
	fmt.Println()

	utils.PrintDivider()
}

func runInteractiveMode(aiAgent *agent.Agent, saveFilePath string) {
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
			if handleCommand(input, aiAgent, saveFilePath, totalTokens, requestCount) {
				return // Выход из программы
			}
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

		// Автоматически сохраняем историю после каждого ответа
		err = aiAgent.AutoSave(saveFilePath)
		if err != nil {
			utils.PrintError(fmt.Sprintf("\n⚠️  Ошибка автосохранения: %v", err))
		}

		// Обновляем статистику
		totalTokens += response.TokensUsed
		requestCount++

		// Показываем метаданные
		fmt.Printf("\n")
		utils.PrintKeyValue("├─ Токены", fmt.Sprintf("%d", response.TokensUsed))
		utils.PrintKeyValue("├─ Время", response.ExecutionTime.String())
		utils.PrintKeyValue("├─ Сообщений в истории", fmt.Sprintf("%d", aiAgent.GetHistorySize()))
		utils.PrintKeyValue("└─ Автосохранение", "✓")
	}
}

func handleCommand(cmd string, aiAgent *agent.Agent, saveFilePath string, totalTokens, requestCount int) bool {
	cmd = strings.ToLower(cmd)

	switch cmd {
	case "/help":
		printHelp()
		return false

	case "/history":
		printHistory(aiAgent)
		return false

	case "/save":
		err := aiAgent.SaveHistory(saveFilePath)
		if err != nil {
			utils.PrintError(fmt.Sprintf("\n❌ Ошибка сохранения: %v", err))
		} else {
			utils.PrintSuccess(fmt.Sprintf("\n✓ История сохранена в %s", saveFilePath))
		}
		return false

	case "/clear":
		if confirmClear() {
			aiAgent.ClearHistory()
			// Сохраняем пустую историю
			err := aiAgent.SaveHistory(saveFilePath)
			if err != nil {
				utils.PrintError(fmt.Sprintf("\n⚠️  История очищена, но не сохранена: %v", err))
			} else {
				utils.PrintSuccess("\n✓ История очищена и сохранена")
			}
		} else {
			utils.PrintInfo("\nОтменено")
		}
		return false

	case "/stats":
		printStats(aiAgent, saveFilePath, totalTokens, requestCount)
		return false

	case "/exit", "/quit":
		fmt.Println()
		// Финальное сохранение
		err := aiAgent.SaveHistory(saveFilePath)
		if err != nil {
			utils.PrintError(fmt.Sprintf("⚠️  Ошибка сохранения перед выходом: %v", err))
		}
		utils.PrintSuccess("История сохранена. До свидания! 👋")
		return true

	default:
		utils.PrintError(fmt.Sprintf("\n❌ Неизвестная команда: %s", cmd))
		fmt.Println("Используйте /help для списка доступных команд")
		return false
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
		{"/save", "Принудительно сохранить историю"},
		{"/clear", "Очистить историю (с подтверждением)"},
		{"/stats", "Показать статистику использования"},
		{"/exit", "Выйти из программы"},
	}

	for _, c := range commands {
		fmt.Printf("  %-12s - %s\n", c.cmd, c.desc)
	}

	fmt.Println()
	fmt.Println("Особенности:")
	fmt.Println("  • История автоматически сохраняется после каждого ответа")
	fmt.Println("  • При перезапуске агент помнит все предыдущие разговоры")
	fmt.Println("  • Используйте /clear для начала нового диалога")
	fmt.Println()
	utils.PrintDivider()
}

func printHistory(aiAgent *agent.Agent) {
	history := aiAgent.GetHistory()

	if len(history) == 0 {
		utils.PrintInfo("\n📭 История диалога пуста")
		return
	}

	fmt.Println()
	utils.PrintSection("📜", fmt.Sprintf("ИСТОРИЯ ДИАЛОГА (%d сообщений)", len(history)))

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

		fmt.Printf("\n%s%s [%s]:\033[0m\n", color, prefix, msg.Timestamp.Format("2006-01-02 15:04:05"))

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

func printStats(aiAgent *agent.Agent, saveFilePath string, totalTokens, requestCount int) {
	fmt.Println()
	utils.PrintSection("📊", "СТАТИСТИКА")

	historySize := aiAgent.GetHistorySize()
	estimatedTokens := aiAgent.GetTotalTokens()

	// Проверяем существование файла сохранения
	fileInfo, err := os.Stat(saveFilePath)
	var fileSize int64
	var lastSaved string
	if err == nil {
		fileSize = fileInfo.Size()
		lastSaved = fileInfo.ModTime().Format("2006-01-02 15:04:05")
	}

	fmt.Println()
	utils.PrintKeyValue("Запросов в этой сессии", fmt.Sprintf("%d", requestCount))
	utils.PrintKeyValue("Сообщений в истории", fmt.Sprintf("%d", historySize))
	utils.PrintKeyValue("Токенов использовано (сессия)", fmt.Sprintf("%d", totalTokens))
	utils.PrintKeyValue("Токенов в памяти (оценка)", fmt.Sprintf("%d", estimatedTokens))

	if requestCount > 0 {
		avgTokens := float64(totalTokens) / float64(requestCount)
		utils.PrintKeyValue("Среднее токенов/запрос", fmt.Sprintf("%.1f", avgTokens))
	}

	// Информация о файле сохранения
	fmt.Println()
	utils.PrintInfo("Файл сохранения:")
	utils.PrintKeyValue("  Путь", saveFilePath)
	if err == nil {
		utils.PrintKeyValue("  Размер", fmt.Sprintf("%.2f KB", float64(fileSize)/1024))
		utils.PrintKeyValue("  Последнее сохранение", lastSaved)
	} else {
		utils.PrintKeyValue("  Статус", "Не создан")
	}

	// Примерная стоимость для GPT-4o-mini
	cost := float64(totalTokens) / 1_000_000 * 0.40
	fmt.Println()
	utils.PrintKeyValue("Примерная стоимость (сессия)", fmt.Sprintf("$%.6f", cost))

	fmt.Println()
	utils.PrintDivider()
}

func confirmClear() bool {
	fmt.Print("\n⚠️  Вы уверены, что хотите очистить историю? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}
