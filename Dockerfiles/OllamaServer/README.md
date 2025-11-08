# 🐳 LCG с Ollama Server - Docker/Podman контейнер

Этот образ содержит Linux Command GPT (LCG) и Ollama сервер, работающие вместе в одном контейнере.

Поддерживается запуск через Docker и Podman.

## 📋 Описание

Контейнер автоматически запускает:
1. **Ollama сервер** (v0.9.5) на порту 11434
2. **LCG веб-сервер** на порту 8080

Ollama используется как провайдер LLM для генерации Linux команд.

## 🚀 Быстрый старт

### Сборка образа

#### Docker
```bash
# Из корня проекта
docker build -f Dockerfiles/OllamaServer/Dockerfile -t lcg-ollama:latest .
```

#### Podman
```bash
# Из корня проекта
podman build -f Dockerfiles/OllamaServer/Dockerfile -t lcg-ollama:latest .
```

### Запуск контейнера

#### Docker
```bash
docker run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  lcg-ollama:latest
```

#### Podman
```bash
podman run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  lcg-ollama:latest
```

### Использование docker-compose / podman-compose

#### Docker Compose
```bash
cd Dockerfiles/OllamaServer
docker-compose up -d
```

#### Podman Compose
```bash
cd Dockerfiles/OllamaServer
podman-compose -f podman-compose.yml up -d
```

Или используйте встроенную поддержку Podman:
```bash
cd Dockerfiles/OllamaServer
podman play kube podman-compose.yml
```

## 🌐 Доступ к сервисам

После запуска контейнера доступны:

- **LCG веб-интерфейс**: http://localhost:8080
- **Ollama API**: http://localhost:11434

## ⚙️ Переменные окружения

### Настройки LCG

| Переменная | Значение по умолчанию | Описание |
|------------|----------------------|----------|
| `LCG_PROVIDER` | `ollama` | Тип провайдера |
| `LCG_HOST` | `http://127.0.0.1:11434/` | URL Ollama API |
| `LCG_MODEL` | `codegeex4` | Модель для использования |
| `LCG_SERVER_HOST` | `0.0.0.0` | Хост LCG сервера |
| `LCG_SERVER_PORT` | `8080` | Порт LCG сервера |
| `LCG_SERVER_ALLOW_HTTP` | `true` | Разрешить HTTP |
| `LCG_RESULT_FOLDER` | `/app/data/results` | Папка для результатов |
| `LCG_PROMPT_FOLDER` | `/app/data/prompts` | Папка для промптов |
| `LCG_CONFIG_FOLDER` | `/app/data/config` | Папка для конфигурации |

### Настройки Ollama

| Переменная | Значение по умолчанию | Описание |
|------------|----------------------|----------|
| `OLLAMA_HOST` | `127.0.0.1` | Хост Ollama сервера |
| `OLLAMA_PORT` | `11434` | Порт Ollama сервера |

### Безопасность

| Переменная | Значение по умолчанию | Описание |
|------------|----------------------|----------|
| `LCG_SERVER_REQUIRE_AUTH` | `false` | Требовать аутентификацию |
| `LCG_SERVER_PASSWORD` | `admin#123456` | Пароль для аутентификации |

## 📦 Volumes

Рекомендуется монтировать volumes для персистентного хранения данных:

```bash
docker run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  -v ollama-data:/home/ollama/.ollama \
  -v lcg-results:/app/data/results \
  -v lcg-prompts:/app/data/prompts \
  -v lcg-config:/app/data/config \
  lcg-ollama:latest
```

### Volumes описание

- `ollama-data`: Модели и данные Ollama
- `lcg-results`: Результаты генерации команд
- `lcg-prompts`: Системные промпты
- `lcg-config`: Конфигурация LCG

## 🔧 Примеры использования

### Запуск с кастомной моделью

```bash
docker run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  -e LCG_MODEL=llama3:8b \
  lcg-ollama:latest
```

### Запуск с аутентификацией

```bash
docker run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  -e LCG_SERVER_REQUIRE_AUTH=true \
  -e LCG_SERVER_PASSWORD=my_secure_password \
  lcg-ollama:latest
```

### Запуск с кастомным портом

```bash
docker run -d \
  --name lcg-ollama \
  -p 9000:9000 \
  -p 11434:11434 \
  -e LCG_SERVER_PORT=9000 \
  lcg-ollama:latest
```

## 📥 Загрузка моделей Ollama

После запуска контейнера можно загрузить модели:

```bash
# Подключиться к контейнеру
docker exec -it lcg-ollama sh

# Загрузить модель
ollama pull codegeex4
ollama pull llama3:8b
ollama pull qwen2.5:7b
```

Или извне контейнера:

```bash
# Убедитесь, что Ollama доступен извне (OLLAMA_HOST=0.0.0.0)
docker exec lcg-ollama ollama pull codegeex4
```

## 🔍 Проверка работоспособности

### Проверка Ollama

```bash
# Проверка health
curl http://localhost:11434/api/tags

# Список моделей
curl http://localhost:11434/api/tags | jq '.models'
```

### Проверка LCG

```bash
# Проверка веб-интерфейса
curl http://localhost:8080/

# Проверка через API
curl -X POST http://localhost:8080/api/execute \
  -H "Content-Type: application/json" \
  -d '{"prompt": "создать директорию test"}'
```

## 🐧 Podman специфичные инструкции

### Запуск с Podman

Podman работает аналогично Docker, но есть несколько отличий:

#### Создание сетей (если нужно)

```bash
podman network create lcg-network
```

#### Запуск с сетью

```bash
podman run -d \
  --name lcg-ollama \
  --network lcg-network \
  -p 8080:8080 \
  -p 11434:11434 \
  lcg-ollama:latest
```

#### Запуск в rootless режиме

Podman по умолчанию работает в rootless режиме, что повышает безопасность:

```bash
# Не требует sudo
podman run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  lcg-ollama:latest
```

#### Использование systemd для автозапуска

Создайте systemd unit файл:

```bash
# Генерируем unit файл
podman generate systemd --name lcg-ollama --files

# Копируем в systemd
sudo cp container-lcg-ollama.service /etc/systemd/system/

# Включаем автозапуск
sudo systemctl enable container-lcg-ollama.service
sudo systemctl start container-lcg-ollama.service
```

#### Проверка статуса

```bash
# Статус контейнера
podman ps

# Логи
podman logs lcg-ollama

# Логи в реальном времени
podman logs -f lcg-ollama
```

## 🐛 Отладка

### Просмотр логов

#### Docker
```bash
# Логи контейнера
docker logs lcg-ollama

# Логи в реальном времени
docker logs -f lcg-ollama
```

#### Podman
```bash
# Логи контейнера
podman logs lcg-ollama

# Логи в реальном времени
podman logs -f lcg-ollama
```

### Подключение к контейнеру

#### Docker
```bash
docker exec -it lcg-ollama sh
```

#### Podman
```bash
podman exec -it lcg-ollama sh
```

### Проверка процессов

#### Docker
```bash
docker exec lcg-ollama ps aux
```

#### Podman
```bash
podman exec lcg-ollama ps aux
```

## 🔒 Безопасность

### Рекомендации для продакшена

1. **Используйте аутентификацию**:
   ```bash
   -e LCG_SERVER_REQUIRE_AUTH=true
   -e LCG_SERVER_PASSWORD=strong_password
   ```

2. **Ограничьте доступ к портам**:
   - Используйте firewall правила
   - Не экспортируйте порты на публичный интерфейс

3. **Используйте HTTPS**:
   - Настройте reverse proxy (nginx, traefik)
   - Используйте SSL сертификаты

4. **Ограничьте ресурсы**:
   ```bash
   docker run -d \
     --name lcg-ollama \
     --memory="4g" \
     --cpus="2" \
     lcg-ollama:latest
   ```

## 📊 Мониторинг

### Healthcheck

Контейнер включает healthcheck, который проверяет доступность LCG сервера:

```bash
docker inspect lcg-ollama | jq '.[0].State.Health'
```

### Метрики

LCG предоставляет Prometheus метрики на `/metrics` endpoint (если включено).

## 🚀 Production Deployment

### С docker-compose

```bash
cd Dockerfiles/OllamaServer
docker-compose up -d
```

### С Kubernetes

Используйте манифесты из папки `deploy/` или `kustomize/`.

## 📝 Примечания

- Ollama версия: 0.9.5
- LCG версия: см. VERSION.txt
- Минимальные требования: 2GB RAM, 2 CPU cores
- Рекомендуется: 4GB+ RAM для больших моделей

## 🔗 Полезные ссылки

- [Ollama документация](https://github.com/ollama/ollama)
- [LCG документация](../../docs/README.md)
- [LCG API Guide](../../docs/API_GUIDE.md)

## ❓ Поддержка

При возникновении проблем:
1. Проверьте логи: `docker logs lcg-ollama`
2. Проверьте переменные окружения
3. Убедитесь, что порты не заняты
4. Проверьте, что модели загружены в Ollama

