package utils

import (
	"fmt"
	"strings"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// PrintHeader выводит заголовок
func PrintHeader(title string) {
	line := strings.Repeat("=", 80)
	fmt.Println(line)
	fmt.Println(title)
	fmt.Println(line)
	fmt.Println()
}

// PrintSubHeader выводит подзаголовок
func PrintSubHeader(title string) {
	line := strings.Repeat("-", 80)
	fmt.Println(line)
	fmt.Println(title)
	fmt.Println(line)
}

// PrintSection выводит секцию с эмодзи
func PrintSection(emoji, title string) {
	fmt.Printf("%s %s\n", emoji, title)
	PrintSubHeader("")
}

// PrintKeyValue выводит пару ключ-значение
func PrintKeyValue(key, value string) {
	fmt.Printf("%s: %s\n", key, value)
}

// PrintTokenStats выводит статистику по токенам
func PrintTokenStats(total, prompt, completion int) {
	fmt.Printf("Токены использовано: %d (промпт: %d, ответ: %d)\n", total, prompt, completion)
}

// PrintDivider выводит разделитель
func PrintDivider() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()
}

// PrintColored выводит цветной текст
func PrintColored(color, text string) {
	fmt.Printf("%s%s%s\n", color, text, ColorReset)
}

// PrintSuccess выводит успешное сообщение
func PrintSuccess(text string) {
	fmt.Printf("%s✓ %s%s\n", ColorGreen, text, ColorReset)
}

// PrintError выводит сообщение об ошибке
func PrintError(text string) {
	fmt.Printf("%s✗ %s%s\n", ColorRed, text, ColorReset)
}

// PrintInfo выводит информационное сообщение
func PrintInfo(text string) {
	fmt.Printf("%sℹ %s%s\n", ColorBlue, text, ColorReset)
}

// Repeat повторяет строку n раз
func Repeat(s string, n int) string {
	return strings.Repeat(s, n)
}

// WrapText переносит текст по словам с заданной шириной
func WrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		// Если добавление слова превысит ширину
		if len(currentLine)+len(word)+1 > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Слово длиннее ширины - добавляем как есть
				lines = append(lines, word)
			}
		} else {
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
	}

	// Добавляем последнюю строку
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// PrintSeparator выводит разделитель
func PrintSeparator() {
	fmt.Println(strings.Repeat("-", 80))
}
