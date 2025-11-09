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
    log "Сервисы остановлены"
    exit 0
}

trap cleanup SIGTERM SIGINT

# Проверка наличия бинарника lcg
if [ ! -f /usr/local/bin/lcg ]; then
    error "Бинарник lcg не найден в /usr/local/bin/lcg"
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
export LCG_HOST="${LCG_HOST:-http://0.0.0.0:11434/}"
export LCG_MODEL="${LCG_MODEL:-qwen2.5-coder:1.5b}"
export LCG_RESULT_FOLDER="${LCG_RESULT_FOLDER:-/app/data/results}"
export LCG_PROMPT_FOLDER="${LCG_PROMPT_FOLDER:-/app/data/prompts}"
export LCG_CONFIG_FOLDER="${LCG_CONFIG_FOLDER:-/app/data/config}"
export LCG_SERVER_HOST="${LCG_SERVER_HOST:-0.0.0.0}"
export LCG_SERVER_PORT="${LCG_SERVER_PORT:-8080}"
export LCG_SERVER_ALLOW_HTTP="${LCG_SERVER_ALLOW_HTTP:-false}"

log "=========================================="
log "Запуск LCG с Ollama сервером"
log "=========================================="
info "LCG Provider: $LCG_PROVIDER"
info "LCG Host: $LCG_HOST"
info "LCG Model: $LCG_MODEL"
info "LCG Server: http://${LCG_SERVER_HOST}:${LCG_SERVER_PORT}"
info "Ollama Host: $OLLAMA_HOST:$OLLAMA_PORT"
log "=========================================="


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
    exit 1
fi

log "LCG сервер запущен на http://${LCG_SERVER_HOST}:${LCG_SERVER_PORT}"

# Запускаем переданные аргументы
exec "$@"
