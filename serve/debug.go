package serve

import (
	"fmt"

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

// PrintWebDebugInfo выводит отладочную информацию для веб-запросов
func PrintWebDebugInfo(operation, prompt, systemPrompt, model string, timeout int) {
	if !config.AppConfig.MainFlags.Debug {
		return
	}

	fmt.Printf("\n🔍 DEBUG WEB %s:\n", operation)
	fmt.Printf("💬 Запрос: %s\n", prompt)
	fmt.Printf("🤖 Системный промпт: %s\n", systemPrompt)
	fmt.Printf("⏱️  Таймаут: %d сек\n", timeout)
	fmt.Printf("🌐 Провайдер: %s\n", config.AppConfig.ProviderType)
	fmt.Printf("🏠 Хост: %s\n", config.AppConfig.Host)
	fmt.Printf("🧠 Модель: %s\n", model)
	fmt.Printf("📝 История: %t\n", !config.AppConfig.MainFlags.NoHistory)
	fmt.Printf("────────────────────────────────────────\n")
}

// PrintWebVerboseDebugInfo выводит отладочную информацию для verbose запросов
func PrintWebVerboseDebugInfo(operation, prompt, verbosePrompt, model string, level int, timeout int) {
	if !config.AppConfig.MainFlags.Debug {
		return
	}

	fmt.Printf("\n🔍 DEBUG WEB %s (v%d):\n", operation, level)
	fmt.Printf("💬 Запрос: %s\n", prompt)
	fmt.Printf("📝 Системный промпт подробности:\n%s\n", verbosePrompt)
	fmt.Printf("⏱️  Таймаут: %d сек\n", timeout)
	fmt.Printf("🌐 Провайдер: %s\n", config.AppConfig.ProviderType)
	fmt.Printf("🏠 Хост: %s\n", config.AppConfig.Host)
	fmt.Printf("🧠 Модель: %s\n", model)
	fmt.Printf("🎯 Уровень подробности: %d\n", level)
	fmt.Printf("────────────────────────────────────────\n")
}

// PrintWebSaveDebugInfo выводит отладочную информацию для сохранения
func PrintWebSaveDebugInfo(operation, prompt, command, explanation, model, file string) {
	if !config.AppConfig.MainFlags.Debug {
		return
	}

	fmt.Printf("\n🔍 DEBUG WEB %s:\n", operation)
	fmt.Printf("💬 Запрос: %s\n", prompt)
	fmt.Printf("⚡ Команда: %s\n", command)
	fmt.Printf("📖 Объяснение: %s\n", explanation)
	fmt.Printf("🧠 Модель: %s\n", model)
	fmt.Printf("📁 Файл: %s\n", file)
	fmt.Printf("────────────────────────────────────────\n")
}
