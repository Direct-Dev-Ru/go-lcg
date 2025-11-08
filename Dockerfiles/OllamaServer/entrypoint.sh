#!/bin/bash
set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Обработка сигналов для корректного завершения
cleanup() {
    log "Получен сигнал завершения, останавливаем сервисы..."
    if [ ! -z "$LCG_PID" ]; then
        kill $LCG_PID 2>/dev/null || true
        wait $LCG_PID 2>/dev/null || true
    fi
    if [ ! -z "$OLLAMA_PID" ]; then
        kill $OLLAMA_PID 2>/dev/null || true
        wait $OLLAMA_PID 2>/dev/null || true
    fi
    log "Сервисы остановлены"
    exit 0
}

trap cleanup SIGTERM SIGINT

# Проверка наличия бинарника lcg
if [ ! -f /usr/local/bin/lcg ]; then
    error "Бинарник lcg не найден в /usr/local/bin/lcg"
    exit 1
fi

# Проверка наличия Ollama
if [ ! -f /usr/local/bin/ollama ]; then
    error "Ollama не найден в /usr/local/bin/ollama"
    exit 1
fi

# Создаем необходимые директории
mkdir -p "${LCG_RESULT_FOLDER:-/app/data/results}"
mkdir -p "${LCG_PROMPT_FOLDER:-/app/data/prompts}"
mkdir -p "${LCG_CONFIG_FOLDER:-/app/data/config}"

# Настройка переменных окружения для Ollama
export OLLAMA_HOST="${OLLAMA_HOST:-0.0.0.0}"
export OLLAMA_PORT="${OLLAMA_PORT:-11434}"
export OLLAMA_ORIGINS="*"

# Настройка переменных окружения для LCG
export LCG_PROVIDER="${LCG_PROVIDER:-ollama}"
export LCG_HOST="${LCG_HOST:-http://127.0.0.1:11434/}"
export LCG_MODEL="${LCG_MODEL:-codegeex4}"
export LCG_RESULT_FOLDER="${LCG_RESULT_FOLDER:-/app/data/results}"
export LCG_PROMPT_FOLDER="${LCG_PROMPT_FOLDER:-/app/data/prompts}"
export LCG_CONFIG_FOLDER="${LCG_CONFIG_FOLDER:-/app/data/config}"
export LCG_SERVER_HOST="${LCG_SERVER_HOST:-0.0.0.0}"
export LCG_SERVER_PORT="${LCG_SERVER_PORT:-8080}"
export LCG_SERVER_ALLOW_HTTP="${LCG_SERVER_ALLOW_HTTP:-true}"

log "=========================================="
log "Запуск LCG с Ollama сервером"
log "=========================================="
info "LCG Provider: $LCG_PROVIDER"
info "LCG Host: $LCG_HOST"
info "LCG Model: $LCG_MODEL"
info "LCG Server: http://${LCG_SERVER_HOST}:${LCG_SERVER_PORT}"
info "Ollama Host: $OLLAMA_HOST:$OLLAMA_PORT"
log "=========================================="

# Запускаем Ollama сервер в фоне
log "Запуск Ollama сервера..."
/usr/local/bin/ollama serve &
OLLAMA_PID=$!

# Ждем, пока Ollama запустится
log "Ожидание запуска Ollama сервера..."
sleep 5

# Проверяем, что Ollama запущен
if ! kill -0 $OLLAMA_PID 2>/dev/null; then
    error "Ollama сервер не запустился"
    exit 1
fi

# Проверяем доступность Ollama API
max_attempts=30
attempt=0
while [ $attempt -lt $max_attempts ]; do
    # Проверяем через localhost, так как OLLAMA_HOST может быть 0.0.0.0
    if curl -s -f "http://127.0.0.1:${OLLAMA_PORT}/api/tags" > /dev/null 2>&1; then
        log "Ollama сервер готов!"
        break
    fi
    attempt=$((attempt + 1))
    if [ $attempt -eq $max_attempts ]; then
        error "Ollama сервер не отвечает после $max_attempts попыток"
        exit 1
    fi
    sleep 1
done

# Запускаем LCG сервер в фоне
log "Запуск LCG сервера..."
/usr/local/bin/lcg serve \
    --host "${LCG_SERVER_HOST}" \
    --port "${LCG_SERVER_PORT}" &
LCG_PID=$!

# Ждем, пока LCG запустится
sleep 3

# Проверяем, что LCG запущен
if ! kill -0 $LCG_PID 2>/dev/null; then
    error "LCG сервер не запустился"
    kill $OLLAMA_PID 2>/dev/null || true
    exit 1
fi

log "LCG сервер запущен на http://${LCG_SERVER_HOST}:${LCG_SERVER_PORT}"
log "Ollama сервер доступен на http://${OLLAMA_HOST}:${OLLAMA_PORT}"
log "=========================================="
log "Сервисы запущены и готовы к работе!"
log "=========================================="

# Функция для проверки здоровья процессов
health_check() {
    while true; do
        # Проверяем Ollama
        if ! kill -0 $OLLAMA_PID 2>/dev/null; then
            error "Ollama процесс завершился неожиданно"
            kill $LCG_PID 2>/dev/null || true
            exit 1
        fi
        
        # Проверяем LCG
        if ! kill -0 $LCG_PID 2>/dev/null; then
            error "LCG процесс завершился неожиданно"
            kill $OLLAMA_PID 2>/dev/null || true
            exit 1
        fi
        
        sleep 10
    done
}

# Запускаем проверку здоровья в фоне
health_check &
HEALTH_CHECK_PID=$!

# Ждем завершения процессов
wait $LCG_PID $OLLAMA_PID
kill $HEALTH_CHECK_PID 2>/dev/null || true

