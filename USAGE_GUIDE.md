# Руководство по использованию (USAGE_GUIDE)

## Что это

Linux Command GPT (`lcg`) преобразует описание на естественном языке в готовую Linux‑команду. Инструмент поддерживает сменные провайдеры LLM (Ollama или Proxy), управление системными промптами, историю за сессию, сохранение результатов и интерактивные действия над сгенерированной командой.

## Требования

- Установленный Go (для сборки из исходников) или готовый бинарник.
- Для функции «скопировать в буфер обмена»: установите `xclip` или `xsel`.

```bash
# Debian/Ubuntu
sudo apt-get install xclip
# или
sudo apt-get install xsel
```

## Установка

Сборка из исходников:

```bash
git clone --depth 1 https://github.com/asrul10/linux-command-gpt.git ~/.linux-command-gpt
cd ~/.linux-command-gpt
go build -o lcg

# Добавьте бинарник в PATH
ln -s ~/.linux-command-gpt/lcg ~/.local/bin
```

Или скачайте готовый бинарник из раздела релизов.

## Быстрый старт

Простой запрос:

```bash
lcg "хочу извлечь файл linux-command-gpt.tar.gz"
```

Смешанный ввод: часть из файла, часть — текстом:

```bash
lcg --file /path/to/context.txt "хочу вывести список директорий с помощью ls"
```

После генерации вы увидите:

```text
🤖 Запрос: <ваше описание>
✅ Выполнено за X.XX сек

📋 Команда:
   <сгенерированная команда>

Действия: (c)копировать, (s)сохранить, (r)перегенерировать, (e)выполнить, (n)ничего:
```

## Переменные окружения

Можно настроить поведение без изменения командной строки.

| Переменная | Значение по умолчанию | Назначение |
| --- | --- | --- |
| `LCG_HOST` | `http://192.168.87.108:11434/` | Базовый URL API провайдера (для Ollama поставьте, например, `http://localhost:11434/`). |
| `LCG_COMPLETIONS_PATH` | `api/chat` | Относительный путь эндпоинта для Ollama. |
| `LCG_MODEL` | `codegeex4` | Имя модели у выбранного провайдера. |
| `LCG_PROMPT` | См. значение в коде | Содержимое системного промпта по умолчанию. |
| `LCG_API_KEY_FILE` | `.openai_api_key` | Файл с API‑ключом (для Ollama/Proxy не требуется). |
| `LCG_RESULT_FOLDER` | `$(pwd)/gpt_results` | Папка для сохранения результатов. |
| `LCG_PROVIDER` | `ollama` | Тип провайдера: `ollama` или `proxy`. |
| `LCG_JWT_TOKEN` | пусто | JWT токен для `proxy` провайдера (альтернатива — файл `~/.proxy_jwt_token`). |
| `LCG_PROMPT_ID` | `1` | ID системного промпта по умолчанию. |
| `LCG_TIMEOUT` | `120` | Таймаут запроса в секундах. |

Примеры настройки:

```bash
# Ollama
export LCG_PROVIDER=ollama
export LCG_HOST=http://localhost:11434/
export LCG_MODEL=codegeex4

# Proxy
export LCG_PROVIDER=proxy
export LCG_HOST=http://localhost:8080
export LCG_MODEL=GigaChat-2
export LCG_JWT_TOKEN=your_jwt_token_here
```

## Базовый синтаксис

```bash
lcg [глобальные опции] <описание команды>
```

Глобальные опции:

- `--file, -f string` — прочитать часть запроса из файла и добавить к описанию.
- `--sys, -s string` — системный промпт (содержимое или ID как строка). Если не задан, используется `--prompt-id` или `LCG_PROMPT`.
- `--prompt-id, --pid int` — ID системного промпта (1–5 для стандартных, либо ваш кастомный ID).
- `--timeout, -t int` — таймаут запроса в секундах (по умолчанию 120).
- `--version, -v` — вывести версию.
- `--help, -h` — помощь.

## Подкоманды

- `lcg update-key` (`-u`): обновить API‑ключ. Для `ollama` и `proxy` не требуется — команда сообщит, что ключ не нужен.
- `lcg delete-key` (`-d`): удалить API‑ключ (не требуется для `ollama`/`proxy`).
- `lcg update-jwt` (`-j`): обновить JWT для `proxy`. Токен будет сохранён в `~/.proxy_jwt_token` (права `0600`).
- `lcg delete-jwt` (`-dj`): удалить JWT файл для `proxy`.
- `lcg models` (`-m`): показать доступные модели у текущего провайдера.
- `lcg health` (`-he`): проверить доступность API провайдера.
- `lcg config` (`-co`): показать текущую конфигурацию и состояние JWT.
- `lcg history` (`-hist`): показать историю запросов за текущий запуск (до 100 записей, не сохраняется между запусками).
- `lcg prompts ...` (`-p`): управление системными промптами:
  - `lcg prompts list` (`-l`) — список всех промптов.
  - `lcg prompts add` (`-a`) — добавить пользовательский промпт (по шагам в интерактиве).
  - `lcg prompts delete <id>` (`-d`) — удалить пользовательский промпт по ID (>5).
- `lcg test-prompt <prompt-id> <описание>` (`-tp`): показать детали выбранного системного промпта и протестировать его на заданном описании.

## Провайдеры

### Ollama (`LCG_PROVIDER=ollama`)

- Требуется запущенный Ollama API (`LCG_HOST`, например `http://localhost:11434/`).
- `models`, `health` и генерация используют REST Ollama (`/api/tags`, `/api/chat`).
- API‑ключ не нужен.

### Proxy (`LCG_PROVIDER=proxy`)

- Требуется доступ к прокси‑серверу (`LCG_HOST`) и JWT (`LCG_JWT_TOKEN` или файл `~/.proxy_jwt_token`).
- Основные эндпоинты: `/api/v1/protected/sberchat/chat` и `/api/v1/protected/sberchat/health`.
- Команды `update-jwt`/`delete-jwt` помогают управлять токеном локально.

## Системные промпты

Встроенные (ID 1–5):

| ID | Name | Описание |
| --- | --- | --- |
| 1 | linux-command | «Ответь только Linux‑командой, без форматирования и объяснений». |
| 2 | linux-command-with-explanation | Сгенерируй команду и кратко объясни, что она делает (формат: COMMAND: explanation). |
| 3 | linux-command-safe | Безопасные команды (без потери данных). Вывод — только команда. |
| 4 | linux-command-verbose | Команда с подробными объяснениями флагов и альтернатив. |
| 5 | linux-command-simple | Простые команды, избегать сложных опций. |

Пользовательские промпты сохраняются в `~/.lcg_prompts.json` и доступны между запусками.

## Сохранение результатов

При выборе действия `s` ответ сохраняется в `LCG_RESULT_FOLDER` (по умолчанию: `./gpt_results`) в файл вида:

```text
gpt_request_<MODEL>_YYYY-MM-DD_HH-MM-SS.md
```

Структура файла:

- `## Prompt:` — ваш запрос (включая системный промпт).
- `## Response:` — полученный ответ.

## Выполнение сгенерированной команды

Действие `e` запустит команду через `bash -c`. Перед запуском потребуется подтверждение `y/yes`. Всегда проверяйте команду вручную, особенно при операциях с файлами и сетью.

## Примеры

1. Базовый запрос с Ollama:

```bash
export LCG_PROVIDER=ollama
export LCG_HOST=http://localhost:11434/
export LCG_MODEL=codegeex4

lcg "хочу извлечь linux-command-gpt.tar.gz"
```

1. Полный ответ от LLM (пример настройки):

```bash
LCG_PROMPT='Provide full response' LCG_MODEL=codellama:13b \
  lcg 'i need bash script to execute command by ssh on array of hosts'
```

1. Proxy‑провайдер:

```bash
export LCG_PROVIDER=proxy
export LCG_HOST=http://localhost:8080
export LCG_MODEL=GigaChat-2
export LCG_JWT_TOKEN=your_jwt_token_here

lcg "I want to extract linux-command-gpt.tar.gz file"

lcg health
lcg config
lcg update-jwt
```

1. Работа с файлами и промптами:

```bash
lcg --file ./context.txt "сгенерируй команду jq для выборки поля name"
lcg --prompt-id 2 "удали все *.tmp в текущем каталоге"
lcg --sys 1 "показать размер каталога в человеко‑читаемом виде"
```

1. Диагностика и модели:

```bash
lcg health
lcg models
```

## История

`lcg history` выводит историю текущего процесса (не сохраняется между запусками, максимум 100 записей):

```bash
lcg history
```

## Типичные проблемы

- Нет ответа/таймаут: увеличьте `--timeout` или `LCG_TIMEOUT`, проверьте `LCG_HOST` и сетевую доступность.
- `health` падает: проверьте, что провайдер запущен и URL верный; для `proxy` — что JWT валиден (`lcg config`).
- Копирование не работает: установите `xclip` или `xsel`.
- Нет допуска к папке результатов: настройте `LCG_RESULT_FOLDER` или права доступа.
- Для `ollama`/`proxy` API‑ключ не нужен; команды `update-key`/`delete-key` просто уведомят об этом.

## Лицензия и исходники

См. README и репозиторий проекта. Предложения и баг‑репорты приветствуются в Issues.
