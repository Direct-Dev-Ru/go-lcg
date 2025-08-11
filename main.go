package main

import (
	_ "embed"
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
	"github.com/direct-dev-ru/linux-command-gpt/reader"
	"github.com/urfave/cli/v2"
)

//go:embed VERSION.txt
var Version string

var (
	cwd, _        = os.Getwd()
	HOST          = getEnv("LCG_HOST", "http://192.168.87.108:11434/")
	COMPLETIONS   = getEnv("LCG_COMPLETIONS_PATH", "api/chat")
	MODEL         = getEnv("LCG_MODEL", "codegeex4")
	PROMPT        = getEnv("LCG_PROMPT", "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols.")
	API_KEY_FILE  = getEnv("LCG_API_KEY_FILE", ".openai_api_key")
	RESULT_FOLDER = getEnv("LCG_RESULT_FOLDER", path.Join(cwd, "gpt_results"))
	PROVIDER_TYPE = getEnv("LCG_PROVIDER", "ollama") // "ollama", "proxy"
	JWT_TOKEN     = getEnv("LCG_JWT_TOKEN", "")
	PROMPT_ID     = getEnv("LCG_PROMPT_ID", "1") // ID промпта по умолчанию
	TIMEOUT       = getEnv("LCG_TIMEOUT", "120") // Таймаут в секундах по умолчанию
)

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
	app := &cli.App{
		Name:     "lcg",
		Usage:    "Linux Command GPT - Генерация Linux команд из описаний",
		Version:  Version,
		Commands: getCommands(),
		UsageText: `
lcg [опции] <описание команды>

Примеры:
  lcg "хочу извлечь файл linux-command-gpt.tar.gz"
  lcg --file /path/to/file.txt "хочу вывести все директории с помощью ls"
`,
		Description: `
Linux Command GPT - инструмент для генерации Linux команд из описаний на естественном языке.
Поддерживает чтение частей промпта из файлов и позволяет сохранять, копировать или перегенерировать результаты.

Переменные окружения:
  LCG_HOST          Endpoint для LLM API (по умолчанию: http://192.168.87.108:11434/)
  LCG_MODEL         Название модели (по умолчанию: codegeex4)
  LCG_PROMPT        Текст промпта по умолчанию
  LCG_PROVIDER      Тип провайдера: "ollama" или "proxy" (по умолчанию: ollama)
  LCG_JWT_TOKEN     JWT токен для proxy провайдера
`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Usage:   "Read part of the command from a file",
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
		},
		Action: func(c *cli.Context) error {
			file := c.String("file")
			system := c.String("sys")
			promptID := c.Int("prompt-id")
			timeout := c.Int("timeout")
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
	return []*cli.Command{
		{
			Name:    "update-key",
			Aliases: []string{"u"},
			Usage:   "Update the API key",
			Action: func(c *cli.Context) error {
				if PROVIDER_TYPE == "ollama" || PROVIDER_TYPE == "proxy" {
					fmt.Println("API key is not needed for ollama and proxy providers")
					return nil
				}
				timeout := 120 // default timeout
				if t, err := strconv.Atoi(TIMEOUT); err == nil {
					timeout = t
				}
				gpt3 := initGPT(PROMPT, timeout)
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
				if PROVIDER_TYPE == "ollama" || PROVIDER_TYPE == "proxy" {
					fmt.Println("API key is not needed for ollama and proxy providers")
					return nil
				}
				timeout := 120 // default timeout
				if t, err := strconv.Atoi(TIMEOUT); err == nil {
					timeout = t
				}
				gpt3 := initGPT(PROMPT, timeout)
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
				if PROVIDER_TYPE != "proxy" {
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
				if PROVIDER_TYPE != "proxy" {
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
				if t, err := strconv.Atoi(TIMEOUT); err == nil {
					timeout = t
				}
				gpt3 := initGPT(PROMPT, timeout)
				models, err := gpt3.GetAvailableModels()
				if err != nil {
					fmt.Printf("Ошибка получения моделей: %v\n", err)
					return err
				}

				fmt.Printf("Доступные модели для провайдера %s:\n", PROVIDER_TYPE)
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
				if t, err := strconv.Atoi(TIMEOUT); err == nil {
					timeout = t
				}
				gpt3 := initGPT(PROMPT, timeout)
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
			Action: func(c *cli.Context) error {
				fmt.Printf("Provider: %s\n", PROVIDER_TYPE)
				fmt.Printf("Host: %s\n", HOST)
				fmt.Printf("Model: %s\n", MODEL)
				fmt.Printf("Prompt: %s\n", PROMPT)
				fmt.Printf("Timeout: %s seconds\n", TIMEOUT)
				if PROVIDER_TYPE == "proxy" {
					fmt.Printf("JWT Token: %s\n", func() string {
						if JWT_TOKEN != "" {
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
				return nil
			},
		},
		{
			Name:    "history",
			Aliases: []string{"hist"},
			Usage:   "Show command history",
			Action: func(c *cli.Context) error {
				showHistory()
				return nil
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
					Action: func(c *cli.Context) error {
						currentUser, _ := user.Current()
						pm := gpt.NewPromptManager(currentUser.HomeDir)
						pm.ListPrompts()
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
					if t, err := strconv.Atoi(TIMEOUT); err == nil {
						timeout = t
					}
					executeMain("", prompt.Content, command, timeout)
				}

				return nil
			},
		},
	}
}

func executeMain(file, system, commandInput string, timeout int) {
	if file != "" {
		if err := reader.FileToPrompt(&commandInput, file); err != nil {
			printColored(fmt.Sprintf("❌ Ошибка чтения файла: %v\n", err), colorRed)
			return
		}
	}

	// Если system пустой, используем дефолтный промпт
	if system == "" {
		system = PROMPT
	}

	if _, err := os.Stat(RESULT_FOLDER); os.IsNotExist(err) {
		if err := os.MkdirAll(RESULT_FOLDER, 0755); err != nil {
			printColored(fmt.Sprintf("❌ Ошибка создания папки результатов: %v\n", err), colorRed)
			return
		}
	}

	gpt3 := initGPT(system, timeout)

	printColored("🤖 Запрос: ", colorCyan)
	fmt.Printf("%s\n", commandInput)

	response, elapsed := getCommand(gpt3, commandInput)
	if response == "" {
		printColored("❌ Ответ не получен. Проверьте подключение к API.\n", colorRed)
		return
	}

	printColored(fmt.Sprintf("✅ Выполнено за %.2f сек\n", elapsed), colorGreen)
	printColored("\n📋 Команда:\n", colorYellow)
	printColored(fmt.Sprintf("   %s\n\n", response), colorBold+colorGreen)

	saveToHistory(commandInput, response)
	handlePostResponse(response, gpt3, system, commandInput)
}

func initGPT(system string, timeout int) gpt.Gpt3 {
	currentUser, _ := user.Current()

	// Загружаем JWT токен в зависимости от провайдера
	var jwtToken string
	if PROVIDER_TYPE == "proxy" {
		jwtToken = JWT_TOKEN
		if jwtToken == "" {
			// Пытаемся загрузить из файла
			jwtFile := currentUser.HomeDir + "/.proxy_jwt_token"
			if data, err := os.ReadFile(jwtFile); err == nil {
				jwtToken = strings.TrimSpace(string(data))
			}
		}
	}

	return *gpt.NewGpt3(PROVIDER_TYPE, HOST, jwtToken, MODEL, system, 0.01, timeout)
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

func handlePostResponse(response string, gpt3 gpt.Gpt3, system, cmd string) {
	fmt.Printf("Действия: (c)копировать, (s)сохранить, (r)перегенерировать, (e)выполнить, (n)ничего: ")
	var choice string
	fmt.Scanln(&choice)

	switch strings.ToLower(choice) {
	case "c":
		clipboard.WriteAll(response)
		fmt.Println("✅ Команда скопирована в буфер обмена")
	case "s":
		saveResponse(response, gpt3, cmd)
	case "r":
		fmt.Println("🔄 Перегенерирую...")
		executeMain("", system, cmd, 120) // Use default timeout for regeneration
	case "e":
		executeCommand(response)
	default:
		fmt.Println(" До свидания!")
	}
}

func saveResponse(response string, gpt3 gpt.Gpt3, cmd string) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("gpt_request_%s_%s.md", gpt3.Model, timestamp)
	filePath := path.Join(RESULT_FOLDER, filename)
	content := fmt.Sprintf("## Prompt:\n\n%s\n\n## Response:\n\n%s\n", cmd+". "+gpt3.Prompt, response)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Println("Failed to save response:", err)
	} else {
		fmt.Printf("Response saved to %s\n", filePath)
	}
}

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

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

type CommandHistory struct {
	Command   string
	Response  string
	Timestamp time.Time
}

var commandHistory []CommandHistory

func saveToHistory(cmd, response string) {
	commandHistory = append(commandHistory, CommandHistory{
		Command:   cmd,
		Response:  response,
		Timestamp: time.Now(),
	})

	// Ограничиваем историю 100 командами
	if len(commandHistory) > 100 {
		commandHistory = commandHistory[1:]
	}
}

func showHistory() {
	if len(commandHistory) == 0 {
		printColored("📝 История пуста\n", colorYellow)
		return
	}

	printColored("📝 История команд:\n", colorYellow)
	for i, hist := range commandHistory {
		fmt.Printf("%d. %s → %s (%s)\n",
			i+1,
			hist.Command,
			hist.Response,
			hist.Timestamp.Format("15:04:05"))
	}
}

func printColored(text, color string) {
	fmt.Printf("%s%s%s", color, text, colorReset)
}

func showTips() {
	printColored("💡 Подсказки:\n", colorCyan)
	fmt.Println("   • Используйте --file для чтения из файла")
	fmt.Println("   • Используйте --sys для изменения системного промпта")
	fmt.Println("   • Используйте --prompt-id для выбора предустановленного промпта")
	fmt.Println("   • Используйте --timeout для установки таймаута запроса")
	fmt.Println("   • Команда 'prompts list' покажет все доступные промпты")
	fmt.Println("   • Команда 'history' покажет историю запросов")
	fmt.Println("   • Команда 'config' покажет текущие настройки")
	fmt.Println("   • Команда 'health' проверит доступность API")
}
