#!/bin/bash

# 🐳 LCG Docker Build Script
# Скрипт для сборки Docker образа с предварительно собранными бинарными файлами

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
VERSION=${2:-""}
PLATFORMS=${3:-"linux/amd64,linux/arm64"}

if [ -z "$VERSION" ]; then
    error "Версия не указана! Использование: $0 <repository> <version>"
    echo "Пример: $0 your-registry.com/lcg v1.0.0 <platforms>"
    exit 1
fi

log "🐳 Сборка Docker образа LCG..."

# Проверяем наличие docker
if ! command -v docker &> /dev/null; then
    error "Docker не найден. Установите Docker для сборки образов."
    exit 1
fi

# Проверяем наличие docker buildx
if ! docker buildx version &> /dev/null; then
    error "Docker Buildx не найден. Установите Docker Buildx для мультиплатформенной сборки."
    exit 1
fi

# Проверяем наличие бинарных файлов в текущей директории (если запускаем из корня)
if [ ! -d "dist" ]; then
    error "Папка dist/ не найдена. Сначала соберите бинарные файлы:"
    echo "  ./deploy/4.build-binaries.sh $VERSION"
    exit 1
fi

# Проверяем наличие бинарных файлов для всех платформ
MISSING_BINARIES=()

# Ищем бинарные файлы с версией в имени
AMD64_BINARY=$(find dist -name "*linux_amd64*" -type d | head -1)
echo "AMD64_BINARY: $AMD64_BINARY"
ARM64_BINARY=$(find dist -name "*linux_arm64*" -type d | head -1)
echo "ARM64_BINARY: $ARM64_BINARY"

# Проверяем наличие бинарных файлов в найденных папках и соответствие версии
if [ -n "$AMD64_BINARY" ]; then
    AMD64_FILE=$(find "$AMD64_BINARY" -name "lcg_*" -type f | head -1)
    if [ -z "$AMD64_FILE" ]; then
        AMD64_BINARY=""
    else
        # Извлекаем версию из имени файла
        FILE_VERSION=$(basename "$AMD64_FILE" | sed 's/lcg_//' | sed 's/-SNAPSHOT.*//')
        # Нормализуем версии для сравнения (убираем префикс 'v' если есть)
        NORMALIZED_FILE_VERSION=$(echo "$FILE_VERSION" | sed 's/^v//')
        NORMALIZED_VERSION=$(echo "$VERSION" | sed 's/^v//')
        if [ "$NORMALIZED_FILE_VERSION" != "$NORMALIZED_VERSION" ]; then
            error "Версия в имени бинарного файла ($FILE_VERSION) не совпадает с переданной версией ($VERSION)"
            echo "Файл: $AMD64_FILE"
            echo "Ожидаемая версия: $VERSION"
            echo "Версия в файле: $FILE_VERSION"
            exit 1
        fi
    fi
fi

if [ -n "$ARM64_BINARY" ]; then
    ARM64_FILE=$(find "$ARM64_BINARY" -name "lcg_*" -type f | head -1)
    if [ -z "$ARM64_FILE" ]; then
        ARM64_BINARY=""
    else
        # Извлекаем версию из имени файла
        FILE_VERSION=$(basename "$ARM64_FILE" | sed 's/lcg_//' | sed 's/-SNAPSHOT.*//')
        # Нормализуем версии для сравнения (убираем префикс 'v' если есть)
        NORMALIZED_FILE_VERSION=$(echo "$FILE_VERSION" | sed 's/^v//')
        NORMALIZED_VERSION=$(echo "$VERSION" | sed 's/^v//')
        if [ "$NORMALIZED_FILE_VERSION" != "$NORMALIZED_VERSION" ]; then
            error "Версия в имени бинарного файла ($FILE_VERSION) не совпадает с переданной версией ($VERSION)"
            echo "Файл: $ARM64_FILE"
            echo "Ожидаемая версия: $VERSION"
            echo "Версия в файле: $FILE_VERSION"
            exit 1
        fi
    fi
fi

if [ -z "$AMD64_BINARY" ]; then
    MISSING_BINARIES+=("linux/amd64")
fi
if [ -z "$ARM64_BINARY" ]; then
    MISSING_BINARIES+=("linux/arm64")
fi

if [ ${#MISSING_BINARIES[@]} -gt 0 ]; then
    error "Отсутствуют бинарные файлы для платформ: ${MISSING_BINARIES[*]}"
    echo "Сначала соберите бинарные файлы:"
    echo "  ./4.build-binaries.sh $VERSION"
    exit 1
fi

# Показываем найденные файлы и их версии
log "📊 Найденные бинарные файлы:"
if [ -n "$AMD64_FILE" ]; then
    echo "  AMD64: $AMD64_FILE"
fi
if [ -n "$ARM64_FILE" ]; then
    echo "  ARM64: $ARM64_FILE"
fi

# Создаем builder если не существует
log "🔧 Настройка Docker Buildx..."
docker buildx create --name lcg-builder --use 2>/dev/null || docker buildx use lcg-builder

# Копируем бинарные файлы и файл версии в папку deploy
log "📋 Копирование бинарных файлов и файла версии..."
cp -r dist ./deploy/dist
cp VERSION.txt ./deploy/VERSION.txt 2>/dev/null || echo "dev" > ./deploy/VERSION.txt

# Сборка для всех платформ
log "🏗️  Сборка образа для платформ: $PLATFORMS"
log "📦 Репозиторий: $REPOSITORY"
log "🏷️  Версия: $VERSION"

# Сборка и push
docker buildx build \
    --platform "$PLATFORMS" \
    --tag "$REPOSITORY:$VERSION" \
    --tag "$REPOSITORY:latest" \
    --push \
    --file deploy/Dockerfile \
    deploy/

# Очищаем скопированные файлы
rm -rf ./deploy/dist

success "🎉 Образ успешно собран и отправлен в репозиторий!"

# Показываем информацию о собранном образе
log "📊 Информация о собранном образе:"
echo "  Репозиторий: $REPOSITORY"
echo "  Версия: $VERSION"
echo "  Платформы: $PLATFORMS"
echo "  Теги: $REPOSITORY:$VERSION, $REPOSITORY:latest"

# Проверяем образы в репозитории
log "🔍 Проверка образов в репозитории..."
docker buildx imagetools inspect "$REPOSITORY:$VERSION" || warning "Не удалось проверить образ в репозитории"

success "🎉 Сборка завершена успешно!"

# Показываем команды для использования
echo ""
log "📝 Полезные команды:"
echo "  docker pull $REPOSITORY:$VERSION"
echo "  docker run -p 8080:8080 $REPOSITORY:$VERSION"
echo "  docker buildx imagetools inspect $REPOSITORY:$VERSION"
