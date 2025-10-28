package config

import (
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
)

type Config struct {
	Cwd            string
	Host           string
	ProxyUrl       string
	AppName        string
	Completions    string
	Model          string
	Prompt         string
	ApiKeyFile     string
	ConfigFolder   string
	ResultFolder   string
	PromptFolder   string
	ProviderType   string
	JwtToken       string
	PromptID       string
	Timeout        string
	ResultHistory  string
	NoHistoryEnv   string
	AllowExecution bool
	MainFlags      MainFlags
	Server         ServerConfig
	Validation     ValidationConfig
}

type MainFlags struct {
	File      string
	NoHistory bool
	Sys       string
	PromptID  int
	Timeout   int
	Debug     bool
}

type ServerConfig struct {
	Port           string
	Host           string
	HealthUrl      string
	ProxyUrl       string
	BasePath       string
	ConfigFolder   string
	AllowHTTP      bool
	SSLCertFile    string
	SSLKeyFile     string
	RequireAuth    bool
	Password       string
	Domain         string
	CookieSecure   bool
	CookiePath     string
	CookieTTLHours int
}

type ValidationConfig struct {
	MaxSystemPromptLength int
	MaxUserMessageLength  int
	MaxPromptNameLength   int
	MaxPromptDescLength   int
	MaxCommandLength      int
	MaxExplanationLength  int
}

func GetEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getServerAllowHTTP() bool {
	// Если переменная явно установлена, используем её
	if value, exists := os.LookupEnv("LCG_SERVER_ALLOW_HTTP"); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}

	// Если переменная не установлена, определяем по умолчанию на основе хоста
	host := getEnv("LCG_SERVER_HOST", "localhost")
	return isSecureHost(host)
}

func isSecureHost(host string) bool {
	secureHosts := []string{"localhost", "127.0.0.1", "::1"}
	return slices.Contains(secureHosts, host)
}

func Load() Config {
	cwd, _ := os.Getwd()

	homedir, err := os.UserHomeDir()
	if err != nil {
		homedir = cwd
	}
	privateResultsDir := path.Join(homedir, ".config", "lcg", "gpt_results")
	os.MkdirAll(privateResultsDir, 0700)
	resultFolder := getEnv("LCG_RESULT_FOLDER", privateResultsDir)

	privatePromptsDir := path.Join(homedir, ".config", "lcg", "gpt_sys_prompts")
	os.MkdirAll(privatePromptsDir, 0700)
	promptFolder := getEnv("LCG_PROMPT_FOLDER", privatePromptsDir)

	privateConfigDir := path.Join(homedir, ".config", "lcg", "config")
	os.MkdirAll(privateConfigDir, 0700)
	configFolder := getEnv("LCG_CONFIG_FOLDER", privateConfigDir)

	return Config{
		Cwd:            cwd,
		AppName:        getEnv("LCG_APP_NAME", "Linux Command GPT"),
		Host:           getEnv("LCG_HOST", "http://192.168.87.108:11434/"),
		ProxyUrl:       getEnv("LCG_PROXY_URL", "/api/v1/protected/sberchat/chat"),
		Completions:    getEnv("LCG_COMPLETIONS_PATH", "api/chat"),
		Model:          getEnv("LCG_MODEL", "hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M"),
		Prompt:         getEnv("LCG_PROMPT", "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols."),
		ApiKeyFile:     getEnv("LCG_API_KEY_FILE", ".openai_api_key"),
		ResultFolder:   resultFolder,
		PromptFolder:   promptFolder,
		ConfigFolder:   configFolder,
		ProviderType:   getEnv("LCG_PROVIDER", "ollama"),
		JwtToken:       getEnv("LCG_JWT_TOKEN", ""),
		PromptID:       getEnv("LCG_PROMPT_ID", "1"),
		Timeout:        getEnv("LCG_TIMEOUT", "300"),
		ResultHistory:  getEnv("LCG_RESULT_HISTORY", path.Join(resultFolder, "lcg_history.json")),
		NoHistoryEnv:   getEnv("LCG_NO_HISTORY", ""),
		AllowExecution: isAllowExecutionEnabled(),
		Server: ServerConfig{
			Port:           getEnv("LCG_SERVER_PORT", "8080"),
			Host:           getEnv("LCG_SERVER_HOST", "localhost"),
			ConfigFolder:   getEnv("LCG_CONFIG_FOLDER", path.Join(homedir, ".config", "lcg", "config")),
			AllowHTTP:      getServerAllowHTTP(),
			SSLCertFile:    getEnv("LCG_SERVER_SSL_CERT_FILE", ""),
			SSLKeyFile:     getEnv("LCG_SERVER_SSL_KEY_FILE", ""),
			RequireAuth:    isServerRequireAuth(),
			Password:       getEnv("LCG_SERVER_PASSWORD", "admin#123456"),
			Domain:         getEnv("LCG_DOMAIN", getEnv("LCG_SERVER_HOST", "localhost")),
			CookieSecure:   isCookieSecure(),
			CookiePath:     getEnv("LCG_COOKIE_PATH", "/lcg"),
			CookieTTLHours: getEnvInt("LCG_COOKIE_TTL_HOURS", 168), // 7 дней по умолчанию
			BasePath:       getEnv("LCG_BASE_URL", "/lcg"),
			HealthUrl:      getEnv("LCG_HEALTH_URL", "/api/v1/protected/sberchat/health"),
			ProxyUrl:       getEnv("LCG_PROXY_URL", "/api/v1/protected/sberchat/chat"),
		},
		Validation: ValidationConfig{
			MaxSystemPromptLength: getEnvInt("LCG_MAX_SYSTEM_PROMPT_LENGTH", 2000),
			MaxUserMessageLength:  getEnvInt("LCG_MAX_USER_MESSAGE_LENGTH", 4000),
			MaxPromptNameLength:   getEnvInt("LCG_MAX_PROMPT_NAME_LENGTH", 2000),
			MaxPromptDescLength:   getEnvInt("LCG_MAX_PROMPT_DESC_LENGTH", 5000),
			MaxCommandLength:      getEnvInt("LCG_MAX_COMMAND_LENGTH", 8000),
			MaxExplanationLength:  getEnvInt("LCG_MAX_EXPLANATION_LENGTH", 20000),
		},
	}
}

func (c Config) IsNoHistoryEnabled() bool {
	v := strings.TrimSpace(c.NoHistoryEnv)
	if v == "" {
		return false
	}
	vLower := strings.ToLower(v)
	return vLower == "1" || vLower == "true"
}

func isAllowExecutionEnabled() bool {
	v := strings.TrimSpace(getEnv("LCG_ALLOW_EXECUTION", ""))
	if v == "" {
		return false
	}
	vLower := strings.ToLower(v)
	return vLower == "1" || vLower == "true"
}

func isServerRequireAuth() bool {
	v := strings.TrimSpace(getEnv("LCG_SERVER_REQUIRE_AUTH", ""))
	if v == "" {
		return false
	}
	vLower := strings.ToLower(v)
	return vLower == "1" || vLower == "true"
}

func isCookieSecure() bool {
	v := strings.TrimSpace(getEnv("LCG_COOKIE_SECURE", ""))
	if v == "" {
		return false
	}
	vLower := strings.ToLower(v)
	return vLower == "1" || vLower == "true"
}

var AppConfig Config

func init() {
	AppConfig = Load()
}
