package main

import (
	_ "embed"
	"encoding/json"
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
	cwd, _         = os.Getwd()
	HOST           = getEnv("LCG_HOST", "http://192.168.87.108:11434/")
	COMPLETIONS    = getEnv("LCG_COMPLETIONS_PATH", "api/chat")
	MODEL          = getEnv("LCG_MODEL", "codegeex4")
	PROMPT         = getEnv("LCG_PROMPT", "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols.")
	API_KEY_FILE   = getEnv("LCG_API_KEY_FILE", ".openai_api_key")
	RESULT_FOLDER  = getEnv("LCG_RESULT_FOLDER", path.Join(cwd, "gpt_results"))
	PROVIDER_TYPE  = getEnv("LCG_PROVIDER", "ollama") // "ollama", "proxy"
	JWT_TOKEN      = getEnv("LCG_JWT_TOKEN", "")
	PROMPT_ID      = getEnv("LCG_PROMPT_ID", "1") // ID промпта по умолчанию
	TIMEOUT        = getEnv("LCG_TIMEOUT", "120") // Таймаут в секундах по умолчанию
	RESULT_HISTORY = getEnv("LCG_RESULT_HISTORY", path.Join(RESULT_FOLDER, "lcg_history.json"))
	NO_HISTORY_ENV = getEnv("LCG_NO_HISTORY", "")
)

// disableHistory управляет записью/обновлением истории на уровне процесса (флаг имеет приоритет над env)
var disableHistory bool

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
		},
		Action: func(c *cli.Context) error {
			file := c.String("file")
			system := c.String("sys")
			disableHistory = c.Bool("no-history") || isNoHistoryEnv()
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
			Subcommands: []*cli.Command{
				{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "List history entries",
					Action: func(c *cli.Context) error {
						showHistory()
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
						viewHistoryEntry(id)
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
						deleteHistoryEntry(id)
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

	// Обеспечим папку результатов заранее (может понадобиться при действиях)
	if _, err := os.Stat(RESULT_FOLDER); os.IsNotExist(err) {
		if err := os.MkdirAll(RESULT_FOLDER, 0755); err != nil {
			printColored(fmt.Sprintf("❌ Ошибка создания папки результатов: %v\n", err), colorRed)
			return
		}
	}

	// Проверка истории: если такой запрос уже встречался — предложить открыть из истории
	if !disableHistory {
		if found, hist := checkAndSuggestFromHistory(commandInput); found && hist != nil {
			gpt3 := initGPT(system, timeout)
			printColored("\nВНИМАНИЕ: ОТВЕТ СФОРМИРОВАН ИИ. ТРЕБУЕТСЯ ПРОВЕРКА И КРИТИЧЕСКИЙ АНАЛИЗ. ВОЗМОЖНЫ ОШИБКИ И ГАЛЛЮЦИНАЦИИ.\n", colorRed)
			printColored("\n📋 Команда (из истории):\n", colorYellow)
			printColored(fmt.Sprintf("   %s\n\n", hist.Response), colorBold+colorGreen)
			if strings.TrimSpace(hist.Explanation) != "" {
				printColored("\n📖 Подробное объяснение (из истории):\n\n", colorYellow)
				fmt.Println(hist.Explanation)
			}
			// Показали из истории — не выполняем запрос к API, сразу меню действий
			handlePostResponse(hist.Response, gpt3, system, commandInput, timeout)
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
	handlePostResponse(response, gpt3, system, commandInput, timeout)
}

// checkAndSuggestFromHistory проверяет файл истории и при совпадении запроса предлагает показать сохраненный результат
func checkAndSuggestFromHistory(cmd string) (bool, *CommandHistory) {
	if disableHistory {
		return false, nil
	}
	data, err := os.ReadFile(RESULT_HISTORY)
	if err != nil || len(data) == 0 {
		return false, nil
	}
	var fileHistory []CommandHistory
	if err := json.Unmarshal(data, &fileHistory); err != nil {
		return false, nil
	}
	for _, h := range fileHistory {
		if strings.TrimSpace(strings.ToLower(h.Command)) == strings.TrimSpace(strings.ToLower(cmd)) {
			fmt.Printf("\nВ истории найден похожий запрос от %s. Показать сохраненный результат? (y/N): ", h.Timestamp.Format("2006-01-02 15:04:05"))
			var ans string
			fmt.Scanln(&ans)
			if strings.ToLower(ans) == "y" || strings.ToLower(ans) == "yes" {
				return true, &h
			}
			break
		}
	}
	return false, nil
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

func handlePostResponse(response string, gpt3 gpt.Gpt3, system, cmd string, timeout int) {
	fmt.Printf("Действия: (c)копировать, (s)сохранить, (r)перегенерировать, (e)выполнить, (v|vv|vvv)подробно, (n)ничего: ")
	var choice string
	fmt.Scanln(&choice)

	switch strings.ToLower(choice) {
	case "c":
		clipboard.WriteAll(response)
		fmt.Println("✅ Команда скопирована в буфер обмена")
		if !disableHistory {
			saveToHistory(cmd, response, gpt3.Prompt)
		}
	case "s":
		saveResponse(response, gpt3, cmd)
		if !disableHistory {
			saveToHistory(cmd, response, gpt3.Prompt)
		}
	case "r":
		fmt.Println("🔄 Перегенерирую...")
		executeMain("", system, cmd, timeout)
	case "e":
		executeCommand(response)
		if !disableHistory {
			saveToHistory(cmd, response, gpt3.Prompt)
		}
	case "v", "vv", "vvv":
		level := len(choice) // 1, 2, 3
		showDetailedExplanation(response, gpt3, system, cmd, timeout, level)
	default:
		fmt.Println(" До свидания!")
		if !disableHistory {
			saveToHistory(cmd, response, gpt3.Prompt)
		}
	}
}

func saveResponse(response string, gpt3 gpt.Gpt3, cmd string) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("gpt_request_%s_%s.md", gpt3.Model, timestamp)
	filePath := path.Join(RESULT_FOLDER, filename)
	// Заголовок — сокращенный текст запроса пользователя
	title := truncateTitle(cmd)
	content := fmt.Sprintf("# %s\n\n## Prompt\n\n%s\n\n## Response\n\n%s\n", title, cmd+". "+gpt3.Prompt, response)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Println("Failed to save response:", err)
	} else {
		fmt.Printf("Response saved to %s\n", filePath)
	}
}

// saveExplanation сохраняет подробное объяснение и альтернативные способы
func saveExplanation(explanation string, model string, originalCmd string, commandResponse string) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("gpt_explanation_%s_%s.md", model, timestamp)
	filePath := path.Join(RESULT_FOLDER, filename)
	title := truncateTitle(originalCmd)
	content := fmt.Sprintf(
		"# %s\n\n## Prompt\n\n%s\n\n## Command\n\n%s\n\n## Explanation and Alternatives (model: %s)\n\n%s\n",
		title,
		originalCmd,
		commandResponse,
		model,
		explanation,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Println("Failed to save explanation:", err)
	} else {
		fmt.Printf("Explanation saved to %s\n", filePath)
	}
}

// truncateTitle сокращает строку до 120 символов (по рунам), добавляя " ..." при усечении
func truncateTitle(s string) string {
	const maxLen = 120
	if runeCount := len([]rune(s)); runeCount <= maxLen {
		return s
	}
	// взять первые 116 рунических символов и добавить " ..."
	const head = 116
	r := []rune(s)
	if len(r) <= head {
		return s
	}
	return string(r[:head]) + " ..."
}

// showDetailedExplanation делает дополнительный запрос с подробным описанием и альтернативами
func showDetailedExplanation(command string, gpt3 gpt.Gpt3, system, originalCmd string, timeout int, level int) {
	// Формируем системный промпт для подробного ответа (на русском)
	var detailedSystem string
	switch level {
	case 1: // v — кратко
		detailedSystem = "Ты опытный Linux-инженер. Объясни КРАТКО, по делу: что делает команда и самые важные ключи. Без сравнений и альтернатив. Минимум текста. Пиши на русском."
	case 2: // vv — средне
		detailedSystem = "Ты опытный Linux-инженер. Дай сбалансированное объяснение: назначение команды, разбор основных ключей, 1-2 примера. Кратко упомяни 1-2 альтернативы без глубокого сравнения. Пиши на русском."
	default: // vvv — максимально подробно
		detailedSystem = "Ты опытный Linux-инженер. Дай подробное объяснение команды с полным разбором ключей, подкоманд, сценариев применения, примеров. Затем предложи альтернативные способы решения задачи другой командой/инструментами (со сравнениями и когда что лучше применять). Пиши на русском."
	}

	// Текст запроса к модели
	ask := fmt.Sprintf("Объясни подробно команду и предложи альтернативы. Исходная команда: %s. Исходное задание пользователя: %s", command, originalCmd)

	// Создаем временный экземпляр с иным системным промптом
	detailed := gpt.NewGpt3(gpt3.ProviderType, HOST, gpt3.ApiKey, gpt3.Model, detailedSystem, 0.2, timeout)

	printColored("\n🧠 Получаю подробное объяснение...\n", colorPurple)
	explanation, elapsed := getCommand(*detailed, ask)
	if explanation == "" {
		printColored("❌ Не удалось получить подробное объяснение.\n", colorRed)
		return
	}

	printColored(fmt.Sprintf("✅ Готово за %.2f сек\n", elapsed), colorGreen)
	// Обязательное предупреждение перед выводом подробного объяснения
	printColored("\nВНИМАНИЕ: ОТВЕТ СФОРМИРОВАН ИИ. ТРЕБУЕТСЯ ПРОВЕРКА И КРИТИЧЕСКИЙ АНАЛИЗ. ВОЗМОЖНЫ ОШИБКИ И ГАЛЛЮЦИНАЦИИ.\n", colorRed)
	printColored("\n📖 Подробное объяснение и альтернативы:\n\n", colorYellow)
	fmt.Println(explanation)

	// Вторичное меню действий
	fmt.Printf("\nДействия: (c)копировать, (s)сохранить, (r)перегенерировать, (n)ничего: ")
	var choice string
	fmt.Scanln(&choice)
	switch strings.ToLower(choice) {
	case "c":
		clipboard.WriteAll(explanation)
		fmt.Println("✅ Объяснение скопировано в буфер обмена")
	case "s":
		saveExplanation(explanation, gpt3.Model, originalCmd, command)
	case "r":
		fmt.Println("🔄 Перегенерирую подробное объяснение...")
		showDetailedExplanation(command, gpt3, system, originalCmd, timeout, level)
	default:
		fmt.Println(" Возврат в основное меню.")
	}

	// После работы с объяснением — сохраняем запись в файл истории, но только если было действие не r
	if !disableHistory && (strings.ToLower(choice) == "c" || strings.ToLower(choice) == "s" || strings.ToLower(choice) == "n") {
		saveToHistory(originalCmd, command, system, explanation)
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

func isNoHistoryEnv() bool {
	v := strings.TrimSpace(NO_HISTORY_ENV)
	vLower := strings.ToLower(v)
	return vLower == "1" || vLower == "true"
}

type CommandHistory struct {
	Index       int       `json:"index"`
	Command     string    `json:"command"`
	Response    string    `json:"response"`
	Explanation string    `json:"explanation,omitempty"`
	System      string    `json:"system_prompt"`
	Timestamp   time.Time `json:"timestamp"`
}

var commandHistory []CommandHistory

func saveToHistory(cmd, response, system string, explanationOptional ...string) {
	if disableHistory {
		return
	}
	var explanation string
	if len(explanationOptional) > 0 {
		explanation = explanationOptional[0]
	}

	entry := CommandHistory{
		Index:       len(commandHistory) + 1,
		Command:     cmd,
		Response:    response,
		Explanation: explanation,
		System:      system,
		Timestamp:   time.Now(),
	}

	commandHistory = append(commandHistory, entry)

	// Ограничиваем историю 100 командами в оперативной памяти
	if len(commandHistory) > 100 {
		commandHistory = commandHistory[1:]
		// Перепривязать индексы после усечения
		for i := range commandHistory {
			commandHistory[i].Index = i + 1
		}
	}

	// Обеспечим существование папки
	if _, err := os.Stat(RESULT_FOLDER); os.IsNotExist(err) {
		_ = os.MkdirAll(RESULT_FOLDER, 0755)
	}

	// Загрузим существующий файл истории
	var fileHistory []CommandHistory
	if data, err := os.ReadFile(RESULT_HISTORY); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &fileHistory)
	}

	// Поиск дубликата по полю Command
	duplicateIndex := -1
	for i, h := range fileHistory {
		if strings.TrimSpace(strings.ToLower(h.Command)) == strings.TrimSpace(strings.ToLower(cmd)) {
			duplicateIndex = i
			break
		}
	}

	if duplicateIndex == -1 {
		// Добавляем молча, если такого запроса не было
		fileHistory = append(fileHistory, entry)
	} else {
		// Спросим о перезаписи
		fmt.Printf("\nЗапрос уже есть в истории от %s. Перезаписать? (y/N): ", fileHistory[duplicateIndex].Timestamp.Format("2006-01-02 15:04:05"))
		var ans string
		fmt.Scanln(&ans)
		if strings.ToLower(ans) == "y" || strings.ToLower(ans) == "yes" {
			entry.Index = fileHistory[duplicateIndex].Index
			fileHistory[duplicateIndex] = entry
		} else {
			// Оставляем как есть, ничего не делаем
		}
	}

	// Пересчитать индексы в файле
	for i := range fileHistory {
		fileHistory[i].Index = i + 1
	}

	if out, err := json.MarshalIndent(fileHistory, "", "  "); err == nil {
		_ = os.WriteFile(RESULT_HISTORY, out, 0644)
	}
}

func showHistory() {
	// Пытаемся прочитать историю из файла
	if disableHistory {
		printColored("📝 История отключена (--no-history / LCG_NO_HISTORY)\n", colorYellow)
		return
	}
	data, err := os.ReadFile(RESULT_HISTORY)
	if err == nil && len(data) > 0 {
		var fileHistory []CommandHistory
		if err := json.Unmarshal(data, &fileHistory); err == nil && len(fileHistory) > 0 {
			printColored("📝 История (из файла):\n", colorYellow)
			for _, hist := range fileHistory {
				ts := hist.Timestamp.Format("2006-01-02 15:04:05")
				fmt.Printf("%d. [%s] %s → %s\n", hist.Index, ts, hist.Command, hist.Response)
			}
			return
		}
	}

	// Фоллбек к памяти процесса
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

func readFileHistory() ([]CommandHistory, error) {
	if disableHistory {
		return nil, fmt.Errorf("history disabled")
	}
	data, err := os.ReadFile(RESULT_HISTORY)
	if err != nil || len(data) == 0 {
		return nil, err
	}
	var fileHistory []CommandHistory
	if err := json.Unmarshal(data, &fileHistory); err != nil {
		return nil, err
	}
	return fileHistory, nil
}

func viewHistoryEntry(id int) {
	fileHistory, err := readFileHistory()
	if err != nil || len(fileHistory) == 0 {
		fmt.Println("История пуста или недоступна")
		return
	}
	var h *CommandHistory
	for i := range fileHistory {
		if fileHistory[i].Index == id {
			h = &fileHistory[i]
			break
		}
	}
	if h == nil {
		fmt.Println("Запись не найдена")
		return
	}
	printColored("\n📋 Команда:\n", colorYellow)
	printColored(fmt.Sprintf("   %s\n\n", h.Response), colorBold+colorGreen)
	if strings.TrimSpace(h.Explanation) != "" {
		printColored("\n📖 Подробное объяснение:\n\n", colorYellow)
		fmt.Println(h.Explanation)
	}
}

func deleteHistoryEntry(id int) {
	fileHistory, err := readFileHistory()
	if err != nil || len(fileHistory) == 0 {
		fmt.Println("История пуста или недоступна")
		return
	}
	// Найти индекс элемента с совпадающим полем Index
	pos := -1
	for i := range fileHistory {
		if fileHistory[i].Index == id {
			pos = i
			break
		}
	}
	if pos == -1 {
		fmt.Println("Запись не найдена")
		return
	}
	// Удаляем элемент
	fileHistory = append(fileHistory[:pos], fileHistory[pos+1:]...)
	// Перенумеровываем индексы
	for i := range fileHistory {
		fileHistory[i].Index = i + 1
	}
	if out, err := json.MarshalIndent(fileHistory, "", "  "); err == nil {
		if err := os.WriteFile(RESULT_HISTORY, out, 0644); err != nil {
			fmt.Println("Ошибка записи истории:", err)
		} else {
			fmt.Println("Запись удалена")
		}
	} else {
		fmt.Println("Ошибка сериализации истории:", err)
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
	fmt.Println("   • Укажите --no-history чтобы не записывать историю (аналог LCG_NO_HISTORY)")
	fmt.Println("   • Команда 'prompts list' покажет все доступные промпты")
	fmt.Println("   • Команда 'history list' покажет историю запросов")
	fmt.Println("   • Команда 'config' покажет текущие настройки")
	fmt.Println("   • Команда 'health' проверит доступность API")
}
