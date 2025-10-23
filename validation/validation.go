package validation

import (
	"fmt"
	"strings"

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

// ValidationError представляет ошибку валидации
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateSystemPrompt проверяет длину системного промпта
func ValidateSystemPrompt(prompt string) error {
	maxLen := config.AppConfig.Validation.MaxSystemPromptLength
	if len(prompt) > maxLen {
		return ValidationError{
			Field:   "system_prompt",
			Message: fmt.Sprintf("системный промпт слишком длинный: %d символов (максимум %d)", len(prompt), maxLen),
		}
	}
	return nil
}

// ValidateUserMessage проверяет длину пользовательского сообщения
func ValidateUserMessage(message string) error {
	maxLen := config.AppConfig.Validation.MaxUserMessageLength
	if len(message) > maxLen {
		return ValidationError{
			Field:   "user_message",
			Message: fmt.Sprintf("пользовательское сообщение слишком длинное: %d символов (максимум %d)", len(message), maxLen),
		}
	}
	return nil
}

// ValidatePromptAndMessage проверяет и системный промпт, и пользовательское сообщение
func ValidatePromptAndMessage(systemPrompt, userMessage string) error {
	if err := ValidateSystemPrompt(systemPrompt); err != nil {
		return err
	}
	if err := ValidateUserMessage(userMessage); err != nil {
		return err
	}
	return nil
}

// TruncateSystemPrompt обрезает системный промпт до максимальной длины
func TruncateSystemPrompt(prompt string) string {
	maxLen := config.AppConfig.Validation.MaxSystemPromptLength
	if len(prompt) <= maxLen {
		return prompt
	}
	return prompt[:maxLen]
}

// TruncateUserMessage обрезает пользовательское сообщение до максимальной длины
func TruncateUserMessage(message string) string {
	maxLen := config.AppConfig.Validation.MaxUserMessageLength
	if len(message) <= maxLen {
		return message
	}
	return message[:maxLen]
}

// GetSystemPromptLength возвращает длину системного промпта
func GetSystemPromptLength(prompt string) int {
	return len(prompt)
}

// GetUserMessageLength возвращает длину пользовательского сообщения
func GetUserMessageLength(message string) int {
	return len(message)
}

// FormatLengthInfo форматирует информацию о длине для отображения
func FormatLengthInfo(systemPrompt, userMessage string) string {
	systemLen := GetSystemPromptLength(systemPrompt)
	userLen := GetUserMessageLength(userMessage)
	maxSystemLen := config.AppConfig.Validation.MaxSystemPromptLength
	maxUserLen := config.AppConfig.Validation.MaxUserMessageLength

	var warnings []string

	if systemLen > maxSystemLen {
		warnings = append(warnings, fmt.Sprintf("⚠️ Системный промпт превышает лимит: %d/%d символов", systemLen, maxSystemLen))
	}

	if userLen > maxUserLen {
		warnings = append(warnings, fmt.Sprintf("⚠️ Пользовательское сообщение превышает лимит: %d/%d символов", userLen, maxUserLen))
	}

	if len(warnings) == 0 {
		return fmt.Sprintf("✅ Длины в пределах нормы: системный промпт %d/%d, сообщение %d/%d",
			systemLen, maxSystemLen, userLen, maxUserLen)
	}

	return strings.Join(warnings, "\n")
}

// ValidatePromptName проверяет длину названия промпта
func ValidatePromptName(name string) error {
	maxLen := config.AppConfig.Validation.MaxPromptNameLength
	if len(name) > maxLen {
		return ValidationError{
			Field:   "prompt_name",
			Message: fmt.Sprintf("название промпта слишком длинное: %d символов (максимум %d)", len(name), maxLen),
		}
	}
	return nil
}

// ValidatePromptDescription проверяет длину описания промпта
func ValidatePromptDescription(description string) error {
	maxLen := config.AppConfig.Validation.MaxPromptDescLength
	if len(description) > maxLen {
		return ValidationError{
			Field:   "prompt_description",
			Message: fmt.Sprintf("описание промпта слишком длинное: %d символов (максимум %d)", len(description), maxLen),
		}
	}
	return nil
}

// ValidateCommand проверяет длину команды
func ValidateCommand(command string) error {
	maxLen := config.AppConfig.Validation.MaxCommandLength
	if len(command) > maxLen {
		return ValidationError{
			Field:   "command",
			Message: fmt.Sprintf("команда слишком длинная: %d символов (максимум %d)", len(command), maxLen),
		}
	}
	return nil
}

// ValidateExplanation проверяет длину объяснения
func ValidateExplanation(explanation string) error {
	maxLen := config.AppConfig.Validation.MaxExplanationLength
	if len(explanation) > maxLen {
		return ValidationError{
			Field:   "explanation",
			Message: fmt.Sprintf("объяснение слишком длинное: %d символов (максимум %d)", len(explanation), maxLen),
		}
	}
	return nil
}
