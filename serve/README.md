# Пакет serve

Этот пакет содержит HTTP сервер для веб-интерфейса LCG (Linux Command GPT).

## Структура файлов

### serve.go

Основной файл пакета. Содержит:

- `StartResultServer()` - функция запуска HTTP сервера
- `registerRoutes()` - регистрация всех маршрутов

### results.go

Обработчики для результатов и файлов:

- `handleResultsPage()` - главная страница со списком файлов результатов
- `handleFileView()` - просмотр конкретного файла
- `handleDeleteFile()` - удаление файла результата
- `getResultFiles()` - получение списка файлов
- `formatFileSize()` - форматирование размера файла

### history.go

Обработчики для работы с историей запросов:

- `handleHistoryPage()` - страница истории запросов
- `handleDeleteHistoryEntry()` - удаление записи из истории
- `handleClearHistory()` - очистка всей истории
- `readHistoryEntries()` - чтение записей истории

### history_utils.go

Утилиты для работы с историей:

- `HistoryEntry` - структура записи истории
- `read()` - чтение истории из файла
- `write()` - запись истории в файл
- `DeleteHistoryEntry()` - удаление записи по индексу

### prompts.go

Обработчики для управления промптами:

- `handlePromptsPage()` - страница управления промптами
- `handleAddPrompt()` - добавление нового промпта
- `handleEditPrompt()` - редактирование промпта
- `handleDeletePrompt()` - удаление промпта
- `handleRestorePrompt()` - восстановление системного промпта к значению по умолчанию
- `handleRestoreVerbosePrompt()` - восстановление verbose промпта
- `handleSaveLang()` - сохранение промптов при переключении языка

### prompts_helpers.go

Вспомогательные функции для работы с промптами:

- `getVerbosePromptsFromFile()` - получение verbose промптов из файла
- `translateVerbosePrompt()` - перевод verbose промпта
- `getVerbosePrompts()` - получение встроенных verbose промптов (fallback)
- `getSystemPromptsWithLang()` - получение системных промптов с учетом языка
- `translateSystemPrompt()` - перевод системного промпта

## Использование

```go
import "github.com/direct-dev-ru/linux-command-gpt/serve"

// Запуск сервера на localhost:8080
err := serve.StartResultServer("localhost", "8080")
```

## Маршруты

### Результаты

- `GET /` - главная страница со списком файлов
- `GET /file/{filename}` - просмотр файла результата
- `DELETE /delete/{filename}` - удаление файла

### История

- `GET /history` - страница истории запросов
- `GET /history/view/{id}` - просмотр записи истории в развернутом виде
- `DELETE /history/delete/{id}` - удаление записи
- `DELETE /history/clear` - очистка всей истории

### Промпты

- `GET /prompts` - страница управления промптами
- `POST /prompts/add` - добавление промпта
- `PUT /prompts/edit/{id}` - редактирование промпта
- `DELETE /prompts/delete/{id}` - удаление промпта
- `POST /prompts/restore/{id}` - восстановление системного промпта
- `POST /prompts/restore-verbose/{mode}` - восстановление verbose промпта (v/vv/vvv)
- `POST /prompts/save-lang` - сохранение языка промптов

### Выполнение запросов

- `GET /run` - веб-страница для выполнения запросов
- `POST /run` - обработка выполнения запроса
- `POST /execute` - API для программного доступа (только curl)

## Особенности

- **Многоязычность**: Поддержка английского и русского языков для промптов
- **Responsive дизайн**: Адаптивный интерфейс для различных устройств
- **Markdown**: Автоматическая конвертация Markdown файлов в HTML
- **История**: Поиск дубликатов с учетом регистра
- **Промпты**: Управление встроенными и пользовательскими промптами
