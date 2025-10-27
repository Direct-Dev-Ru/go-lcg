package gpt

import (
	_ "embed"
	"runtime"

	"gopkg.in/yaml.v3"
)

//go:embed builtin_prompts.yaml
var builtinPromptsYAML string

//go:embed builtin_prompts_windows.yaml
var builtinPromptsWindowsYAML string

var builtinPrompts string

// BuiltinPromptsData структура для YAML файла
type BuiltinPromptsData struct {
	Prompts []BuiltinPrompt `yaml:"prompts"`
}

// BuiltinPrompt структура для встроенных промптов с поддержкой языков
type BuiltinPrompt struct {
	ID          int               `yaml:"id"`
	Name        string            `yaml:"name"`
	Description map[string]string `yaml:"description"`
	Content     map[string]string `yaml:"content"`
}

// ToSystemPrompt конвертирует BuiltinPrompt в SystemPrompt для указанного языка
func (bp *BuiltinPrompt) ToSystemPrompt(lang string) SystemPrompt {
	// Если язык не найден, используем английский по умолчанию
	if _, exists := bp.Description[lang]; !exists {
		lang = "en"
	}

	return SystemPrompt{
		ID:          bp.ID,
		Name:        bp.Name,
		Description: bp.Description[lang],
		Content:     bp.Content[lang],
	}
}

// GetBuiltinPrompts возвращает встроенные промпты из YAML (по умолчанию английские)
func GetBuiltinPrompts() []SystemPrompt {
	return GetBuiltinPromptsByLanguage("en")
}

// GetBuiltinPromptsByLanguage возвращает встроенные промпты для указанного языка
func GetBuiltinPromptsByLanguage(lang string) []SystemPrompt {
	var data BuiltinPromptsData
	if err := yaml.Unmarshal([]byte(builtinPrompts), &data); err != nil {
		// В случае ошибки возвращаем пустой массив
		return []SystemPrompt{}
	}

	var result []SystemPrompt
	for _, prompt := range data.Prompts {
		result = append(result, prompt.ToSystemPrompt(lang))
	}
	return result
}

// IsBuiltinPrompt проверяет, является ли промпт встроенным
func IsBuiltinPrompt(prompt SystemPrompt) bool {
	// Проверяем английскую версию
	englishPrompts := GetBuiltinPromptsByLanguage("en")
	for _, builtin := range englishPrompts {
		if builtin.ID == prompt.ID {
			if builtin.Content == prompt.Content &&
				builtin.Name == prompt.Name &&
				builtin.Description == prompt.Description {
				return true
			}
		}
	}

	// Проверяем русскую версию
	russianPrompts := GetBuiltinPromptsByLanguage("ru")
	for _, builtin := range russianPrompts {
		if builtin.ID == prompt.ID {
			if builtin.Content == prompt.Content &&
				builtin.Name == prompt.Name &&
				builtin.Description == prompt.Description {
				return true
			}
		}
	}

	return false
}

// GetBuiltinPromptByID возвращает встроенный промпт по ID (английская версия)
func GetBuiltinPromptByID(id int) *SystemPrompt {
	builtinPrompts := GetBuiltinPrompts()

	for _, prompt := range builtinPrompts {
		if prompt.ID == id {
			return &prompt
		}
	}

	return nil
}

// GetBuiltinPromptByIDAndLanguage возвращает встроенный промпт по ID и языку
func GetBuiltinPromptByIDAndLanguage(id int, lang string) *SystemPrompt {
	builtinPrompts := GetBuiltinPromptsByLanguage(lang)

	for _, prompt := range builtinPrompts {
		if prompt.ID == id {
			return &prompt
		}
	}

	return nil
}

func InitBuiltinPrompts(embeddedBuiltinPromptsYAML string) {
	// Используем встроенный YAML, если переданный параметр пустой
	if embeddedBuiltinPromptsYAML == "" {
		// Выбираем промпты в зависимости от операционной системы
		if runtime.GOOS == "windows" {
			builtinPrompts = builtinPromptsWindowsYAML
		} else {
			builtinPrompts = builtinPromptsYAML
		}
	} else {
		builtinPrompts = embeddedBuiltinPromptsYAML
	}
}
