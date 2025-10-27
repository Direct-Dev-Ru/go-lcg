#!/bin/bash

# 🚀 LCG Build Script (Root)
# Скрипт для сборки из корневой директории проекта

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

# Параметры
REPOSITORY=${1:-"your-registry.com/lcg"}
VERSION=${2:-"latest"}
PLATFORMS=${3:-"linux/amd64,linux/arm64"}

log "🚀 Сборка LCG из корневой директории..."

# Проверяем, что мы в корневой директории
if [ ! -f "go.mod" ]; then
    error "Этот скрипт должен запускаться из корневой директории проекта (где находится go.mod)"
    exit 1
fi

# Записываем версию в файл VERSION.txt
echo "$VERSION" > VERSION.txt
log "📝 Версия записана в VERSION.txt: $VERSION"

# Запускаем полную сборку
log "🚀 Запуск полной сборки..."
./deploy/full-build.sh "$REPOSITORY" "$VERSION" "$PLATFORMS"

if [ $? -eq 0 ]; then
    success "🎉 Сборка завершена успешно!"
else
    error "Ошибка при сборке"
    exit 1
fi
