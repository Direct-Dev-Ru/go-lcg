# Контракт API для провайдеров (proxy и ollama)

Этот документ описывает минимально необходимый API, который должен предоставлять сервер-провайдер (режимы: "proxy" и "ollama"), чтобы CLI-приложение работало корректно.

## Общие требования

- **Базовый URL** берётся из `config.AppConfig.Host`. Трейлинг-слэш на стороне клиента обрезается.
- **Таймаут** HTTP-запросов задаётся в секундах через конфигурацию (см. `config.AppConfig.Timeout`).
- **Кодирование**: все тела запросов и ответов — `application/json; charset=utf-8`.
- **Стриминг**: на данный момент клиент всегда запрашивает `stream=false`; стриминг не используется.

---

## Режим proxy

### Аутентификация

- Все защищённые эндпоинты требуют заголовок: `Authorization: Bearer <JWT>`.
- Токен берётся из `config.AppConfig.JwtToken`, либо из файла `~/.proxy_jwt_token`.

### 1) POST `/api/v1/protected/sberchat/chat`

- **Назначение**: получить единственный текстовый ответ LLM.
- **Заголовки**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <JWT>` (обязательно)
- **Тело запроса** (минимально необходимые поля):

```json
{
  "messages": [
    { "role": "system", "content": "<system_prompt>" },
    { "role": "user",   "content": "<ask>" }
  ],
  "model": "<model_name>",
  "temperature": 0.5,
  "top_p": 0.5,
  "stream": false,
  "random_words": ["linux", "command", "gpt"],
  "fallback_string": "I'm sorry, I can't help with that. Please try again."
}
```

- **Ответ 200 OK**:

```json
{
  "response": "<string>",
  "usage": {
    "prompt_tokens": 0,
    "completion_tokens": 0,
    "total_tokens": 0
  },
  "error": "",
  "model": "<model_name>",
  "timeout_seconds": 0
}
```

- **Ошибки**: любой статус != 200 воспринимается как ошибка. Желательно вернуть JSON вида:

```json
{ "error": "<message>" }
```

- **Пример cURL**:

```bash
curl -sS -X POST \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  "$HOST/api/v1/protected/sberchat/chat" \
  -d '{
    "messages": [
      {"role":"system","content":"system prompt"},
      {"role":"user","content":"user ask"}
    ],
    "model":"GigaChat-2-Max",
    "temperature":0.5,
    "top_p":0.5,
    "stream":false,
    "random_words":["linux","command","gpt"],
    "fallback_string":"I'm sorry, I can't help with that. Please try again."
  }'
```

### 2) GET `/api/v1/protected/sberchat/health`

- **Назначение**: health-check API и получение части метаданных по умолчанию.
- **Заголовки**:
  - `Authorization: Bearer <JWT>` (если сервер требует авторизацию на health)
- **Ответ 200 OK**:

```json
{
  "status": "ok",
  "message": "<string>",
  "default_model": "<string>",
  "default_timeout_seconds": 120
}
```

- **Ошибки**: любой статус != 200 считается падением health.

### Модели

- В текущей реализации клиент не запрашивает список моделей у proxy и использует фиксированный набор.
- Опционально можно реализовать эндпоинт для списка моделей (например, `GET /api/v1/protected/sberchat/models`) и расширить клиента позже.

---

## Режим ollama

### 1) POST `/api/chat`

- **Назначение**: синхронная генерация одного ответа (без стрима).
- **Заголовки**:
  - `Content-Type: application/json`
- **Тело запроса**:

```json
{
  "model": "<model_name>",
  "stream": false,
  "messages": [
    { "role": "system", "content": "<system_prompt>" },
    { "role": "user",   "content": "<ask>" }
  ],
  "options": {"temperature": 0.2}
}
```

- **Ответ 200 OK** (минимальный, который поддерживает клиент):

```json
{
  "model": "<model_name>",
  "message": { "role": "assistant", "content": "<string>" },
  "done": true
}
```

- Прочие поля ответа (`total_duration`, `eval_count` и т.д.) допускаются, но клиент использует только `message.content`.

- **Ошибки**: любой статус != 200 считается ошибкой. Желательно возвращать читаемое тело.

### 2) GET `/api/tags`

- **Назначение**: используется как health-check и для получения списка моделей.
- **Ответ 200 OK**:

```json
{
  "models": [
    { "name": "llama3:8b", "modified_at": "2024-01-01T00:00:00Z", "size": 123456789 },
    { "name": "qwen2.5:7b", "modified_at": "2024-01-02T00:00:00Z", "size": 987654321 }
  ]
}
```

- Любой другой статус трактуется как ошибка health.

---

## Семантика сообщений

- `messages` — массив объектов `{ "role": "system"|"user"|"assistant", "content": "<string>" }`.
- Клиент всегда отправляет как минимум 2 сообщения: системное и пользовательское.
- Ответ должен содержать один финальный текст в виде `response` (proxy) или `message.content` (ollama).

## Поведение при таймаутах

- Сервер должен завершать запрос в пределах `config.AppConfig.Timeout` секунд (значение передаётся клиентом в настройки HTTP-клиента; отдельным полем в запросе оно не отправляется, исключение — `proxy` может возвращать `timeout_seconds` в ответе как справочную информацию).

## Коды ответов и ошибки

- 200 — успешный ответ с телом согласно контракту.
- !=200 — ошибка; тело желательно в JSON с полем `error`.

## Изменения контракта

- Добавление новых полей в ответах, не используемых клиентом, допустимо при сохранении существующих.
- Переименование или удаление полей `response` (proxy) и `message.content` (ollama) нарушит совместимость.

---

Дополнительно: для HTTP API веб‑сервера (эндпоинт `POST /execute`, только `curl`) см. `API_GUIDE.md` с примерами и подробной схемой запроса/ответа.
