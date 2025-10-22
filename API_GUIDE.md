# API Guide - Linux Command GPT

## Обзор

API позволяет выполнять запросы к Linux Command GPT через HTTP POST запросы с помощью curl. API принимает только запросы от curl (проверка User-Agent).

## Endpoint

``` curl
POST /execute
```

## Запуск сервера

```bash
# Запуск сервера
lcg serve

# Запуск на другом порту
lcg serve --port 9000

# Запуск с автоматическим открытием браузера
lcg serve --browser
```

## Структура запроса

### JSON Payload

```json
{
  "prompt": "создать директорию test",
  "system_id": 1,
  "system": "альтернативный системный промпт",
  "verbose": "vv",
  "timeout": 120
}
```

### Поля запроса

| Поле | Тип | Обязательное | Описание |
|------|-----|--------------|----------|
| `prompt` | string | ✅ | Пользовательский запрос |
| `system_id` | int | ❌ | ID системного промпта (1-5) |
| `system` | string | ❌ | Текст системного промпта (альтернатива system_id) |
| `verbose` | string | ❌ | Степень подробности: "v", "vv", "vvv" |
| `timeout` | int | ❌ | Таймаут в секундах (по умолчанию: 120) |

### Структура ответа

```json
{
  "success": true,
  "command": "mkdir test",
  "explanation": "Команда mkdir создает новую директорию...",
  "model": "hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M",
  "elapsed": 2.34
}
```

## Примеры использования

### 1. Базовый запрос

```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "создать директорию test"
  }'
```

**Ответ:**

```json
{
  "success": true,
  "command": "mkdir test",
  "model": "hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M",
  "elapsed": 1.23
}
```

### 2. Запрос с системным промптом по ID

```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "найти все файлы .txt",
    "system_id": 2
  }'
```

### 3. Запрос с кастомным системным промптом

```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "показать использование памяти",
    "system": "Ты эксперт по Linux. Отвечай только командами без объяснений."
  }'
```

### 4. Запрос с подробным объяснением

```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "архивировать папку documents",
    "verbose": "vv",
    "timeout": 180
  }'
```

**Ответ:**

```json
{
  "success": true,
  "command": "tar -czf documents.tar.gz documents/",
  "explanation": "Команда tar создает архив в формате gzip...",
  "model": "hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M",
  "elapsed": 3.45
}
```

### 5. Запрос с максимальной подробностью

```bash
curl -X POST http://localhost:8080/execute \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "настроить SSH сервер",
    "system_id": 3,
    "verbose": "vvv",
    "timeout": 300
  }'
```

## Системные промпты

| ID | Название | Описание |
|----|----------|----------|
| 1 | basic | Базовые команды Linux |
| 2 | advanced | Продвинутые команды |
| 3 | system | Системное администрирование |
| 4 | network | Сетевые команды |
| 5 | security | Безопасность |

## Степени подробности

| Уровень | Описание |
|---------|----------|
| `v` | Краткое объяснение |
| `vv` | Подробное объяснение с альтернативами |
| `vvv` | Максимально подробное объяснение с примерами |

## Обработка ошибок

### Ошибка валидации

```json
{
  "success": false,
  "error": "Prompt is required"
}
```

### Ошибка AI

```json
{
  "success": false,
  "error": "Failed to get response from AI"
}
```

### Ошибка доступа

``` text
HTTP 403 Forbidden
Only curl requests are allowed
```

## Переменные окружения

Убедитесь, что настроены необходимые переменные:

```bash
# Основные настройки
export LCG_PROVIDER="ollama"
export LCG_HOST="http://localhost:11434"
export LCG_MODEL="hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M"

# Для proxy провайдера
export LCG_PROVIDER="proxy"
export LCG_HOST="https://your-proxy-server.com"
export LCG_JWT_TOKEN="your-jwt-token"
```

## Безопасность

- ✅ **Только curl**: API принимает только запросы от curl
- ✅ **POST только**: Только POST запросы к `/execute`
- ✅ **JSON валидация**: Строгая проверка входных данных
- ✅ **Таймауты**: Ограничение времени выполнения запросов

## Примеры скриптов

### Bash скрипт для автоматизации

```bash
#!/bin/bash

API_URL="http://localhost:8080/execute"

# Функция для выполнения запроса
execute_command() {
    local prompt="$1"
    local verbose="${2:-}"
    
    curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -d "{\"prompt\": \"$prompt\", \"verbose\": \"$verbose\"}" | \
        jq -r '.command'
}

# Использование
echo "Команда: $(execute_command "создать директорию backup")"
```

### Python скрипт

```python
import requests
import json

def execute_command(prompt, system_id=None, verbose=None):
    url = "http://localhost:8080/execute"
    payload = {"prompt": prompt}
    
    if system_id:
        payload["system_id"] = system_id
    if verbose:
        payload["verbose"] = verbose
    
    response = requests.post(url, json=payload)
    return response.json()

# Использование
result = execute_command("показать использование диска", verbose="vv")
print(f"Команда: {result['command']}")
if 'explanation' in result:
    print(f"Объяснение: {result['explanation']}")
```
