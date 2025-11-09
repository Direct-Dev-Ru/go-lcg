package serve

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims представляет claims для JWT токена
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthRequest представляет запрос на аутентификацию
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse представляет ответ на аутентификацию
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// JWTSecretKey генерирует или загружает секретный ключ для JWT
func getJWTSecretKey() ([]byte, error) {
	// Пытаемся загрузить из переменной окружения
	if secret := os.Getenv("LCG_JWT_SECRET"); secret != "" {
		return []byte(secret), nil
	}

	// Пытаемся загрузить из файла
	secretFile := fmt.Sprintf("%s/server/jwt_secret", config.AppConfig.Server.ConfigFolder)
	if data, err := os.ReadFile(secretFile); err == nil {
		return data, nil
	}

	// Генерируем новый секретный ключ
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("failed to generate JWT secret: %v", err)
	}

	// Создаем директорию если не существует
	if err := os.MkdirAll(fmt.Sprintf("%s/server", config.AppConfig.Server.ConfigFolder), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	// Сохраняем секретный ключ в файл
	if err := os.WriteFile(secretFile, secret, 0600); err != nil {
		return nil, fmt.Errorf("failed to save JWT secret: %v", err)
	}

	return secret, nil
}

// generateJWTToken создает JWT токен для пользователя
func generateJWTToken(username string) (string, error) {
	secret, err := getJWTSecretKey()
	if err != nil {
		return "", err
	}

	// Создаем claims
	claims := JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Токен действителен 24 часа
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "lcg-server",
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// validateJWTToken проверяет JWT токен
func validateJWTToken(tokenString string) (*JWTClaims, error) {
	secret, err := getJWTSecretKey()
	if err != nil {
		return nil, err
	}

	// Парсим токен
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// getTokenFromCookie извлекает JWT токен из cookies
func getTokenFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// setAuthCookie устанавливает HTTP-only cookie с JWT токеном
func setAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     config.AppConfig.Server.CookiePath,
		HttpOnly: true,
		Secure:   config.AppConfig.Server.CookieSecure,
		SameSite: http.SameSiteLaxMode, // Более мягкий режим для reverse proxy
		MaxAge:   config.AppConfig.Server.CookieTTLHours * 60 * 60,
	}

	// Добавляем домен если указан
	if config.AppConfig.Server.Domain != "" {
		cookie.Domain = config.AppConfig.Server.Domain
	}

	http.SetCookie(w, cookie)
}

// clearAuthCookie удаляет cookie с токеном
func clearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     config.AppConfig.Server.CookiePath,
		HttpOnly: true,
		Secure:   config.AppConfig.Server.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // Удаляем cookie
	}

	// Добавляем домен если указан
	if config.AppConfig.Server.Domain != "" {
		cookie.Domain = config.AppConfig.Server.Domain
	}

	http.SetCookie(w, cookie)
}

// handleLogin обрабатывает запрос на вход
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiJsonResponse(w, AuthResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	// Проверяем пароль
	if req.Password != config.AppConfig.Server.Password {
		apiJsonResponse(w, AuthResponse{
			Success: false,
			Error:   "Неверный пароль",
		})
		return
	}

	// Генерируем JWT токен
	token, err := generateJWTToken(req.Username)
	if err != nil {
		apiJsonResponse(w, AuthResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	// Устанавливаем cookie
	setAuthCookie(w, token)

	apiJsonResponse(w, AuthResponse{
		Success: true,
		Message: "Успешная авторизация",
	})
}

// handleLogout обрабатывает запрос на выход
func handleLogout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleValidateToken обрабатывает проверку валидности токена
func handleValidateToken(w http.ResponseWriter, r *http.Request) {
	token, err := getTokenFromCookie(r)
	if err != nil {
		apiJsonResponse(w, AuthResponse{
			Success: false,
			Error:   "Token not found",
		})
		return
	}

	_, err = validateJWTToken(token)
	if err != nil {
		apiJsonResponse(w, AuthResponse{
			Success: false,
			Error:   "Invalid token",
		})
		return
	}

	apiJsonResponse(w, AuthResponse{
		Success: true,
		Message: "Token is valid",
	})
}

// requireAuth middleware проверяет аутентификацию
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, требуется ли аутентификация
		if !config.AppConfig.Server.RequireAuth {
			next(w, r)
			return
		}

		// Получаем токен из cookie
		token, err := getTokenFromCookie(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Проверяем валидность токена
		_, err = validateJWTToken(token)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Токен валиден, продолжаем
		next(w, r)
	}
}
