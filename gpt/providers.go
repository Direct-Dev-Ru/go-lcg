package gpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Provider интерфейс для работы с разными LLM провайдерами
type Provider interface {
	Chat(messages []Chat) (string, error)
	Health() error
}

// ProxyAPIProvider реализация для прокси API (gin-restapi)
type ProxyAPIProvider struct {
	BaseURL    string
	JWTToken   string
	Model      string
	HTTPClient *http.Client
}

// ProxyChatRequest структура запроса к прокси API
type ProxyChatRequest struct {
	Messages       []Chat   `json:"messages"`
	Model          string   `json:"model,omitempty"`
	Temperature    float64  `json:"temperature,omitempty"`
	TopP           float64  `json:"top_p,omitempty"`
	Stream         bool     `json:"stream,omitempty"`
	SystemContent  string   `json:"system_content,omitempty"`
	UserContent    string   `json:"user_content,omitempty"`
	RandomWords    []string `json:"random_words,omitempty"`
	FallbackString string   `json:"fallback_string,omitempty"`
}

// ProxyChatResponse структура ответа от прокси API
type ProxyChatResponse struct {
	Response string `json:"response"`
	Usage    struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Error   string `json:"error,omitempty"`
	Model   string `json:"model,omitempty"`
	Timeout int    `json:"timeout_seconds,omitempty"`
}

// ProxyHealthResponse структура ответа health check
type ProxyHealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Model   string `json:"default_model,omitempty"`
	Timeout int    `json:"default_timeout_seconds,omitempty"`
}

// OllamaProvider реализация для Ollama API
type OllamaProvider struct {
	BaseURL     string
	Model       string
	Temperature float64
	HTTPClient  *http.Client
}

func NewProxyAPIProvider(baseURL, jwtToken, model string) *ProxyAPIProvider {
	return &ProxyAPIProvider{
		BaseURL:    strings.TrimSuffix(baseURL, "/"),
		JWTToken:   jwtToken,
		Model:      model,
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
	}
}

func NewOllamaProvider(baseURL, model string, temperature float64) *OllamaProvider {
	return &OllamaProvider{
		BaseURL:     strings.TrimSuffix(baseURL, "/"),
		Model:       model,
		Temperature: temperature,
		HTTPClient:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Chat для ProxyAPIProvider
func (p *ProxyAPIProvider) Chat(messages []Chat) (string, error) {
	// Используем основной endpoint /api/v1/protected/sberchat/chat
	payload := ProxyChatRequest{
		Messages:    messages,
		Model:       p.Model,
		Temperature: 0.5,
		TopP:        0.5,
		Stream:      false,
		RandomWords: []string{"linux", "command", "gpt"},
		FallbackString: "I'm sorry, I can't help with that. Please try again.",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	req, err := http.NewRequest("POST", p.BaseURL+"/api/v1/protected/sberchat/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if p.JWTToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.JWTToken)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка API: %d - %s", resp.StatusCode, string(body))
	}

	var response ProxyChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	if response.Error != "" {
		return "", fmt.Errorf("ошибка прокси API: %s", response.Error)
	}

	if response.Response == "" {
		return "", fmt.Errorf("пустой ответ от API")
	}

	return strings.TrimSpace(response.Response), nil
}

// Health для ProxyAPIProvider
func (p *ProxyAPIProvider) Health() error {
	req, err := http.NewRequest("GET", p.BaseURL+"/api/v1/protected/sberchat/health", nil)
	if err != nil {
		return fmt.Errorf("ошибка создания health check запроса: %w", err)
	}

	if p.JWTToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.JWTToken)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}

	var healthResponse ProxyHealthResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения health check ответа: %w", err)
	}

	if err := json.Unmarshal(body, &healthResponse); err != nil {
		return fmt.Errorf("ошибка парсинга health check ответа: %w", err)
	}

	if healthResponse.Status != "ok" {
		return fmt.Errorf("health check status: %s - %s", healthResponse.Status, healthResponse.Message)
	}

	return nil
}

// Chat для OllamaProvider
func (o *OllamaProvider) Chat(messages []Chat) (string, error) {
	payload := Gpt3Request{
		Model:    o.Model,
		Messages: messages,
		Stream:   false,
		Options:  Gpt3Options{o.Temperature},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("ошибка маршалинга запроса: %w", err)
	}

	req, err := http.NewRequest("POST", o.BaseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ошибка API: %d - %s", resp.StatusCode, string(body))
	}

	var response OllamaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return strings.TrimSpace(response.Message.Content), nil
}

// Health для OllamaProvider
func (o *OllamaProvider) Health() error {
	req, err := http.NewRequest("GET", o.BaseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("ошибка создания health check запроса: %w", err)
	}

	resp, err := o.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка health check: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}

	return nil
}
