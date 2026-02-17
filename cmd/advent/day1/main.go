package main

import (
	"fmt"
	"log"

	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/client"
	"github.com/georgijter-grigoranc/ai-advent-challenge/internal/config"
	"github.com/georgijter-grigoranc/ai-advent-challenge/pkg/utils"
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
	utils.PrintHeader("Day 1: Первый запрос к OpenAI API")

	// Запрос
	prompt := "Привет! Расскажи, что ты умеешь делать?"
	utils.PrintSection("📝", "ЗАПРОС")
	fmt.Printf("Промпт: %s\n\n", prompt)

	// Выполнение запроса
	resp, err := aiClient.CreateCompletion(client.CompletionRequest{
		Prompt:      prompt,
		Temperature: 0.7,
	})

	if err != nil {
		log.Fatalf("Ошибка при выполнении запроса: %v", err)
	}

	// Вывод ответа
	utils.PrintSection("💬", "ОТВЕТ")
	fmt.Printf("%s\n\n", resp.Content)

	// Статистика
	utils.PrintSection("📊", "СТАТИСТИКА")
	utils.PrintTokenStats(resp.TotalTokens, resp.PromptTokens, resp.CompletionTokens)
	utils.PrintKeyValue("Модель", resp.Model)
	utils.PrintKeyValue("Finish reason", resp.FinishReason)

	utils.PrintDivider()
	utils.PrintSuccess("Задание Day 1 выполнено!")
}
