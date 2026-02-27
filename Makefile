.PHONY: help day1 day2 day3 day4 day5 day6 day7 day8 day9 build clean test tidy install

help: ## Показать эту справку
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

day1: ## Запустить Day 1
	@echo "🚀 Запуск Day 1..."
	@go run cmd/advent/day1/main.go

day2: ## Запустить Day 2
	@echo "🚀 Запуск Day 2..."
	set -a && source .env && set +a && go run cmd/advent/day2/main.go

day3: ## Запустить Day 3
	@echo "🚀 Запуск Day 3..."
	set -a && source .env && set +a && go run cmd/advent/day3/main.go

day4: ## Запустить Day 4
	@echo "🚀 Запуск Day 4..."
	set -a && source .env && set +a && go run cmd/advent/day4/main.go

day5: ## Запустить Day 5
	@echo "🚀 Запуск Day 5..."
	set -a && source .env && set +a && go run cmd/advent/day5/main.go

day6: ## Запустить Day 6 (интерактивный агент)
	@echo "🚀 Запуск Day 6..."
	set -a && source .env && set +a && go run cmd/advent/day6/main.go

day7: ## Запустить Day 7 (агент с сохранением контекста)
	@echo "🚀 Запуск Day 7..."
	set -a && source .env && set +a && go run cmd/advent/day7/main.go

day8: ## Запустить Day 8 (работа с токенами)
	@echo "🚀 Запуск Day 8..."
	set -a && source .env && set +a && go run cmd/advent/day8/main.go

day9: ## Запустить Day 9 (управление контекстом, сжатие истории)
	@echo "🚀 Запуск Day 9..."
	set -a && source .env && set +a && go run cmd/advent/day9/main.go

build: ## Собрать все бинарники
	@echo "🔨 Сборка всех бинарников..."
	@mkdir -p bin
	@go build -o bin/day1 cmd/advent/day1/main.go
	@go build -o bin/day2 cmd/advent/day2/main.go
	@go build -o bin/day3 cmd/advent/day3/main.go
	@go build -o bin/day4 cmd/advent/day4/main.go
	@go build -o bin/day5 cmd/advent/day5/main.go
	@go build -o bin/day6 cmd/advent/day6/main.go
	@go build -o bin/day7 cmd/advent/day7/main.go
	@go build -o bin/day8 cmd/advent/day8/main.go
	@go build -o bin/day9 cmd/advent/day9/main.go
	@echo "✅ Бинарники собраны в директории bin/"

clean: ## Удалить собранные бинарники
	@echo "🧹 Очистка..."
	@rm -rf bin/
	@echo "✅ Очистка завершена"

test: ## Запустить тесты
	@echo "🧪 Запуск тестов..."
	@go test -v ./...

tidy: ## Обновить зависимости
	@echo "📦 Обновление зависимостей..."
	@go mod tidy
	@echo "✅ Зависимости обновлены"

install: ## Установить зависимости
	@echo "📦 Установка зависимостей..."
	@go mod download
	@echo "✅ Зависимости установлены"
