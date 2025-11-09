package serve

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/direct-dev-ru/linux-command-gpt/config"
)

const (
	// CSRFTokenLifetimeHours минимальное время жизни CSRF токена в часах (не менее 12 часов)
	CSRFTokenLifetimeHours = 12
	// CSRFTokenLifetimeSeconds минимальное время жизни CSRF токена в секундах
	CSRFTokenLifetimeSeconds = CSRFTokenLifetimeHours * 60 * 60
)

// CSRFManager управляет CSRF токенами
type CSRFManager struct {
	secretKey []byte
}

// CSRFData содержит данные для CSRF токена
type CSRFData struct {
	Token     string
	Timestamp int64
	UserID    string
}

// NewCSRFManager создает новый менеджер CSRF
func NewCSRFManager() (*CSRFManager, error) {
	secret, err := getCSRFSecretKey()
	if err != nil {
		return nil, err
	}
	return &CSRFManager{secretKey: secret}, nil
}

// getCSRFSecretKey получает или генерирует секретный ключ для CSRF
func getCSRFSecretKey() ([]byte, error) {
	// Пытаемся загрузить из переменной окружения
	if secret := os.Getenv("LCG_CSRF_SECRET"); secret != "" {
		return []byte(secret), nil
	}

	// Пытаемся загрузить из файла
	secretFile := fmt.Sprintf("%s/server/csrf_secret", config.AppConfig.Server.ConfigFolder)
	if data, err := os.ReadFile(secretFile); err == nil {
		return data, nil
	}

	// Генерируем новый секретный ключ
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("failed to generate CSRF secret: %v", err)
	}

	// Создаем директорию если не существует
	if err := os.MkdirAll(fmt.Sprintf("%s/server", config.AppConfig.Server.ConfigFolder), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	// Сохраняем секретный ключ в файл
	if err := os.WriteFile(secretFile, secret, 0600); err != nil {
		return nil, fmt.Errorf("failed to save CSRF secret: %v", err)
	}

	return secret, nil
}

// GenerateToken генерирует CSRF токен для пользователя
func (c *CSRFManager) GenerateToken(userID string) (string, error) {
	fmt.Printf("[CSRF DEBUG] Генерация нового токена для UserID: %s\n", userID)

	// Создаем данные токена
	data := CSRFData{
		Token:     generateRandomString(32),
		Timestamp: time.Now().Unix(),
		UserID:    userID,
	}

	fmt.Printf("[CSRF DEBUG] Созданные данные токена: Token (первые 20 символов): %s..., Timestamp: %d, UserID: %s\n",
		safeSubstring(data.Token, 0, 20), data.Timestamp, data.UserID)

	// Создаем подпись
	signature := c.createSignature(data)
	fmt.Printf("[CSRF DEBUG] Созданная подпись (первые 20 символов): %s...\n", safeSubstring(signature, 0, 20))

	// Кодируем данные в base64
	encodedData := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%d:%s", data.Token, data.Timestamp, data.UserID)))
	fmt.Printf("[CSRF DEBUG] Закодированные данные (первые 30 символов): %s...\n", safeSubstring(encodedData, 0, 30))

	token := fmt.Sprintf("%s.%s", encodedData, signature)
	fmt.Printf("[CSRF DEBUG] Итоговый токен сгенерирован (первые 50 символов): %s...\n", safeSubstring(token, 0, 50))

	return token, nil
}

// ValidateToken проверяет CSRF токен
func (c *CSRFManager) ValidateToken(token, userID string) bool {
	fmt.Printf("[CSRF DEBUG] Начало валидации токена. UserID из запроса: %s\n", userID)
	fmt.Printf("[CSRF DEBUG] Токен (первые 50 символов): %s...\n", safeSubstring(token, 0, 50))

	// Разделяем токен на данные и подпись
	parts := splitToken(token)
	if len(parts) != 2 {
		fmt.Printf("[CSRF DEBUG] ❌ ОШИБКА: Токен не может быть разделен на 2 части. Получено частей: %d\n", len(parts))
		return false
	}

	encodedData, signature := parts[0], parts[1]
	fmt.Printf("[CSRF DEBUG] Токен разделен на encodedData (первые 30 символов): %s... и signature (первые 20 символов): %s...\n",
		safeSubstring(encodedData, 0, 30), safeSubstring(signature, 0, 20))

	// Декодируем данные
	dataBytes, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		fmt.Printf("[CSRF DEBUG] ❌ ОШИБКА: Не удалось декодировать base64 данные: %v\n", err)
		return false
	}

	fmt.Printf("[CSRF DEBUG] Данные декодированы. Длина: %d байт\n", len(dataBytes))

	// Парсим данные
	dataParts := splitString(string(dataBytes), ":")
	if len(dataParts) != 3 {
		fmt.Printf("[CSRF DEBUG] ❌ ОШИБКА: Данные не могут быть разделены на 3 части. Получено частей: %d. Данные: %s\n", len(dataParts), string(dataBytes))
		return false
	}

	tokenValue, timestampStr, tokenUserID := dataParts[0], dataParts[1], dataParts[2]
	fmt.Printf("[CSRF DEBUG] Распарсены данные: tokenValue (первые 20 символов): %s..., timestamp: %s, tokenUserID: %s\n",
		safeSubstring(tokenValue, 0, 20), timestampStr, tokenUserID)

	// Проверяем пользователя
	if tokenUserID != userID {
		fmt.Printf("[CSRF DEBUG] ❌ ОШИБКА: UserID не совпадает! Ожидался: '%s', получен из токена: '%s'\n", userID, tokenUserID)
		return false
	}
	fmt.Printf("[CSRF DEBUG] ✅ UserID совпадает: %s\n", userID)

	// Проверяем время жизни токена (минимум 12 часов)
	timestamp, err := parseInt64(timestampStr)
	if err != nil {
		fmt.Printf("[CSRF DEBUG] ❌ ОШИБКА: Не удалось распарсить timestamp '%s': %v\n", timestampStr, err)
		return false
	}

	now := time.Now().Unix()
	age := now - timestamp
	ageHours := float64(age) / 3600.0
	fmt.Printf("[CSRF DEBUG] Текущее время: %d, timestamp токена: %d, возраст токена: %d сек (%.2f часов)\n", now, timestamp, age, ageHours)

	// Минимальное время жизни токена: 12 часов (не менее 12 часов согласно требованиям)
	if age > CSRFTokenLifetimeSeconds {
		fmt.Printf("[CSRF DEBUG] ❌ ОШИБКА: Токен устарел! Возраст: %d сек (%.2f часов), максимум: %d сек (%.2f часов)\n",
			age, ageHours, CSRFTokenLifetimeSeconds, float64(CSRFTokenLifetimeSeconds)/3600.0)
		return false
	}
	fmt.Printf("[CSRF DEBUG] ✅ Токен не устарел (возраст в пределах лимита)\n")

	// Создаем данные для проверки подписи
	data := CSRFData{
		Token:     tokenValue,
		Timestamp: timestamp,
		UserID:    tokenUserID,
	}

	// Проверяем подпись
	expectedSignature := c.createSignature(data)
	signatureMatch := signature == expectedSignature
	if !signatureMatch {
		fmt.Printf("[CSRF DEBUG] ❌ ОШИБКА: Подпись не совпадает!\n")
		fmt.Printf("[CSRF DEBUG] Ожидаемая подпись (первые 20 символов): %s...\n", safeSubstring(expectedSignature, 0, 20))
		fmt.Printf("[CSRF DEBUG] Полученная подпись (первые 20 символов): %s...\n", safeSubstring(signature, 0, 20))
		fmt.Printf("[CSRF DEBUG] Данные для подписи: Token=%s (первые 20), Timestamp=%d, UserID=%s\n",
			safeSubstring(tokenValue, 0, 20), timestamp, tokenUserID)
	} else {
		fmt.Printf("[CSRF DEBUG] ✅ Подпись совпадает\n")
	}
	fmt.Printf("[CSRF DEBUG] Результат валидации: %t\n", signatureMatch)
	return signatureMatch
}

// safeSubstring безопасно обрезает строку
func safeSubstring(s string, start, length int) string {
	if start >= len(s) {
		return ""
	}
	end := start + length
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

// createSignature создает подпись для данных
func (c *CSRFManager) createSignature(data CSRFData) string {
	message := fmt.Sprintf("%s:%d:%s", data.Token, data.Timestamp, data.UserID)
	hash := sha256.Sum256(append(c.secretKey, []byte(message)...))
	return hex.EncodeToString(hash[:])
}

// getTokenFromCookie извлекает CSRF токен из cookie
func GetCSRFTokenFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("csrf_token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// setCSRFCookie устанавливает CSRF токен в cookie
func setCSRFCookie(w http.ResponseWriter, token string) {
	fmt.Printf("[CSRF DEBUG] Установка CSRF cookie. Токен (первые 50 символов): %s...\n", safeSubstring(token, 0, 50))
	fmt.Printf("[CSRF DEBUG] Cookie настройки: Path=%s, Secure=%t, Domain=%s, MaxAge=%d сек\n",
		config.AppConfig.Server.CookiePath,
		config.AppConfig.Server.CookieSecure,
		config.AppConfig.Server.Domain,
		CSRFTokenLifetimeSeconds)

	// Минимальное время жизни токена: 12 часов (не менее 12 часов согласно требованиям)
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		Path:     config.AppConfig.Server.CookiePath,
		HttpOnly: true,
		Secure:   config.AppConfig.Server.CookieSecure,
		SameSite: http.SameSiteLaxMode,     // Более мягкий режим для reverse proxy
		MaxAge:   CSRFTokenLifetimeSeconds, // Минимум 12 часов в секундах
	}

	// Добавляем домен если указан
	if config.AppConfig.Server.Domain != "" {
		cookie.Domain = config.AppConfig.Server.Domain
		fmt.Printf("[CSRF DEBUG] Cookie Domain установлен: %s\n", cookie.Domain)
	} else {
		fmt.Printf("[CSRF DEBUG] Cookie Domain не установлен (пустой)\n")
	}

	http.SetCookie(w, cookie)
	fmt.Printf("[CSRF DEBUG] ✅ CSRF cookie установлен: Name=%s, Path=%s, Domain=%s, Secure=%t, HttpOnly=%t, SameSite=%v, MaxAge=%d\n",
		cookie.Name, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly, cookie.SameSite, cookie.MaxAge)
}

// clearCSRFCookie удаляет CSRF cookie
func СlearCSRFCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     config.AppConfig.Server.CookiePath,
		HttpOnly: true,
		Secure:   config.AppConfig.Server.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	}

	// Добавляем домен если указан
	if config.AppConfig.Server.Domain != "" {
		cookie.Domain = config.AppConfig.Server.Domain
	}

	http.SetCookie(w, cookie)
}

// generateRandomString генерирует случайную строку
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}

// splitToken разделяет токен на части
func splitToken(token string) []string {
	// Ищем последнюю точку
	lastDot := -1
	for i := len(token) - 1; i >= 0; i-- {
		if token[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot == -1 {
		return []string{token}
	}

	return []string{token[:lastDot], token[lastDot+1:]}
}

// splitString разделяет строку по разделителю
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// parseInt64 парсит строку в int64
func parseInt64(s string) (int64, error) {
	var result int64
	for _, char := range s {
		if char < '0' || char > '9' {
			return 0, fmt.Errorf("invalid number: %s", s)
		}
		result = result*10 + int64(char-'0')
	}
	return result, nil
}

// Глобальный экземпляр CSRF менеджера
var csrfManager *CSRFManager

// InitCSRFManager инициализирует глобальный CSRF менеджер
func InitCSRFManager() error {
	var err error
	csrfManager, err = NewCSRFManager()
	return err
}

// GetCSRFManager возвращает глобальный CSRF менеджер
func GetCSRFManager() *CSRFManager {
	return csrfManager
}
