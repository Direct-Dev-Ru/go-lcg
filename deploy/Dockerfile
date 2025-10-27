# Однофазный build для LCG с предварительно собранным бинарным файлом
FROM alpine:3.22.2

# Устанавливаем зависимости
RUN apk --no-cache add ca-certificates tzdata

# Создаем пользователя
RUN adduser -D -s /bin/sh lcg

# Создаем директории и файлы
RUN mkdir -p /app/data /app/config /home/lcg/.config/lcg/gpt_results /home/lcg/.config/lcg/gpt_sys_prompts && \
    echo '[]' > /home/lcg/.config/lcg/gpt_results/lcg_history.json && \
    chown -R lcg:lcg /app /home/lcg/.config

# Копируем файл версии
COPY VERSION.txt /app/VERSION.txt

# Копируем предварительно собранный бинарный файл
# Ищем папку с бинарным файлом для текущей архитектуры
COPY dist/lcg_linux_${TARGETARCH}*/lcg_* /app/lcg

# Устанавливаем права
RUN chmod +x /app/lcg

# Переключаемся на пользователя lcg
USER lcg

# Устанавливаем рабочую директорию
WORKDIR /app

# Открываем порт
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/login || exit 1

# Запускаем приложение
CMD ["./lcg", "serve", "-H", "0.0.0.0", "-p", "8080"]
