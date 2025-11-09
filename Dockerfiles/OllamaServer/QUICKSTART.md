# 🚀 Быстрый старт - LCG с Ollama

## Подготовка

1. Убедитесь, что у вас установлен Docker или Podman
2. Клонируйте репозиторий (если еще не сделали)
3. Соберите бинарники (требуется перед сборкой образа)

```bash
# Из корня проекта
goreleaser build --snapshot --clean

# Или используйте скрипт
./deploy/4.build-binaries.sh v2.0.15
```

4. Перейдите в папку с Dockerfile

```bash
cd Dockerfiles/OllamaServer
```

## Запуск с Docker

### Вариант 1: Docker Compose (рекомендуется)

```bash
# Важно: убедитесь, что бинарники собраны в ../../dist/
docker-compose up -d
```

### Вариант 2: Ручная сборка и запуск

```bash
# Сборка образа (контекст должен быть корень проекта)
cd ../..  # Переходим в корень проекта
docker build -f Dockerfiles/OllamaServer/Dockerfile -t lcg-ollama:latest .

# Запуск контейнера
docker run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  -v ollama-data:/home/ollama/.ollama \
  -v lcg-results:/app/data/results \
  lcg-ollama:latest
```

## Запуск с Podman

### Вариант 1: Podman Compose

```bash
podman-compose -f podman-compose.yml up -d
```

### Вариант 2: Ручная сборка и запуск

```bash
# Сборка образа (контекст должен быть корень проекта)
cd ../..  # Переходим в корень проекта
podman build -f Dockerfiles/OllamaServer/Dockerfile -t lcg-ollama:latest .

# Запуск контейнера
podman run -d \
  --name lcg-ollama \
  -p 8080:8080 \
  -p 11434:11434 \
  -v ollama-data:/home/ollama/.ollama \
  -v lcg-results:/app/data/results \
  lcg-ollama:latest
```

## Проверка запуска

### Проверка логов

```bash
# Docker
docker logs -f lcg-ollama

# Podman
podman logs -f lcg-ollama
```

Дождитесь сообщений:
- `Ollama сервер готов!`
- `LCG сервер запущен на http://0.0.0.0:8080`

### Проверка доступности

```bash
# Проверка Ollama
curl http://localhost:11434/api/tags

# Проверка LCG
curl http://localhost:8080/
```

## Загрузка модели

После запуска контейнера загрузите модель:

```bash
# Docker
docker exec lcg-ollama ollama pull codegeex4

# Podman
podman exec lcg-ollama ollama pull codegeex4
```

Или используйте модель по умолчанию, указанную в переменных окружения.

## Доступ к веб-интерфейсу

Откройте в браузере: http://localhost:8080

## Остановка

```bash
# Docker
docker-compose down

# Podman
podman-compose -f podman-compose.yml down
```

Или для ручного запуска:

```bash
# Docker
docker stop lcg-ollama
docker rm lcg-ollama

# Podman
podman stop lcg-ollama
podman rm lcg-ollama
```

## Решение проблем

### Порт занят

Измените порты в docker-compose.yml или команде run:

```bash
-p 9000:8080  # LCG на порту 9000
-p 11435:11434  # Ollama на порту 11435
```

### Контейнер не запускается

Проверьте логи:

```bash
docker logs lcg-ollama
# или
podman logs lcg-ollama
```

### Модель не загружена

Убедитесь, что модель существует:

```bash
docker exec lcg-ollama ollama list
# или
podman exec lcg-ollama ollama list
```

Если модели нет, загрузите её:

```bash
docker exec lcg-ollama ollama pull codegeex4
# или
podman exec lcg-ollama ollama pull codegeex4
```

## Следующие шаги

- Прочитайте полную документацию в [README.md](README.md)
- Настройте аутентификацию для продакшена
- Настройте reverse proxy для HTTPS
- Загрузите нужные модели Ollama

