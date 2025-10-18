#!/bin/bash

# Простой скрипт для создания релиза на GitHub
# Использование: GITHUB_TOKEN=your_token ./release.sh

set -e

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log() { echo -e "${GREEN}[INFO]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
debug() { echo -e "${BLUE}[DEBUG]${NC} $1"; }

# Конфигурация
REPO="direct-dev-ru/go-lcg"
VERSION_FILE="VERSION.txt"
BINARIES_DIR="binaries-for-upload"

# Проверки
if [[ -z "$GITHUB_TOKEN" ]]; then
    error "GITHUB_TOKEN не установлен"
    exit 1
fi

if [[ ! -f "$VERSION_FILE" ]]; then
    error "Файл $VERSION_FILE не найден"
    exit 1
fi

if [[ ! -d "$BINARIES_DIR" ]]; then
    error "Директория $BINARIES_DIR не найдена"
    exit 1
fi

# Получение версии
VERSION=$(cat "$VERSION_FILE" | tr -d ' \t\n\r')
TAG="lcg.$VERSION"

log "Версия: $VERSION"
log "Тег: $TAG"

# Проверяем, существует ли уже релиз
log "Проверяем существующий релиз..."
EXISTING_RELEASE=$(curl -s -H "Authorization: token $GITHUB_TOKEN" \
    "https://api.github.com/repos/$REPO/releases/tags/$TAG")

if echo "$EXISTING_RELEASE" | grep -q '"id":'; then
    log "Реліз $TAG уже существует, получаем upload_url..."
    UPLOAD_URL=$(echo "$EXISTING_RELEASE" | grep '"upload_url"' | cut -d'"' -f4 | sed 's/{?name,label}//')
else
    log "Создаем новый релиз $TAG..."
    
    # Создаем релиз
    RELEASE_DATA="{\"tag_name\":\"$TAG\",\"name\":\"$TAG\",\"body\":\"Release $TAG\"}"
    
    RELEASE_RESPONSE=$(curl -s -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Content-Type: application/json" \
        "https://api.github.com/repos/$REPO/releases" \
        -d "$RELEASE_DATA")
    
    if echo "$RELEASE_RESPONSE" | grep -q '"message"'; then
        error "Ошибка создания релиза:"
        echo "$RELEASE_RESPONSE" | grep '"message"' | cut -d'"' -f4
        exit 1
    fi
    
    UPLOAD_URL=$(echo "$RELEASE_RESPONSE" | grep '"upload_url"' | cut -d'"' -f4 | sed 's/{?name,label}//')
    log "Реліз создан успешно"
fi

if [[ -z "$UPLOAD_URL" ]]; then
    error "Не удалось получить upload_url"
    exit 1
fi

log "Upload URL: $UPLOAD_URL"

# Проверяем файлы в директории
log "Проверяем файлы в директории $BINARIES_DIR:"
ls -la "$BINARIES_DIR"

# Загружаем файлы
log "Загружаем файлы..."
UPLOADED=0
FAILED=0

# Простой цикл по всем файлам в директории
for file in "$BINARIES_DIR"/*; do
    if [[ -f "$file" ]]; then
        filename=$(basename "$file")
        log "Обрабатываем файл: $file"
        debug "Имя файла: $filename"
        
        log "Загружаем: $filename"
        
        response=$(curl -s -X POST \
            -H "Authorization: token $GITHUB_TOKEN" \
            -H "Content-Type: application/octet-stream" \
            "$UPLOAD_URL?name=$filename" \
            --data-binary @"$file")
        
        debug "Ответ API: $response"
        
        if echo "$response" | grep -q '"message"'; then
            error "Ошибка загрузки $filename:"
            echo "$response" | grep '"message"' | cut -d'"' -f4
            ((FAILED++))
        else
            log "✓ $filename загружен"
            ((UPLOADED++))
        fi
    else
        warn "Пропускаем не-файл: $file"
    fi
done

# Результат
log "=== РЕЗУЛЬТАТ ==="
log "Успешно загружено: $UPLOADED"
if [[ $FAILED -gt 0 ]]; then
    warn "Ошибок: $FAILED"
else
    log "Все файлы загружены успешно!"
fi

log "Реліз доступен: https://github.com/$REPO/releases/tag/$TAG"
