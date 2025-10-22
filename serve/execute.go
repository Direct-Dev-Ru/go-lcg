package serve

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
)

// ExecuteRequest представляет запрос на выполнение
type ExecuteRequest struct {
	Prompt     string `json:"prompt"`    // Пользовательский промпт
	SystemID   int    `json:"system_id"` // ID системного промпта (1-5)
	SystemText string `json:"system"`    // Текст системного промпта (альтернатива system_id)
	Verbose    string `json:"verbose"`   // Степень подробности: "v", "vv", "vvv" или пустая строка
	Timeout    int    `json:"timeout"`   // Таймаут в секундах (опционально)
}

// ExecuteResponse представляет ответ
type ExecuteResponse struct {
	Success     bool    `json:"success"`
	Command     string  `json:"command,omitempty"`
	Explanation string  `json:"explanation,omitempty"`
	Error       string  `json:"error,omitempty"`
	Model       string  `json:"model,omitempty"`
	Elapsed     float64 `json:"elapsed,omitempty"`
}

// handleExecute обрабатывает POST запросы на выполнение
func handleExecute(w http.ResponseWriter, r *http.Request) {
	// Проверяем User-Agent - только curl
	userAgent := r.Header.Get("User-Agent")
	if !strings.Contains(strings.ToLower(userAgent), "curl") {
		http.Error(w, "Only curl requests are allowed", http.StatusForbidden)
		return
	}

	// Проверяем метод
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Парсим JSON
	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация обязательных полей
	if req.Prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	// Определяем системный промпт
	systemPrompt := ""
	if req.SystemText != "" {
		systemPrompt = req.SystemText
	} else if req.SystemID > 0 && req.SystemID <= 5 {
		// Получаем системный промпт по ID
		pm := gpt.NewPromptManager(config.AppConfig.PromptFolder)
		prompt, err := pm.GetPromptByID(req.SystemID)
		if err != nil {
			http.Error(w, "Failed to get system prompt", http.StatusInternalServerError)
			return
		}
		systemPrompt = prompt.Content
	} else {
		// Используем промпт по умолчанию
		systemPrompt = config.AppConfig.Prompt
	}

	// Устанавливаем таймаут
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 120 // По умолчанию 2 минуты
	}

	// Создаем GPT клиент
	gpt3 := gpt.NewGpt3(
		config.AppConfig.ProviderType,
		config.AppConfig.Host,
		config.AppConfig.JwtToken,
		config.AppConfig.Model,
		systemPrompt,
		0.01,
		timeout,
	)

	// Выполняем запрос
	response, elapsed := getCommand(*gpt3, req.Prompt)
	if response == "" {
		jsonResponse(w, ExecuteResponse{
			Success: false,
			Error:   "Failed to get response from AI",
		})
		return
	}

	// Если запрошено подробное объяснение
	if req.Verbose != "" {
		explanation, err := getDetailedExplanation(req.Prompt, req.Verbose, timeout)
		if err != nil {
			jsonResponse(w, ExecuteResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to get explanation: %v", err),
			})
			return
		}

		jsonResponse(w, ExecuteResponse{
			Success:     true,
			Command:     response,
			Explanation: explanation,
			Model:       config.AppConfig.Model,
			Elapsed:     elapsed,
		})
	} else {
		jsonResponse(w, ExecuteResponse{
			Success: true,
			Command: response,
			Model:   config.AppConfig.Model,
			Elapsed: elapsed,
		})
	}
}

// getCommand выполняет запрос к AI
func getCommand(gpt3 gpt.Gpt3, prompt string) (string, float64) {
	gpt3.InitKey()
	start := time.Now()
	response := gpt3.Completions(prompt)
	elapsed := time.Since(start).Seconds()
	return response, elapsed
}

// getDetailedExplanation получает подробное объяснение
func getDetailedExplanation(prompt, verbose string, timeout int) (string, error) {
	level := len(verbose) // 1, 2, 3

	// Получаем системный промпт для подробного объяснения
	detailedSystem := gpt.GetVerbosePromptByLevel(level)
	if detailedSystem == "" {
		return "", fmt.Errorf("invalid verbose level: %s", verbose)
	}

	// Создаем GPT клиент для объяснения
	explanationGpt := gpt.NewGpt3(
		config.AppConfig.ProviderType,
		config.AppConfig.Host,
		config.AppConfig.JwtToken,
		config.AppConfig.Model,
		detailedSystem,
		0.2,
		timeout,
	)

	explanationGpt.InitKey()
	explanation := explanationGpt.Completions(prompt)

	if explanation == "" {
		return "", fmt.Errorf("failed to get explanation")
	}

	return explanation, nil
}

// jsonResponse отправляет JSON ответ
func jsonResponse(w http.ResponseWriter, response ExecuteResponse) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
