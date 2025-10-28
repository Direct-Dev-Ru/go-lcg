package serve

import (
	"net/http"
	"strings"

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

// AuthMiddleware проверяет аутентификацию для всех запросов
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, требуется ли аутентификация
		if !config.AppConfig.Server.RequireAuth {
			next(w, r)
			return
		}

		// Исключаем страницу входа и API логина из проверки (с учетом BasePath)
		if r.URL.Path == makePath("/login") || r.URL.Path == makePath("/api/login") || r.URL.Path == makePath("/api/validate-token") {
			next(w, r)
			return
		}

		// Проверяем аутентификацию
		if !isAuthenticated(r) {
			// Для API запросов возвращаем JSON ошибку
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"success": false, "error": "Authentication required"}`))
				return
			}

			// Для веб-запросов перенаправляем на страницу входа (с учетом BasePath)
			http.Redirect(w, r, makePath("/login"), http.StatusSeeOther)
			return
		}

		// Пользователь аутентифицирован, продолжаем
		next(w, r)
	}
}

// CSRFMiddleware проверяет CSRF токены для POST/PUT/DELETE запросов
func CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем только изменяющие запросы
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next(w, r)
			return
		}

		// Исключаем некоторые API endpoints (с учетом BasePath)
		if r.URL.Path == makePath("/api/login") || r.URL.Path == makePath("/api/logout") {
			next(w, r)
			return
		}

		// Получаем CSRF токен из заголовка или формы
		csrfToken := r.Header.Get("X-CSRF-Token")
		if csrfToken == "" {
			csrfToken = r.FormValue("csrf_token")
		}

		if csrfToken == "" {
			// Для API запросов возвращаем JSON ошибку
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"success": false, "error": "CSRF token required"}`))
				return
			}

			// Для веб-запросов возвращаем ошибку
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}

		// Получаем сессионный ID
		sessionID := getSessionID(r)

		// Проверяем CSRF токен
		csrfManager := GetCSRFManager()
		if csrfManager == nil || !csrfManager.ValidateToken(csrfToken, sessionID) {
			// Для API запросов возвращаем JSON ошибку
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"success": false, "error": "Invalid CSRF token"}`))
				return
			}

			// Для веб-запросов возвращаем ошибку
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		// CSRF токен валиден, продолжаем
		next(w, r)
	}
}

// isAPIRequest проверяет, является ли запрос API запросом
func isAPIRequest(r *http.Request) bool {
	path := r.URL.Path
	apiPrefix := makePath("/api")
	return strings.HasPrefix(path, apiPrefix)
}

// RequireAuth обертка для requireAuth из auth.go
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return requireAuth(next)
}
