package serve

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/ssl"
)

// StartResultServer запускает HTTP/HTTPS сервер для просмотра сохраненных результатов
func StartResultServer(host, port string) error {
	addr := fmt.Sprintf("%s:%s", host, port)

	// Проверяем, нужно ли использовать HTTPS
	useHTTPS := ssl.ShouldUseHTTPS(host)

	if useHTTPS {
		// Регистрируем HTTPS маршруты (включая редирект)
		registerHTTPSRoutes()

		// Создаем директорию для SSL сертификатов
		sslDir := fmt.Sprintf("%s/server/ssl", config.AppConfig.Server.ConfigFolder)
		if err := os.MkdirAll(sslDir, 0755); err != nil {
			return fmt.Errorf("failed to create SSL directory: %v", err)
		}

		// Загружаем или генерируем SSL сертификат
		cert, err := ssl.LoadOrGenerateCert(host)
		if err != nil {
			return fmt.Errorf("failed to load/generate SSL certificate: %v", err)
		}

		// Настраиваем TLS
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{*cert},
			MinVersion:   tls.VersionTLS12,
			MaxVersion:   tls.VersionTLS13,
			// Отключаем проверку клиентских сертификатов
			ClientAuth: tls.NoClientCert,
			// Добавляем логирование для отладки
			GetCertificate: func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
				if config.AppConfig.MainFlags.Debug {
					fmt.Printf("🔍 TLS запрос от %s (SNI: %s)\n", clientHello.Conn.RemoteAddr(), clientHello.ServerName)
				}
				return cert, nil
			},
		}

		// Создаем HTTPS сервер
		server := &http.Server{
			Addr:      addr,
			TLSConfig: tlsConfig,
		}

		fmt.Printf("🔒 Сервер запущен на https://%s (SSL включен)\n", addr)
		fmt.Println("Нажмите Ctrl+C для остановки")

		// Тестовое логирование для проверки debug флага
		if config.AppConfig.MainFlags.Debug {
			fmt.Printf("🔍 DEBUG РЕЖИМ ВКЛЮЧЕН - веб-операции будут логироваться\n")
		} else {
			fmt.Printf("🔍 DEBUG РЕЖИМ ОТКЛЮЧЕН - веб-операции не будут логироваться\n")
		}

		return server.ListenAndServeTLS("", "")
	} else {
		// Регистрируем обычные маршруты для HTTP
		registerRoutes()

		fmt.Printf("🌐 Сервер запущен на http://%s (HTTP режим)\n", addr)
		fmt.Println("Нажмите Ctrl+C для остановки")

		// Тестовое логирование для проверки debug флага
		if config.AppConfig.MainFlags.Debug {
			fmt.Printf("🔍 DEBUG РЕЖИМ ВКЛЮЧЕН - веб-операции будут логироваться\n")
		} else {
			fmt.Printf("🔍 DEBUG РЕЖИМ ОТКЛЮЧЕН - веб-операции не будут логироваться\n")
		}

		return http.ListenAndServe(addr, nil)
	}
}

// handleHTTPSRedirect обрабатывает редирект с HTTP на HTTPS
func handleHTTPSRedirect(w http.ResponseWriter, r *http.Request) {
	// Определяем протокол и хост
	host := r.Host
	if host == "" {
		host = r.Header.Get("Host")
	}

	// Редиректим на HTTPS
	httpsURL := fmt.Sprintf("https://%s%s", host, r.RequestURI)
	http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}

// registerHTTPSRoutes регистрирует маршруты для HTTPS сервера
func registerHTTPSRoutes() {
	// Регистрируем все маршруты кроме главной страницы
	registerRoutesExceptHome()

	// Регистрируем главную страницу с проверкой HTTPS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, пришел ли запрос по HTTP (не HTTPS)
		if r.TLS == nil {
			handleHTTPSRedirect(w, r)
			return
		}
		// Если уже HTTPS, обрабатываем как обычно
		handleResultsPage(w, r)
	})
}

// registerRoutesExceptHome регистрирует все маршруты кроме главной страницы
func registerRoutesExceptHome() {
	// Файлы
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
	http.HandleFunc("/prompts/edit-verbose/", handleEditVerbosePrompt)
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
	http.HandleFunc("/prompts/edit-verbose/", handleEditVerbosePrompt)
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
