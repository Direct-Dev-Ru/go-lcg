package config

import (
	"os"
	"path"
	"strings"
)

type Config struct {
	Cwd            string
	Host           string
	ProxyUrl       string
	Completions    string
	Model          string
	Prompt         string
	ApiKeyFile     string
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
	Port string
	Host string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func Load() Config {
	cwd, _ := os.Getwd()

	homedir, err := os.UserHomeDir()
	if err != nil {
		homedir = cwd
	}
	os.MkdirAll(path.Join(homedir, ".config", "lcg", "gpt_results"), 0755)
	resultFolder := getEnv("LCG_RESULT_FOLDER", path.Join(homedir, ".config", "lcg", "gpt_results"))

	os.MkdirAll(path.Join(homedir, ".config", "lcg", "gpt_sys_prompts"), 0755)
	promptFolder := getEnv("LCG_PROMPT_FOLDER", path.Join(homedir, ".config", "lcg", "gpt_sys_prompts"))

	return Config{
		Cwd:            cwd,
		Host:           getEnv("LCG_HOST", "http://192.168.87.108:11434/"),
		ProxyUrl:       getEnv("LCG_PROXY_URL", "/api/v1/protected/sberchat/chat"),
		Completions:    getEnv("LCG_COMPLETIONS_PATH", "api/chat"),
		Model:          getEnv("LCG_MODEL", "hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M"),
		Prompt:         getEnv("LCG_PROMPT", "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols."),
		ApiKeyFile:     getEnv("LCG_API_KEY_FILE", ".openai_api_key"),
		ResultFolder:   resultFolder,
		PromptFolder:   promptFolder,
		ProviderType:   getEnv("LCG_PROVIDER", "ollama"),
		JwtToken:       getEnv("LCG_JWT_TOKEN", ""),
		PromptID:       getEnv("LCG_PROMPT_ID", "1"),
		Timeout:        getEnv("LCG_TIMEOUT", "300"),
		ResultHistory:  getEnv("LCG_RESULT_HISTORY", path.Join(resultFolder, "lcg_history.json")),
		NoHistoryEnv:   getEnv("LCG_NO_HISTORY", ""),
		AllowExecution: isAllowExecutionEnabled(),
		Server: ServerConfig{
			Port: getEnv("LCG_SERVER_PORT", "8080"),
			Host: getEnv("LCG_SERVER_HOST", "localhost"),
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

var AppConfig Config

func init() {
	AppConfig = Load()
}
