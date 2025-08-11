package gpt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SystemPrompt представляет системный промпт
type SystemPrompt struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// PromptManager управляет системными промптами
type PromptManager struct {
	Prompts    []SystemPrompt
	ConfigFile string
	HomeDir    string
}

// NewPromptManager создает новый менеджер промптов
func NewPromptManager(homeDir string) *PromptManager {
	configFile := filepath.Join(homeDir, ".lcg_prompts.json")

	pm := &PromptManager{
		ConfigFile: configFile,
		HomeDir:    homeDir,
	}

	// Загружаем предустановленные промпты
	pm.loadDefaultPrompts()

	// Загружаем пользовательские промпты
	pm.loadCustomPrompts()

	return pm
}

// loadDefaultPrompts загружает предустановленные промпты
func (pm *PromptManager) loadDefaultPrompts() {
	defaultPrompts := []SystemPrompt{
		{
			ID:          1,
			Name:        "linux-command",
			Description: "Generate Linux commands (default)",
			Content:     "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols.",
		},
		{
			ID:          2,
			Name:        "linux-command-with-explanation",
			Description: "Generate Linux commands with explanation",
			Content:     "Generate a Linux command and provide a brief explanation of what it does. Format: COMMAND: explanation",
		},
		{
			ID:          3,
			Name:        "linux-command-safe",
			Description: "Generate safe Linux commands",
			Content:     "Generate a safe Linux command that won't cause data loss or system damage. Reply with linux command and nothing else. Output with plain response - no need formatting.",
		},
		{
			ID:          4,
			Name:        "linux-command-verbose",
			Description: "Generate Linux commands with detailed explanation",
			Content:     "Generate a Linux command and provide detailed explanation including what each flag does and potential alternatives.",
		},
		{
			ID:          5,
			Name:        "linux-command-simple",
			Description: "Generate simple Linux commands",
			Content:     "Generate a simple, easy-to-understand Linux command. Avoid complex flags and options when possible.",
		},
	}

	pm.Prompts = defaultPrompts
}

// loadCustomPrompts загружает пользовательские промпты из файла
func (pm *PromptManager) loadCustomPrompts() {
	if _, err := os.Stat(pm.ConfigFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(pm.ConfigFile)
	if err != nil {
		return
	}

	var customPrompts []SystemPrompt
	if err := json.Unmarshal(data, &customPrompts); err != nil {
		return
	}

	// Добавляем пользовательские промпты с новыми ID
	for i, prompt := range customPrompts {
		prompt.ID = len(pm.Prompts) + i + 1
		pm.Prompts = append(pm.Prompts, prompt)
	}
}

// saveCustomPrompts сохраняет пользовательские промпты
func (pm *PromptManager) saveCustomPrompts() error {
	// Находим пользовательские промпты (ID > 5)
	var customPrompts []SystemPrompt
	for _, prompt := range pm.Prompts {
		if prompt.ID > 5 {
			customPrompts = append(customPrompts, prompt)
		}
	}

	data, err := json.MarshalIndent(customPrompts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(pm.ConfigFile, data, 0644)
}

// GetPromptByID возвращает промпт по ID
func (pm *PromptManager) GetPromptByID(id int) (*SystemPrompt, error) {
	for _, prompt := range pm.Prompts {
		if prompt.ID == id {
			return &prompt, nil
		}
	}
	return nil, fmt.Errorf("промпт с ID %d не найден", id)
}

// GetPromptByName возвращает промпт по имени
func (pm *PromptManager) GetPromptByName(name string) (*SystemPrompt, error) {
	for _, prompt := range pm.Prompts {
		if strings.EqualFold(prompt.Name, name) {
			return &prompt, nil
		}
	}
	return nil, fmt.Errorf("промпт с именем '%s' не найден", name)
}

// ListPrompts выводит список всех доступных промптов
func (pm *PromptManager) ListPrompts() {
	fmt.Println("Available system prompts:")
	fmt.Println("ID | Name                      | Description")
	fmt.Println("---+---------------------------+--------------------------------")

	for _, prompt := range pm.Prompts {
		description := prompt.Description
		if len(description) > 80 {
			description = description[:77] + "..."
		}
		fmt.Printf("%-2d | %-25s | %s\n",
			prompt.ID,
			truncateString(prompt.Name, 25),
			description)
	}
}

// AddCustomPrompt добавляет новый пользовательский промпт
func (pm *PromptManager) AddCustomPrompt(name, description, content string) error {
	// Проверяем, что имя уникально
	for _, prompt := range pm.Prompts {
		if strings.EqualFold(prompt.Name, name) {
			return fmt.Errorf("промпт с именем '%s' уже существует", name)
		}
	}

	newPrompt := SystemPrompt{
		ID:          len(pm.Prompts) + 1,
		Name:        name,
		Description: description,
		Content:     content,
	}

	pm.Prompts = append(pm.Prompts, newPrompt)
	return pm.saveCustomPrompts()
}

// DeleteCustomPrompt удаляет пользовательский промпт
func (pm *PromptManager) DeleteCustomPrompt(id int) error {
	if id <= 5 {
		return fmt.Errorf("нельзя удалить предустановленный промпт")
	}

	for i, prompt := range pm.Prompts {
		if prompt.ID == id {
			pm.Prompts = append(pm.Prompts[:i], pm.Prompts[i+1:]...)
			return pm.saveCustomPrompts()
		}
	}

	return fmt.Errorf("промпт с ID %d не найден", id)
}

// truncateString обрезает строку до указанной длины
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
