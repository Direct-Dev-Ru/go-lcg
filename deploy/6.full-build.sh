#!/bin/bash

# 🚀 LCG Full Build Script
# Полный скрипт сборки: бинарные файлы + Docker образ

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

REPOSITORY=${1:-"kuznetcovay/lcg"}
VERSION=${2:-""}
PLATFORMS=${3:-"linux/amd64,linux/arm64"}

if [ -z "$VERSION" ]; then
    error "Версия не указана! Использование: $0 <repository> <version> <platforms>"
    echo "Пример: $0 kuznetcovay/lcg v1.0.0 linux/amd64,linux/arm64"
    exit 1
fi

# Записываем версию в файл VERSION.txt (в корневой директории проекта)
cd "$(dirname "$0")/.."
echo "$VERSION" > VERSION.txt
log "📝 Версия записана в VERSION.txt: $VERSION"

log "🚀 Полная сборка LCG (бинарные файлы + Docker образ)..."

# Этап 1: Сборка бинарных файлов
log "📦 Этап 1: Сборка бинарных файлов с goreleaser..."
if ! ./deploy/4.build-binaries.sh "$VERSION"; then
    error "Ошибка при сборке бинарных файлов"
    exit 1
fi

success "✅ Бинарные файлы собраны успешно"

# Этап 2: Сборка Docker образа
log "🐳 Этап 2: Сборка Docker образа..."
if ! ./deploy/5.build-docker.sh "$REPOSITORY" "$VERSION" "$PLATFORMS"; then
    error "Ошибка при сборке Docker образа"
    exit 1
fi

success "✅ Docker образы собраны успешно"

# Этап 3: Генерация deployment.yaml
log "📝 Этап 3: Генерация deployment.yaml..."
# Generate deployment.yaml with env substitution
export REPOSITORY=$REPOSITORY
export VERSION=$VERSION
export PLATFORMS=$PLATFORMS
export KUBECONFIG="${HOME}/.kube/config_hlab" && kubectx default

if ! envsubst < deploy/1.configmap.tmpl.yaml > kustomize/configmap.yaml; then
    error "Ошибка при генерации deploy/1.configmap.yaml"
    exit 1
fi

success "✅ kustomize/configmap.yaml сгенерирован успешно"

if ! envsubst < deploy/deployment.tmpl.yaml > kustomize/deployment.yaml; then
    error "Ошибка при генерации kustomize/deployment.yaml"
    exit 1
fi
success "✅ kustomize/deployment.yaml сгенерирован успешно"

if ! envsubst < deploy/ingress-route.tmpl.yaml > kustomize/ingress-route.yaml; then
    error "Ошибка при генерации kustomize/ingress-route.yaml"
    exit 1
fi
success "✅ kustomize/ingress-route.yaml сгенерирован успешно"

if ! envsubst < deploy/service.tmpl.yaml > kustomize/service.yaml; then
    error "Ошибка при генерации kustomize/service.yaml"
    exit 1
fi
success "✅ kustomize/service.yaml сгенерирован успешно"

if ! envsubst < deploy/kustomization.tmpl.yaml > kustomize/kustomization.yaml; then
    error "Ошибка при генерации kustomize/kustomization.yaml"
    exit 1
fi
success "✅ kustomize/kustomization.yaml сгенерирован успешно"

# отключить reconciliation flux
if kubectl get kustomization lcg -n flux-system > /dev/null 2>&1; then
    kubectl patch kustomization lcg -n flux-system --type=merge -p '{"spec":{"suspend":true}}'
else
    echo "ℹ️  Kustomization 'lcg' does not exist in 'flux-system' namespace. Skipping suspend."
fi
sleep 5


# зафиксировать изменения в текущей ветке, если она не main
current_branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$current_branch" != "main" ]; then
    log "🔧 Исправления в текущей ветке: $current_branch"
    # считать, что изменения уже сделаны
    git add .
    git commit -m "Исправления в ветке $current_branch"
fi

# переключиться на ветку main и слить с текущей веткой, если не находимся на main
if [ "$current_branch" != "main" ]; then
    git checkout main
    git merge --no-ff -m "Merged branch '$current_branch' into main while building $VERSION" "$current_branch"
elif [ "$current_branch" = "main" ]; then
    log "🔄 Вы находитесь на ветке main. Слияние с release..."
    git add .
    git commit -m "Исправления в ветке $current_branch"
fi

# переключиться на ветку release и слить с веткой main
git checkout release
git merge --no-ff -m "Merged main into release while building $VERSION" main

# если тег $VERSION существует, удалить его и принудительно запушить
tag_exists=$(git tag -l "$VERSION")
if [ "$tag_exists" ]; then
    log "🗑️ Удаление существующего тега $VERSION"
    git tag -d "$VERSION"
    git push origin ":refs/tags/$VERSION"
fi

# Create tag $VERSION and push to remote release branch and all tags
git tag "$VERSION"
git push origin release
git push origin --tags

# Push main branch
git checkout main
git push origin main


# Включить reconciliation flux
if kubectl get kustomization lcg -n flux-system > /dev/null 2>&1; then
    kubectl patch kustomization lcg -n flux-system --type=merge -p '{"spec":{"suspend":false}}'
else
    echo "ℹ️  Kustomization 'lcg' does not exist in 'flux-system' namespace. Skipping suspend."
fi
echo "🔄 Flux will automatically deploy $VERSION version in ~4-6 minutes..."

# Итоговая информация
echo ""
log "🎉 Полная сборка завершена успешно!"
echo ""
log "📊 Результат:"
echo "  Репозиторий: $REPOSITORY"
echo "  Версия: $VERSION"
echo "  Платформы: $PLATFORMS"
echo "  Теги: $REPOSITORY:$VERSION, $REPOSITORY:latest"
echo ""
echo ""
log "🔍 Информация о git коммитах:"
git_log=$(git log release -1 --pretty=format:"%H - %s")
echo "$git_log"
echo ""


log "📝 Команды для использования:"
echo "  docker pull $REPOSITORY:$VERSION"
echo "  docker run -p 8080:8080 $REPOSITORY:$VERSION"
echo "  docker buildx imagetools inspect $REPOSITORY:$VERSION"
echo ""
log "🔍 Проверка образа:"
echo "  docker run --rm $REPOSITORY:$VERSION /app/lcg --version"
echo ""
log "📝 Команды для использования:"
echo "  kubectl apply -k kustomize"
echo "  kubectl get pods"
echo "  kubectl get services"
echo "  kubectl get ingress"
echo "  kubectl get hpa"
echo "  kubectl get servicemonitor"
echo "  kubectl get pods"
echo "  kubectl get services"
echo "  kubectl get ingress"
echo "  kubectl get hpa"
echo "  kubectl get servicemonitor"
