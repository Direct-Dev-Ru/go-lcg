package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

// SaveResultRequest представляет запрос на сохранение результата
type SaveResultRequest struct {
	Prompt      string `json:"prompt"`
	Command     string `json:"command"`
	Explanation string `json:"explanation,omitempty"`
	Model       string `json:"model"`
}

// SaveResultResponse представляет ответ на сохранение результата
type SaveResultResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	File    string `json:"file,omitempty"`
	Error   string `json:"error,omitempty"`
}

// AddToHistoryRequest представляет запрос на добавление в историю
type AddToHistoryRequest struct {
	Prompt      string `json:"prompt"`
	Command     string `json:"command"`
	Response    string `json:"response"`
	Explanation string `json:"explanation,omitempty"`
	System      string `json:"system"`
}

// AddToHistoryResponse представляет ответ на добавление в историю
type AddToHistoryResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// handleSaveResult обрабатывает сохранение результата
func handleSaveResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SaveResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" || req.Command == "" {
		http.Error(w, "Prompt and command are required", http.StatusBadRequest)
		return
	}

	// Создаем папку результатов если не существует
	if err := os.MkdirAll(config.AppConfig.ResultFolder, 0755); err != nil {
		apiJsonResponse(w, SaveResultResponse{
			Success: false,
			Error:   "Failed to create result folder",
		})
		return
	}

	// Генерируем имя файла
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("gpt_request_%s_%s.md", req.Model, timestamp)
	filePath := path.Join(config.AppConfig.ResultFolder, filename)
	title := truncateTitle(req.Prompt)

	// Формируем содержимое
	var content string
	if strings.TrimSpace(req.Explanation) != "" {
		content = fmt.Sprintf("# %s\n\n## Prompt\n\n%s\n\n## Response\n\n%s\n\n## Explanation\n\n%s\n",
			title, req.Prompt, req.Command, req.Explanation)
	} else {
		content = fmt.Sprintf("# %s\n\n## Prompt\n\n%s\n\n## Response\n\n%s\n",
			title, req.Prompt, req.Command)
	}

	// Сохраняем файл
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		apiJsonResponse(w, SaveResultResponse{
			Success: false,
			Error:   "Failed to save file",
		})
		return
	}

	// Debug вывод для сохранения результата
	PrintWebSaveDebugInfo("SAVE_RESULT", req.Prompt, req.Command, req.Explanation, req.Model, filename)

	apiJsonResponse(w, SaveResultResponse{
		Success: true,
		Message: "Result saved successfully",
		File:    filename,
	})
}

// handleAddToHistory обрабатывает добавление в историю
func handleAddToHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddToHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" || req.Command == "" || req.Response == "" {
		http.Error(w, "Prompt, command and response are required", http.StatusBadRequest)
		return
	}

	// Проверяем, есть ли уже такой запрос в истории
	entries, err := Read(config.AppConfig.ResultHistory)
	if err != nil {
		// Если файл не существует, создаем пустой массив
		entries = []HistoryEntry{}
	}

	// Ищем дубликат
	duplicateIndex := -1
	for i, entry := range entries {
		if strings.EqualFold(strings.TrimSpace(entry.Command), strings.TrimSpace(req.Prompt)) {
			duplicateIndex = i
			break
		}
	}

	// Создаем новую запись
	newEntry := HistoryEntry{
		Index:       len(entries) + 1,
		Command:     req.Prompt,
		Response:    req.Response,
		Explanation: req.Explanation,
		System:      req.System,
		Timestamp:   time.Now(),
	}

	if duplicateIndex == -1 {
		// Добавляем новую запись
		entries = append(entries, newEntry)
	} else {
		// Перезаписываем существующую
		newEntry.Index = entries[duplicateIndex].Index
		entries[duplicateIndex] = newEntry
	}

	// Сохраняем историю
	if err := Write(config.AppConfig.ResultHistory, entries); err != nil {
		apiJsonResponse(w, AddToHistoryResponse{
			Success: false,
			Error:   "Failed to save to history",
		})
		return
	}

	message := "Added to history successfully"
	if duplicateIndex != -1 {
		message = "Updated existing history entry"
	}

	// Debug вывод для добавления в историю
	PrintWebSaveDebugInfo("ADD_TO_HISTORY", req.Prompt, req.Command, req.Explanation, req.System, "")

	apiJsonResponse(w, AddToHistoryResponse{
		Success: true,
		Message: message,
	})
}

// apiJsonResponse отправляет JSON ответ
func apiJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// truncateTitle сокращает строку до 120 символов (по рунам), добавляя " ..." при усечении
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
