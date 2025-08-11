#!/bin/bash

# Включаем строгий режим для лучшей отладки
set -euo pipefail

# Конфигурация
readonly REPO="kuznetcovay/go-lcg"
readonly BRANCH="main"
readonly BINARY_NAME="lcg"

# Получаем версию из аргумента или используем значение по умолчанию
VERSION="${1:-v1.1.0}"

# Цвета для вывода
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly NC='\033[0m' # No Color

# Функции для логирования
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Функция для обработки ошибок
handle_error() {
    local exit_code=$?
    log_error "Скрипт завершился с ошибкой (код: $exit_code)"
    exit $exit_code
}

# Функция для восстановления ветки
restore_branch() {
    if [[ -n "${CURRENT_BRANCH:-}" ]]; then
        log_info "Восстанавливаем исходную ветку: ${CURRENT_BRANCH}"
        git checkout "${CURRENT_BRANCH}" || log_warn "Не удалось переключиться на ${CURRENT_BRANCH}"
    fi
}

# Функция для сборки бинарного файла
build_binary() {
    local platform=$1
    local output_dir="bin-linux-${platform}"
    local dockerfile="Dockerfiles/LocalCompile/Dockerfile"
    
    log_info "Собираем для ${platform}..."
    
    if docker build -f "$dockerfile" --target bin-linux --output "$output_dir/" --platform "linux/${platform}" .; then
        cp "$output_dir/$BINARY_NAME" "binaries-for-upload/$BINARY_NAME.${platform}.${VERSION}"
        log_info "Сборка для ${platform} завершена успешно"
    else
        log_error "Сборка для ${platform} не удалась"
        return 1
    fi
}

# Функция для git операций
git_operations() {
    log_info "Выполняем git операции..."
    
    git add -A . || { log_error "git add не удался"; return 1; }
    git commit -m "release $VERSION" || { log_error "git commit не удался"; return 1; }
    git tag -a "$VERSION" -m "release $VERSION" || { log_error "git tag не удался"; return 1; }
    git push -u origin main --tags || { log_error "git push не удался"; return 1; }
    
    log_info "Git операции завершены успешно"
}

# Основная функция
main() {
    log_info "Начинаем сборку версии: $VERSION"
    
    # Записываем версию в файл
    echo "$VERSION" > VERSION.txt
    
    # Настраиваем кэш Go
    export GOCACHE="${HOME}/.cache/go-build"
    
    # Сохраняем текущую ветку
    CURRENT_BRANCH=$(git branch --show-current)
    
    # Настраиваем обработчик ошибок
    trap handle_error ERR
    trap restore_branch EXIT
    
    # Переключаемся на нужную ветку если необходимо
    if [[ "$CURRENT_BRANCH" != "$BRANCH" ]]; then
        log_info "Переключаемся на ветку: $BRANCH"
        git checkout "$BRANCH"
    fi
    
    # Получаем теги
    log_info "Получаем теги из удаленного репозитория..."
    git fetch --tags
    
    # Проверяем существование тега
    if git rev-parse "refs/tags/${VERSION}" >/dev/null 2>&1; then
        log_error "Тег ${VERSION} уже существует. Прерываем выполнение."
        exit 1
    fi
    
    # Создаем директорию для бинарных файлов
    mkdir -p binaries-for-upload
    
    # Собираем бинарные файлы для обеих платформ
    build_binary "amd64"
    build_binary "arm64"
    
    # Собираем и пушим Docker образы
    log_info "Собираем и пушим multi-platform Docker образы..."
    if docker buildx build -f Dockerfiles/ImageBuild/Dockerfile --push --platform linux/amd64,linux/arm64 -t "${REPO}:${VERSION}" .; then
        log_info "Docker образы успешно собраны и запушены"
    else
        log_error "Сборка Docker образов не удалась"
        exit 1
    fi
    
    # Выполняем git операции
    git_operations
    
    log_info "Сборка версии $VERSION завершена успешно!"
}

# Запускаем основную функцию
main "$@"

