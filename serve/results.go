package serve

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/serve/templates"
	"github.com/russross/blackfriday/v2"
)

// generateAbbreviation создает аббревиатуру из первых букв слов в названии приложения
func generateAbbreviation(appName string) string {
	words := strings.Fields(appName)
	var abbreviation strings.Builder

	for _, word := range words {
		if len(word) > 0 {
			// Берем первую букву слова, если это буква
			firstRune := []rune(word)[0]
			if unicode.IsLetter(firstRune) {
				abbreviation.WriteRune(unicode.ToUpper(firstRune))
			}
		}
	}

	result := abbreviation.String()
	if result == "" {
		return "LCG" // Fallback если не удалось сгенерировать аббревиатуру
	}
	return result
}

// FileInfo содержит информацию о файле
type FileInfo struct {
	Name        string
	DisplayName string
	Size        string
	ModTime     string
	Preview     template.HTML
	Content     string // Полное содержимое для поиска
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
		Files           []FileInfo
		TotalFiles      int
		RecentFiles     int
		BasePath        string
		AppName         string
		AppAbbreviation string
	}{
		Files:           files,
		TotalFiles:      len(files),
		RecentFiles:     recentCount,
		BasePath:        getBasePath(),
		AppName:         config.AppConfig.AppName,
		AppAbbreviation: generateAbbreviation(config.AppConfig.AppName),
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

		// Читаем превью файла (первые 200 символов) как обычный текст
		preview := ""
		fullContent := ""
		if content, err := os.ReadFile(filepath.Join(config.AppConfig.ResultFolder, entry.Name())); err == nil {
			// Сохраняем полное содержимое для поиска
			fullContent = string(content)

			// Берем первые 200 символов как превью
			preview = string(content)
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
			Name:        entry.Name(),
			DisplayName: formatFileDisplayName(entry.Name()),
			Size:        formatFileSize(info.Size()),
			ModTime:     info.ModTime().Format("02.01.2006 15:04"),
			Preview:     template.HTML(preview),
			Content:     fullContent,
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

// formatFileDisplayName преобразует имя файла вида
// gpt_request_GigaChat-2-Max_2025-10-22_13-50-13.md
// в "Gpt Request GigaChat 2 Max — 2025-10-22 13:50:13"
func formatFileDisplayName(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	// Разделим на части по '_'
	parts := strings.Split(name, "_")
	if len(parts) == 0 {
		return filename
	}

	// Первая часть может быть префиксом gpt/request — заменим '_' на пробел и приведем регистр
	var words []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		// Заменяем '-' на пробел в словах модели/текста
		p = strings.ReplaceAll(p, "-", " ")
		// Разбиваем по пробелам и капитализуем каждое слово
		for _, w := range strings.Fields(p) {
			if w == "" {
				continue
			}
			r := []rune(w)
			r[0] = unicode.ToUpper(r[0])
			words = append(words, string(r))
		}
	}

	// Попробуем распознать хвост как дату и время
	// Ищем шаблон YYYY-MM-DD_HH-MM-SS в исходном имени
	var pretty string
	// ожидаем последние две части — дата и время
	if len(parts) >= 3 {
		datePart := parts[len(parts)-2]
		timePart := parts[len(parts)-1]
		// заменить '-' в времени на ':'
		timePretty := strings.ReplaceAll(timePart, "-", ":")
		if len(datePart) == 10 && len(timePart) == 8 { // примитивная проверка
			// Собираем текст до датных частей
			text := strings.Join(words[:len(words)-2], " ")
			pretty = strings.TrimSpace(text)
			if pretty != "" {
				pretty += " — " + datePart + " " + timePretty
			} else {
				pretty = datePart + " " + timePretty
			}
			return pretty
		}
	}

	if len(words) > 0 {
		pretty = strings.Join(words, " ")
		return pretty
	}
	return filename
}

// handleFileView обрабатывает просмотр конкретного файла
func handleFileView(w http.ResponseWriter, r *http.Request) {
	// Учитываем BasePath при извлечении имени файла
	basePath := config.AppConfig.Server.BasePath
	var filename string
	if basePath != "" && basePath != "/" {
		basePath = strings.TrimSuffix(basePath, "/")
		filename = strings.TrimPrefix(r.URL.Path, basePath+"/file/")
	} else {
		filename = strings.TrimPrefix(r.URL.Path, "/file/")
	}
	if filename == "" {
		renderNotFound(w, "Файл не указан", getBasePath())
		return
	}

	// Проверяем, что файл существует и находится в папке результатов
	filePath := filepath.Join(config.AppConfig.ResultFolder, filename)
	if !strings.HasPrefix(filePath, config.AppConfig.ResultFolder) {
		renderNotFound(w, "Запрошенный файл недоступен", getBasePath())
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		renderNotFound(w, "Файл не найден или был удален", getBasePath())
		return
	}

	// Конвертируем Markdown в HTML
	htmlContent := blackfriday.Run(content)

	// Создаем данные для шаблона
	data := struct {
		Filename string
		Content  template.HTML
		BasePath string
	}{
		Filename: filename,
		Content:  template.HTML(htmlContent),
		BasePath: getBasePath(),
	}

	// Парсим и выполняем шаблон
	tmpl := templates.FileViewTemplate
	t, err := template.New("file_view").Parse(tmpl)
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки для отображения HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// handleDeleteFile обрабатывает удаление файла
func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Учитываем BasePath при извлечении имени файла
	basePath := config.AppConfig.Server.BasePath
	var filename string
	if basePath != "" && basePath != "/" {
		basePath = strings.TrimSuffix(basePath, "/")
		filename = strings.TrimPrefix(r.URL.Path, basePath+"/delete/")
	} else {
		filename = strings.TrimPrefix(r.URL.Path, "/delete/")
	}
	if filename == "" {
		renderNotFound(w, "Файл не указан", getBasePath())
		return
	}

	// Проверяем, что файл существует и находится в папке результатов
	filePath := filepath.Join(config.AppConfig.ResultFolder, filename)
	if !strings.HasPrefix(filePath, config.AppConfig.ResultFolder) {
		renderNotFound(w, "Запрошенный файл недоступен", getBasePath())
		return
	}

	// Проверяем, что файл существует
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		renderNotFound(w, "Файл не найден или уже удален", getBasePath())
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
