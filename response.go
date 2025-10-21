package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

func nowTimestamp() string {
	return time.Now().Format("2006-01-02_15-04-05")
}

func pathJoin(base, name string) string {
	return path.Join(base, name)
}

func writeFile(filePath, content string) {
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Println("Failed to save response:", err)
	} else {
		fmt.Printf("Saved to %s\n", filePath)
	}
}

func saveResponse(response string, gpt3Model string, prompt string, cmd string, explanation ...string) {
	timestamp := nowTimestamp()
	filename := fmt.Sprintf("gpt_request_%s_%s.md", gpt3Model, timestamp)
	filePath := pathJoin(config.AppConfig.ResultFolder, filename)
	title := truncateTitle(cmd)

	var content string
	if len(explanation) > 0 && strings.TrimSpace(explanation[0]) != "" {
		// Если есть объяснение, сохраняем полную структуру
		content = fmt.Sprintf("# %s\n\n## Prompt\n\n%s\n\n## Response\n\n%s\n\n## Explanation\n\n%s\n",
			title, cmd+". "+prompt, response, explanation[0])
	} else {
		// Если объяснения нет, сохраняем базовую структуру
		content = fmt.Sprintf("# %s\n\n## Prompt\n\n%s\n\n## Response\n\n%s\n",
			title, cmd+". "+prompt, response)
	}
	writeFile(filePath, content)
}

func truncateTitle(s string) string {
	const maxLen = 120
	if runeCount := len([]rune(s)); runeCount <= maxLen {
		return s
	}
	const head = 116
	r := []rune(s)
	if len(r) <= head {
		return s
	}
	return string(r[:head]) + " ..."
}
