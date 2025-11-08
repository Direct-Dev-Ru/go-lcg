# 🔧 Конфигурация валидации длины полей

## 📋 Переменные окружения

Все настройки валидации можно настроить через переменные окружения:

### Основные лимиты

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `LCG_MAX_SYSTEM_PROMPT_LENGTH` | Максимальная длина системного промпта | 2000 |
| `LCG_MAX_USER_MESSAGE_LENGTH` | Максимальная длина пользовательского сообщения | 4000 |
| `LCG_MAX_PROMPT_NAME_LENGTH` | Максимальная длина названия промпта | 2000 |
| `LCG_MAX_PROMPT_DESC_LENGTH` | Максимальная длина описания промпта | 5000 |
| `LCG_MAX_COMMAND_LENGTH` | Максимальная длина команды/ответа | 8000 |
| `LCG_MAX_EXPLANATION_LENGTH` | Максимальная длина объяснения | 20000 |

## 🚀 Примеры использования

### Установка через переменные окружения

```bash
# Увеличить лимит системного промпта до 3к символов
export LCG_MAX_SYSTEM_PROMPT_LENGTH=3000

# Уменьшить лимит пользовательского сообщения до 2к символов
export LCG_MAX_USER_MESSAGE_LENGTH=2000

# Увеличить лимит названия промпта до 3000 символов
export LCG_MAX_PROMPT_NAME_LENGTH=3000
```

### Установка в .env файле

```bash
# .env файл
LCG_MAX_SYSTEM_PROMPT_LENGTH=2000
LCG_MAX_USER_MESSAGE_LENGTH=4000
LCG_MAX_PROMPT_NAME_LENGTH=2000
LCG_MAX_PROMPT_DESC_LENGTH=5000
LCG_MAX_COMMAND_LENGTH=8000
LCG_MAX_EXPLANATION_LENGTH=20000
```

### Установка в systemd сервисе

```ini
[Unit]
Description=Linux Command GPT
After=network.target

[Service]
Type=simple
User=lcg
WorkingDirectory=/opt/lcg
ExecStart=/opt/lcg/lcg serve
Environment=LCG_MAX_SYSTEM_PROMPT_LENGTH=2000
Environment=LCG_MAX_USER_MESSAGE_LENGTH=4000
Environment=LCG_MAX_PROMPT_NAME_LENGTH=2000
Restart=always

[Install]
WantedBy=multi-user.target
```

### Установка в Docker

```dockerfile
FROM golang:1.21-alpine AS builder
# ... build steps ...

FROM alpine:latest
COPY --from=builder /app/lcg /usr/local/bin/
ENV LCG_MAX_SYSTEM_PROMPT_LENGTH=2000
ENV LCG_MAX_USER_MESSAGE_LENGTH=4000
CMD ["lcg", "serve"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  lcg:
    image: lcg:latest
    environment:
      - LCG_MAX_SYSTEM_PROMPT_LENGTH=2000
      - LCG_MAX_USER_MESSAGE_LENGTH=4000
      - LCG_MAX_PROMPT_NAME_LENGTH=2000
    ports:
      - "8080:8080"
```

## 🔍 Где применяется валидация

### 1. Консольная часть (main.go)
- ✅ Валидация пользовательского сообщения
- ✅ Валидация системного промпта
- ✅ Цветные сообщения об ошибках

### 2. API эндпоинты
- ✅ `/execute` - валидация промпта и системного промпта
- ✅ `/api/save-result` - валидация всех полей
- ✅ `/api/add-to-history` - валидация всех полей

### 3. Веб-интерфейс
- ✅ Страница выполнения - валидация в JavaScript и на сервере
- ✅ Управление промптами - валидация всех полей формы

### 4. JavaScript валидация
- ✅ Клиентская валидация перед отправкой
- ✅ Динамические лимиты из конфигурации
- ✅ Понятные сообщения об ошибках

## 🛠️ Технические детали

### Структура конфигурации

```go
type ValidationConfig struct {
    MaxSystemPromptLength int  // LCG_MAX_SYSTEM_PROMPT_LENGTH
    MaxUserMessageLength  int  // LCG_MAX_USER_MESSAGE_LENGTH
    MaxPromptNameLength   int  // LCG_MAX_PROMPT_NAME_LENGTH
    MaxPromptDescLength   int  // LCG_MAX_PROMPT_DESC_LENGTH
    MaxCommandLength      int  // LCG_MAX_COMMAND_LENGTH
    MaxExplanationLength  int  // LCG_MAX_EXPLANATION_LENGTH
}
```

### Функции валидации

```go
// Основные функции
validation.ValidateSystemPrompt(prompt)
validation.ValidateUserMessage(message)
validation.ValidatePromptName(name)
validation.ValidatePromptDescription(description)
validation.ValidateCommand(command)
validation.ValidateExplanation(explanation)

// Вспомогательные функции
validation.TruncateSystemPrompt(prompt)
validation.TruncateUserMessage(message)
validation.FormatLengthInfo(systemPrompt, userMessage)
```

### Обработка ошибок

- **API**: HTTP 400 с JSON сообщением об ошибке
- **Веб-интерфейс**: HTTP 400 с текстовым сообщением
- **Консоль**: Цветные сообщения об ошибках
- **JavaScript**: Alert с предупреждением

## 📝 Примеры сообщений об ошибках

```
❌ Ошибка: system_prompt: системный промпт слишком длинный: 2500 символов (максимум 2000)
❌ Ошибка: user_message: пользовательское сообщение слишком длинное: 4500 символов (максимум 4000)
❌ Ошибка: prompt_name: название промпта слишком длинное: 2500 символов (максимум 2000)
```

## 🔄 Миграция с жестко заданных значений

Если ранее использовались жестко заданные значения в коде, теперь они автоматически заменяются на значения из конфигурации:

```go
// Старый код
if len(prompt) > 2000 {
    return errors.New("too long")
}

// Новый код
if err := validation.ValidateSystemPrompt(prompt); err != nil {
    return err
}
```

## 🎯 Рекомендации по настройке

### Для разработки
```bash
export LCG_MAX_SYSTEM_PROMPT_LENGTH=2000
export LCG_MAX_USER_MESSAGE_LENGTH=4000
export LCG_MAX_PROMPT_NAME_LENGTH=2000
export LCG_MAX_PROMPT_DESC_LENGTH=5000
```

### Для продакшена
```bash
export LCG_MAX_SYSTEM_PROMPT_LENGTH=2000
export LCG_MAX_USER_MESSAGE_LENGTH=4000
export LCG_MAX_PROMPT_NAME_LENGTH=2000
export LCG_MAX_PROMPT_DESC_LENGTH=5000
```

### Для высоконагруженных систем
```bash
export LCG_MAX_SYSTEM_PROMPT_LENGTH=1000
export LCG_MAX_USER_MESSAGE_LENGTH=2000
export LCG_MAX_PROMPT_NAME_LENGTH=1000
export LCG_MAX_PROMPT_DESC_LENGTH=2500
```

---

**Примечание**: Все значения настраиваются через переменные окружения и применяются ко всем частям приложения (консоль, веб-интерфейс, API).
