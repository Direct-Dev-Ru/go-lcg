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
		// Проверяем, нужно ли пропустить CSRF проверку
		if config.AppConfig.Server.ForceNoCSRF {
			csrfDebugPrint("[CSRF MIDDLEWARE] ⚠️  CSRF проверка отключена через LCG_FORCE_NO_CSRF\n")
			next(w, r)
			return
		}

		csrfDebugPrint("\n[CSRF MIDDLEWARE] ==========================================\n")
		csrfDebugPrint("[CSRF MIDDLEWARE] Обработка запроса: %s %s\n", r.Method, r.URL.Path)
		csrfDebugPrint("[CSRF MIDDLEWARE] RemoteAddr: %s\n", r.RemoteAddr)
		csrfDebugPrint("[CSRF MIDDLEWARE] Host: %s\n", r.Host)

		// Выводим все заголовки
		csrfDebugPrint("[CSRF MIDDLEWARE] Заголовки:\n")
		for name, values := range r.Header {
			if name == "Cookie" {
				// Cookie выводим отдельно, разбирая их
				csrfDebugPrint("[CSRF MIDDLEWARE]   %s: %s\n", name, strings.Join(values, "; "))
			} else {
				csrfDebugPrint("[CSRF MIDDLEWARE]   %s: %s\n", name, strings.Join(values, ", "))
			}
		}

		// Выводим все cookies
		csrfDebugPrint("[CSRF MIDDLEWARE] Все cookies:\n")
		if len(r.Cookies()) == 0 {
			csrfDebugPrint("[CSRF MIDDLEWARE]   (нет cookies)\n")
		} else {
			for _, cookie := range r.Cookies() {
				csrfDebugPrint("[CSRF MIDDLEWARE]   %s = %s (Path: %s, Domain: %s, Secure: %t, HttpOnly: %t, SameSite: %v, MaxAge: %d)\n",
					cookie.Name,
					safeSubstring(cookie.Value, 0, 50),
					cookie.Path,
					cookie.Domain,
					cookie.Secure,
					cookie.HttpOnly,
					cookie.SameSite,
					cookie.MaxAge)
			}
		}

		// Проверяем только изменяющие запросы
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			csrfDebugPrint("[CSRF MIDDLEWARE] Пропускаем проверку CSRF для метода %s\n", r.Method)
			next(w, r)
			return
		}

		// Исключаем некоторые API endpoints (с учетом BasePath)
		if r.URL.Path == makePath("/api/login") || r.URL.Path == makePath("/api/logout") {
			csrfDebugPrint("[CSRF MIDDLEWARE] Пропускаем проверку CSRF для пути %s\n", r.URL.Path)
			next(w, r)
			return
		}

		// Получаем CSRF токен из заголовка или формы
		csrfTokenFromHeader := r.Header.Get("X-CSRF-Token")
		csrfTokenFromForm := r.FormValue("csrf_token")

		csrfDebugPrint("[CSRF MIDDLEWARE] CSRF токен из заголовка X-CSRF-Token: %s\n",
			safeSubstring(csrfTokenFromHeader, 0, 50))
		csrfDebugPrint("[CSRF MIDDLEWARE] CSRF токен из формы csrf_token: %s\n",
			safeSubstring(csrfTokenFromForm, 0, 50))

		csrfToken := csrfTokenFromHeader
		if csrfToken == "" {
			csrfToken = csrfTokenFromForm
		}

		if csrfToken == "" {
			csrfDebugPrint("[CSRF MIDDLEWARE] ❌ ОШИБКА: CSRF токен не найден ни в заголовке, ни в форме!\n")
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

		csrfDebugPrint("[CSRF MIDDLEWARE] Используемый CSRF токен (первые 50 символов): %s...\n",
			safeSubstring(csrfToken, 0, 50))

		// Получаем сессионный ID
		sessionID := getSessionID(r)
		csrfDebugPrint("[CSRF MIDDLEWARE] SessionID: %s\n", sessionID)

		// Получаем CSRF токен из cookie для сравнения
		csrfTokenFromCookie := GetCSRFTokenFromCookie(r)
		valid := true
		if csrfTokenFromCookie != "" {
			csrfDebugPrint("[CSRF MIDDLEWARE] CSRF токен из cookie (первые 50 символов): %s...\n",
				safeSubstring(csrfTokenFromCookie, 0, 50))
			if csrfTokenFromCookie != csrfToken {
				csrfDebugPrint("[CSRF MIDDLEWARE] ⚠️  ВНИМАНИЕ: Токен из cookie отличается от токена в запросе!\n")
				valid = false
			} else {
				csrfDebugPrint("[CSRF MIDDLEWARE] ✅ Токен из cookie совпадает с токеном в запросе\n")
				valid = true
			}
		} else {
			csrfDebugPrint("[CSRF MIDDLEWARE] ⚠️  ВНИМАНИЕ: CSRF токен не найден в cookie!\n")
			valid = false
		}

		// Проверяем CSRF токен

		if !valid {
			csrfDebugPrint("[CSRF MIDDLEWARE] ❌ ОШИБКА: Валидация CSRF токена не прошла!\n")
			// Для API запросов возвращаем JSON ошибку
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"success": false, "error": "Invalid CSRF token"}`))
				return
			}

			// Для веб-запросов возвращаем ошибку
			http.Error(w, "Invalid OR Empty CSRF token", http.StatusForbidden)
			return
		}

		csrfManager := GetCSRFManager()
		if csrfManager == nil {
			csrfDebugPrint("[CSRF MIDDLEWARE] ❌ ОШИБКА: CSRF менеджер не инициализирован!\n")
			if isAPIRequest(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"success": false, "error": "Invalid CSRF token"}`))
				return
			}
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		csrfDebugPrint("[CSRF MIDDLEWARE] Вызов ValidateToken с токеном и sessionID: %s\n", sessionID)
		valid = csrfManager.ValidateToken(csrfToken, sessionID)
		csrfDebugPrint("[CSRF MIDDLEWARE] Результат ValidateToken: %t\n", valid)

		if !valid {
			csrfDebugPrint("[CSRF MIDDLEWARE] ❌ ОШИБКА: Валидация CSRF токена не прошла!\n")
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

		csrfDebugPrint("[CSRF MIDDLEWARE] ✅ CSRF токен валиден, продолжаем обработку запроса\n")
		csrfDebugPrint("[CSRF MIDDLEWARE] ==========================================\n\n")
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
