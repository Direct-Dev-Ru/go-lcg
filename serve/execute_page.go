package serve

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
	"github.com/direct-dev-ru/linux-command-gpt/serve/templates"
	"github.com/direct-dev-ru/linux-command-gpt/validation"
	"github.com/russross/blackfriday/v2"
)

// ExecutePageData содержит данные для страницы выполнения
type ExecutePageData struct {
	Title          string
	Header         string
	CurrentPrompt  string
	SystemOptions  []SystemPromptOption
	ResultSection  template.HTML
	VerboseButtons template.HTML
	ActionButtons  template.HTML
	CSRFToken      string
	ProviderType   string
	Model          string
	Host           string
	BasePath       string
	AppName        string
	// Поля конфигурации для валидации
	MaxUserMessageLength int
}

// SystemPromptOption представляет опцию системного промпта
type SystemPromptOption struct {
	ID          int
	Name        string
	Description string
}

// ExecuteResultData содержит результат выполнения
type ExecuteResultData struct {
	Success     bool
	Command     string
	Explanation string
	Error       string
	Model       string
	Elapsed     float64
	Verbose     string
}

// handleExecutePage обрабатывает страницу выполнения
func handleExecutePage(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Показываем форму
		showExecuteForm(w, r)
	case http.MethodPost:
		// Обрабатываем выполнение
		handleExecuteRequest(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// showExecuteForm показывает форму выполнения
func showExecuteForm(w http.ResponseWriter, r *http.Request) {
	// Генерируем CSRF токен
	csrfManager := GetCSRFManager()
	if csrfManager == nil {
		http.Error(w, "CSRF manager not initialized", http.StatusInternalServerError)
		return
	}

	// Получаем сессионный ID
	sessionID := getSessionID(r)
	csrfToken, err := csrfManager.GenerateToken(sessionID)
	if err != nil {
		http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем CSRF токен в cookie
	setCSRFCookie(w, csrfToken)

	// Получаем системные промпты
	pm := gpt.NewPromptManager(config.AppConfig.PromptFolder)

	var systemOptions []SystemPromptOption
	for i := 1; i <= 5; i++ {
		prompt, err := pm.GetPromptByID(i)
		if err == nil {
			systemOptions = append(systemOptions, SystemPromptOption{
				ID:          prompt.ID,
				Name:        prompt.Name,
				Description: prompt.Description,
			})
		}
	}

	data := ExecutePageData{
		Title:                "Выполнение запроса",
		Header:               "Выполнение запроса",
		CurrentPrompt:        "",
		SystemOptions:        systemOptions,
		ResultSection:        template.HTML(""),
		VerboseButtons:       template.HTML(""),
		ActionButtons:        template.HTML(""),
		CSRFToken:            csrfToken,
		ProviderType:         config.AppConfig.ProviderType,
		Model:                config.AppConfig.Model,
		Host:                 config.AppConfig.Host,
		BasePath:             getBasePath(),
		AppName:              config.AppConfig.AppName,
		MaxUserMessageLength: config.AppConfig.Validation.MaxUserMessageLength,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecutePageTemplate.Execute(w, data)
}

// handleExecuteRequest обрабатывает запрос на выполнение
func handleExecuteRequest(w http.ResponseWriter, r *http.Request) {
	// Парсим форму
	prompt := r.FormValue("prompt")
	systemIDStr := r.FormValue("system_id")
	verbose := r.FormValue("verbose")

	// Получаем системные промпты
	pm := gpt.NewPromptManager(config.AppConfig.PromptFolder)

	if prompt == "" {
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	// Валидация длины пользовательского сообщения
	if err := validation.ValidateUserMessage(prompt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	systemID := 1
	if systemIDStr != "" {
		if id, err := strconv.Atoi(systemIDStr); err == nil && id >= 1 && id <= 5 {
			systemID = id
		}
	}

	// Получаем системный промпт
	systemPrompt, err := pm.GetPromptByID(systemID)
	if err != nil {
		http.Error(w, "Failed to get system prompt", http.StatusInternalServerError)
		return
	}

	// Валидация длины системного промпта
	if err := validation.ValidateSystemPrompt(systemPrompt.Content); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем GPT клиент
	gpt3 := gpt.NewGpt3(
		config.AppConfig.ProviderType,
		config.AppConfig.Host,
		config.AppConfig.JwtToken,
		config.AppConfig.Model,
		systemPrompt.Content,
		0.01,
		120,
	)

	// Debug вывод для основного запроса
	PrintWebDebugInfo("EXECUTE", prompt, systemPrompt.Content, config.AppConfig.Model, 120)

	// Выполняем запрос
	response, elapsed := getCommand(*gpt3, prompt)

	var result ExecuteResultData
	if response == "" {
		result = ExecuteResultData{
			Success: false,
			Error:   "Failed to get response from AI",
		}
	} else {
		result = ExecuteResultData{
			Success: true,
			Command: response,
			Model:   config.AppConfig.Model,
			Elapsed: elapsed,
		}
	}

	// Если запрошено подробное объяснение
	if verbose != "" {
		level := len(verbose)
		verbosePrompt := gpt.GetVerbosePromptByLevel(level)

		// Debug вывод для verbose запроса
		PrintWebVerboseDebugInfo("VERBOSE", prompt, verbosePrompt, config.AppConfig.Model, level, 120)

		explanation, err := getDetailedExplanation(prompt, verbose, 120)
		if err == nil {
			// Конвертируем Markdown в HTML
			explanationHTML := blackfriday.Run([]byte(explanation))
			result.Explanation = string(explanationHTML)
			result.Verbose = verbose
		}
	}

	// Получаем системные промпты для dropdown
	var systemOptions []SystemPromptOption
	for i := 1; i <= 5; i++ {
		prompt, err := pm.GetPromptByID(i)
		if err == nil {
			systemOptions = append(systemOptions, SystemPromptOption{
				ID:          prompt.ID,
				Name:        prompt.Name,
				Description: prompt.Description,
			})
		}
	}

	// Генерируем CSRF токен для результата
	csrfManager := GetCSRFManager()
	sessionID := getSessionID(r)
	csrfToken, err := csrfManager.GenerateToken(sessionID)
	if err != nil {
		http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем CSRF токен в cookie после обработки запроса
	setCSRFCookie(w, csrfToken)

	data := ExecutePageData{
		Title:                "Результат выполнения",
		Header:               "Результат выполнения",
		CurrentPrompt:        prompt,
		SystemOptions:        systemOptions,
		ResultSection:        template.HTML(formatResultSection(result)),
		VerboseButtons:       template.HTML(formatVerboseButtons(result)),
		ActionButtons:        template.HTML(formatActionButtons(result)),
		CSRFToken:            csrfToken,
		ProviderType:         config.AppConfig.ProviderType,
		Model:                config.AppConfig.Model,
		Host:                 config.AppConfig.Host,
		BasePath:             getBasePath(),
		AppName:              config.AppConfig.AppName,
		MaxUserMessageLength: config.AppConfig.Validation.MaxUserMessageLength,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.ExecutePageTemplate.Execute(w, data)
}

// formatResultSection форматирует секцию результата
func formatResultSection(result ExecuteResultData) string {
	if !result.Success {
		return fmt.Sprintf(`
			<div class="result-section">
				<div class="error-message">
					<h3>❌ Ошибка</h3>
					<p>%s</p>
				</div>
			</div>`, result.Error)
	}

	explanationSection := ""
	if result.Explanation != "" {
		explanationSection = fmt.Sprintf(`
			<div class="explanation-section">
				<h3>📖 Подробное объяснение (%s):</h3>
				<div class="explanation-content">%s</div>
			</div>
			<script>
				// Показываем кнопку "Наверх" после загрузки объяснения
				showScrollToTopButton();
			</script>`, result.Verbose, result.Explanation)
	}

	// Определяем, содержит ли результат Markdown/многострочный текст
	useMarkdown := false
	if strings.Contains(result.Command, "```") || strings.Contains(result.Command, "\n") || strings.Contains(result.Command, "#") || strings.Contains(result.Command, "*") || strings.Contains(result.Command, "_") {
		useMarkdown = true
	}

	commandBlock := ""
	if useMarkdown {
		// Рендерим Markdown в HTML
		cmdHTML := blackfriday.Run([]byte(result.Command))
		commandBlock = fmt.Sprintf(`<div class="command-md">%s</div>`, string(cmdHTML))
	} else {
		// Оставляем как простой однострочный вывод команды
		commandBlock = fmt.Sprintf(`<div class="command-code">%s</div>`, result.Command)
	}

	return fmt.Sprintf(`
        <div class="result-section">
            <div class="command-result">
                <h3>✅ Команда:</h3>
                %s
                <div class="result-meta">
                    <span>Модель: %s</span>
                    <span>Время: %.2f сек</span>
                </div>
            </div>
            %s
        </div>
        <script>
            // Сохраняем результаты в скрытое поле
            (function() {
                const resultData = {
                    command: %s,
                    explanation: %s,
                    model: %s
                };
                const resultDataField = document.getElementById('resultData');
                if (resultDataField) {
                    resultDataField.value = JSON.stringify(resultData);
                }
            })();
        </script>`,
		commandBlock, result.Model, result.Elapsed, explanationSection,
		fmt.Sprintf(`"%s"`, strings.ReplaceAll(result.Command, `"`, `\"`)),
		fmt.Sprintf(`"%s"`, strings.ReplaceAll(result.Explanation, `"`, `\"`)),
		fmt.Sprintf(`"%s"`, result.Model))
}

// formatVerboseButtons форматирует кнопки подробности
func formatVerboseButtons(result ExecuteResultData) string {
	if !result.Success || result.Explanation != "" {
		return "" // Скрываем кнопки если есть ошибка или уже есть объяснение
	}

	return `
		<div class="verbose-buttons">
			<button onclick="requestExplanation('v')" class="verbose-btn v-btn">v - Краткое объяснение</button>
			<button onclick="requestExplanation('vv')" class="verbose-btn vv-btn">vv - Подробное объяснение</button>
			<button onclick="requestExplanation('vvv')" class="verbose-btn vvv-btn">vvv - Максимально подробное</button>
		</div>`
}

// formatActionButtons форматирует кнопки действий
func formatActionButtons(result ExecuteResultData) string {
	if !result.Success {
		return "" // Скрываем кнопки если есть ошибка
	}

	return `
		<div class="action-buttons">
			<button onclick="saveResult()" class="action-btn">💾 Сохранить результат</button>
			<button onclick="addToHistory()" class="action-btn">📝 Добавить в историю</button>
		</div>`
}
