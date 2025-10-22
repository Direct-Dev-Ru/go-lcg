package serve

import (
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
)

// getVerbosePromptsFromFile возвращает промпты подробности из файла sys_prompts
func getVerbosePromptsFromFile(prompts []gpt.SystemPrompt, lang string) []VerbosePrompt {
	var verbosePrompts []VerbosePrompt

	// Ищем промпты подробности в загруженных промптах (ID 6, 7, 8)
	for _, prompt := range prompts {
		if prompt.ID >= 6 && prompt.ID <= 8 {
			// Определяем режим по ID
			var mode string
			switch prompt.ID {
			case 6:
				mode = "v"
			case 7:
				mode = "vv"
			case 8:
				mode = "vvv"
			}

			// Переводим на нужный язык если необходимо
			translatedPrompt := translateVerbosePrompt(prompt, lang)

			verbosePrompts = append(verbosePrompts, VerbosePrompt{
				Mode:        mode,
				Name:        translatedPrompt.Name,
				Description: translatedPrompt.Description,
				Content:     translatedPrompt.Content,
				IsDefault:   gpt.IsBuiltinPrompt(translatedPrompt), // Проверяем, является ли промпт встроенным
			})
		}
	}

	// Если промпты подробности не найдены в файле, используем встроенные
	if len(verbosePrompts) == 0 {
		return getVerbosePrompts(lang)
	}

	return verbosePrompts
}

// translateVerbosePrompt переводит промпт подробности на указанный язык
func translateVerbosePrompt(prompt gpt.SystemPrompt, lang string) gpt.SystemPrompt {
	// Получаем встроенный промпт для указанного языка из YAML
	if builtinPrompt := gpt.GetBuiltinPromptByIDAndLanguage(prompt.ID, lang); builtinPrompt != nil {
		return *builtinPrompt
	}

	// Если перевод не найден, возвращаем оригинал
	return prompt
}

// getVerbosePrompts возвращает промпты для режимов v/vv/vvv (fallback)
func getVerbosePrompts(lang string) []VerbosePrompt {
	// Английские версии (по умолчанию)
	enPrompts := []VerbosePrompt{
		{
			Mode:        "v",
			Name:        "Verbose Mode",
			Description: "Detailed explanation of the command",
			Content:     "Provide a brief explanation of what this Linux command does, including what each flag and option means, and give examples of usage.",
			IsDefault:   true,
		},
		{
			Mode:        "vv",
			Name:        "Very Verbose Mode",
			Description: "Comprehensive explanation with alternatives",
			Content:     "Provide a comprehensive explanation of this Linux command, including detailed descriptions of all flags and options, alternative approaches, common use cases, and potential pitfalls to avoid.",
			IsDefault:   true,
		},
		{
			Mode:        "vvv",
			Name:        "Maximum Verbose Mode",
			Description: "Complete guide with examples and best practices",
			Content:     "Provide a complete guide for this Linux command, including detailed explanations of all options, multiple examples with different scenarios, alternative commands that achieve similar results, best practices, troubleshooting tips, and related commands that work well together.",
			IsDefault:   true,
		},
	}

	// Русские версии
	ruPrompts := []VerbosePrompt{
		{
			Mode:        "v",
			Name:        "Подробный режим",
			Description: "Подробное объяснение команды",
			Content:     "Предоставь краткое объяснение того, что делает эта Linux команда, включая значение каждого флага и опции, и приведи примеры использования.",
			IsDefault:   true,
		},
		{
			Mode:        "vv",
			Name:        "Очень подробный режим",
			Description: "Исчерпывающее объяснение с альтернативами",
			Content:     "Предоставь исчерпывающее объяснение этой Linux команды, включая подробные описания всех флагов и опций, альтернативные подходы, распространенные случаи использования и потенциальные подводные камни, которых следует избегать.",
			IsDefault:   true,
		},
		{
			Mode:        "vvv",
			Name:        "Максимально подробный режим",
			Description: "Полное руководство с примерами и лучшими практиками",
			Content:     "Предоставь полное руководство по этой Linux команде, включая подробные объяснения всех опций, множественные примеры с различными сценариями, альтернативные команды, которые дают аналогичные результаты, лучшие практики, советы по устранению неполадок и связанные команды, которые хорошо работают вместе.",
			IsDefault:   true,
		},
	}

	if lang == "ru" {
		return ruPrompts
	}
	return enPrompts
}

// getSystemPromptsWithLang возвращает системные промпты с учетом языка
func getSystemPromptsWithLang(prompts []gpt.SystemPrompt, lang string) []gpt.SystemPrompt {
	// Если язык английский, возвращаем оригинальные промпты
	if lang == "en" {
		return prompts
	}

	// Для русского языка переводим только встроенные промпты
	var translatedPrompts []gpt.SystemPrompt
	for _, prompt := range prompts {
		// Проверяем, является ли это встроенным промптом
		if gpt.IsBuiltinPrompt(prompt) {
			// Переводим встроенные промпты на русский
			translated := translateSystemPrompt(prompt, lang)
			translatedPrompts = append(translatedPrompts, translated)
		} else {
			translatedPrompts = append(translatedPrompts, prompt)
		}
	}

	return translatedPrompts
}

// translateSystemPrompt переводит системный промпт на указанный язык
func translateSystemPrompt(prompt gpt.SystemPrompt, lang string) gpt.SystemPrompt {
	// Получаем встроенный промпт для указанного языка из YAML
	if builtinPrompt := gpt.GetBuiltinPromptByIDAndLanguage(prompt.ID, lang); builtinPrompt != nil {
		return *builtinPrompt
	}

	// Если перевод не найден, возвращаем оригинал
	return prompt
}
