package config

import (
	"os"
	"path"
	"strings"
)

type Config struct {
	Cwd           string
	Host          string
	ProxyUrl      string
	Completions   string
	Model         string
	Prompt        string
	ApiKeyFile    string
	ResultFolder  string
	ProviderType  string
	JwtToken      string
	PromptID      string
	Timeout       string
	ResultHistory string
	NoHistoryEnv  string
	MainFlags     MainFlags
}

type MainFlags struct {
	File      string
	NoHistory bool
	Sys       string
	PromptID  int
	Timeout   int
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

	return Config{
		Cwd:           cwd,
		Host:          getEnv("LCG_HOST", "http://192.168.87.108:11434/"),
		ProxyUrl:      getEnv("LCG_PROXY_URL", "/api/v1/protected/sberchat/chat"),
		Completions:   getEnv("LCG_COMPLETIONS_PATH", "api/chat"),
		Model:         getEnv("LCG_MODEL", "hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M"),
		Prompt:        getEnv("LCG_PROMPT", "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols."),
		ApiKeyFile:    getEnv("LCG_API_KEY_FILE", ".openai_api_key"),
		ResultFolder:  resultFolder,
		ProviderType:  getEnv("LCG_PROVIDER", "ollama"),
		JwtToken:      getEnv("LCG_JWT_TOKEN", ""),
		PromptID:      getEnv("LCG_PROMPT_ID", "1"),
		Timeout:       getEnv("LCG_TIMEOUT", "300"),
		ResultHistory: getEnv("LCG_RESULT_HISTORY", path.Join(resultFolder, "lcg_history.json")),
		NoHistoryEnv:  getEnv("LCG_NO_HISTORY", ""),
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

var AppConfig Config

func init() {
	AppConfig = Load()
}
