package serve

import (
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/serve/templates"
)

// handleLoginPage обрабатывает страницу входа
func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	// Если пользователь уже авторизован, перенаправляем на главную
	if isAuthenticated(r) {
		http.Redirect(w, r, makePath("/"), http.StatusSeeOther)
		return
	}

	// Генерируем CSRF токен
	csrfManager := GetCSRFManager()
	if csrfManager == nil {
		http.Error(w, "CSRF manager not initialized", http.StatusInternalServerError)
		return
	}

	// Для неавторизованных пользователей используем сессионный ID
	sessionID := getSessionID(r)
	csrfToken, err := csrfManager.GenerateToken(sessionID)
	if err != nil {
		http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
		return
	}

	// Устанавливаем CSRF токен в cookie
	setCSRFCookie(w, csrfToken)

	data := LoginPageData{
		Title:     "Авторизация - LCG",
		Message:   "",
		Error:     "",
		CSRFToken: csrfToken,
		BasePath:  getBasePath(),
	}

	if err := RenderLoginPage(w, data); err != nil {
		http.Error(w, "Failed to render login page", http.StatusInternalServerError)
		return
	}
}

// isAuthenticated проверяет, авторизован ли пользователь
func isAuthenticated(r *http.Request) bool {
	// Проверяем, требуется ли аутентификация
	if !config.AppConfig.Server.RequireAuth {
		return true
	}

	// Получаем токен из cookie
	token, err := getTokenFromCookie(r)
	if err != nil {
		return false
	}

	// Проверяем валидность токена
	_, err = validateJWTToken(token)
	return err == nil
}

// LoginPageData представляет данные для страницы входа
type LoginPageData struct {
	Title     string
	Message   string
	Error     string
	CSRFToken string
	BasePath  string
}

// RenderLoginPage рендерит страницу входа
func RenderLoginPage(w http.ResponseWriter, data LoginPageData) error {
	tmpl, err := template.New("login").Parse(templates.LoginPageTemplate)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

// getSessionID получает или создает сессионный ID для пользователя
func getSessionID(r *http.Request) string {
	// Пытаемся получить из cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}

	// Если нет cookie, генерируем новый ID на основе IP и User-Agent
	ip := r.RemoteAddr
	userAgent := r.Header.Get("User-Agent")

	// Создаем простой хеш для сессии
	hash := sha256.Sum256([]byte(ip + userAgent))
	return hex.EncodeToString(hash[:])[:16]
}
