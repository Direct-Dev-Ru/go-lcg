package serve

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/serve/templates"
	"github.com/russross/blackfriday/v2"
)

// HistoryEntryInfo содержит информацию о записи истории для отображения
type HistoryEntryInfo struct {
	Index     int
	Command   string
	Response  string
	Timestamp string
}

// handleHistoryPage обрабатывает страницу истории запросов
func handleHistoryPage(w http.ResponseWriter, r *http.Request) {
	historyEntries, err := readHistoryEntries()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения истории: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl := templates.HistoryPageTemplate

	t, err := template.New("history").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	data := struct {
		Entries  []HistoryEntryInfo
		BasePath string
		AppName  string
	}{
		Entries:  historyEntries,
		BasePath: getBasePath(),
		AppName:  config.AppConfig.AppName,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// readHistoryEntries читает записи истории
func readHistoryEntries() ([]HistoryEntryInfo, error) {
	entries, err := Read(config.AppConfig.ResultHistory)
	if err != nil {
		return nil, err
	}

	// Сортируем записи по времени в убывающем порядке (новые сначала)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	var result []HistoryEntryInfo
	for _, entry := range entries {
		result = append(result, HistoryEntryInfo{
			Index:     entry.Index,
			Command:   entry.Command,
			Response:  entry.Response,
			Timestamp: entry.Timestamp.Format("02.01.2006 15:04:05"),
		})
	}

	return result, nil
}

// handleDeleteHistoryEntry обрабатывает удаление записи истории
func handleDeleteHistoryEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Убираем BasePath из URL перед извлечением индекса
	basePath := config.AppConfig.Server.BasePath
	var indexStr string
	if basePath != "" && basePath != "/" {
		basePath = strings.TrimSuffix(basePath, "/")
		indexStr = strings.TrimPrefix(r.URL.Path, basePath+"/history/delete/")
	} else {
		indexStr = strings.TrimPrefix(r.URL.Path, "/history/delete/")
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid index", http.StatusBadRequest)
		return
	}

	err = DeleteHistoryEntry(config.AppConfig.ResultHistory, index)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка удаления: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Запись успешно удалена"))
}

// handleClearHistory обрабатывает очистку всей истории
func handleClearHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := os.WriteFile(config.AppConfig.ResultHistory, []byte("[]"), 0644)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка очистки: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("История успешно очищена"))
}

// handleHistoryView обрабатывает просмотр записи истории
func handleHistoryView(w http.ResponseWriter, r *http.Request) {
	// Получаем индекс из URL, учитывая BasePath
	basePath := config.AppConfig.Server.BasePath
	var indexStr string
	if basePath != "" && basePath != "/" {
		basePath = strings.TrimSuffix(basePath, "/")
		indexStr = strings.TrimPrefix(r.URL.Path, basePath+"/history/view/")
	} else {
		indexStr = strings.TrimPrefix(r.URL.Path, "/history/view/")
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Читаем записи истории
	entries, err := Read(config.AppConfig.ResultHistory)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения истории: %v", err), http.StatusInternalServerError)
		return
	}

	// Ищем запись с нужным индексом
	var targetEntry *HistoryEntry
	for _, entry := range entries {
		if entry.Index == index {
			targetEntry = &entry
			break
		}
	}

	if targetEntry == nil {
		http.NotFound(w, r)
		return
	}

	// Формируем объяснение, если оно есть
	explanationSection := ""
	if strings.TrimSpace(targetEntry.Explanation) != "" {
		// Конвертируем Markdown в HTML
		explanationHTML := blackfriday.Run([]byte(targetEntry.Explanation))
		explanationSection = fmt.Sprintf(`
			<div class="history-explanation">
				<h3>📖 Подробное объяснение:</h3>
				<div class="history-explanation-content">%s</div>
			</div>`, string(explanationHTML))
	}

	// Создаем данные для шаблона
	data := struct {
		Index           int
		Timestamp       string
		Command         string
		Response        string
		ExplanationHTML template.HTML
		BasePath        string
	}{
		Index:           index,
		Timestamp:       targetEntry.Timestamp.Format("02.01.2006 15:04:05"),
		Command:         targetEntry.Command,
		Response:        targetEntry.Response,
		ExplanationHTML: template.HTML(explanationSection),
		BasePath:        getBasePath(),
	}

	// Парсим и выполняем шаблон
	tmpl := templates.HistoryViewTemplate
	t, err := template.New("history_view").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}
