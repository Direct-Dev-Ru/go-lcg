# Используем готовый образ Ollama
FROM localhost/ollama_packed:latest

# Устанавливаем bash если его нет (базовый образ ollama может быть на разных дистрибутивах)
RUN if ! command -v bash >/dev/null 2>&1; then \
        if command -v apk >/dev/null 2>&1; then \
            apk add --no-cache bash; \
        elif command -v apt-get >/dev/null 2>&1; then \
            apt-get update && apt-get install -y --no-install-recommends bash && rm -rf /var/lib/apt/lists/*; \
        fi; \
    fi

# Определяем архитектуру для копирования правильного бинарника
ARG TARGETARCH
ARG TARGETOS=linux

# Копируем папку dist с бинарниками
# Структура: dist/lcg_linux_amd64_v1/lcg_* или dist/lcg_linux_arm64_v8.0/lcg_*
COPY dist/ /tmp/dist/

# Выбираем правильный бинарник в зависимости от архитектуры
# Если TARGETARCH не установлен, определяем архитектуру через uname
RUN ARCH="${TARGETARCH:-$(uname -m)}" && \
    case "${ARCH}" in \
    amd64|x86_64) \
        BIN_FILE=$(find /tmp/dist/lcg_linux_amd64_v* -name "lcg_*" -type f 2>/dev/null | head -1) && \
        if [ -n "$BIN_FILE" ]; then \
            cp "$BIN_FILE" /usr/local/bin/lcg && \
            echo "Установлен бинарник для amd64: $BIN_FILE"; \
        else \
            echo "Бинарник для amd64 не найден в /tmp/dist/" && exit 1; \
        fi ;; \
    arm64|aarch64|arm) \
        BIN_FILE=$(find /tmp/dist/lcg_linux_arm64_v* -name "lcg_*" -type f 2>/dev/null | head -1) && \
        if [ -n "$BIN_FILE" ]; then \
            cp "$BIN_FILE" /usr/local/bin/lcg && \
            echo "Установлен бинарник для arm64: $BIN_FILE"; \
        else \
            echo "Бинарник для arm64 не найден в /tmp/dist/" && exit 1; \
        fi ;; \
    *) \
        echo "Unsupported architecture: ${ARCH}" && \
        echo "Доступные бинарники:" && \
        find /tmp/dist -name "lcg_*" -type f 2>/dev/null && \
        exit 1 ;; \
    esac && \
    chmod +x /usr/local/bin/lcg && \
    rm -rf /tmp/dist && \
    (lcg --version || echo "Бинарник lcg установлен")

# Копируем entrypoint скрипт
COPY --chmod=755 Dockerfiles/OllamaServer/entrypoint.sh /entrypoint.sh

# Создаем директории для данных LCG
# В базовом образе ollama уже есть пользователь ollama
RUN mkdir -p /app/data/results /app/data/prompts /app/data/config

# Устанавливаем права доступа (пользователь ollama должен существовать в базовом образе)
RUN chown -R ollama:ollama /app/data 2>/dev/null || \
    (chown -R 1000:1000 /app/data 2>/dev/null || true)

# Настройки по умолчанию
ENV LCG_PROVIDER=ollama
ENV LCG_HOST=http://127.0.0.1:11434/
ENV LCG_MODEL=qwen2.5-coder:1.5b
ENV LCG_RESULT_FOLDER=/app/data/results
ENV LCG_PROMPT_FOLDER=/app/data/prompts
ENV LCG_CONFIG_FOLDER=/app/data/config
ENV LCG_SERVER_HOST=0.0.0.0
ENV LCG_SERVER_PORT=8080
# ENV LCG_SERVER_ALLOW_HTTP=true
# ENV OLLAMA_HOST=127.0.0.1
# ENV OLLAMA_PORT=11434

# Expose порты
EXPOSE 8080

WORKDIR /home/ollama

# Запускаем entrypoint
ENTRYPOINT ["/entrypoint.sh"]
CMD []

