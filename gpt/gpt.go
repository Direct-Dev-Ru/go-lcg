package gpt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProxySimpleChatRequest структура для простого запроса
type ProxySimpleChatRequest struct {
	Message string `json:"message"`
	Model   string `json:"model,omitempty"`
}

// ProxySimpleChatResponse структура ответа для простого запроса
type ProxySimpleChatResponse struct {
	Response string `json:"response"`
	Usage    struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Model   string `json:"model,omitempty"`
	Timeout int    `json:"timeout_seconds,omitempty"`
}

// Gpt3 обновленная структура с поддержкой разных провайдеров
type Gpt3 struct {
	Provider     Provider
	Prompt       string
	Model        string
	HomeDir      string
	ApiKeyFile   string
	ApiKey       string
	Temperature  float64
	ProviderType string // "ollama", "proxy"
}

type Chat struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Gpt3Request struct {
	Model    string      `json:"model"`
	Stream   bool        `json:"stream"`	
	Messages []Chat      `json:"messages"`
	Options  Gpt3Options `json:"options"`
}

type Gpt3ThinkRequest struct {
	Model  string `json:"model"`
	Stream bool   `json:"stream"`
	Think  bool   `json:"think"`
	Messages []Chat `json:"messages"`
	Options  Gpt3Options `json:"options"`
}

type Gpt3Options struct {
	Temperature float64 `json:"temperature"`
}

type Gpt3Response struct {
	Choices []struct {
		Message Chat `json:"message"`
	} `json:"choices"`
}

// LlamaResponse represents the response structure.
type OllamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Message            Chat   `json:"message"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int64  `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int64  `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
}

func (gpt3 *Gpt3) deleteApiKey() {
	filePath := gpt3.HomeDir + string(filepath.Separator) + gpt3.ApiKeyFile
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	}
	err := os.Remove(filePath)
	if err != nil {
		panic(err)
	}
}

func (gpt3 *Gpt3) updateApiKey(apiKey string) {
	filePath := gpt3.HomeDir + string(filepath.Separator) + gpt3.ApiKeyFile
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	apiKey = strings.TrimSpace(apiKey)
	_, err = file.WriteString(apiKey)
	if err != nil {
		panic(err)
	}
	gpt3.ApiKey = apiKey
}

func (gpt3 *Gpt3) storeApiKey(apiKey string) {
	if apiKey == "" {
		return
	}
	filePath := gpt3.HomeDir + string(filepath.Separator) + gpt3.ApiKeyFile
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	apiKey = strings.TrimSpace(apiKey)
	_, err = file.WriteString(apiKey)
	if err != nil {
		panic(err)
	}
	gpt3.ApiKey = apiKey
}

func (gpt3 *Gpt3) loadApiKey() bool {
	dirSeparator := string(filepath.Separator)
	apiKeyFile := gpt3.HomeDir + dirSeparator + gpt3.ApiKeyFile
	if _, err := os.Stat(apiKeyFile); os.IsNotExist(err) {
		return false
	}
	apiKey, err := os.ReadFile(apiKeyFile)
	if err != nil {
		return false
	}
	gpt3.ApiKey = string(apiKey)

	return true
}

func (gpt3 *Gpt3) UpdateKey() {
	var apiKey string
	fmt.Print("OpenAI API Key: ")
	fmt.Scanln(&apiKey)
	gpt3.updateApiKey(apiKey)
}

func (gpt3 *Gpt3) DeleteKey() {
	var c string
	fmt.Print("Are you sure you want to delete the API key? (y/N): ")
	fmt.Scanln(&c)
	if c == "Y" || c == "y" {
		gpt3.deleteApiKey()
	}
}

func (gpt3 *Gpt3) InitKey() {
	// Для ollama и proxy провайдеров не нужен API ключ
	if gpt3.ProviderType == "ollama" || gpt3.ProviderType == "proxy" {
		return
	}

	load := gpt3.loadApiKey()
	if load {
		return
	}
	var apiKey string
	fmt.Print("OpenAI API Key: ")
	fmt.Scanln(&apiKey)
	gpt3.storeApiKey(apiKey)
}

// NewGpt3 создает новый экземпляр GPT с выбранным провайдером
func NewGpt3(providerType, host, apiKey, model, prompt string, temperature float64, timeout int) *Gpt3 {
	var provider Provider

	switch providerType {
	case "proxy":
		provider = NewProxyAPIProvider(host, apiKey, model, timeout) // apiKey используется как JWT токен
	case "ollama":
		provider = NewOllamaProvider(host, model, temperature, timeout)
	default:
		provider = NewOllamaProvider(host, model, temperature, timeout)
	}

	return &Gpt3{
		Provider:     provider,
		Prompt:       prompt,
		Model:        model,
		ApiKey:       apiKey,
		Temperature:  temperature,
		ProviderType: providerType,
	}
}

// Completions обновленный метод с поддержкой разных провайдеров
func (gpt3 *Gpt3) Completions(ask string) string {
	messages := []Chat{
		{"system", gpt3.Prompt},
		{"user", ask + ". " + gpt3.Prompt},
	}

	response, err := gpt3.Provider.Chat(messages)
	if err != nil {
		fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
		return ""
	}

	return response
}

// Health проверяет состояние провайдера
func (gpt3 *Gpt3) Health() error {
	return gpt3.Provider.Health()
}

// GetAvailableModels возвращает список доступных моделей
func (gpt3 *Gpt3) GetAvailableModels() ([]string, error) {
	return gpt3.Provider.GetAvailableModels()
}
