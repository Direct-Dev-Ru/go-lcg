package serve

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/ssl"
)

// makePath создает путь с учетом BasePath
func makePath(path string) string {
	basePath := config.AppConfig.Server.BasePath
	if basePath == "" || basePath == "/" {
		return path
	}

	// Убираем слэш в конце basePath если есть
	basePath = strings.TrimSuffix(basePath, "/")

	// Убираем слэш в начале path если есть
	path = strings.TrimPrefix(path, "/")

	// Если path пустой, возвращаем basePath с слэшем в конце
	if path == "" {
		return basePath + "/"
	}

	return basePath + "/" + path
}

// getBasePath возвращает BasePath для использования в шаблонах
func getBasePath() string {
	basePath := config.AppConfig.Server.BasePath
	if basePath == "" || basePath == "/" {
		return ""
	}
	return strings.TrimSuffix(basePath, "/")
}

// StartResultServer запускает HTTP/HTTPS сервер для просмотра сохраненных результатов
func StartResultServer(host, port string) error {
	// Инициализируем CSRF менеджер
	if err := InitCSRFManager(); err != nil {
		return fmt.Errorf("failed to initialize CSRF manager: %v", err)
	}

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
	http.HandleFunc(makePath("/"), func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, пришел ли запрос по HTTP (не HTTPS)
		if r.TLS == nil {
			handleHTTPSRedirect(w, r)
			return
		}
		// Если уже HTTPS, обрабатываем как обычно
		AuthMiddleware(handleResultsPage)(w, r)
	})

	// Регистрируем главную страницу без слэша в конце для BasePath
	basePath := config.AppConfig.Server.BasePath
	if basePath != "" && basePath != "/" {
		basePath = strings.TrimSuffix(basePath, "/")
		http.HandleFunc(basePath, func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, пришел ли запрос по HTTP (не HTTPS)
			if r.TLS == nil {
				handleHTTPSRedirect(w, r)
				return
			}
			// Если уже HTTPS, обрабатываем как обычно
			AuthMiddleware(handleResultsPage)(w, r)
		})
	}
}

// registerRoutesExceptHome регистрирует все маршруты кроме главной страницы
func registerRoutesExceptHome() {
	// Страница входа (без аутентификации)
	http.HandleFunc(makePath("/login"), handleLoginPage)

	// API для аутентификации (без аутентификации)
	http.HandleFunc(makePath("/api/login"), handleLogin)
	http.HandleFunc(makePath("/api/logout"), handleLogout)
	http.HandleFunc(makePath("/api/validate-token"), handleValidateToken)

	// Файлы
	http.HandleFunc(makePath("/file/"), AuthMiddleware(handleFileView))
	http.HandleFunc(makePath("/delete/"), AuthMiddleware(handleDeleteFile))

	// История запросов
	http.HandleFunc(makePath("/history"), AuthMiddleware(handleHistoryPage))
	http.HandleFunc(makePath("/history/view/"), AuthMiddleware(handleHistoryView))
	http.HandleFunc(makePath("/history/delete/"), AuthMiddleware(handleDeleteHistoryEntry))
	http.HandleFunc(makePath("/history/clear"), AuthMiddleware(handleClearHistory))

	// Управление промптами
	http.HandleFunc(makePath("/prompts"), AuthMiddleware(handlePromptsPage))
	http.HandleFunc(makePath("/prompts/add"), AuthMiddleware(handleAddPrompt))
	http.HandleFunc(makePath("/prompts/edit/"), AuthMiddleware(handleEditPrompt))
	http.HandleFunc(makePath("/prompts/edit-verbose/"), AuthMiddleware(handleEditVerbosePrompt))
	http.HandleFunc(makePath("/prompts/delete/"), AuthMiddleware(handleDeletePrompt))
	http.HandleFunc(makePath("/prompts/restore/"), AuthMiddleware(handleRestorePrompt))
	http.HandleFunc(makePath("/prompts/restore-verbose/"), AuthMiddleware(handleRestoreVerbosePrompt))
	http.HandleFunc(makePath("/prompts/save-lang"), AuthMiddleware(handleSaveLang))

	// Веб-страница для выполнения запросов
	http.HandleFunc(makePath("/run"), AuthMiddleware(CSRFMiddleware(handleExecutePage)))

	// API для выполнения запросов
	http.HandleFunc(makePath("/api/execute"), AuthMiddleware(CSRFMiddleware(handleExecute)))
	// API для сохранения результатов и истории
	http.HandleFunc(makePath("/api/save-result"), AuthMiddleware(CSRFMiddleware(handleSaveResult)))
	http.HandleFunc(makePath("/api/add-to-history"), AuthMiddleware(CSRFMiddleware(handleAddToHistory)))
}

// registerRoutes регистрирует все маршруты сервера
func registerRoutes() {
	// Страница входа (без аутентификации)
	http.HandleFunc(makePath("/login"), handleLoginPage)

	// API для аутентификации (без аутентификации)
	http.HandleFunc(makePath("/api/login"), handleLogin)
	http.HandleFunc(makePath("/api/logout"), handleLogout)
	http.HandleFunc(makePath("/api/validate-token"), handleValidateToken)

	// Главная страница и файлы
	http.HandleFunc(makePath("/"), AuthMiddleware(handleResultsPage))
	http.HandleFunc(makePath("/file/"), AuthMiddleware(handleFileView))
	http.HandleFunc(makePath("/delete/"), AuthMiddleware(handleDeleteFile))

	// История запросов
	http.HandleFunc(makePath("/history"), AuthMiddleware(handleHistoryPage))
	http.HandleFunc(makePath("/history/view/"), AuthMiddleware(handleHistoryView))
	http.HandleFunc(makePath("/history/delete/"), AuthMiddleware(handleDeleteHistoryEntry))
	http.HandleFunc(makePath("/history/clear"), AuthMiddleware(handleClearHistory))

	// Управление промптами
	http.HandleFunc(makePath("/prompts"), AuthMiddleware(handlePromptsPage))
	http.HandleFunc(makePath("/prompts/add"), AuthMiddleware(handleAddPrompt))
	http.HandleFunc(makePath("/prompts/edit/"), AuthMiddleware(handleEditPrompt))
	http.HandleFunc(makePath("/prompts/edit-verbose/"), AuthMiddleware(handleEditVerbosePrompt))
	http.HandleFunc(makePath("/prompts/delete/"), AuthMiddleware(handleDeletePrompt))
	http.HandleFunc(makePath("/prompts/restore/"), AuthMiddleware(handleRestorePrompt))
	http.HandleFunc(makePath("/prompts/restore-verbose/"), AuthMiddleware(handleRestoreVerbosePrompt))
	http.HandleFunc(makePath("/prompts/save-lang"), AuthMiddleware(handleSaveLang))

	// Веб-страница для выполнения запросов
	http.HandleFunc(makePath("/run"), AuthMiddleware(CSRFMiddleware(handleExecutePage)))

	// API для выполнения запросов
	http.HandleFunc(makePath("/api/execute"), AuthMiddleware(CSRFMiddleware(handleExecute)))
	// API для сохранения результатов и истории
	http.HandleFunc(makePath("/api/save-result"), AuthMiddleware(CSRFMiddleware(handleSaveResult)))
	http.HandleFunc(makePath("/api/add-to-history"), AuthMiddleware(CSRFMiddleware(handleAddToHistory)))

	// Регистрируем главную страницу без слэша в конце для BasePath
	basePath := config.AppConfig.Server.BasePath
	if basePath != "" && basePath != "/" {
		basePath = strings.TrimSuffix(basePath, "/")
		http.HandleFunc(basePath, AuthMiddleware(handleResultsPage))
	}
}
