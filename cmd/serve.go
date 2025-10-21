package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
	"github.com/russross/blackfriday/v2"
)

// StartResultServer запускает HTTP сервер для просмотра сохраненных результатов
func StartResultServer(host, port string) error {
	http.HandleFunc("/", handleResultsPage)
	http.HandleFunc("/file/", handleFileView)
	http.HandleFunc("/delete/", handleDeleteFile)
	http.HandleFunc("/history", handleHistoryPage)
	http.HandleFunc("/history/delete/", handleDeleteHistoryEntry)
	http.HandleFunc("/history/clear", handleClearHistory)
	http.HandleFunc("/prompts", handlePromptsPage)
	http.HandleFunc("/prompts/add", handleAddPrompt)
	http.HandleFunc("/prompts/edit/", handleEditPrompt)
	http.HandleFunc("/prompts/delete/", handleDeletePrompt)
	http.HandleFunc("/prompts/restore/", handleRestorePrompt)
	http.HandleFunc("/prompts/restore-verbose/", handleRestoreVerbosePrompt)
	http.HandleFunc("/prompts/save-lang", handleSaveLang)

	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Сервер запущен на http://%s\n", addr)
	fmt.Println("Нажмите Ctrl+C для остановки")

	return http.ListenAndServe(addr, nil)
}

// handleResultsPage обрабатывает главную страницу со списком файлов
func handleResultsPage(w http.ResponseWriter, r *http.Request) {
	files, err := getResultFiles()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения папки: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LCG Results - Linux Command GPT</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #56ab2f 0%, #a8e6cf 100%);
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2d5016 0%, #4a7c59 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
            font-size: 1.1em;
        }
        .content {
            padding: 30px;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: #f0f8f0;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
            border-left: 4px solid #2d5016;
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            color: #2d5016;
        }
        .stat-label {
            color: #666;
            margin-top: 5px;
        }
        .files-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 20px;
        }
        .file-card {
            background: white;
            border: 1px solid #e1e5e9;
            border-radius: 8px;
            padding: 20px;
            transition: all 0.3s ease;
            position: relative;
        }
        .file-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(45,80,22,0.2);
            border-color: #2d5016;
        }
        .file-card-content {
            cursor: pointer;
        }
        .file-actions {
            position: absolute;
            top: 10px;
            right: 10px;
            display: flex;
            gap: 8px;
        }
        .delete-btn {
            background: #e74c3c;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            transition: background 0.3s ease;
        }
        .delete-btn:hover {
            background: #c0392b;
        }
        .file-name {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            font-size: 1.1em;
        }
        .file-info {
            color: #666;
            font-size: 0.9em;
            margin-bottom: 10px;
        }
        .file-preview {
            background: #f0f8f0;
            padding: 10px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.85em;
            color: #2d5016;
            max-height: 100px;
            overflow: hidden;
            border-left: 3px solid #2d5016;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }
        .empty-state h3 {
            color: #333;
            margin-bottom: 10px;
        }
        .nav-button {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 1em;
            text-decoration: none;
            transition: background 0.3s ease;
            display: inline-block;
            text-align: center;
        }
        .nav-button:hover {
            background: #2980b9;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 LCG Results</h1>
            <p>Просмотр сохраненных результатов Linux Command GPT</p>
        </div>
        <div class="content">
            <div style="display: flex; gap: 10px; margin-bottom: 20px;">
                <button class="nav-button" onclick="location.reload()">🔄 Обновить</button>
                <a href="/history" class="nav-button">📝 История</a>
                <a href="/prompts" class="nav-button">⚙️ Промпты</a>
            </div>
            
            <div class="stats">
                <div class="stat-card">
                    <div class="stat-number">{{.TotalFiles}}</div>
                    <div class="stat-label">Всего файлов</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">{{.RecentFiles}}</div>
                    <div class="stat-label">За последние 7 дней</div>
                </div>
            </div>

            {{if .Files}}
            <div class="files-grid">
                {{range .Files}}
                <div class="file-card">
                    <div class="file-actions">
                        <button class="delete-btn" onclick="deleteFile('{{.Name}}')" title="Удалить файл">🗑️</button>
                    </div>
                    <div class="file-card-content" onclick="window.open('/file/{{.Name}}', '_blank')">
                        <div class="file-name">{{.Name}}</div>
                        <div class="file-info">
                            📅 {{.ModTime}} | 📏 {{.Size}}
                        </div>
                        <div class="file-preview">{{.Preview}}</div>
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="empty-state">
                <h3>📁 Папка пуста</h3>
                <p>Здесь будут отображаться сохраненные результаты после использования команды lcg</p>
            </div>
            {{end}}
        </div>
    </div>
    
    <script>
        function deleteFile(filename) {
            if (confirm('Вы уверены, что хотите удалить файл "' + filename + '"?\\n\\nЭто действие нельзя отменить.')) {
                fetch('/delete/' + encodeURIComponent(filename), {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при удалении файла');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при удалении файла');
                });
            }
        }
    </script>
</body>
</html>`

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

// FileInfo содержит информацию о файле
type FileInfo struct {
	Name    string
	Size    string
	ModTime string
	Preview string
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
		if content, err := os.ReadFile(filepath.Join(config.AppConfig.ResultFolder, entry.Name())); err == nil {
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
	htmlPage := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - LCG Results</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #56ab2f 0%%, #a8e6cf 100%%);
            min-height: 100vh;
        }
        .container {
            max-width: 1000px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2d5016 0%%, #4a7c59 100%%);
            color: white;
            padding: 20px 30px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            margin: 0;
            font-size: 1.5em;
            font-weight: 300;
        }
        .back-btn {
            background: rgba(255,255,255,0.2);
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            text-decoration: none;
            transition: background 0.3s ease;
        }
        .back-btn:hover {
            background: rgba(255,255,255,0.3);
        }
        .content {
            padding: 30px;
            line-height: 1.6;
        }
        .content h1 {
            color: #2d5016;
            border-bottom: 2px solid #2d5016;
            padding-bottom: 10px;
        }
        .content h2 {
            color: #4a7c59;
            margin-top: 30px;
        }
        .content h3 {
            color: #2d5016;
        }
        .content code {
            background: #f0f8f0;
            padding: 2px 6px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            color: #2d5016;
            border: 1px solid #a8e6cf;
        }
        .content pre {
            background: #f0f8f0;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #2d5016;
            overflow-x: auto;
        }
        .content pre code {
            background: none;
            padding: 0;
            border: none;
            color: #2d5016;
        }
        .content blockquote {
            border-left: 4px solid #4a7c59;
            margin: 20px 0;
            padding: 10px 20px;
            background: #f0f8f0;
            border-radius: 0 8px 8px 0;
        }
        .content ul, .content ol {
            padding-left: 20px;
        }
        .content li {
            margin: 5px 0;
        }
        .content strong {
            color: #2d5016;
        }
        .content em {
            color: #4a7c59;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📄 %s</h1>
            <a href="/" class="back-btn">← Назад к списку</a>
        </div>
        <div class="content">
            %s
        </div>
    </div>
</body>
</html>`, filename, filename, string(htmlContent))

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

// handleHistoryPage обрабатывает страницу истории запросов
func handleHistoryPage(w http.ResponseWriter, r *http.Request) {
	historyEntries, err := readHistoryEntries()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка чтения истории: %v", err), http.StatusInternalServerError)
		return
	}

	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>История запросов - LCG Results</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #56ab2f 0%, #a8e6cf 100%);
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2d5016 0%, #4a7c59 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        .content {
            padding: 30px;
        }
        .nav-buttons {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        .nav-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 1em;
            text-decoration: none;
            transition: background 0.3s ease;
            display: inline-block;
            text-align: center;
        }
        .nav-btn:hover {
            background: #2980b9;
        }
        .clear-btn {
            background: #e74c3c;
        }
        .clear-btn:hover {
            background: #c0392b;
        }
        .history-item {
            background: #f0f8f0;
            border: 1px solid #a8e6cf;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 15px;
            position: relative;
        }
        .history-item:hover {
            border-color: #2d5016;
        }
        .history-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .history-index {
            background: #2d5016;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-weight: bold;
        }
        .history-timestamp {
            color: #666;
            font-size: 0.9em;
        }
        .history-command {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
        }
        .history-response {
            background: #f8f9fa;
            padding: 10px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9em;
            color: #2d5016;
            border-left: 3px solid #2d5016;
        }
        .delete-btn {
            background: #e74c3c;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            transition: background 0.3s ease;
        }
        .delete-btn:hover {
            background: #c0392b;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📝 История запросов</h1>
            <p>Управление историей запросов Linux Command GPT</p>
        </div>
        <div class="content">
            <div class="nav-buttons">
                <a href="/" class="nav-btn">🏠 Главная</a>
                <a href="/prompts" class="nav-btn">⚙️ Промпты</a>
                <button class="nav-btn clear-btn" onclick="clearHistory()">🗑️ Очистить всю историю</button>
            </div>

            {{if .Entries}}
            {{range .Entries}}
            <div class="history-item">
                <div class="history-header">
                    <div>
                        <span class="history-index">#{{.Index}}</span>
                        <span class="history-timestamp">{{.Timestamp}}</span>
                    </div>
                    <button class="delete-btn" onclick="deleteHistoryEntry({{.Index}})">🗑️ Удалить</button>
                </div>
                <div class="history-command">{{.Command}}</div>
                <div class="history-response">{{.Response}}</div>
            </div>
            {{end}}
            {{else}}
            <div class="empty-state">
                <h3>📝 История пуста</h3>
                <p>Здесь будут отображаться запросы после использования команды lcg</p>
            </div>
            {{end}}
        </div>
    </div>
    
    <script>
        function deleteHistoryEntry(index) {
            if (confirm('Вы уверены, что хотите удалить запись #' + index + '?')) {
                fetch('/history/delete/' + index, {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при удалении записи');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при удалении записи');
                });
            }
        }
        
        function clearHistory() {
            if (confirm('Вы уверены, что хотите очистить всю историю?\\n\\nЭто действие нельзя отменить.')) {
                fetch('/history/clear', {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при очистке истории');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при очистке истории');
                });
            }
        }
    </script>
</body>
</html>`

	t, err := template.New("history").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	data := struct {
		Entries []HistoryEntryInfo
	}{
		Entries: historyEntries,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// HistoryEntryInfo содержит информацию о записи истории для отображения
type HistoryEntryInfo struct {
	Index     int
	Command   string
	Response  string
	Timestamp string
}

// readHistoryEntries читает записи истории
func readHistoryEntries() ([]HistoryEntryInfo, error) {
	entries, err := read(config.AppConfig.ResultHistory)
	if err != nil {
		return nil, err
	}

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

	indexStr := strings.TrimPrefix(r.URL.Path, "/history/delete/")
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

	tmpl := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Системные промпты - LCG Results</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #56ab2f 0%, #a8e6cf 100%);
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2d5016 0%, #4a7c59 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        .content {
            padding: 30px;
        }
        .nav-buttons {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
        }
        .nav-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 1em;
            text-decoration: none;
            transition: background 0.3s ease;
            display: inline-block;
            text-align: center;
        }
        .nav-btn:hover {
            background: #2980b9;
        }
        .add-btn {
            background: #27ae60;
        }
        .add-btn:hover {
            background: #229954;
        }
        .prompt-item {
            background: #f0f8f0;
            border: 1px solid #a8e6cf;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 15px;
            position: relative;
        }
        .prompt-item:hover {
            border-color: #2d5016;
        }
        .prompt-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .prompt-id {
            background: #2d5016;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-weight: bold;
        }
        .prompt-name {
            font-weight: 600;
            color: #333;
            font-size: 1.2em;
        }
        .prompt-description {
            color: #666;
            margin-bottom: 10px;
        }
        .prompt-content {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9em;
            color: #2d5016;
            border-left: 3px solid #2d5016;
            white-space: pre-wrap;
        }
        .prompt-actions {
            position: absolute;
            top: 10px;
            right: 10px;
            display: flex;
            gap: 8px;
        }
        .action-btn {
            background: #4a7c59;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            transition: background 0.3s ease;
        }
        .action-btn:hover {
            background: #2d5016;
        }
        .delete-btn {
            background: #e74c3c;
        }
        .delete-btn:hover {
            background: #c0392b;
        }
        .restore-btn {
            background: #3498db;
        }
        .restore-btn:hover {
            background: #2980b9;
        }
        .default-badge {
            background: #28a745;
            color: white;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 0.7em;
            margin-left: 8px;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }
        .lang-switcher {
            display: flex;
            gap: 5px;
            margin-left: auto;
        }
        .lang-btn {
            background: #6c757d;
            color: white;
            border: none;
            padding: 8px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9em;
            transition: background 0.3s ease;
        }
        .lang-btn:hover {
            background: #5a6268;
        }
        .lang-btn.active {
            background: #3498db;
        }
        .lang-btn.active:hover {
            background: #2980b9;
        }
        .tabs {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            border-bottom: 2px solid #e9ecef;
        }
        .tab-btn {
            background: #f8f9fa;
            color: #6c757d;
            border: none;
            padding: 12px 20px;
            border-radius: 6px 6px 0 0;
            cursor: pointer;
            font-size: 1em;
            transition: all 0.3s ease;
            border-bottom: 3px solid transparent;
        }
        .tab-btn:hover {
            background: #e9ecef;
            color: #495057;
        }
        .tab-btn.active {
            background: #3498db;
            color: white;
            border-bottom-color: #2980b9;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚙️ Системные промпты</h1>
            <p>Управление системными промптами Linux Command GPT</p>
        </div>
        <div class="content">
            <div class="nav-buttons">
                <a href="/" class="nav-btn">🏠 Главная</a>
                <a href="/history" class="nav-btn">📝 История</a>
                <button class="nav-btn add-btn" onclick="showAddForm()">➕ Добавить промпт</button>
                <div class="lang-switcher">
                    <button class="lang-btn {{if eq .Lang "ru"}}active{{end}}" onclick="switchLang('ru')">🇷🇺 RU</button>
                    <button class="lang-btn {{if eq .Lang "en"}}active{{end}}" onclick="switchLang('en')">🇺🇸 EN</button>
                </div>
            </div>
            
            <!-- Вкладки -->
            <div class="tabs">
                <button class="tab-btn active" onclick="switchTab('system')">⚙️ Системные промпты</button>
                <button class="tab-btn" onclick="switchTab('verbose')">📝 Промпты подробности (v/vv/vvv)</button>
            </div>

            <!-- Вкладка системных промптов -->
            <div id="system-tab" class="tab-content active">
                {{if .Prompts}}
                {{range .Prompts}}
                <div class="prompt-item">
                    <div class="prompt-actions">
                        <button class="action-btn" onclick="editPrompt({{.ID}}, '{{.Name}}', '{{.Description}}', '{{.Content}}')">✏️</button>
                        <button class="action-btn restore-btn" onclick="restorePrompt({{.ID}})" title="Восстановить к значению по умолчанию">🔄</button>
                        <button class="action-btn delete-btn" onclick="deletePrompt({{.ID}})">🗑️</button>
                    </div>
                    <div class="prompt-header">
                        <div>
                            <span class="prompt-id">#{{.ID}}</span>
                            <span class="prompt-name">{{.Name}}</span>
                            {{if .IsDefault}}<span class="default-badge">Встроенный</span>{{end}}
                        </div>
                    </div>
                    <div class="prompt-description">{{.Description}}</div>
                    <div class="prompt-content">{{.Content}}</div>
                </div>
                {{end}}
                {{else}}
                <div class="empty-state">
                    <h3>⚙️ Промпты не найдены</h3>
                    <p>Добавьте пользовательские промпты для настройки поведения системы</p>
                </div>
                {{end}}
            </div>
            
            <!-- Вкладка промптов подробности -->
            <div id="verbose-tab" class="tab-content">
                {{if .VerbosePrompts}}
                {{range .VerbosePrompts}}
                <div class="prompt-item">
                    <div class="prompt-actions">
                        <button class="action-btn" onclick="editVerbosePrompt('{{.Mode}}', '{{.Content}}')">✏️</button>
                        <button class="action-btn restore-btn" onclick="restoreVerbosePrompt('{{.Mode}}')" title="Восстановить к значению по умолчанию">🔄</button>
                    </div>
                    <div class="prompt-header">
                        <div>
                            <span class="prompt-id">#{{.Mode}}</span>
                            <span class="prompt-name">{{.Name}}</span>
                            {{if .IsDefault}}<span class="default-badge">Встроенный</span>{{end}}
                        </div>
                    </div>
                    <div class="prompt-description">{{.Description}}</div>
                    <div class="prompt-content">{{.Content}}</div>
                </div>
                {{end}}
                {{else}}
                <div class="empty-state">
                    <h3>📝 Промпты подробности</h3>
                    <p>Промпты для режимов v, vv, vvv</p>
                </div>
                {{end}}
            </div>
        </div>
    </div>
    
    <!-- Форма добавления/редактирования -->
    <div id="promptForm" style="display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000;">
        <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); background: white; padding: 30px; border-radius: 12px; max-width: 600px; width: 90%;">
            <h3 id="formTitle">Добавить промпт</h3>
            <form id="promptFormData">
                <input type="hidden" id="promptId" name="id">
                <div style="margin-bottom: 15px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 600;">Название:</label>
                    <input type="text" id="promptName" name="name" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;" required>
                </div>
                <div style="margin-bottom: 15px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 600;">Описание:</label>
                    <input type="text" id="promptDescription" name="description" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;" required>
                </div>
                <div style="margin-bottom: 20px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 600;">Содержание:</label>
                    <textarea id="promptContent" name="content" rows="6" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; font-family: monospace;" required></textarea>
                </div>
                <div style="text-align: right;">
                    <button type="button" onclick="hideForm()" style="background: #6c757d; color: white; border: none; padding: 8px 16px; border-radius: 4px; margin-right: 10px; cursor: pointer;">Отмена</button>
                    <button type="submit" style="background: #2d5016; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer;">Сохранить</button>
                </div>
            </form>
        </div>
    </div>
    
    <script>
        function showAddForm() {
            document.getElementById('formTitle').textContent = 'Добавить промпт';
            document.getElementById('promptFormData').reset();
            document.getElementById('promptId').value = '';
            document.getElementById('promptForm').style.display = 'block';
        }
        
        function editPrompt(id, name, description, content) {
            document.getElementById('formTitle').textContent = 'Редактировать промпт';
            document.getElementById('promptId').value = id;
            document.getElementById('promptName').value = name;
            document.getElementById('promptDescription').value = description;
            document.getElementById('promptContent').value = content;
            document.getElementById('promptForm').style.display = 'block';
        }
        
        function hideForm() {
            document.getElementById('promptForm').style.display = 'none';
        }
        
        function switchTab(tabName) {
            // Скрываем все вкладки
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // Убираем активный класс с кнопок
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            
            // Показываем нужную вкладку
            document.getElementById(tabName + '-tab').classList.add('active');
            
            // Активируем нужную кнопку
            event.target.classList.add('active');
        }
        
        function switchLang(lang) {
            // Сохраняем текущие промпты перед переключением языка
            saveCurrentPrompts(lang);
            
            // Перезагружаем страницу с новым языком
            const url = new URL(window.location);
            url.searchParams.set('lang', lang);
            window.location.href = url.toString();
        }
        
        function saveCurrentPrompts(lang) {
            // Отправляем запрос для сохранения текущих промптов с новым языком
            fetch('/prompts/save-lang', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    lang: lang
                })
            })
            .catch(error => {
                console.error('Error saving prompts:', error);
            });
        }
        
        function editVerbosePrompt(mode, content) {
            // Редактирование промпта подробности
            alert('Редактирование промптов подробности будет реализовано');
        }
        
        function deletePrompt(id) {
            if (confirm('Вы уверены, что хотите удалить промпт #' + id + '?')) {
                fetch('/prompts/delete/' + id, {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при удалении промпта');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при удалении промпта');
                });
            }
        }
        
        document.getElementById('promptFormData').addEventListener('submit', function(e) {
            e.preventDefault();
            
            const formData = new FormData(this);
            const id = formData.get('id');
            const url = id ? '/prompts/edit/' + id : '/prompts/add';
            const method = id ? 'PUT' : 'POST';
            
            fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: formData.get('name'),
                    description: formData.get('description'),
                    content: formData.get('content')
                })
            })
            .then(response => {
                if (response.ok) {
                    location.reload();
                } else {
                    alert('Ошибка при сохранении промпта');
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Ошибка при сохранении промпта');
            });
        });

        // Функция восстановления системного промпта
        function restorePrompt(id) {
            if (confirm('Восстановить промпт к значению по умолчанию?')) {
                fetch('/prompts/restore/' + id, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert('Промпт восстановлен');
                        location.reload();
                    } else {
                        alert('Ошибка: ' + data.error);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при восстановлении промпта');
                });
            }
        }

        // Функция восстановления verbose промпта
        function restoreVerbosePrompt(mode) {
            if (confirm('Восстановить промпт к значению по умолчанию?')) {
                fetch('/prompts/restore-verbose/' + mode, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert('Промпт восстановлен');
                        location.reload();
                    } else {
                        alert('Ошибка: ' + data.error);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при восстановлении промпта');
                });
            }
        }
    </script>
</body>
</html>`

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
		Prompts        []PromptWithDefault
		VerbosePrompts []VerbosePrompt
		Lang           string
	}{
		Prompts:        promptsWithDefault,
		VerbosePrompts: verbosePrompts,
		Lang:           lang,
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

	// Обновляем промпт
	if err := pm.UpdatePrompt(id, promptData.Name, promptData.Description, promptData.Content); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка обновления промпта: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Промпт успешно обновлен"))
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

// VerbosePrompt структура для промптов подробности
type VerbosePrompt struct {
	Mode        string
	Name        string
	Description string
	Content     string
	IsDefault   bool
}

// getVerbosePromptsFromFile возвращает промпты подробности из файла sys_prompts
func getVerbosePromptsFromFile(prompts []gpt.SystemPrompt, lang string) []VerbosePrompt {
	var verbosePrompts []VerbosePrompt

	// Ищем промпты подробности в загруженных промптах (ID 6, 7, 8)
	for _, prompt := range prompts {
		if prompt.ID >= 6 && prompt.ID <= 8 {
			// Определяем режим по ID
			var mode string
			switch prompt.ID {
			case 6:
				mode = "v"
			case 7:
				mode = "vv"
			case 8:
				mode = "vvv"
			}

			// Переводим на нужный язык если необходимо
			translatedPrompt := translateVerbosePrompt(prompt, lang)

			verbosePrompts = append(verbosePrompts, VerbosePrompt{
				Mode:        mode,
				Name:        translatedPrompt.Name,
				Description: translatedPrompt.Description,
				Content:     translatedPrompt.Content,
				IsDefault:   gpt.IsBuiltinPrompt(translatedPrompt), // Проверяем, является ли промпт встроенным
			})
		}
	}

	// Если промпты подробности не найдены в файле, используем встроенные
	if len(verbosePrompts) == 0 {
		return getVerbosePrompts(lang)
	}

	return verbosePrompts
}

// translateVerbosePrompt переводит промпт подробности на указанный язык
func translateVerbosePrompt(prompt gpt.SystemPrompt, lang string) gpt.SystemPrompt {
	// Получаем встроенный промпт для указанного языка из YAML
	if builtinPrompt := gpt.GetBuiltinPromptByIDAndLanguage(prompt.ID, lang); builtinPrompt != nil {
		return *builtinPrompt
	}

	// Если перевод не найден, возвращаем оригинал
	return prompt
}

// getVerbosePrompts возвращает промпты для режимов v/vv/vvv (fallback)
func getVerbosePrompts(lang string) []VerbosePrompt {
	// Английские версии (по умолчанию)
	enPrompts := []VerbosePrompt{
		{
			Mode:        "v",
			Name:        "Verbose Mode",
			Description: "Detailed explanation of the command",
			Content:     "Provide a brief explanation of what this Linux command does, including what each flag and option means, and give examples of usage.",
			IsDefault:   true,
		},
		{
			Mode:        "vv",
			Name:        "Very Verbose Mode",
			Description: "Comprehensive explanation with alternatives",
			Content:     "Provide a comprehensive explanation of this Linux command, including detailed descriptions of all flags and options, alternative approaches, common use cases, and potential pitfalls to avoid.",
			IsDefault:   true,
		},
		{
			Mode:        "vvv",
			Name:        "Maximum Verbose Mode",
			Description: "Complete guide with examples and best practices",
			Content:     "Provide a complete guide for this Linux command, including detailed explanations of all options, multiple examples with different scenarios, alternative commands that achieve similar results, best practices, troubleshooting tips, and related commands that work well together.",
			IsDefault:   true,
		},
	}

	// Русские версии
	ruPrompts := []VerbosePrompt{
		{
			Mode:        "v",
			Name:        "Подробный режим",
			Description: "Подробное объяснение команды",
			Content:     "Предоставь краткое объяснение того, что делает эта Linux команда, включая значение каждого флага и опции, и приведи примеры использования.",
			IsDefault:   true,
		},
		{
			Mode:        "vv",
			Name:        "Очень подробный режим",
			Description: "Исчерпывающее объяснение с альтернативами",
			Content:     "Предоставь исчерпывающее объяснение этой Linux команды, включая подробные описания всех флагов и опций, альтернативные подходы, распространенные случаи использования и потенциальные подводные камни, которых следует избегать.",
			IsDefault:   true,
		},
		{
			Mode:        "vvv",
			Name:        "Максимально подробный режим",
			Description: "Полное руководство с примерами и лучшими практиками",
			Content:     "Предоставь полное руководство по этой Linux команде, включая подробные объяснения всех опций, множественные примеры с различными сценариями, альтернативные команды, которые дают аналогичные результаты, лучшие практики, советы по устранению неполадок и связанные команды, которые хорошо работают вместе.",
			IsDefault:   true,
		},
	}

	if lang == "ru" {
		return ruPrompts
	}
	return enPrompts
}

// getSystemPromptsWithLang возвращает системные промпты с учетом языка
func getSystemPromptsWithLang(prompts []gpt.SystemPrompt, lang string) []gpt.SystemPrompt {
	// Если язык английский, возвращаем оригинальные промпты
	if lang == "en" {
		return prompts
	}

	// Для русского языка переводим только встроенные промпты
	var translatedPrompts []gpt.SystemPrompt
	for _, prompt := range prompts {
		// Проверяем, является ли это встроенным промптом
		if gpt.IsBuiltinPrompt(prompt) {
			// Переводим встроенные промпты на русский
			translated := translateSystemPrompt(prompt, lang)
			translatedPrompts = append(translatedPrompts, translated)
		} else {
			translatedPrompts = append(translatedPrompts, prompt)
		}
	}

	return translatedPrompts
}

// translateSystemPrompt переводит системный промпт на указанный язык
func translateSystemPrompt(prompt gpt.SystemPrompt, lang string) gpt.SystemPrompt {
	// Получаем встроенный промпт для указанного языка из YAML
	if builtinPrompt := gpt.GetBuiltinPromptByIDAndLanguage(prompt.ID, lang); builtinPrompt != nil {
		return *builtinPrompt
	}

	// Если перевод не найден, возвращаем оригинал
	return prompt
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
