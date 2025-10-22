package serve

import (
	"fmt"
	"net/http"

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

// StartResultServer запускает HTTP сервер для просмотра сохраненных результатов
func StartResultServer(host, port string) error {
	// Регистрируем все маршруты
	registerRoutes()

	addr := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("Сервер запущен на http://%s\n", addr)
	fmt.Println("Нажмите Ctrl+C для остановки")

	// Тестовое логирование для проверки debug флага
	if config.AppConfig.MainFlags.Debug {
		fmt.Printf("🔍 DEBUG РЕЖИМ ВКЛЮЧЕН - веб-операции будут логироваться\n")
	} else {
		fmt.Printf("🔍 DEBUG РЕЖИМ ОТКЛЮЧЕН - веб-операции не будут логироваться\n")
	}

	return http.ListenAndServe(addr, nil)
}

// registerRoutes регистрирует все маршруты сервера
func registerRoutes() {
	// Главная страница и файлы
	http.HandleFunc("/", handleResultsPage)
	http.HandleFunc("/file/", handleFileView)
	http.HandleFunc("/delete/", handleDeleteFile)

	// История запросов
	http.HandleFunc("/history", handleHistoryPage)
	http.HandleFunc("/history/view/", handleHistoryView)
	http.HandleFunc("/history/delete/", handleDeleteHistoryEntry)
	http.HandleFunc("/history/clear", handleClearHistory)

	// Управление промптами
	http.HandleFunc("/prompts", handlePromptsPage)
	http.HandleFunc("/prompts/add", handleAddPrompt)
	http.HandleFunc("/prompts/edit/", handleEditPrompt)
	http.HandleFunc("/prompts/delete/", handleDeletePrompt)
	http.HandleFunc("/prompts/restore/", handleRestorePrompt)
	http.HandleFunc("/prompts/restore-verbose/", handleRestoreVerbosePrompt)
	http.HandleFunc("/prompts/save-lang", handleSaveLang)

	// Веб-страница для выполнения запросов
	http.HandleFunc("/run", handleExecutePage)

	// API для выполнения запросов
	http.HandleFunc("/api/execute", handleExecute)
	// API для сохранения результатов и истории
	http.HandleFunc("/api/save-result", handleSaveResult)
	http.HandleFunc("/api/add-to-history", handleAddToHistory)
}
