# 🔒 Функции безопасности LCG

## 🛡️ Автоматическое принуждение к HTTPS

### Логика безопасности

Приложение автоматически определяет, нужно ли использовать HTTPS:

1. **Небезопасные хосты** (не localhost/127.0.0.1) → **принудительно HTTPS**
2. **Безопасные хосты** (localhost/127.0.0.1) → HTTP (если не указано иное)
3. **Переменная `LCG_SERVER_ALLOW_HTTP=true`** → разрешает HTTP для любых хостов

### Примеры

```bash
# Небезопасно - принудительно HTTPS
LCG_SERVER_HOST=192.168.1.100 lcg serve
# Результат: https://192.168.1.100:8080

# Безопасно - HTTP по умолчанию
LCG_SERVER_HOST=localhost lcg serve  
# Результат: http://localhost:8080

# Принудительно HTTP для любого хоста
LCG_SERVER_HOST=192.168.1.100 LCG_SERVER_ALLOW_HTTP=true lcg serve
# Результат: http://192.168.1.100:8080
```

## 🔐 SSL/TLS сертификаты

### Автоматическая генерация

Приложение автоматически генерирует самоподписанный сертификат если:

1. Не указаны переменные `LCG_SERVER_SSL_CERT_FILE` и `LCG_SERVER_SSL_KEY_FILE`
2. Не найдены файлы в `~/.config/lcg/server/ssl/cert.pem` и `~/.config/lcg/server/ssl/key.pem`

### Расположение сертификатов

``` text
~/.config/lcg/
├── config/
│   └── server/
│       └── ssl/
│           ├── cert.pem    # Сертификат
│           └── key.pem     # Приватный ключ
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `LCG_CONFIG_FOLDER` | Папка конфигурации | `~/.config/lcg/config` |
| `LCG_SERVER_ALLOW_HTTP` | Разрешить HTTP для любых хостов | `false` |
| `LCG_SERVER_SSL_CERT_FILE` | Путь к сертификату | `""` (авто) |
| `LCG_SERVER_SSL_KEY_FILE` | Путь к ключу | `""` (авто) |

## 🚀 Примеры использования

### Безопасный режим (по умолчанию)

```bash
# Локальный сервер - HTTP
lcg serve

# Внешний сервер - принудительно HTTPS
LCG_SERVER_HOST=192.168.1.100 lcg serve
```

### Настройка SSL сертификатов

```bash
# Использовать собственные сертификаты
LCG_SERVER_SSL_CERT_FILE=/path/to/cert.pem \
LCG_SERVER_SSL_KEY_FILE=/path/to/key.pem \
lcg serve

# Разрешить HTTP для внешних хостов
LCG_SERVER_HOST=192.168.1.100 \
LCG_SERVER_ALLOW_HTTP=true \
lcg serve
```

### Docker контейнер

```dockerfile
FROM golang:1.21-alpine AS builder
# ... build steps ...

FROM alpine:latest
COPY --from=builder /app/lcg /usr/local/bin/
ENV LCG_SERVER_HOST=0.0.0.0
ENV LCG_SERVER_ALLOW_HTTP=false
CMD ["lcg", "serve"]
```

### Systemd сервис

```ini
[Unit]
Description=LCG Server
After=network.target

[Service]
Type=simple
User=lcg
WorkingDirectory=/opt/lcg
ExecStart=/opt/lcg/lcg serve
Environment=LCG_SERVER_HOST=0.0.0.0
Environment=LCG_SERVER_ALLOW_HTTP=false
Restart=always

[Install]
WantedBy=multi-user.target
```

## 🔧 Технические детали

### Генерация сертификата

Самоподписанный сертификат генерируется с:

- **Размер ключа**: 2048 бит RSA
- **Срок действия**: 1 год
- **Поддерживаемые хосты**: localhost, 127.0.0.1, указанный хост
- **Использование**: Server Authentication

### Безопасные хосты

Следующие хосты считаются безопасными для HTTP:

- `localhost`
- `127.0.0.1`
- `::1` (IPv6 localhost)

### Проверка безопасности

```go
// Проверка хоста
if !ssl.IsSecureHost(host) {
    // Принудительно HTTPS
    useHTTPS = true
}

// Проверка разрешения HTTP
if config.AppConfig.Server.AllowHTTP {
    useHTTPS = false
}
```

## 🛠️ Отладка

### Проверка конфигурации

```bash
# Показать текущую конфигурацию
lcg config --full | jq '.server'

# Проверить SSL сертификаты
ls -la ~/.config/lcg/config/server/ssl/

# Проверить переменные окружения
env | grep LCG_SERVER
```

### Логи безопасности

```bash
# Запуск с отладкой
LCG_SERVER_HOST=192.168.1.100 lcg serve --debug

# Проверка SSL
openssl x509 -in ~/.config/lcg/config/server/ssl/cert.pem -text -noout
```

## ⚠️ Важные замечания

### Безопасность

1. **Самоподписанные сертификаты** - браузеры будут показывать предупреждение
2. **Продакшен** - используйте настоящие SSL сертификаты от CA
3. **Сетевой доступ** - HTTPS защищает трафик, но не аутентификацию

### Производительность

1. **HTTPS** - небольшая нагрузка на CPU для шифрования
2. **Сертификаты** - генерируются один раз, затем кэшируются
3. **Память** - сертификаты загружаются в память при запуске

## 📚 Связанные файлы

- `config/config.go` - конфигурация безопасности
- `ssl/ssl.go` - генерация и управление сертификатами  
- `serve/serve.go` - HTTP/HTTPS сервер
- `SECURITY_FEATURES.md` - эта документация

---

**Результат**: Приложение теперь автоматически обеспечивает безопасность соединения в зависимости от конфигурации хоста!
