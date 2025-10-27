# 🔧 Команда config - Управление конфигурацией

## 📋 Описание

Команда `config` позволяет просматривать текущую конфигурацию приложения, включая все настройки, переменные окружения и значения по умолчанию.

## 🚀 Использование

### Краткий вывод конфигурации (по умолчанию)

```bash
lcg config
# или
lcg co
```

**Вывод:**

``` text
Provider: ollama
Host: http://192.168.87.108:11434/
Model: hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M
Prompt: Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols.
Timeout: 300 seconds
```

### Полный вывод конфигурации

```bash
lcg config --full
# или
lcg config -f
# или
lcg co --full
# или
lcg co -f
```

**Вывод (JSON формат):**

```json
{
  "cwd": "/home/user/projects/golang/linux-command-gpt",
  "host": "http://192.168.87.108:11434/",
  "proxy_url": "/api/v1/protected/sberchat/chat",
  "completions": "api/chat",
  "model": "hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M",
  "prompt": "Reply with linux command and nothing else. Output with plain response - no need formatting. No need explanation. No need code blocks. No need ` symbols.",
  "api_key_file": ".openai_api_key",
  "result_folder": "/home/user/.config/lcg/gpt_results",
  "prompt_folder": "/home/user/.config/lcg/gpt_sys_prompts",
  "provider_type": "ollama",
  "jwt_token": "***not set***",
  "prompt_id": "1",
  "timeout": "300",
  "result_history": "/home/user/.config/lcg/gpt_results/lcg_history.json",
  "no_history_env": "",
  "allow_execution": false,
  "main_flags": {
    "file": "",
    "no_history": false,
    "sys": "",
    "prompt_id": 0,
    "timeout": 0,
    "debug": false
  },
  "server": {
    "port": "8080",
    "host": "localhost"
  },
  "validation": {
    "max_system_prompt_length": 1000,
    "max_user_message_length": 2000,
    "max_prompt_name_length": 2000,
    "max_prompt_desc_length": 5000,
    "max_command_length": 8000,
    "max_explanation_length": 20000
  }
}
```

## 📊 Структура полной конфигурации

### Основные настройки

- **cwd** - текущая рабочая директория
- **host** - адрес API сервера
- **proxy_url** - путь к API эндпоинту
- **completions** - путь к эндпоинту completions
- **model** - используемая модель ИИ
- **prompt** - системный промпт по умолчанию
- **api_key_file** - файл с API ключом
- **result_folder** - папка для сохранения результатов
- **prompt_folder** - папка с системными промптами
- **provider_type** - тип провайдера (ollama/proxy)
- **jwt_token** - статус JWT токена (***set***/***from file***/***not set***)
- **prompt_id** - ID промпта по умолчанию
- **timeout** - таймаут запросов в секундах
- **result_history** - файл истории запросов
- **no_history_env** - переменная окружения для отключения истории
- **allow_execution** - разрешение выполнения команд

### Флаги командной строки (main_flags)

- **file** - файл для чтения
- **no_history** - отключение истории
- **sys** - системный промпт
- **prompt_id** - ID промпта
- **timeout** - таймаут
- **debug** - отладочный режим

### Настройки сервера (server)

- **port** - порт веб-сервера
- **host** - хост веб-сервера

### Настройки валидации (validation)

- **max_system_prompt_length** - максимальная длина системного промпта
- **max_user_message_length** - максимальная длина пользовательского сообщения
- **max_prompt_name_length** - максимальная длина названия промпта
- **max_prompt_desc_length** - максимальная длина описания промпта
- **max_command_length** - максимальная длина команды/ответа
- **max_explanation_length** - максимальная длина объяснения

## 🔒 Безопасность

При выводе полной конфигурации чувствительные данные маскируются:

- **JWT токены** - показывается статус (***set***/***from file***/***not set***)
- **API ключи** - не выводятся в открытом виде
- **Пароли** - не сохраняются в конфигурации

## 📝 Примеры использования

### Просмотр текущих настроек

```bash
# Краткий вывод
lcg config

# Полный вывод
lcg config --full
```

### Проверка настроек валидации

```bash
# Показать только настройки валидации
lcg config --full | jq '.validation'
```

### Проверка настроек сервера

```bash
# Показать только настройки сервера
lcg config --full | jq '.server'
```

### Проверка переменных окружения

```bash
# Показать все переменные окружения LCG
env | grep LCG
```

## 🔧 Интеграция с другими инструментами

### Использование с jq

```bash
# Получить только модель
lcg config --full | jq -r '.model'

# Получить настройки валидации
lcg config --full | jq '.validation'

# Получить все пути
lcg config --full | jq '{result_folder, prompt_folder, result_history}'
```

### Использование с grep

```bash
# Найти все настройки с "timeout"
lcg config --full | grep -i timeout

# Найти все пути
lcg config --full | grep -E "(folder|history)"
```

### Сохранение конфигурации в файл

```bash
# Сохранить полную конфигурацию
lcg config --full > config.json

# Сохранить только настройки валидации
lcg config --full | jq '.validation' > validation.json
```

## 🐛 Отладка

### Проверка загрузки конфигурации

```bash
# Показать все настройки
lcg config --full

# Проверить переменные окружения
env | grep LCG

# Проверить файлы конфигурации
ls -la ~/.config/lcg/
```

### Типичные проблемы

1. **Неправильные пути** - проверьте `result_folder` и `prompt_folder`
2. **Отсутствующие токены** - проверьте `jwt_token` статус
3. **Неправильные лимиты** - проверьте секцию `validation`

## 📚 Связанные команды

- `lcg --help` - общая справка
- `lcg config --help` - справка по команде config
- `lcg serve` - запуск веб-сервера
- `lcg prompts list` - список промптов

---

**Примечание**: Команда `config` показывает актуальное состояние конфигурации после применения всех переменных окружения и значений по умолчанию.
