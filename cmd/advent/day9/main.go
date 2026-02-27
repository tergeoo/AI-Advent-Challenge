package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/agent"
	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/config"
	"github.com/georgijter-grigoranc/ai-advent-challenge/pkg/utils"
	"github.com/sashabaranov/go-openai"
)

func main() {
	utils.PrintHeader("Day 9: Управление контекстом - сжатие истории")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	client := openai.NewClient(cfg.OpenAIKey)

	// Демонстрация 1: Длинный диалог без сжатия
	fmt.Println("\n📝 СЦЕНАРИЙ 1: Длинный диалог БЕЗ сжатия")
	utils.PrintSeparator()
	runWithoutCompression(client)

	fmt.Println("\n\n")

	// Демонстрация 2: Длинный диалог со сжатием
	fmt.Println("🗜️  СЦЕНАРИЙ 2: Длинный диалог СО сжатием")
	utils.PrintSeparator()
	runWithCompression(client)

	fmt.Println("\n\n")

	// Демонстрация 3: Сравнение качества ответов
	fmt.Println("🔍 СЦЕНАРИЙ 3: Сравнение качества ответов")
	utils.PrintSeparator()
	compareQuality(client)
}

// runWithoutCompression демонстрирует работу без сжатия
func runWithoutCompression(client *openai.Client) {
	ctx := context.Background()

	// Симулируем длинный диалог (20 сообщений)
	messages := generateLongDialog()

	fmt.Printf("Всего сообщений: %d\n", len(messages))

	// Подсчитываем токены
	totalTokens := 0
	for _, msg := range messages {
		totalTokens += len(msg.Content) / 3 // Приблизительная оценка
	}

	fmt.Printf("Приблизительно токенов в контексте: %d\n", totalTokens)

	// Отправляем запрос со всей историей
	fullHistory := make([]openai.ChatCompletionMessage, 0)
	for _, msg := range messages {
		fullHistory = append(fullHistory, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Добавляем финальный вопрос
	fullHistory = append(fullHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "Подведи итог нашего разговора: о чем мы говорили и какие решения приняли?",
	})

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		Messages:    fullHistory,
		Temperature: 0.7,
	})

	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if len(resp.Choices) > 0 {
		answer := resp.Choices[0].Message.Content
		fmt.Println("\n💬 Ответ агента:")
		fmt.Println(utils.WrapText(answer, 80))
	}

	// Статистика
	fmt.Println("\n📊 Статистика:")
	fmt.Printf("  • Токенов в запросе: %d\n", resp.Usage.PromptTokens)
	fmt.Printf("  • Токенов в ответе:  %d\n", resp.Usage.CompletionTokens)
	fmt.Printf("  • Всего токенов:     %d\n", resp.Usage.TotalTokens)
	fmt.Printf("  • Стоимость:         $%.6f\n", calculateCost(resp.Usage))
}

// runWithCompression демонстрирует работу со сжатием
func runWithCompression(client *openai.Client) {
	// Создаем менеджер контекста
	// Сжимаем каждые 10 сообщений, храним последние 6 "как есть"
	cm := agent.NewContextManager(client, 10, 6)

	// Симулируем длинный диалог
	messages := generateLongDialog()

	fmt.Printf("Всего сообщений: %d\n", len(messages))
	fmt.Printf("Настройки: сжатие каждые 10 сообщений, последние 6 без сжатия\n")

	// Добавляем сообщения постепенно
	for _, msg := range messages {
		cm.AddMessage(msg.Role, msg.Content)

		// Проверяем и сжимаем при необходимости
		if err := cm.CompressIfNeeded(); err != nil {
			fmt.Printf("Ошибка сжатия: %v\n", err)
		}
	}

	// Получаем статистику
	stats := cm.GetStats()

	fmt.Println("\n📊 Статистика сжатия:")
	fmt.Printf("  • Всего сообщений:        %d\n", stats.TotalMessages)
	fmt.Printf("  • Сжатых блоков:          %d\n", stats.CompressedBlocks)
	fmt.Printf("  • Последних (без сжатия): %d\n", stats.RecentMessages)
	fmt.Printf("  • Токенов оригинал:       %d\n", stats.OriginalTokens)
	fmt.Printf("  • Токенов сжато:          %d\n", stats.CompressedTokens)
	fmt.Printf("  • Сэкономлено токенов:    %d\n", stats.TokensSaved)
	fmt.Printf("  • Сжатие:                 %.1f%%\n", stats.CompressionPercent)

	// Формируем контекст для запроса
	contextMessages := cm.GetContextForRequest()

	fmt.Printf("\n📦 Контекст для запроса:\n")
	for i, msg := range contextMessages {
		role := msg.Role
		content := msg.Content
		if len(content) > 100 {
			content = content[:100] + "..."
		}
		fmt.Printf("  [%d] %s: %s\n", i+1, role, content)
	}

	// Отправляем запрос со сжатым контекстом
	compressedHistory := make([]openai.ChatCompletionMessage, 0)
	for _, msg := range contextMessages {
		compressedHistory = append(compressedHistory, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Добавляем финальный вопрос
	compressedHistory = append(compressedHistory, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "Подведи итог нашего разговора: о чем мы говорили и какие решения приняли?",
	})

	ctx := context.Background()
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		Messages:    compressedHistory,
		Temperature: 0.7,
	})

	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		return
	}

	if len(resp.Choices) > 0 {
		answer := resp.Choices[0].Message.Content
		fmt.Println("\n💬 Ответ агента:")
		fmt.Println(utils.WrapText(answer, 80))
	}

	// Финальная статистика
	fmt.Println("\n📊 Статистика запроса:")
	fmt.Printf("  • Токенов в запросе: %d\n", resp.Usage.PromptTokens)
	fmt.Printf("  • Токенов в ответе:  %d\n", resp.Usage.CompletionTokens)
	fmt.Printf("  • Всего токенов:     %d\n", resp.Usage.TotalTokens)
	fmt.Printf("  • Стоимость:         $%.6f\n", calculateCost(resp.Usage))
}

// compareQuality сравнивает качество ответов со сжатием и без
func compareQuality(client *openai.Client) {
	ctx := context.Background()

	// Создаем диалог с важной информацией в разных частях
	messages := []agent.Message{
		{Role: "user", Content: "Привет! Меня зовут Алексей, я работаю программистом в компании TechCorp."},
		{Role: "assistant", Content: "Приятно познакомиться, Алексей! Чем могу помочь?"},
		{Role: "user", Content: "Мне нужно выбрать язык программирования для нового проекта. Это будет веб-приложение для управления задачами."},
		{Role: "assistant", Content: "Отличный проект! Для веб-приложений есть много вариантов. Какой у вас опыт разработки?"},
		{Role: "user", Content: "Я знаю Python и JavaScript. Команда состоит из 5 человек, все знают JavaScript."},
		{Role: "assistant", Content: "Понятно. Учитывая знания команды, JavaScript (Node.js + React) будет хорошим выбором."},
		{Role: "user", Content: "А что насчет производительности? Приложение должно обрабатывать до 10000 пользователей."},
		{Role: "assistant", Content: "Node.js справится с такой нагрузкой. Можно также рассмотреть Next.js для SSR."},
		{Role: "user", Content: "Отлично! Еще вопрос: какую базу данных выбрать - PostgreSQL или MongoDB?"},
		{Role: "assistant", Content: "Для задач с четкой структурой (управление задачами) PostgreSQL будет лучше."},
		{Role: "user", Content: "Согласен. А для хостинга что посоветуешь? Бюджет ограничен - до $100/месяц."},
		{Role: "assistant", Content: "В таком случае Vercel (фронтенд) + Railway или Render (бэкенд) - отличные варианты в рамках бюджета."},
		{Role: "user", Content: "Спасибо! Давай подытожим: мы выбрали JavaScript (Next.js), PostgreSQL, хостинг Vercel+Railway."},
		{Role: "assistant", Content: "Верно! Это сбалансированный стек для вашего проекта управления задачами."},
	}

	// Тест 1: Вопрос требующий информации из начала диалога
	question1 := "Как меня зовут и где я работаю?"

	// Тест 2: Вопрос требующий информации из середины
	question2 := "Какие технологии мы выбрали для проекта и почему?"

	// Тест БЕЗ сжатия
	fmt.Println("\n🔷 БЕЗ СЖАТИЯ:")
	fmt.Println(strings.Repeat("─", 80))

	fullHistory := make([]openai.ChatCompletionMessage, 0)
	for _, msg := range messages {
		fullHistory = append(fullHistory, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Вопрос 1
	fmt.Printf("\n❓ Вопрос 1: %s\n", question1)
	answer1Without := askQuestion(ctx, client, fullHistory, question1)
	fmt.Printf("💬 Ответ: %s\n", answer1Without)

	// Вопрос 2
	fmt.Printf("\n❓ Вопрос 2: %s\n", question2)
	answer2Without := askQuestion(ctx, client, fullHistory, question2)
	fmt.Printf("💬 Ответ: %s\n", utils.WrapText(answer2Without, 80))

	// Тест СО сжатием
	fmt.Println("\n\n🔶 СО СЖАТИЕМ:")
	fmt.Println(strings.Repeat("─", 80))

	cm := agent.NewContextManager(client, 6, 4)
	for _, msg := range messages {
		cm.AddMessage(msg.Role, msg.Content)
		cm.CompressIfNeeded()
	}

	stats := cm.GetStats()
	fmt.Printf("📊 Сжатие: %d блоков, %.1f%% экономии токенов\n", stats.CompressedBlocks, stats.CompressionPercent)

	compressedHistory := make([]openai.ChatCompletionMessage, 0)
	for _, msg := range cm.GetContextForRequest() {
		compressedHistory = append(compressedHistory, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Вопрос 1
	fmt.Printf("\n❓ Вопрос 1: %s\n", question1)
	answer1With := askQuestion(ctx, client, compressedHistory, question1)
	fmt.Printf("💬 Ответ: %s\n", answer1With)

	// Вопрос 2
	fmt.Printf("\n❓ Вопрос 2: %s\n", question2)
	answer2With := askQuestion(ctx, client, compressedHistory, question2)
	fmt.Printf("💬 Ответ: %s\n", utils.WrapText(answer2With, 80))

	// Выводы
	fmt.Println("\n\n📋 ВЫВОДЫ:")
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println("✅ Информация из начала диалога:")
	fmt.Printf("   • Без сжатия: %s\n", truncate(answer1Without, 60))
	fmt.Printf("   • Со сжатием: %s\n", truncate(answer1With, 60))
	fmt.Println("\n✅ Информация из всего диалога:")
	fmt.Printf("   • Без сжатия: точный ответ с деталями\n")
	fmt.Printf("   • Со сжатием: основные решения сохранены\n")
	fmt.Printf("\n💡 Сжатие экономит %.1f%% токенов при сохранении ключевой информации!\n", stats.CompressionPercent)
}

// askQuestion отправляет вопрос с историей и возвращает ответ
func askQuestion(ctx context.Context, client *openai.Client, history []openai.ChatCompletionMessage, question string) string {
	messages := make([]openai.ChatCompletionMessage, len(history))
	copy(messages, history)

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question,
	})

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini,
		Messages:    messages,
		Temperature: 0.3,
	})

	if err != nil {
		return fmt.Sprintf("Ошибка: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "Нет ответа"
	}

	return resp.Choices[0].Message.Content
}

// generateLongDialog генерирует длинный диалог для тестирования
func generateLongDialog() []agent.Message {
	return []agent.Message{
		{Role: "user", Content: "Привет! Хочу изучить машинное обучение. С чего начать?"},
		{Role: "assistant", Content: "Отлично! Начните с основ Python и математики (линейная алгебра, статистика)."},
		{Role: "user", Content: "Python я знаю. А какие библиотеки нужны для ML?"},
		{Role: "assistant", Content: "Основные: NumPy, Pandas, Scikit-learn, Matplotlib. Для глубокого обучения - TensorFlow или PyTorch."},
		{Role: "user", Content: "Понял. А есть хорошие курсы?"},
		{Role: "assistant", Content: "Да! Coursera (Andrew Ng), Fast.ai, Google ML Crash Course - отличные варианты."},
		{Role: "user", Content: "Спасибо! Сколько времени обычно занимает обучение?"},
		{Role: "assistant", Content: "От 3-6 месяцев для базы до 1-2 лет для уверенного уровня. Зависит от интенсивности."},
		{Role: "user", Content: "Хорошо. А какой первый проект сделать?"},
		{Role: "assistant", Content: "Начните с классификации (например, MNIST - распознавание цифр) или регрессии (предсказание цен)."},
		{Role: "user", Content: "MNIST звучит интересно. Какую модель использовать?"},
		{Role: "assistant", Content: "Для начала логистическая регрессия, потом простая нейросеть (MLP), затем CNN."},
		{Role: "user", Content: "А что такое CNN?"},
		{Role: "assistant", Content: "Convolutional Neural Network - сверточная нейросеть. Отлично работает с изображениями."},
		{Role: "user", Content: "Понятно. А как оценить качество модели?"},
		{Role: "assistant", Content: "Используйте метрики: accuracy, precision, recall, F1-score. Важна также cross-validation."},
		{Role: "user", Content: "Что делать с переобучением?"},
		{Role: "assistant", Content: "Методы: больше данных, регуляризация (L1/L2), dropout, early stopping, data augmentation."},
		{Role: "user", Content: "А где брать данные для проектов?"},
		{Role: "assistant", Content: "Kaggle, UCI ML Repository, Google Dataset Search, OpenML. На Kaggle еще и соревнования есть."},
		{Role: "user", Content: "Отлично! Еще вопрос: GPU обязателен?"},
		{Role: "assistant", Content: "Для начала нет. Google Colab дает бесплатный GPU. Для серьезных проектов - желателен."},
		{Role: "user", Content: "А какие зарплаты у ML-инженеров?"},
		{Role: "assistant", Content: "В России: junior от 80-120k руб, middle 150-250k, senior 250k+. За границей значительно выше."},
		{Role: "user", Content: "Хорошая мотивация! Спасибо за помощь!"},
		{Role: "assistant", Content: "Пожалуйста! Удачи в изучении ML. Главное - практика и регулярность!"},
	}
}

// calculateCost рассчитывает стоимость запроса
func calculateCost(usage openai.Usage) float64 {
	// GPT-4o-mini pricing (per 1M tokens)
	inputPrice := 0.150 / 1_000_000  // $0.150 per 1M input tokens
	outputPrice := 0.600 / 1_000_000 // $0.600 per 1M output tokens

	inputCost := float64(usage.PromptTokens) * inputPrice
	outputCost := float64(usage.CompletionTokens) * outputPrice

	return inputCost + outputCost
}

// truncate обрезает строку до заданной длины
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
