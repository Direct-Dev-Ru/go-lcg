package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type HistoryEntry struct {
	Index       int       `json:"index"`
	Command     string    `json:"command"`
	Response    string    `json:"response"`
	Explanation string    `json:"explanation,omitempty"`
	System      string    `json:"system_prompt"`
	Timestamp   time.Time `json:"timestamp"`
}

func read(historyPath string) ([]HistoryEntry, error) {
	data, err := os.ReadFile(historyPath)
	if err != nil || len(data) == 0 {
		return nil, err
	}
	var items []HistoryEntry
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func write(historyPath string, entries []HistoryEntry) error {
	for i := range entries {
		entries[i].Index = i + 1
	}
	out, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(historyPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(historyPath, out, 0644)
}

func ShowHistory(historyPath string, printColored func(string, string), colorYellow string) {
	items, err := read(historyPath)
	if err != nil || len(items) == 0 {
		printColored("📝 История пуста\n", colorYellow)
		return
	}
	printColored("📝 История (из файла):\n", colorYellow)
	for _, h := range items {
		ts := h.Timestamp.Format("2006-01-02 15:04:05")
		fmt.Printf("%d. [%s] %s → %s\n", h.Index, ts, h.Command, h.Response)
	}
}

func ViewHistoryEntry(historyPath string, id int, printColored func(string, string), colorYellow, colorBold, colorGreen string) {
	items, err := read(historyPath)
	if err != nil || len(items) == 0 {
		fmt.Println("История пуста или недоступна")
		return
	}
	var h *HistoryEntry
	for i := range items {
		if items[i].Index == id {
			h = &items[i]
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

func DeleteHistoryEntry(historyPath string, id int) error {
	items, err := read(historyPath)
	if err != nil || len(items) == 0 {
		return fmt.Errorf("история пуста или недоступна")
	}
	pos := -1
	for i := range items {
		if items[i].Index == id {
			pos = i
			break
		}
	}
	if pos == -1 {
		return fmt.Errorf("запись не найдена")
	}
	items = append(items[:pos], items[pos+1:]...)
	return write(historyPath, items)
}

func SaveToHistory(historyPath, resultFolder, cmdText, response, system string, explanationOptional ...string) error {
	var explanation string
	if len(explanationOptional) > 0 {
		explanation = explanationOptional[0]
	}
	items, _ := read(historyPath)
	duplicateIndex := -1
	for i, h := range items {
		if strings.EqualFold(strings.TrimSpace(h.Command), strings.TrimSpace(cmdText)) {
			duplicateIndex = i
			break
		}
	}
	entry := HistoryEntry{
		Index:       len(items) + 1,
		Command:     cmdText,
		Response:    response,
		Explanation: explanation,
		System:      system,
		Timestamp:   time.Now(),
	}
	if duplicateIndex == -1 {
		items = append(items, entry)
		return write(historyPath, items)
	}
	fmt.Printf("\nЗапрос уже есть в истории от %s. Перезаписать? (y/N): ", items[duplicateIndex].Timestamp.Format("2006-01-02 15:04:05"))
	var ans string
	fmt.Scanln(&ans)
	if strings.ToLower(ans) == "y" || strings.ToLower(ans) == "yes" {
		entry.Index = items[duplicateIndex].Index
		items[duplicateIndex] = entry
		return write(historyPath, items)
	}
	return nil
}

// SaveToHistoryFromHistory сохраняет запись из истории без запроса о перезаписи
func SaveToHistoryFromHistory(historyPath, resultFolder, cmdText, response, system, explanation string) error {
	items, _ := read(historyPath)
	duplicateIndex := -1
	for i, h := range items {
		if strings.EqualFold(strings.TrimSpace(h.Command), strings.TrimSpace(cmdText)) {
			duplicateIndex = i
			break
		}
	}
	entry := HistoryEntry{
		Index:       len(items) + 1,
		Command:     cmdText,
		Response:    response,
		Explanation: explanation,
		System:      system,
		Timestamp:   time.Now(),
	}
	if duplicateIndex == -1 {
		items = append(items, entry)
		return write(historyPath, items)
	}
	// Если дубликат найден, перезаписываем без запроса
	entry.Index = items[duplicateIndex].Index
	items[duplicateIndex] = entry
	return write(historyPath, items)
}

func CheckAndSuggestFromHistory(historyPath, cmdText string) (bool, *HistoryEntry) {
	items, err := read(historyPath)
	if err != nil || len(items) == 0 {
		return false, nil
	}
	for _, h := range items {
		if strings.EqualFold(strings.TrimSpace(h.Command), strings.TrimSpace(cmdText)) {
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
