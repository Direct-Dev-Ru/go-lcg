package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	cmdPackage "github.com/direct-dev-ru/linux-command-gpt/cmd"
	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
	"github.com/direct-dev-ru/linux-command-gpt/reader"
	"github.com/direct-dev-ru/linux-command-gpt/serve"
	"github.com/direct-dev-ru/linux-command-gpt/validation"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

//go:embed VERSION.txt
var Version string

//go:embed build-conditions.yaml
var BuildConditionsFromYaml string

type buildConditions struct {
	NoServe bool `yaml:"no-serve"`
}

var CompileConditions buildConditions

// disableHistory управляет записью/обновлением истории на уровне процесса (флаг имеет приоритет над env)
var disableHistory bool

// fromHistory указывает, что текущий ответ взят из истории
var fromHistory bool

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
)

func main() {

	if err := yaml.Unmarshal([]byte(BuildConditionsFromYaml), &CompileConditions); err != nil {
		fmt.Println("Error parsing build conditions:", err)
		CompileConditions.NoServe = false
	}

	// fmt.Println("Build conditions:", CompileConditions)

	_ = colorBlue

	gpt.InitBuiltinPrompts("")

	// Авто-инициализация sys_prompts при старте CLI (создаст файл при отсутствии)
	if currentUser, err := user.Current(); err == nil {
		_ = gpt.NewPromptManager(currentUser.HomeDir)
	}

	app := &cli.App{
		Name:     "lcg",
		Usage:    config.AppConfig.AppName + " - Генерация Linux команд из описаний",
		Version:  Version,
		Commands: getCommands(),
		UsageText: `
lcg [опции] <описание команды>

Примеры:
  lcg "хочу извлечь файл linux-command-gpt.tar.gz"
  lcg --file /path/to/file.txt "хочу вывести все директории с помощью ls"
`,
		Description: `
{{.AppName}} - инструмент для генерации Linux команд из описаний на естественном языке.
Поддерживает чтение частей промпта из файлов и позволяет сохранять, копировать или перегенерировать результаты.
может задавать системный промпт или выбирать из предустановленных промптов.

Переменные окружения:

Основные настройки:
  LCG_HOST                Endpoint для LLM API (по умолчанию: http://192.168.87.108:11434/)
  LCG_MODEL               Название модели (по умолчанию: hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M)
  LCG_PROMPT              Текст промпта по умолчанию
  LCG_PROVIDER            Тип провайдера: "ollama" или "proxy" (по умолчанию: ollama)
  LCG_JWT_TOKEN           JWT токен для proxy провайдера
  LCG_PROMPT_ID           ID промпта по умолчанию (по умолчанию: 1)
  LCG_TIMEOUT             Таймаут запроса в секундах (по умолчанию: 300)
  LCG_COMPLETIONS_PATH    Путь к API для завершений (по умолчанию: api/chat)
  LCG_PROXY_URL           URL прокси для proxy провайдера (по умолчанию: /api/v1/protected/sberchat/chat)
  LCG_API_KEY_FILE        Файл с API ключом (по умолчанию: .openai_api_key)
  LCG_APP_NAME            Название приложения (по умолчанию: Linux Command GPT)

Настройки истории и выполнения:
  LCG_NO_HISTORY          Отключить запись истории ("1" или "true" = отключено, пусто = включено)
  LCG_ALLOW_EXECUTION     Разрешить выполнение команд ("1" или "true" = разрешено, пусто = запрещено)
  LCG_RESULT_FOLDER       Папка для сохранения результатов (по умолчанию: ~/.config/lcg/gpt_results)
  LCG_RESULT_HISTORY      Файл истории результатов (по умолчанию: <result_folder>/lcg_history.json)
  LCG_PROMPT_FOLDER       Папка для системных промптов (по умолчанию: ~/.config/lcg/gpt_sys_prompts)
  LCG_CONFIG_FOLDER       Папка для конфигурации (по умолчанию: ~/.config/lcg/config)

Настройки сервера (команда serve):
  LCG_SERVER_PORT         Порт сервера (по умолчанию: 8080)
  LCG_SERVER_HOST         Хост сервера (по умолчанию: localhost)
  LCG_SERVER_ALLOW_HTTP   Разрешить HTTP соединения ("true" для localhost, "false" для других хостов)
  LCG_SERVER_REQUIRE_AUTH Требовать аутентификацию ("1" или "true" = требуется, пусто = не требуется)
  LCG_SERVER_PASSWORD     Пароль администратора (по умолчанию: admin#123456)
  LCG_SERVER_SSL_CERT_FILE Путь к SSL сертификату
  LCG_SERVER_SSL_KEY_FILE  Путь к приватному ключу SSL
  LCG_DOMAIN              Домен для сервера (по умолчанию: значение LCG_SERVER_HOST)
  LCG_COOKIE_SECURE       Безопасные cookie ("1" или "true" = включено, пусто = выключено)
  LCG_COOKIE_PATH         Путь для cookie (по умолчанию: /lcg)
  LCG_COOKIE_TTL_HOURS    Время жизни cookie в часах (по умолчанию: 168)
  LCG_BASE_URL            Базовый URL приложения (по умолчанию: /lcg)
  LCG_HEALTH_URL          URL для проверки здоровья API (по умолчанию: /api/v1/protected/sberchat/health)

Настройки валидации:
  LCG_MAX_SYSTEM_PROMPT_LENGTH    Максимальная длина системного промпта (по умолчанию: 2000)
  LCG_MAX_USER_MESSAGE_LENGTH     Максимальная длина пользовательского сообщения (по умолчанию: 4000)
  LCG_MAX_PROMPT_NAME_LENGTH      Максимальная длина названия промпта (по умолчанию: 2000)
  LCG_MAX_PROMPT_DESC_LENGTH      Максимальная длина описания промпта (по умолчанию: 5000)
  LCG_MAX_COMMAND_LENGTH          Максимальная длина команды (по умолчанию: 8000)
  LCG_MAX_EXPLANATION_LENGTH      Максимальная длина объяснения (по умолчанию: 20000)

Отладка и браузер:
  LCG_DEBUG               Включить режим отладки ("1" или "true" = включено, пусто = выключено)
  LCG_BROWSER_PATH        Путь к браузеру для автоматического открытия (команда serve --browser)
`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Read part of the command from a file",
			},
			&cli.BoolFlag{
				Name:    "no-history",
				Aliases: []string{"nh"},
				Usage:   "Disable writing/updating command history (overrides LCG_NO_HISTORY)",
				Value:   false,
			},
			&cli.StringFlag{
				Name:        "sys",
				Aliases:     []string{"s"},
				Usage:       "System prompt content or ID",
				DefaultText: "Use prompt ID from LCG_PROMPT_ID or default prompt",
				Value:       "",
			},
			&cli.IntFlag{
				Name:        "prompt-id",
				Aliases:     []string{"pid"},
				Usage:       "System prompt ID (1-5 for default prompts)",
				DefaultText: "1",
				Value:       1,
			},
			&cli.IntFlag{
				Name:        "timeout",
				Aliases:     []string{"t"},
				Usage:       "Request timeout in seconds",
				DefaultText: "120",
				Value:       120,
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Show debug information (request parameters and prompts)",
				Value:   false,
			},
		},
		Action: func(c *cli.Context) error {
			file := c.String("file")
			system := c.String("sys")
			// обновляем конфиг на основе флагов
			if system != "" {
				config.AppConfig.Prompt = system
			}
			if c.IsSet("timeout") {
				config.AppConfig.Timeout = fmt.Sprintf("%d", c.Int("timeout"))
			}
			promptID := c.Int("prompt-id")
			timeout := c.Int("timeout")
			// сохраняем конкретные значения флагов
			config.AppConfig.MainFlags = config.MainFlags{
				File:      file,
				NoHistory: c.Bool("no-history"),
				Sys:       system,
				PromptID:  promptID,
				Timeout:   timeout,
				Debug:     c.Bool("debug"),
			}
			disableHistory = config.AppConfig.MainFlags.NoHistory || config.AppConfig.IsNoHistoryEnabled()

			config.AppConfig.MainFlags.Debug = config.AppConfig.MainFlags.Debug || config.GetEnvBool("LCG_DEBUG", false)

			// fmt.Println("Debug:", config.AppConfig.MainFlags.Debug)
			// fmt.Println("LCG_DEBUG:", config.GetEnvBool("LCG_DEBUG", false))

			args := c.Args().Slice()

			if len(args) == 0 {
				cli.ShowAppHelp(c)
				showTips()
				return nil
			}

			// Если указан prompt-id, загружаем соответствующий промпт
			if system == "" && promptID > 0 {
				currentUser, _ := user.Current()
				pm := gpt.NewPromptManager(currentUser.HomeDir)
				if prompt, err := pm.GetPromptByID(promptID); err == nil {
					system = prompt.Content
				} else {
					fmt.Printf("Warning: Prompt ID %d not found, using default prompt\n", promptID)
				}
			}

			if CompileConditions.NoServe {
				if len(args) > 1 && args[0] == "serve" {
					printColored("❌ Error: serve command is disabled in this build\n", colorRed)
					os.Exit(1)
				}
			}
			executeMain(file, system, strings.Join(args, " "), timeout)
			return nil
		},
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V", "v"},
		Usage:   "prints out version",
	}
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf("%s\n", cCtx.App.Version)
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func getCommands() []*cli.Command {
	commands := []*cli.Command{
		{
			Name:    "update-key",
			Aliases: []string{"u"},
			Usage:   "Update the API key",
			Action: func(c *cli.Context) error {
				if config.AppConfig.ProviderType == "ollama" || config.AppConfig.ProviderType == "proxy" {
					fmt.Println("API key is not needed for ollama and proxy providers")
					return nil
				}
				timeout := 120 // default timeout
				if t, err := strconv.Atoi(config.AppConfig.Timeout); err == nil {
					timeout = t
				}
				gpt3 := initGPT(config.AppConfig.Prompt, timeout)
				gpt3.UpdateKey()
				fmt.Println("API key updated.")
				return nil
			},
		},
		{
			Name:    "delete-key",
			Aliases: []string{"d"},
			Usage:   "Delete the API key",
			Action: func(c *cli.Context) error {
				if config.AppConfig.ProviderType == "ollama" || config.AppConfig.ProviderType == "proxy" {
					fmt.Println("API key is not needed for ollama and proxy providers")
					return nil
				}
				timeout := 120 // default timeout
				if t, err := strconv.Atoi(config.AppConfig.Timeout); err == nil {
					timeout = t
				}
				gpt3 := initGPT(config.AppConfig.Prompt, timeout)
				gpt3.DeleteKey()
				fmt.Println("API key deleted.")
				return nil
			},
		},
		{
			Name:    "update-jwt",
			Aliases: []string{"j"},
			Usage:   "Update the JWT token for proxy API",
			Action: func(c *cli.Context) error {
				if config.AppConfig.ProviderType != "proxy" {
					fmt.Println("JWT token is only needed for proxy provider")
					return nil
				}

				var jwtToken string
				fmt.Print("JWT Token: ")
				fmt.Scanln(&jwtToken)

				currentUser, _ := user.Current()
				jwtFile := currentUser.HomeDir + "/.proxy_jwt_token"
				if err := os.WriteFile(jwtFile, []byte(strings.TrimSpace(jwtToken)), 0600); err != nil {
					fmt.Printf("Ошибка сохранения JWT токена: %v\n", err)
					return err
				}

				fmt.Println("JWT token updated.")
				return nil
			},
		},
		{
			Name:    "delete-jwt",
			Aliases: []string{"dj"},
			Usage:   "Delete the JWT token for proxy API",
			Action: func(c *cli.Context) error {
				if config.AppConfig.ProviderType != "proxy" {
					fmt.Println("JWT token is only needed for proxy provider")
					return nil
				}

				currentUser, _ := user.Current()
				jwtFile := currentUser.HomeDir + "/.proxy_jwt_token"
				if err := os.Remove(jwtFile); err != nil && !os.IsNotExist(err) {
					fmt.Printf("Ошибка удаления JWT токена: %v\n", err)
					return err
				}

				fmt.Println("JWT token deleted.")
				return nil
			},
		},
		{
			Name:    "models",
			Aliases: []string{"m"},
			Usage:   "Show available models",
			Action: func(c *cli.Context) error {
				timeout := 120 // default timeout
				if t, err := strconv.Atoi(config.AppConfig.Timeout); err == nil {
					timeout = t
				}
				gpt3 := initGPT(config.AppConfig.Prompt, timeout)
				models, err := gpt3.GetAvailableModels()
				if err != nil {
					fmt.Printf("Ошибка получения моделей: %v\n", err)
					return err
				}

				fmt.Printf("Доступные модели для провайдера %s:\n", config.AppConfig.ProviderType)
				for i, model := range models {
					fmt.Printf("  %d. %s\n", i+1, model)
				}
				return nil
			},
		},
		{
			Name:    "health",
			Aliases: []string{"he"}, // Изменено с "h" на "he"
			Usage:   "Check API health",
			Action: func(c *cli.Context) error {
				timeout := 120 // default timeout
				if t, err := strconv.Atoi(config.AppConfig.Timeout); err == nil {
					timeout = t
				}
				gpt3 := initGPT(config.AppConfig.Prompt, timeout)
				if err := gpt3.Health(); err != nil {
					fmt.Printf("Health check failed: %v\n", err)
					return err
				}
				fmt.Println("API is healthy.")
				return nil
			},
		},
		{
			Name:    "config",
			Aliases: []string{"co"}, // Изменено с "c" на "co"
			Usage:   "Show current configuration",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "full",
					Aliases: []string{"f"},
					Usage:   "Show full configuration object",
				},
			},
			Action: func(c *cli.Context) error {
				if c.Bool("full") {
					// Выводим полную конфигурацию в JSON формате
					showFullConfig()
				} else {
					// Выводим краткую конфигурацию
					showShortConfig()
				}
				return nil
			},
		},
		{
			Name:    "history",
			Aliases: []string{"hist"},
			Usage:   "Show command history",
			Subcommands: []*cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "List history entries",
					Action: func(c *cli.Context) error {
						if disableHistory {
							printColored("📝 История отключена (--no-history / LCG_NO_HISTORY)\n", colorYellow)
						} else {
							cmdPackage.ShowHistory(config.AppConfig.ResultHistory, printColored, colorYellow)
						}
						return nil
					},
				},
				{
					Name:    "view",
					Aliases: []string{"v"},
					Usage:   "View history entry by ID",
					Action: func(c *cli.Context) error {
						if c.NArg() == 0 {
							fmt.Println("Укажите ID записи истории")
							return nil
						}
						var id int
						if _, err := fmt.Sscanf(c.Args().First(), "%d", &id); err != nil || id <= 0 {
							fmt.Println("Неверный ID")
							return nil
						}
						if disableHistory {
							fmt.Println("История отключена")
						} else {
							cmdPackage.ViewHistoryEntry(config.AppConfig.ResultHistory, id, printColored, colorYellow, colorBold, colorGreen)
						}
						return nil
					},
				},
				{
					Name:    "delete",
					Aliases: []string{"d"},
					Usage:   "Delete history entry by ID",
					Action: func(c *cli.Context) error {
						if c.NArg() == 0 {
							fmt.Println("Укажите ID записи истории")
							return nil
						}
						var id int
						if _, err := fmt.Sscanf(c.Args().First(), "%d", &id); err != nil || id <= 0 {
							fmt.Println("Неверный ID")
							return nil
						}
						if disableHistory {
							fmt.Println("История отключена")
						} else if err := cmdPackage.DeleteHistoryEntry(config.AppConfig.ResultHistory, id); err != nil {
							fmt.Println(err)
						}
						return nil
					},
				},
			},
		},
		{
			Name:    "prompts",
			Aliases: []string{"p"},
			Usage:   "Manage system prompts",
			Subcommands: []*cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "List all available prompts",
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "full",
							Aliases: []string{"f"},
							Usage:   "Show full content without truncation",
						},
					},
					Action: func(c *cli.Context) error {
						currentUser, _ := user.Current()
						pm := gpt.NewPromptManager(currentUser.HomeDir)
						full := c.Bool("full")
						pm.ListPromptsWithFull(full)
						return nil
					},
				},
				{
					Name:    "add",
					Aliases: []string{"a"},
					Usage:   "Add a new custom prompt",
					Action: func(c *cli.Context) error {
						currentUser, _ := user.Current()
						pm := gpt.NewPromptManager(currentUser.HomeDir)

						var name, description, content string

						fmt.Print("Название промпта: ")
						fmt.Scanln(&name)

						fmt.Print("Описание: ")
						fmt.Scanln(&description)

						fmt.Print("Содержание промпта: ")
						fmt.Scanln(&content)

						if err := pm.AddCustomPrompt(name, description, content); err != nil {
							fmt.Printf("Ошибка добавления промпта: %v\n", err)
							return err
						}

						fmt.Println("Промпт успешно добавлен!")
						return nil
					},
				},
				{
					Name:    "delete",
					Aliases: []string{"d"},
					Usage:   "Delete a custom prompt",
					Action: func(c *cli.Context) error {
						if c.NArg() == 0 {
							fmt.Println("Укажите ID промпта для удаления")
							return nil
						}

						var id int
						if _, err := fmt.Sscanf(c.Args().First(), "%d", &id); err != nil {
							fmt.Println("Неверный ID промпта")
							return err
						}

						currentUser, _ := user.Current()
						pm := gpt.NewPromptManager(currentUser.HomeDir)

						if err := pm.DeleteCustomPrompt(id); err != nil {
							fmt.Printf("Ошибка удаления промпта: %v\n", err)
							return err
						}

						fmt.Println("Промпт успешно удален!")
						return nil
					},
				},
			},
		},
		{
			Name:    "test-prompt",
			Aliases: []string{"tp"},
			Usage:   "Test a specific prompt ID",
			Action: func(c *cli.Context) error {
				if c.NArg() == 0 {
					fmt.Println("Usage: lcg test-prompt <prompt-id> <command>")
					return nil
				}

				var promptID int
				if _, err := fmt.Sscanf(c.Args().First(), "%d", &promptID); err != nil {
					fmt.Println("Invalid prompt ID")
					return err
				}

				currentUser, _ := user.Current()
				pm := gpt.NewPromptManager(currentUser.HomeDir)

				prompt, err := pm.GetPromptByID(promptID)
				if err != nil {
					fmt.Printf("Prompt ID %d not found\n", promptID)
					return err
				}

				fmt.Printf("Testing prompt ID %d: %s\n", promptID, prompt.Name)
				fmt.Printf("Description: %s\n", prompt.Description)
				fmt.Printf("Content: %s\n", prompt.Content)

				if len(c.Args().Slice()) > 1 {
					command := strings.Join(c.Args().Slice()[1:], " ")
					fmt.Printf("\nTesting with command: %s\n", command)
					timeout := 120 // default timeout
					if t, err := strconv.Atoi(config.AppConfig.Timeout); err == nil {
						timeout = t
					}
					executeMain("", prompt.Content, command, timeout)
				}

				return nil
			},
		},
		{
			Name:  "serve",
			Usage: "Start HTTP server to browse saved results",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    "port",
					Aliases: []string{"p"},
					Usage:   "Server port",
					Value:   config.AppConfig.Server.Port,
				},
				&cli.StringFlag{
					Name:    "host",
					Aliases: []string{"H"},
					Usage:   "Server host",
					Value:   config.AppConfig.Server.Host,
				},
				&cli.BoolFlag{
					Name:    "browser",
					Aliases: []string{"b"},
					Usage:   "Open browser automatically after starting server",
					Value:   false,
				},
			},
			Action: func(c *cli.Context) error {
				port := c.String("port")
				host := c.String("host")
				openBrowser := c.Bool("browser")

				// Пробрасываем debug: флаг или переменная окружения LCG_DEBUG
				// Позволяет запускать: LCG_DEBUG=1 lcg serve ... или lcg -d serve ...
				config.AppConfig.MainFlags.Debug = c.Bool("debug") || config.GetEnvBool("LCG_DEBUG", false)

				// Обновляем конфигурацию сервера с новыми параметрами
				config.AppConfig.Server.Host = host
				config.AppConfig.Server.Port = port
				// Пересчитываем AllowHTTP на основе нового хоста
				config.AppConfig.Server.AllowHTTP = getServerAllowHTTPForHost(host)

				// Определяем протокол на основе хоста
				useHTTPS := !config.AppConfig.Server.AllowHTTP
				protocol := "http"
				if useHTTPS {
					protocol = "https"
				}

				printColored(fmt.Sprintf("🌐 Запускаю %s сервер на %s:%s\n", strings.ToUpper(protocol), host, port), colorCyan)
				printColored(fmt.Sprintf("📁 Папка результатов: %s\n", config.AppConfig.ResultFolder), colorYellow)

				// Предупреждение о самоподписанном сертификате
				if useHTTPS {
					printColored("⚠️  Используется самоподписанный SSL сертификат\n", colorYellow)
					printColored("   Браузер может показать предупреждение о безопасности\n", colorYellow)
					printColored("   Нажмите 'Дополнительно' → 'Перейти на сайт' для продолжения\n", colorYellow)
				}

				// Для автооткрытия браузера заменяем 0.0.0.0 на localhost
				browserHost := host
				if host == "0.0.0.0" {
					browserHost = "localhost"
				}

				// Учитываем BasePath в URL
				basePath := config.AppConfig.Server.BasePath
				if basePath == "" || basePath == "/" {
					basePath = ""
				} else {
					basePath = strings.TrimSuffix(basePath, "/")
				}
				url := fmt.Sprintf("%s://%s:%s%s", protocol, browserHost, port, basePath)

				if openBrowser {
					printColored("🌍 Открываю браузер...\n", colorGreen)
					if err := openBrowserURL(url); err != nil {
						printColored(fmt.Sprintf("⚠️  Не удалось открыть браузер: %v\n", err), colorYellow)
						printColored("📱 Откройте браузер вручную и перейдите по адресу: ", colorGreen)
						printColored(url+"\n", colorYellow)
					}
				} else {
					printColored("🔗 Откройте в браузере: ", colorGreen)
					printColored(url+"\n", colorYellow)
				}

				return serve.StartResultServer(host, port)
			},
		},
	}

	if CompileConditions.NoServe {
		filteredCommands := []*cli.Command{}
		for _, cmd := range commands {
			if cmd.Name != "serve" {
				filteredCommands = append(filteredCommands, cmd)
			}
		}
		commands = filteredCommands
	}

	return commands

}

func executeMain(file, system, commandInput string, timeout int) {
	// Валидация длины пользовательского сообщения
	if err := validation.ValidateUserMessage(commandInput); err != nil {
		printColored(fmt.Sprintf("❌ Ошибка: %s\n", err.Error()), colorRed)
		return
	}

	// Валидация длины системного промпта
	if err := validation.ValidateSystemPrompt(system); err != nil {
		printColored(fmt.Sprintf("❌ Ошибка: %s\n", err.Error()), colorRed)
		return
	}

	// Выводим debug информацию если включен флаг
	if config.AppConfig.MainFlags.Debug {
		printDebugInfo(file, system, commandInput, timeout)
	}
	if file != "" {
		if err := reader.FileToPrompt(&commandInput, file); err != nil {
			printColored(fmt.Sprintf("❌ Ошибка чтения файла: %v\n", err), colorRed)
			return
		}
	}

	// Если system пустой, используем дефолтный промпт
	if system == "" {
		system = config.AppConfig.Prompt
	}

	// Обеспечим папку результатов заранее (может понадобиться при действиях)
	if _, err := os.Stat(config.AppConfig.ResultFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(config.AppConfig.ResultFolder, 0755); err != nil {
			printColored(fmt.Sprintf("❌ Ошибка создания папки результатов: %v\n", err), colorRed)
			return
		}
	}

	// Проверка истории: если такой запрос уже встречался — предложить открыть из истории
	if !disableHistory {
		if found, hist := cmdPackage.CheckAndSuggestFromHistory(config.AppConfig.ResultHistory, commandInput); found && hist != nil {
			fromHistory = true // Устанавливаем флаг, что ответ из истории
			gpt3 := initGPT(system, timeout)
			printColored("\nВНИМАНИЕ: ОТВЕТ СФОРМИРОВАН ИИ. ТРЕБУЕТСЯ ПРОВЕРКА И КРИТИЧЕСКИЙ АНАЛИЗ. ВОЗМОЖНЫ ОШИБКИ И ГАЛЛЮЦИНАЦИИ.\n", colorRed)
			printColored("\n📋 Команда (из истории):\n", colorYellow)
			printColored(fmt.Sprintf("   %s\n\n", hist.Response), colorBold+colorGreen)
			if strings.TrimSpace(hist.Explanation) != "" {
				printColored("\n📖 Подробное объяснение (из истории):\n\n", colorYellow)
				fmt.Println(hist.Explanation)
			}
			// Показали из истории — не выполняем запрос к API, сразу меню действий
			handlePostResponse(hist.Response, gpt3, system, commandInput, timeout, hist.Explanation)
			return
		}
	}

	// Папка уже создана выше

	gpt3 := initGPT(system, timeout)

	printColored("🤖 Запрос: ", colorCyan)
	fmt.Printf("%s\n", commandInput)

	response, elapsed := getCommand(gpt3, commandInput)
	if response == "" {
		printColored("❌ Ответ не получен. Проверьте подключение к API.\n", colorRed)
		return
	}

	printColored(fmt.Sprintf("✅ Выполнено за %.2f сек\n", elapsed), colorGreen)
	// Обязательное предупреждение перед первым ответом
	printColored("\nВНИМАНИЕ: ОТВЕТ СФОРМИРОВАН ИИ. ТРЕБУЕТСЯ ПРОВЕРКА И КРИТИЧЕСКИЙ АНАЛИЗ. ВОЗМОЖНЫ ОШИБКИ И ГАЛЛЮЦИНАЦИИ.\n", colorRed)
	printColored("\n📋 Команда:\n", colorYellow)
	printColored(fmt.Sprintf("   %s\n\n", response), colorBold+colorGreen)

	// Сохраняем в историю (после завершения работы – т.е. позже, в зависимости от выбора действия)
	// Здесь не сохраняем, чтобы учесть правило: сохранять после действия, отличного от v/vv/vvv
	fromHistory = false // Сбрасываем флаг для новых запросов
	handlePostResponse(response, gpt3, system, commandInput, timeout, "")
}

// checkAndSuggestFromHistory проверяет файл истории и при совпадении запроса предлагает показать сохраненный результат
// moved to history.go

func initGPT(system string, timeout int) gpt.Gpt3 {
	currentUser, _ := user.Current()

	// Загружаем JWT токен в зависимости от провайдера
	var jwtToken string
	if config.AppConfig.ProviderType == "proxy" {
		jwtToken = config.AppConfig.JwtToken
		if jwtToken == "" {
			// Пытаемся загрузить из файла
			jwtFile := currentUser.HomeDir + "/.proxy_jwt_token"
			if data, err := os.ReadFile(jwtFile); err == nil {
				jwtToken = strings.TrimSpace(string(data))
			}
		}
	}

	return *gpt.NewGpt3(config.AppConfig.ProviderType, config.AppConfig.Host, jwtToken, config.AppConfig.Model, system, 0.01, timeout)
}

func getCommand(gpt3 gpt.Gpt3, cmd string) (string, float64) {
	gpt3.InitKey()
	start := time.Now()
	done := make(chan bool)

	go func() {
		loadingChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Printf("\r%s", strings.Repeat(" ", 50))
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\r%s Обрабатываю запрос...", loadingChars[i])
				i = (i + 1) % len(loadingChars)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	response := gpt3.Completions(cmd)
	done <- true
	elapsed := math.Round(time.Since(start).Seconds()*100) / 100

	return response, elapsed
}

func handlePostResponse(response string, gpt3 gpt.Gpt3, system, cmd string, timeout int, explanation string) {
	// Формируем меню действий
	menu := "Действия: (c)копировать, (s)сохранить, (r)перегенерировать"
	if config.AppConfig.AllowExecution {
		menu += ", (e)выполнить"
	}
	menu += ", (v|vv|vvv)подробно, (n)ничего: "

	fmt.Print(menu)
	var choice string
	fmt.Scanln(&choice)

	switch strings.ToLower(choice) {
	case "c":
		clipboard.WriteAll(response)
		fmt.Println("✅ Команда скопирована в буфер обмена")
		if !disableHistory {
			if fromHistory {
				cmdPackage.SaveToHistoryFromHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt, explanation)
			} else {
				cmdPackage.SaveToHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt)
			}
		}
	case "s":
		if fromHistory && strings.TrimSpace(explanation) != "" {
			saveResponse(response, gpt3.Model, gpt3.Prompt, cmd, explanation)
		} else {
			saveResponse(response, gpt3.Model, gpt3.Prompt, cmd)
		}
		if !disableHistory {
			if fromHistory {
				cmdPackage.SaveToHistoryFromHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt, explanation)
			} else {
				cmdPackage.SaveToHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt)
			}
		}
	case "r":
		fmt.Println("🔄 Перегенерирую...")
		executeMain("", system, cmd, timeout)
	case "e":
		if config.AppConfig.AllowExecution {
			executeCommand(response)
			if !disableHistory {
				if fromHistory {
					cmdPackage.SaveToHistoryFromHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt, explanation)
				} else {
					cmdPackage.SaveToHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt)
				}
			}
		} else {
			fmt.Println("⚠️  Выполнение команд отключено. Установите LCG_ALLOW_EXECUTION=1 для включения этой функции.")
		}
	case "v", "vv", "vvv":
		level := len(choice) // 1, 2, 3
		deps := cmdPackage.ExplainDeps{
			DisableHistory: disableHistory,
			PrintColored:   printColored,
			ColorPurple:    colorPurple,
			ColorGreen:     colorGreen,
			ColorRed:       colorRed,
			ColorYellow:    colorYellow,
			GetCommand:     getCommand,
		}
		cmdPackage.ShowDetailedExplanation(response, gpt3, system, cmd, timeout, level, deps)
	default:
		fmt.Println(" До свидания!")
		if !disableHistory {
			if fromHistory {
				cmdPackage.SaveToHistoryFromHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt, explanation)
			} else {
				cmdPackage.SaveToHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, cmd, response, gpt3.Prompt)
			}
		}
	}
}

// moved to response.go

// saveExplanation сохраняет подробное объяснение и альтернативные способы
// moved to explain.go

// truncateTitle сокращает строку до 120 символов (по рунам), добавляя " ..." при усечении
// moved to response.go

// moved to explain.go

func executeCommand(command string) {
	fmt.Printf("🚀 Выполняю: %s\n", command)
	fmt.Print("Продолжить? (y/N): ")
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) == "y" || strings.ToLower(confirm) == "yes" {
		cmd := exec.Command("bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Ошибка выполнения: %v\n", err)
		} else {
			fmt.Println("✅ Команда выполнена успешно")
		}
	} else {
		fmt.Println("❌ Выполнение отменено")
	}
}

// env helpers moved to config package

// moved to history.go

func printColored(text, color string) {
	fmt.Printf("%s%s%s", color, text, colorReset)
}

func showTips() {
	printColored("💡 Подсказки:\n", colorCyan)
	fmt.Println("   • Используйте --file для чтения из файла")
	fmt.Println("   • Используйте --sys для изменения системного промпта")
	fmt.Println("   • Используйте --prompt-id для выбора предустановленного промпта")
	fmt.Println("   • Используйте --timeout для установки таймаута запроса")
	fmt.Println("   • Укажите --no-history чтобы не записывать историю (аналог LCG_NO_HISTORY)")
	fmt.Println("   • Команда 'prompts list' покажет все доступные промпты")
	fmt.Println("   • Команда 'history list' покажет историю запросов")
	fmt.Println("   • Команда 'config' покажет текущие настройки")
	fmt.Println("   • Команда 'health' проверит доступность API")
	fmt.Println("   • Команда 'serve' запустит HTTP сервер для просмотра результатов")
	fmt.Println("   • Используйте --browser для автоматического открытия браузера")
	fmt.Println("   • Установите LCG_BROWSER_PATH для указания конкретного браузера")
}

// printDebugInfo выводит отладочную информацию о параметрах запроса
func printDebugInfo(file, system, commandInput string, timeout int) {
	printColored("\n🔍 DEBUG ИНФОРМАЦИЯ:\n", colorCyan)
	fmt.Printf("📁 Файл: %s\n", file)
	fmt.Printf("🤖 Системный промпт: %s\n", system)
	fmt.Printf("💬 Запрос: %s\n", commandInput)
	fmt.Printf("⏱️  Таймаут: %d сек\n", timeout)
	fmt.Printf("🌐 Провайдер: %s\n", config.AppConfig.ProviderType)
	fmt.Printf("🏠 Хост: %s\n", config.AppConfig.Host)
	fmt.Printf("🧠 Модель: %s\n", config.AppConfig.Model)
	fmt.Printf("📝 История: %t\n", !config.AppConfig.MainFlags.NoHistory)
	printColored("────────────────────────────────────────\n", colorCyan)
}

// openBrowserURL открывает URL в браузере
func openBrowserURL(url string) error {
	// Проверяем переменную окружения LCG_BROWSER_PATH
	if browserPath := os.Getenv("LCG_BROWSER_PATH"); browserPath != "" {
		return exec.Command(browserPath, url).Start()
	}

	// Список браузеров в порядке приоритета
	browsers := []string{
		"yandex-browser",        // Яндекс.Браузер
		"yandex-browser-stable", // Яндекс.Браузер (стабильная версия)
		"firefox",               // Mozilla Firefox
		"firefox-esr",           // Firefox ESR
		"google-chrome",         // Google Chrome
		"google-chrome-stable",  // Google Chrome (стабильная версия)
		"chromium",              // Chromium
		"chromium-browser",      // Chromium (Ubuntu/Debian)
	}

	// Стандартные пути для поиска браузеров
	paths := []string{
		"/usr/bin",
		"/usr/local/bin",
		"/opt/google/chrome",
		"/opt/yandex/browser",
		"/snap/bin",
		"/usr/lib/chromium-browser",
	}

	// Ищем браузер в указанном порядке
	for _, browser := range browsers {
		for _, path := range paths {
			fullPath := filepath.Join(path, browser)
			if _, err := os.Stat(fullPath); err == nil {
				return exec.Command(fullPath, url).Start()
			}
		}
		// Также пробуем найти в PATH
		if _, err := exec.LookPath(browser); err == nil {
			return exec.Command(browser, url).Start()
		}
	}

	return fmt.Errorf("не найден ни один из поддерживаемых браузеров")
}

// getServerAllowHTTPForHost определяет AllowHTTP для конкретного хоста
func getServerAllowHTTPForHost(host string) bool {
	// Если переменная явно установлена, используем её
	if value, exists := os.LookupEnv("LCG_SERVER_ALLOW_HTTP"); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}

	// Если переменная не установлена, определяем по умолчанию на основе хоста
	return isSecureHost(host)
}

// isSecureHost проверяет, является ли хост безопасным для HTTP
func isSecureHost(host string) bool {
	secureHosts := []string{"localhost", "127.0.0.1", "::1"}
	for _, secureHost := range secureHosts {
		if host == secureHost {
			return true
		}
	}
	return false
}

// showShortConfig показывает краткую конфигурацию
func showShortConfig() {
	fmt.Printf("Provider: %s\n", config.AppConfig.ProviderType)
	fmt.Printf("Host: %s\n", config.AppConfig.Host)
	fmt.Printf("Model: %s\n", config.AppConfig.Model)
	fmt.Printf("Prompt: %s\n", config.AppConfig.Prompt)
	fmt.Printf("Timeout: %s seconds\n", config.AppConfig.Timeout)
	if config.AppConfig.ProviderType == "proxy" {
		fmt.Printf("JWT Token: %s\n", func() string {
			if config.AppConfig.JwtToken != "" {
				return "***set***"
			}
			currentUser, _ := user.Current()
			jwtFile := currentUser.HomeDir + "/.proxy_jwt_token"
			if _, err := os.Stat(jwtFile); err == nil {
				return "***from file***"
			}
			return "***not set***"
		}())
	}
}

// showFullConfig показывает полную конфигурацию в JSON формате
func showFullConfig() {
	// Создаем структуру для безопасного вывода (скрываем чувствительные данные)
	type SafeConfig struct {
		Cwd            string                  `json:"cwd"`
		Host           string                  `json:"host"`
		Completions    string                  `json:"completions"`
		Model          string                  `json:"model"`
		Prompt         string                  `json:"prompt"`
		ApiKeyFile     string                  `json:"api_key_file"`
		ResultFolder   string                  `json:"result_folder"`
		PromptFolder   string                  `json:"prompt_folder"`
		ProviderType   string                  `json:"provider_type"`
		JwtToken       string                  `json:"jwt_token"` // Показываем статус, не сам токен
		PromptID       string                  `json:"prompt_id"`
		Timeout        string                  `json:"timeout"`
		ResultHistory  string                  `json:"result_history"`
		NoHistoryEnv   string                  `json:"no_history_env"`
		AllowExecution bool                    `json:"allow_execution"`
		MainFlags      config.MainFlags        `json:"main_flags"`
		Server         config.ServerConfig     `json:"server"`
		Validation     config.ValidationConfig `json:"validation"`
	}

	// Создаем безопасную копию конфигурации
	safeConfig := SafeConfig{
		Cwd:          config.AppConfig.Cwd,
		Host:         config.AppConfig.Host,
		Completions:  config.AppConfig.Completions,
		Model:        config.AppConfig.Model,
		Prompt:       config.AppConfig.Prompt,
		ApiKeyFile:   config.AppConfig.ApiKeyFile,
		ResultFolder: config.AppConfig.ResultFolder,
		PromptFolder: config.AppConfig.PromptFolder,
		ProviderType: config.AppConfig.ProviderType,
		JwtToken: func() string {
			if config.AppConfig.JwtToken != "" {
				return "***set***"
			}
			currentUser, _ := user.Current()
			jwtFile := currentUser.HomeDir + "/.proxy_jwt_token"
			if _, err := os.Stat(jwtFile); err == nil {
				return "***from file***"
			}
			return "***not set***"
		}(),
		PromptID:       config.AppConfig.PromptID,
		Timeout:        config.AppConfig.Timeout,
		ResultHistory:  config.AppConfig.ResultHistory,
		NoHistoryEnv:   config.AppConfig.NoHistoryEnv,
		AllowExecution: config.AppConfig.AllowExecution,
		MainFlags:      config.AppConfig.MainFlags,
		Server:         config.AppConfig.Server,
		Validation:     config.AppConfig.Validation,
	}

	safeConfig.Server.Password = "***"

	// Выводим JSON с отступами
	jsonData, err := json.MarshalIndent(safeConfig, "", "  ")
	if err != nil {
		fmt.Printf("Ошибка сериализации конфигурации: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}
