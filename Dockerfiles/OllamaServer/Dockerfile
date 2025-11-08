# Multi-stage build для LCG с Ollama
FROM golang:1.24.6-alpine3.22 AS builder

WORKDIR /build

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s -buildid=" -trimpath -o /build/lcg .

# Финальный образ с Ollama
FROM alpine:3.22

# Устанавливаем необходимые пакеты
RUN apk add --no-cache \
    curl \
    ca-certificates \
    bash \
    && rm -rf /var/cache/apk/*

# Устанавливаем Ollama 0.9.5 (поддержка разных архитектур)
ARG TARGETARCH
RUN case ${TARGETARCH} in \
    amd64) OLLAMA_ARCH=amd64 ;; \
    arm64) OLLAMA_ARCH=arm64 ;; \
    arm) OLLAMA_ARCH=arm64 ;; \
    *) echo "Unsupported architecture: ${TARGETARCH}" && exit 1 ;; \
    esac && \
    curl -L https://github.com/ollama/ollama/releases/download/v0.9.5/ollama-linux-${OLLAMA_ARCH} -o /usr/local/bin/ollama \
    && chmod +x /usr/local/bin/ollama

# Создаем пользователя для запуска сервисов
RUN addgroup -g 1000 ollama && \
    adduser -D -u 1000 -G ollama ollama && \
    mkdir -p /home/ollama/.ollama && \
    chown -R ollama:ollama /home/ollama

# Копируем бинарник lcg
COPY --from=builder /build/lcg /usr/local/bin/lcg
RUN chmod +x /usr/local/bin/lcg

# Копируем entrypoint скрипт
COPY --chmod=755 Dockerfiles/OllamaServer/entrypoint.sh /entrypoint.sh

# Создаем директории для данных
RUN mkdir -p /app/data/results /app/data/prompts /app/data/config \
    && chown -R ollama:ollama /app/data

# Настройки по умолчанию
ENV LCG_PROVIDER=ollama
ENV LCG_HOST=http://127.0.0.1:11434/
ENV LCG_MODEL=codegeex4
ENV LCG_RESULT_FOLDER=/app/data/results
ENV LCG_PROMPT_FOLDER=/app/data/prompts
ENV LCG_CONFIG_FOLDER=/app/data/config
ENV LCG_SERVER_HOST=0.0.0.0
ENV LCG_SERVER_PORT=8080
ENV LCG_SERVER_ALLOW_HTTP=true
ENV OLLAMA_HOST=0.0.0.0
ENV OLLAMA_PORT=11434

# Expose порты
EXPOSE 8080 11434

# Переключаемся на пользователя ollama
USER ollama

WORKDIR /home/ollama

# Запускаем entrypoint
ENTRYPOINT ["/entrypoint.sh"]
CMD []

