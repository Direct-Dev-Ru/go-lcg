package serve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// HistoryEntry представляет запись в истории
type HistoryEntry struct {
	Index       int       `json:"index"`
	Command     string    `json:"command"`
	Response    string    `json:"response"`
	Explanation string    `json:"explanation,omitempty"`
	System      string    `json:"system_prompt"`
	Timestamp   time.Time `json:"timestamp"`
}

// read читает записи истории из файла
func Read(historyPath string) ([]HistoryEntry, error) {
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

// write записывает записи истории в файл
func Write(historyPath string, entries []HistoryEntry) error {
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

// DeleteHistoryEntry удаляет запись из истории по индексу
func DeleteHistoryEntry(historyPath string, id int) error {
	items, err := Read(historyPath)
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
	return Write(historyPath, items)
}
