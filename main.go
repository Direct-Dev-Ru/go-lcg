package main

import (
	_ "embed"
	"fmt"
	"math"
	"os"
	"os/user"
	"path"
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
)

func main() {
	app := &cli.App{
		Name:     "lcg",
		Usage:    "Linux Command GPT - Generate Linux commands from descriptions",
		Version:  Version,
		Commands: getCommands(),
		UsageText: `
lcg [global options] <command description>

Examples:
  lcg "I want to extract linux-command-gpt.tar.gz file"
  lcg --file /path/to/file.txt "I want to list all directories with ls"
`,
		Description: `
Linux Command GPT is a tool for generating Linux commands from natural language descriptions. 
It supports reading parts of the prompt from files and allows saving, copying, or regenerating results.
Additional commands are available for managing API keys.

Environment Variables:
  LCG_HOST          Endpoint for LLM API (default: http://192.168.87.108:11434/)
  LCG_COMPLETIONS_PATH  Relative API path (default: api/chat)
  LCG_MODEL         Model name (default: codegeex4)
  LCG_PROMPT        Default prompt text
  LCG_API_KEY_FILE  API key storage file (default: ~/.openai_api_key)
  LCG_RESULT_FOLDER Results folder (default: ./gpt_results)
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
				Usage:       "System prompt",
				DefaultText: getEnv("LCG_PROMPT", "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks"),
				Value:       getEnv("LCG_PROMPT", "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks"),
			},
		},
		Action: func(c *cli.Context) error {
			file := c.String("file")
			system := c.String("sys")
			args := c.Args().Slice()
			if len(args) == 0 {
				cli.ShowAppHelp(c)
				return nil
			}
			executeMain(file, system, strings.Join(args, " "))
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
				gpt3 := initGPT()
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
				gpt3 := initGPT()
				gpt3.DeleteKey()
				fmt.Println("API key deleted.")
				return nil
			},
		},
	}
}

func executeMain(file, system, commandInput string) {
	// fmt.Println(system, commandInput)
	// os.Exit(0)
	if file != "" {
		if err := reader.FileToPrompt(&commandInput, file); err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
	}

	if _, err := os.Stat(RESULT_FOLDER); os.IsNotExist(err) {
		os.MkdirAll(RESULT_FOLDER, 0755)
	}

	gpt3 := initGPT()

	response, elapsed := getCommand(gpt3, commandInput)
	if response == "" {
		fmt.Println("No response received.")
		return
	}

	fmt.Printf("Completed in %v seconds\n\n%s\n", elapsed, response)
	handlePostResponse(response, gpt3, system, commandInput)
}

func initGPT() gpt.Gpt3 {
	currentUser, _ := user.Current()
	return gpt.Gpt3{
		CompletionUrl: HOST + COMPLETIONS,
		Model:         MODEL,
		Prompt:        PROMPT,
		HomeDir:       currentUser.HomeDir,
		ApiKeyFile:    API_KEY_FILE,
		Temperature:   0.01,
	}
}

func getCommand(gpt3 gpt.Gpt3, cmd string) (string, float64) {
	gpt3.InitKey()
	start := time.Now()
	done := make(chan bool)
	go func() {
		loadingChars := []rune{'-', '\\', '|', '/'}
		i := 0
		for {
			select {
			case <-done:
				fmt.Printf("\r")
				return
			default:
				fmt.Printf("\rLoading %c", loadingChars[i])
				i = (i + 1) % len(loadingChars)
				time.Sleep(30 * time.Millisecond)
			}
		}
	}()

	response := gpt3.Completions(cmd)
	done <- true
	elapsed := math.Round(time.Since(start).Seconds()*100) / 100

	return response, elapsed
}

func handlePostResponse(response string, gpt3 gpt.Gpt3, system, cmd string) {
	fmt.Print("\nOptions: (c)opy, (s)ave, (r)egenerate, (n)one: ")
	var choice string
	fmt.Scanln(&choice)

	switch strings.ToLower(choice) {
	case "c":
		clipboard.WriteAll(response)
		fmt.Println("Response copied to clipboard.")
	case "s":
		saveResponse(response, gpt3, cmd)
	case "r":
		executeMain("", system, cmd)
	default:
		fmt.Println("No action taken.")
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

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
