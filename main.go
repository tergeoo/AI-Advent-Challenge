package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Provider интерфейс для разных LLM провайдеров
type Provider interface {
	SendRequest(prompt string) (string, error)
}

// OpenAIProvider реализация для OpenAI API
type OpenAIProvider struct {
	APIKey string
	Model  string
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
}

func (p *OpenAIProvider) SendRequest(prompt string) (string, error) {
	reqBody := openAIRequest{
		Model: p.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ошибка при создании JSON: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка при отправке запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка при чтении ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка API (код %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("ошибка при парсинге ответа: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("пустой ответ от API")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

func main() {
	// Получаем API ключ из переменной окружения
	godotenv.Load()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Ошибка: установите переменную окружения OPENAI_API_KEY")
		os.Exit(1)
	}

	// Создаем провайдер
	provider := &OpenAIProvider{
		APIKey: apiKey,
		Model:  "gpt-3.5-turbo", // можно изменить на gpt-4
	}

	// Режим работы: CLI
	fmt.Println("🤖 LLM CLI Client (OpenAI)")
	fmt.Println("Введите ваш запрос (или 'exit' для выхода):")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("👉 Вы: ")
		if !scanner.Scan() {
			break
		}

		prompt := strings.TrimSpace(scanner.Text())
		if prompt == "" {
			continue
		}

		if prompt == "exit" {
			fmt.Println("Пока! 👋")
			break
		}

		fmt.Println("\n⏳ Отправка запроса...")
		response, err := provider.SendRequest(prompt)
		if err != nil {
			fmt.Printf("❌ Ошибка: %v\n\n", err)
			continue
		}

		fmt.Println("\n🤖 Ответ:")
		fmt.Println(response)
		fmt.Println()
	}
}
