package serve

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
	"github.com/direct-dev-ru/linux-command-gpt/serve/templates"
	"github.com/direct-dev-ru/linux-command-gpt/validation"
)

// VerbosePrompt структура для промптов подробности
type VerbosePrompt struct {
	Mode        string
	Name        string
	Description string
	Content     string
	IsDefault   bool
}

// handlePromptsPage обрабатывает страницу управления промптами
func handlePromptsPage(w http.ResponseWriter, r *http.Request) {
	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов (использует конфигурацию из config.AppConfig.PromptFolder)
	pm := gpt.NewPromptManager(homeDir)

	// Получаем язык из параметра запроса, если не указан - берем из файла
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = pm.GetCurrentLanguage()
	}

	tmpl := templates.PromptsPageTemplate

	t, err := template.New("prompts").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	// Создаем структуру с дополнительным полем IsDefault
	type PromptWithDefault struct {
		gpt.SystemPrompt
		IsDefault bool
	}

	// Получаем текущий язык из файла
	currentLang := pm.GetCurrentLanguage()

	// Если язык не указан в URL, используем язык из файла
	if lang == "" {
		lang = currentLang
	}

	// Получаем системные промпты с учетом языка
	systemPrompts := getSystemPromptsWithLang(pm.Prompts, lang)

	var promptsWithDefault []PromptWithDefault
	for _, prompt := range systemPrompts {
		// Показываем только системные промпты (ID 1-5) на первой вкладке
		if prompt.ID >= 1 && prompt.ID <= 5 {
			// Проверяем, является ли промпт встроенным и неизмененным
			isDefault := gpt.IsBuiltinPrompt(prompt)
			promptsWithDefault = append(promptsWithDefault, PromptWithDefault{
				SystemPrompt: prompt,
				IsDefault:    isDefault,
			})
		}
	}

	// Получаем промпты подробности из файла sys_prompts
	verbosePrompts := getVerbosePromptsFromFile(pm.Prompts, lang)

	data := struct {
		Prompts               []PromptWithDefault
		VerbosePrompts        []VerbosePrompt
		Lang                  string
		MaxSystemPromptLength int
		MaxPromptNameLength   int
		MaxPromptDescLength   int
	}{
		Prompts:               promptsWithDefault,
		VerbosePrompts:        verbosePrompts,
		Lang:                  lang,
		MaxSystemPromptLength: config.AppConfig.Validation.MaxSystemPromptLength,
		MaxPromptNameLength:   config.AppConfig.Validation.MaxPromptNameLength,
		MaxPromptDescLength:   config.AppConfig.Validation.MaxPromptDescLength,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// handleAddPrompt обрабатывает добавление нового промпта
func handleAddPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов (использует конфигурацию из config.AppConfig.PromptFolder)
	pm := gpt.NewPromptManager(homeDir)

	// Парсим JSON данные
	var promptData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&promptData); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	// Валидация длины полей
	if err := validation.ValidateSystemPrompt(promptData.Content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidatePromptName(promptData.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidatePromptDescription(promptData.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Добавляем промпт
	if err := pm.AddPrompt(promptData.Name, promptData.Description, promptData.Content); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка добавления промпта: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Промпт успешно добавлен"))
}

// handleEditPrompt обрабатывает редактирование промпта
func handleEditPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/prompts/edit/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID промпта", http.StatusBadRequest)
		return
	}

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов (использует конфигурацию из config.AppConfig.PromptFolder)
	pm := gpt.NewPromptManager(homeDir)

	// Парсим JSON данные
	var promptData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&promptData); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	// Валидация длины полей
	if err := validation.ValidateSystemPrompt(promptData.Content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidatePromptName(promptData.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidatePromptDescription(promptData.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Обновляем промпт
	if err := pm.UpdatePrompt(id, promptData.Name, promptData.Description, promptData.Content); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка обновления промпта: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Промпт успешно обновлен"))
}

// handleEditVerbosePrompt обрабатывает редактирование промпта подробности
func handleEditVerbosePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем режим из URL
	mode := strings.TrimPrefix(r.URL.Path, "/prompts/edit-verbose/")

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов
	pm := gpt.NewPromptManager(homeDir)

	// Парсим JSON данные
	var promptData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&promptData); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	// Валидация длины полей
	if err := validation.ValidateSystemPrompt(promptData.Content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidatePromptName(promptData.Name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validation.ValidatePromptDescription(promptData.Description); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Определяем ID по режиму
	var id int
	switch mode {
	case "v":
		id = 6
	case "vv":
		id = 7
	case "vvv":
		id = 8
	default:
		http.Error(w, "Неверный режим промпта", http.StatusBadRequest)
		return
	}

	// Обновляем промпт
	if err := pm.UpdatePrompt(id, promptData.Name, promptData.Description, promptData.Content); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка обновления промпта: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Промпт подробности успешно обновлен"))
}

// handleDeletePrompt обрабатывает удаление промпта
func handleDeletePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/prompts/delete/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный ID промпта", http.StatusBadRequest)
		return
	}

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов (использует конфигурацию из config.AppConfig.PromptFolder)
	pm := gpt.NewPromptManager(homeDir)

	// Удаляем промпт
	if err := pm.DeletePrompt(id); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка удаления промпта: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Промпт успешно удален"))
}

// handleRestorePrompt восстанавливает системный промпт к значению по умолчанию
func handleRestorePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем ID из URL
	idStr := strings.TrimPrefix(r.URL.Path, "/prompts/restore/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid prompt ID", http.StatusBadRequest)
		return
	}

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов
	pm := gpt.NewPromptManager(homeDir)

	// Получаем текущий язык
	currentLang := pm.GetCurrentLanguage()

	// Получаем встроенный промпт для текущего языка
	builtinPrompt := gpt.GetBuiltinPromptByIDAndLanguage(id, currentLang)
	if builtinPrompt == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Промпт не найден в встроенных",
		})
		return
	}

	// Обновляем промпт в списке
	for i, prompt := range pm.Prompts {
		if prompt.ID == id {
			pm.Prompts[i] = *builtinPrompt
			break
		}
	}

	// Сохраняем изменения
	if err := pm.SaveAllPrompts(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Ошибка сохранения: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleRestoreVerbosePrompt восстанавливает verbose промпт к значению по умолчанию
func handleRestoreVerbosePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем режим из URL
	mode := strings.TrimPrefix(r.URL.Path, "/prompts/restore-verbose/")

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов
	pm := gpt.NewPromptManager(homeDir)

	// Получаем текущий язык
	currentLang := pm.GetCurrentLanguage()

	// Определяем ID по режиму
	var id int
	switch mode {
	case "v":
		id = 6
	case "vv":
		id = 7
	case "vvv":
		id = 8
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Неверный режим промпта",
		})
		return
	}

	// Получаем встроенный промпт для текущего языка
	builtinPrompt := gpt.GetBuiltinPromptByIDAndLanguage(id, currentLang)
	if builtinPrompt == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Промпт не найден в встроенных",
		})
		return
	}

	// Обновляем промпт в списке
	for i, prompt := range pm.Prompts {
		if prompt.ID == id {
			pm.Prompts[i] = *builtinPrompt
			break
		}
	}

	// Сохраняем изменения
	if err := pm.SaveAllPrompts(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Ошибка сохранения: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}

// handleSaveLang обрабатывает сохранение промптов при переключении языка
func handleSaveLang(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем домашнюю директорию пользователя
	homeDir, err := os.UserHomeDir()
	if err != nil {
		http.Error(w, "Ошибка получения домашней директории", http.StatusInternalServerError)
		return
	}

	// Создаем менеджер промптов
	pm := gpt.NewPromptManager(homeDir)

	// Парсим JSON данные
	var langData struct {
		Lang string `json:"lang"`
	}

	if err := json.NewDecoder(r.Body).Decode(&langData); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	// Устанавливаем язык файла
	pm.SetLanguage(langData.Lang)

	// Переводим только встроенные промпты (по ID), а пользовательские оставляем как есть
	var translatedPrompts []gpt.SystemPrompt
	for _, p := range pm.Prompts {
		// Проверяем, является ли промпт встроенным по ID (1-8)
		if pm.IsDefaultPromptByID(p) {
			// System (1-5) и Verbose (6-8)
			if p.ID >= 1 && p.ID <= 5 {
				translatedPrompts = append(translatedPrompts, translateSystemPrompt(p, langData.Lang))
			} else if p.ID >= 6 && p.ID <= 8 {
				translatedPrompts = append(translatedPrompts, translateVerbosePrompt(p, langData.Lang))
			} else {
				translatedPrompts = append(translatedPrompts, p)
			}
		} else {
			// Пользовательские промпты (ID > 8) не трогаем
			translatedPrompts = append(translatedPrompts, p)
		}
	}

	// Обновляем в pm и сохраняем
	pm.Prompts = translatedPrompts
	if err := pm.SaveAllPrompts(); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка сохранения: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Промпты сохранены"))
}
