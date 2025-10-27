package gpt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/direct-dev-ru/linux-command-gpt/config"
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
	Language   string // Текущий язык для файла sys_prompts (en/ru)
}

// NewPromptManager создает новый менеджер промптов
func NewPromptManager(homeDir string) *PromptManager {
	// Используем конфигурацию из модуля config
	promptFolder := config.AppConfig.PromptFolder

	// Путь к файлу sys_prompts
	sysPromptsFile := filepath.Join(promptFolder, "sys_prompts")

	pm := &PromptManager{
		ConfigFile: sysPromptsFile,
		HomeDir:    homeDir,
	}

	// Проверяем, существует ли файл sys_prompts
	if _, err := os.Stat(sysPromptsFile); os.IsNotExist(err) {
		// Если файла нет, создаем его с системными промптами и промптами подробности
		pm.createInitialPromptsFile()
	}

	// Загружаем все промпты из файла
	pm.loadAllPrompts()

	return pm
}

// createInitialPromptsFile создает начальный файл с системными промптами и промптами подробности
func (pm *PromptManager) createInitialPromptsFile() {
	// Устанавливаем язык по умолчанию как русский
	pm.Language = "ru"

	// Загружаем все встроенные промпты из YAML на русском языке
	// Функция GetBuiltinPromptsByLanguage уже учитывает операционную систему
	pm.Prompts = GetBuiltinPromptsByLanguage("ru")

	// Сохраняем все промпты в файл
	pm.saveAllPrompts()
}

// loadDefaultPrompts загружает предустановленные промпты
func (pm *PromptManager) LoadDefaultPrompts() {
	// Используем встроенные промпты, которые автоматически выбираются по ОС
	pm.Prompts = GetBuiltinPromptsByLanguage("en")
}

// loadAllPrompts загружает все промпты из файла sys_prompts
func (pm *PromptManager) loadAllPrompts() {
	if _, err := os.Stat(pm.ConfigFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(pm.ConfigFile)
	if err != nil {
		return
	}

	// Новый формат: объект с полями language и prompts
	var pf promptsFile
	if err := json.Unmarshal(data, &pf); err == nil && len(pf.Prompts) > 0 {
		pm.Language = pf.Language
		pm.Prompts = pf.Prompts
		return
	}

	// Старый формат: просто массив промптов
	var prompts []SystemPrompt
	if err := json.Unmarshal(data, &prompts); err == nil {
		pm.Prompts = prompts
		pm.Language = "en"
		// Миграция в новый формат при следующем сохранении
	}
}

// saveAllPrompts сохраняет все промпты в файл sys_prompts
// внутренний формат хранения файла sys_prompts
type promptsFile struct {
	Language string         `json:"language,omitempty"`
	Prompts  []SystemPrompt `json:"prompts"`
}

func (pm *PromptManager) saveAllPrompts() error {
	pf := promptsFile{
		Language: pm.Language,
		Prompts:  pm.Prompts,
	}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pm.ConfigFile, data, 0644)
}

// SaveAllPrompts экспортированная версия saveAllPrompts
func (pm *PromptManager) SaveAllPrompts() error {
	return pm.saveAllPrompts()
}

// GetCurrentLanguage возвращает текущий язык из файла промптов
func (pm *PromptManager) GetCurrentLanguage() string {
	if pm.Language == "" {
		return "en"
	}
	return pm.Language
}

// SetLanguage устанавливает язык для всех промптов
func (pm *PromptManager) SetLanguage(lang string) {
	pm.Language = lang
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

// AddPrompt добавляет новый промпт
func (pm *PromptManager) AddPrompt(name, description, content string) error {
	// Находим максимальный ID
	maxID := 0
	for _, prompt := range pm.Prompts {
		if prompt.ID > maxID {
			maxID = prompt.ID
		}
	}

	newPrompt := SystemPrompt{
		ID:          maxID + 1,
		Name:        name,
		Description: description,
		Content:     content,
	}

	pm.Prompts = append(pm.Prompts, newPrompt)
	return pm.saveAllPrompts()
}

// UpdatePrompt обновляет существующий промпт
func (pm *PromptManager) UpdatePrompt(id int, name, description, content string) error {
	for i, prompt := range pm.Prompts {
		if prompt.ID == id {
			pm.Prompts[i].Name = name
			pm.Prompts[i].Description = description
			pm.Prompts[i].Content = content
			return pm.saveAllPrompts()
		}
	}
	return fmt.Errorf("промпт с ID %d не найден", id)
}

// DeletePrompt удаляет промпт по ID
func (pm *PromptManager) DeletePrompt(id int) error {
	for i, prompt := range pm.Prompts {
		if prompt.ID == id {
			pm.Prompts = append(pm.Prompts[:i], pm.Prompts[i+1:]...)
			return pm.saveAllPrompts()
		}
	}
	return fmt.Errorf("промпт с ID %d не найден", id)
}

// ListPrompts выводит список всех доступных промптов
func (pm *PromptManager) ListPrompts() {
	pm.ListPromptsWithFull(false)
}

// ListPromptsWithFull выводит список промптов с опцией полного вывода
func (pm *PromptManager) ListPromptsWithFull(full bool) {
	fmt.Println("📝 Доступные системные промпты:")
	fmt.Println()

	for i, prompt := range pm.Prompts {
		// Разделитель между промптами
		if i > 0 {
			fmt.Println("─" + strings.Repeat("─", 60))
		}

		// Проверяем, является ли промпт встроенным и неизмененным
		isDefault := pm.isDefaultPrompt(prompt)

		// Заголовок промпта
		if isDefault {
			fmt.Printf("🔹 ID: %d | Название: %s | Встроенный\n", prompt.ID, prompt.Name)
		} else {
			fmt.Printf("🔹 ID: %d | Название: %s\n", prompt.ID, prompt.Name)
		}

		// Описание
		if prompt.Description != "" {
			fmt.Printf("📋 Описание: %s\n", prompt.Description)
		}

		// Содержимое промпта
		fmt.Println("📄 Содержимое:")
		fmt.Println("┌" + strings.Repeat("─", 58) + "┐")

		// Разбиваем содержимое на строки и выводим с отступами
		lines := strings.Split(prompt.Content, "\n")
		for _, line := range lines {
			if full {
				// Полный вывод без обрезки - разбиваем длинные строки
				if len(line) > 56 {
					// Разбиваем длинную строку на части
					for i := 0; i < len(line); i += 56 {
						end := i + 56
						if end > len(line) {
							end = len(line)
						}
						fmt.Printf("│ %-56s │\n", line[i:end])
					}
				} else {
					fmt.Printf("│ %-56s │\n", line)
				}
			} else {
				// Обычный вывод с обрезкой
				fmt.Printf("│ %-56s │\n", truncateString(line, 56))
			}
		}

		fmt.Println("└" + strings.Repeat("─", 58) + "┘")
		fmt.Println()
	}
}

// isDefaultPrompt проверяет, является ли промпт встроенным и неизмененным
func (pm *PromptManager) isDefaultPrompt(prompt SystemPrompt) bool {
	// Используем новую функцию из builtin_prompts.go
	return IsBuiltinPrompt(prompt)
}

// IsDefaultPromptByID проверяет, является ли промпт встроенным только по ID (игнорирует содержимое)
func (pm *PromptManager) IsDefaultPromptByID(prompt SystemPrompt) bool {
	// Проверяем, что ID находится в диапазоне встроенных промптов (1-8)
	return prompt.ID >= 1 && prompt.ID <= 8
}

// GetRussianDefaultPrompts возвращает русские версии встроенных промптов
func GetRussianDefaultPrompts() []SystemPrompt {
	return GetBuiltinPromptsByLanguage("ru")
}

// getDefaultPrompts возвращает оригинальные встроенные промпты
func (pm *PromptManager) GetDefaultPrompts() []SystemPrompt {
	return GetBuiltinPrompts()
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

// GetVerbosePromptByLevel возвращает промпт для подробного объяснения по уровню
func GetVerbosePromptByLevel(level int) string {
	// Создаем PromptManager для получения текущего языка из sys_prompts (без принудительной загрузки дефолтов)
	pm := NewPromptManager("")
	currentLang := pm.GetCurrentLanguage()

	var prompt *SystemPrompt
	switch level {
	case 1:
		prompt = GetBuiltinPromptByIDAndLanguage(6, currentLang) // v
	case 2:
		prompt = GetBuiltinPromptByIDAndLanguage(7, currentLang) // vv
	case 3:
		prompt = GetBuiltinPromptByIDAndLanguage(8, currentLang) // vvv
	default:
		return ""
	}

	if prompt != nil {
		return prompt.Content
	}
	return ""
}
