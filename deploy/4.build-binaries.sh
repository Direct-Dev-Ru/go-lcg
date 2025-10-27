#!/bin/bash

# 🚀 LCG Binary Build Script
# Скрипт для сборки бинарных файлов с помощью goreleaser на хосте

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
VERSION=${1:-"dev"}
# CLEAN=${2:-"true"}

# Записываем версию в файл VERSION.txt (в корневой директории проекта)
echo "$VERSION" > VERSION.txt
log "📝 Версия записана в VERSION.txt: $VERSION"

log "🚀 Сборка бинарных файлов LCG с goreleaser..."

# Проверяем наличие goreleaser
if ! command -v goreleaser &> /dev/null; then
    error "goreleaser не найден. Установите goreleaser:"
    echo "  curl -sL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz | tar -xz -C /usr/local/bin goreleaser"
    exit 1
fi

# Проверяем наличие Go
if ! command -v go &> /dev/null; then
    error "Go не найден. Установите Go для сборки."
    exit 1
fi

# Переходим в корневую директорию проекта
cd "$(dirname "$0")/.."

log "📁 Рабочая директория: $(pwd)"
log "📁 Папка dist будет создана в: $(pwd)/dist"

# Очищаем предыдущие сборки если нужно
# if [ "$CLEAN" = "true" ]; then
#     log "🧹 Очистка предыдущих сборок..."
#     rm -rf dist/
#     goreleaser clean
# fi

# Проверяем наличие .goreleaser.yaml
if [ ! -f "deploy/.goreleaser.yaml" ]; then
    error "Файл .goreleaser.yaml не найден в папке deploy/"
    exit 1
fi

# Копируем конфигурацию goreleaser в корень проекта
log "📋 Копирование конфигурации goreleaser..."
cp deploy/.goreleaser.yaml .goreleaser.yaml

# Устанавливаем переменные окружения для версии
export GORELEASER_CURRENT_TAG="$VERSION"

# Собираем бинарные файлы
log "🏗️  Сборка бинарных файлов для всех платформ..."
goreleaser build --snapshot --clean

# Проверяем результат
if [ -d "dist" ]; then
    log "📊 Собранные бинарные файлы:"
    find dist -name "lcg_*" -type f | while read -r binary; do
        echo "  $binary ($(stat -c%s "$binary") bytes, $(file "$binary" | cut -d: -f2))"
    done
    
    success "🎉 Бинарные файлы успешно собраны!"
    
    # Показываем структуру dist/
    log "📁 Структура папки dist/:"
    tree -h dist/ 2>/dev/null || find dist -type f | sort
    
else
    error "Папка dist/ не создана. Проверьте конфигурацию goreleaser."
    exit 1
fi

# Очищаем временный файл конфигурации
rm -f .goreleaser.yaml

success "🎉 Сборка бинарных файлов завершена!"

# Показываем команды для Docker сборки
echo ""
log "📝 Следующие шаги:"
echo "  cd deploy"
echo "  docker buildx build --platform linux/amd64,linux/arm64 --tag your-registry.com/lcg:$VERSION --push ."
echo "  # или используйте скрипт:"
echo "  ./5.build-docker.sh your-registry.com/lcg $VERSION"
