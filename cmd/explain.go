package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/direct-dev-ru/linux-command-gpt/config"
	"github.com/direct-dev-ru/linux-command-gpt/gpt"
)

// ExplainDeps инъекция зависимостей для вывода и окружения
type ExplainDeps struct {
	DisableHistory bool
	PrintColored   func(string, string)
	ColorPurple    string
	ColorGreen     string
	ColorRed       string
	ColorYellow    string
	GetCommand     func(gpt.Gpt3, string) (string, float64)
}

// ShowDetailedExplanation делает дополнительный запрос с подробным описанием и альтернативами
func ShowDetailedExplanation(command string, gpt3 gpt.Gpt3, system, originalCmd string, timeout int, level int, deps ExplainDeps) {
	var detailedSystem string
	switch level {
	case 1: // v — кратко
		detailedSystem = "Ты опытный Linux-инженер. Объясни КРАТКО, по делу: что делает команда и самые важные ключи. Без сравнений и альтернатив. Минимум текста. Пиши на русском."
	case 2: // vv — средне
		detailedSystem = "Ты опытный Linux-инженер. Дай сбалансированное объяснение: назначение команды, разбор основных ключей, 1-2 примера. Кратко упомяни 1-2 альтернативы без глубокого сравнения. Пиши на русском."
	default: // vvv — максимально подробно
		detailedSystem = "Ты опытный Linux-инженер. Дай подробное объяснение команды с полным разбором ключей, подкоманд, сценариев применения, примеров. Затем предложи альтернативные способы решения задачи другой командой/инструментами (со сравнениями и когда что лучше применять). Пиши на русском."
	}

	ask := fmt.Sprintf("Объясни подробно команду и предложи альтернативы. Исходная команда: %s. Исходное задание пользователя: %s", command, originalCmd)
	detailed := gpt.NewGpt3(gpt3.ProviderType, config.AppConfig.Host, gpt3.ApiKey, gpt3.Model, detailedSystem, 0.2, timeout)

	deps.PrintColored("\n🧠 Получаю подробное объяснение...\n", deps.ColorPurple)
	explanation, elapsed := deps.GetCommand(*detailed, ask)
	if explanation == "" {
		deps.PrintColored("❌ Не удалось получить подробное объяснение.\n", deps.ColorRed)
		return
	}

	deps.PrintColored(fmt.Sprintf("✅ Готово за %.2f сек\n", elapsed), deps.ColorGreen)
	deps.PrintColored("\nВНИМАНИЕ: ОТВЕТ СФОРМИРОВАН ИИ. ТРЕБУЕТСЯ ПРОВЕРКА И КРИТИЧЕСКИЙ АНАЛИЗ. ВОЗМОЖНЫ ОШИБКИ И ГАЛЛЮЦИНАЦИИ.\n", deps.ColorRed)
	deps.PrintColored("\n📖 Подробное объяснение и альтернативы:\n\n", deps.ColorYellow)
	fmt.Println(explanation)

	fmt.Printf("\nДействия: (c)копировать, (s)сохранить, (r)перегенерировать, (n)ничего: ")
	var choice string
	fmt.Scanln(&choice)
	switch strings.ToLower(choice) {
	case "c":
		clipboard.WriteAll(explanation)
		fmt.Println("✅ Объяснение скопировано в буфер обмена")
	case "s":
		saveExplanation(explanation, gpt3.Model, originalCmd, command, config.AppConfig.ResultFolder)
	case "r":
		fmt.Println("🔄 Перегенерирую подробное объяснение...")
		ShowDetailedExplanation(command, gpt3, system, originalCmd, timeout, level, deps)
	default:
		fmt.Println(" Возврат в основное меню.")
	}

	if !deps.DisableHistory && (strings.ToLower(choice) == "c" || strings.ToLower(choice) == "s" || strings.ToLower(choice) == "n") {
		SaveToHistory(config.AppConfig.ResultHistory, config.AppConfig.ResultFolder, originalCmd, command, system, explanation)
	}
}

// saveExplanation сохраняет подробное объяснение и альтернативные способы
func saveExplanation(explanation string, model string, originalCmd string, commandResponse string, resultFolder string) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("gpt_explanation_%s_%s.md", model, timestamp)
	filePath := path.Join(resultFolder, filename)
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
		fmt.Printf("Saved to %s\n", filePath)
	}
}

// truncateTitle сокращает строку до 120 символов (по рунам), добавляя " ..." при усечении
func truncateTitle(s string) string {
	const maxLen = 120
	if runeCount := len([]rune(s)); runeCount <= maxLen {
		return s
	}
	const head = 116
	r := []rune(s)
	if len(r) <= head {
		return s
	}
	return string(r[:head]) + " ..."
}
