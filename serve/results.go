package serve

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/serve/templates"
	"github.com/russross/blackfriday/v2"
)

// FileInfo содержит информацию о файле
type FileInfo struct {
	Name    string
	Size    string
	ModTime string
	Preview string
	Content string // Полное содержимое для поиска
}

// handleResultsPage обрабатывает главную страницу со списком файлов
func handleResultsPage(w http.ResponseWriter, r *http.Request) {
	files, err := getResultFiles()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения папки: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl := templates.ResultsPageTemplate

	t, err := template.New("results").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	// Подсчитываем статистику
	recentCount := 0
	weekAgo := time.Now().AddDate(0, 0, -7)
	for _, file := range files {
		// Парсим время из строки для сравнения
		if modTime, err := time.Parse("02.01.2006 15:04", file.ModTime); err == nil {
			if modTime.After(weekAgo) {
				recentCount++
			}
		}
	}

	data := struct {
		Files       []FileInfo
		TotalFiles  int
		RecentFiles int
	}{
		Files:       files,
		TotalFiles:  len(files),
		RecentFiles: recentCount,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// getResultFiles возвращает список файлов из папки результатов
func getResultFiles() ([]FileInfo, error) {
	entries, err := os.ReadDir(config.AppConfig.ResultFolder)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Читаем превью файла (первые 200 символов) и конвертируем Markdown
		preview := ""
		fullContent := ""
		if content, err := os.ReadFile(filepath.Join(config.AppConfig.ResultFolder, entry.Name())); err == nil {
			// Сохраняем полное содержимое для поиска
			fullContent = string(content)
			// Конвертируем Markdown в HTML для превью
			htmlContent := blackfriday.Run(content)
			preview = strings.TrimSpace(string(htmlContent))
			// Удаляем HTML теги для превью
			preview = strings.ReplaceAll(preview, "<h1>", "")
			preview = strings.ReplaceAll(preview, "</h1>", "")
			preview = strings.ReplaceAll(preview, "<h2>", "")
			preview = strings.ReplaceAll(preview, "</h2>", "")
			preview = strings.ReplaceAll(preview, "<h3>", "")
			preview = strings.ReplaceAll(preview, "</h3>", "")
			preview = strings.ReplaceAll(preview, "<p>", "")
			preview = strings.ReplaceAll(preview, "</p>", "")
			preview = strings.ReplaceAll(preview, "<code>", "")
			preview = strings.ReplaceAll(preview, "</code>", "")
			preview = strings.ReplaceAll(preview, "<pre>", "")
			preview = strings.ReplaceAll(preview, "</pre>", "")
			preview = strings.ReplaceAll(preview, "<strong>", "")
			preview = strings.ReplaceAll(preview, "</strong>", "")
			preview = strings.ReplaceAll(preview, "<em>", "")
			preview = strings.ReplaceAll(preview, "</em>", "")
			preview = strings.ReplaceAll(preview, "<ul>", "")
			preview = strings.ReplaceAll(preview, "</ul>", "")
			preview = strings.ReplaceAll(preview, "<li>", "• ")
			preview = strings.ReplaceAll(preview, "</li>", "")
			preview = strings.ReplaceAll(preview, "<ol>", "")
			preview = strings.ReplaceAll(preview, "</ol>", "")
			preview = strings.ReplaceAll(preview, "<blockquote>", "")
			preview = strings.ReplaceAll(preview, "</blockquote>", "")
			preview = strings.ReplaceAll(preview, "<br>", "")
			preview = strings.ReplaceAll(preview, "<br/>", "")
			preview = strings.ReplaceAll(preview, "<br />", "")

			// Очищаем от лишних пробелов и переносов
			preview = strings.ReplaceAll(preview, "\n", " ")
			preview = strings.ReplaceAll(preview, "\r", "")
			preview = strings.ReplaceAll(preview, "  ", " ")
			preview = strings.TrimSpace(preview)

			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    formatFileSize(info.Size()),
			ModTime: info.ModTime().Format("02.01.2006 15:04"),
			Preview: preview,
			Content: fullContent,
		})
	}

	// Сортируем по времени изменения (новые сверху)
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[i].ModTime < files[j].ModTime {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	return files, nil
}

// formatFileSize форматирует размер файла в читаемый вид
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// handleFileView обрабатывает просмотр конкретного файла
func handleFileView(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/file/")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Проверяем, что файл существует и находится в папке результатов
	filePath := filepath.Join(config.AppConfig.ResultFolder, filename)
	if !strings.HasPrefix(filePath, config.AppConfig.ResultFolder) {
		http.NotFound(w, r)
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Конвертируем Markdown в HTML
	htmlContent := blackfriday.Run(content)

	// Создаем HTML страницу с красивым отображением
	htmlPage := fmt.Sprintf(templates.FileViewTemplate, filename, filename, string(htmlContent))

	// Устанавливаем заголовки для отображения HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlPage))
}

// handleDeleteFile обрабатывает удаление файла
func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/delete/")
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Проверяем, что файл существует и находится в папке результатов
	filePath := filepath.Join(config.AppConfig.ResultFolder, filename)
	if !strings.HasPrefix(filePath, config.AppConfig.ResultFolder) {
		http.NotFound(w, r)
		return
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Удаляем файл
	err := os.Remove(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка удаления файла: %v", err), http.StatusInternalServerError)
		return
	}

	// Возвращаем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Файл успешно удален"))
}
