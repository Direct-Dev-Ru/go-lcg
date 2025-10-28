#!/usr/bin/env bash
set -euo pipefail

# release-goreleaser.sh
# Копирует deploy/.goreleaser.yaml в корень, запускает релиз и удаляет файл.
#
# Использование:
#   deploy/release-goreleaser.sh            # обычный релиз на GitHub (нужен GITHUB_TOKEN)
#   deploy/release-goreleaser.sh --snapshot # локальный снепшот без публикации

ROOT_DIR="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
SRC_CFG="$ROOT_DIR/deploy/.goreleaser.yaml"
DST_CFG="$ROOT_DIR/.goreleaser.yaml"

log() { echo -e "\033[36m[release]\033[0m $*"; }
err() { echo -e "\033[31m[error]\033[0m $*" >&2; }

if ! command -v goreleaser >/dev/null 2>&1; then
  err "goreleaser не найден. Установите: https://goreleaser.com/install/"
  exit 1
fi

if [[ ! -f "$SRC_CFG" ]]; then
  err "Не найден файл конфигурации: $SRC_CFG"
  exit 1
fi

MODE="release"
if [[ "${1:-}" == "--snapshot" ]]; then
  MODE="snapshot"
  shift || true
fi

if [[ -f "$DST_CFG" ]]; then
  err "В корне уже существует .goreleaser.yaml. Удалите/переименуйте перед запуском."
  exit 1
fi

cleanup() {
  if [[ -f "$DST_CFG" ]]; then
    rm -f "$DST_CFG" || true
    log "Удалил временный $DST_CFG"
  fi
}
trap cleanup EXIT

log "Копирую конфиг: $SRC_CFG -> $DST_CFG"
cp "$SRC_CFG" "$DST_CFG"

pushd "$ROOT_DIR" >/dev/null

EXTRA_FLAGS=()
PREV_HEAD="$(git rev-parse HEAD 2>/dev/null || echo "")"

git add .
git commit --amend --no-edit || true

## Версию берём из deploy/VERSION.txt или VERSION.txt в корне
VERSION_FILE="$ROOT_DIR/deploy/VERSION.txt"
[[ -f "$VERSION_FILE" ]] || VERSION_FILE="$ROOT_DIR/VERSION.txt"
if [[ -f "$VERSION_FILE" ]]; then
  VERSION_RAW="$(head -n1 "$VERSION_FILE" | tr -d ' \t\r\n')"
  if [[ -n "$VERSION_RAW" ]]; then
    TAG="$VERSION_RAW"
    [[ "$TAG" == v* ]] || TAG="v$TAG"
    export GORELEASER_CURRENT_TAG="$TAG"
    log "Версия релиза: $TAG (из $(realpath --relative-to="$ROOT_DIR" "$VERSION_FILE" 2>/dev/null || echo "$VERSION_FILE"))"
  fi
fi

create_and_push_tag() {
  local tag="$1"
  if git rev-parse "$tag" >/dev/null 2>&1; then
    log "Git tag уже существует: $tag"
  else
    log "Создаю git tag: $tag"
    git tag -a "$tag" -m "Release $tag"
    if [[ "${NO_GIT_PUSH:-false}" != "true" ]]; then
      log "Пушу тег $tag на origin"
      git push origin "$tag"
    else
      log "Пропущен пуш тега (NO_GIT_PUSH=true)"
    fi
  fi
}

move_tag_to_head() {
  local tag="$1"
  if [[ -z "$tag" ]]; then
    return 0
  fi
  if git rev-parse "$tag" >/dev/null 2>&1; then
    log "Переношу тег $tag на текущий коммит (HEAD)"
    git tag -f "$tag" HEAD
    if [[ "${NO_GIT_PUSH:-false}" != "true" ]]; then
      log "Форс‑пуш тега $tag на origin"
      git push -f origin "$tag"
    else
      log "Пропущен пуш тега (NO_GIT_PUSH=true)"
    fi
  else
    log "Тега $tag нет — пропускаю перенос"
  fi
}

fetch_token_from_k8s() {
  export KUBECONFIG=/home/su/.kube/config_hlab
  local ns="${K8S_NAMESPACE:-flux-system}"
  local name="${K8S_SECRET_NAME:-git-secrets}"
  # Предпочитаем jq (как в примере), при отсутствии используем jsonpath + base64 -d
  if command -v jq >/dev/null 2>&1; then
    kubectl get secret "$name" -n "$ns" -o json \
      | jq -r '.data.password | @base64d'
  else
    kubectl get secret "$name" -n "$ns" -o jsonpath='{.data.password}' \
      | base64 -d 2>/dev/null || true
  fi
}

if [[ "$MODE" == "snapshot" ]]; then
  log "Запуск goreleaser (snapshot, без публикации)"
  goreleaser release --snapshot --clean --config "$DST_CFG" "${EXTRA_FLAGS[@]}"
else
  # Если версия определена и тега нет — создадим (goreleaser ориентируется на теги)
  if [[ -n "${GORELEASER_CURRENT_TAG:-}" ]]; then
    create_and_push_tag "$GORELEASER_CURRENT_TAG"
    # Перемещаем тег на текущий HEAD (если существовал ранее, закрепим на последнем коммите)
    move_tag_to_head "$GORELEASER_CURRENT_TAG"
  else
    # Если версия не задана, попробуем взять последний существующий тег и перенести его на HEAD
    LAST_TAG="$(git describe --tags --abbrev=0 2>/dev/null || true)"
    if [[ -n "$LAST_TAG" ]]; then
      move_tag_to_head "$LAST_TAG"
      export GORELEASER_CURRENT_TAG="$LAST_TAG"
      log "Использую последний тег: $LAST_TAG"
    fi
  fi
  if [[ -z "${GITHUB_TOKEN:-}" ]]; then
    log "GITHUB_TOKEN не задан — пробую получить из k8s секрета (${K8S_NAMESPACE:-flux-system}/${K8S_SECRET_NAME:-git-secrets}, ключ: password)"
    if ! command -v kubectl >/dev/null 2>&1; then
      err "kubectl не найден, а GITHUB_TOKEN не задан. Установите kubectl или экспортируйте GITHUB_TOKEN."
      exit 1
    fi
    TOKEN_FROM_K8S="$(fetch_token_from_k8s || true)"
    if [[ -n "$TOKEN_FROM_K8S" && "$TOKEN_FROM_K8S" != "null" ]]; then
      export GITHUB_TOKEN="$TOKEN_FROM_K8S"
      log "GITHUB_TOKEN получен из секрета Kubernetes."
    else
      err "Не удалось получить GITHUB_TOKEN из секрета Kubernetes. Экспортируйте GITHUB_TOKEN и повторите."
      exit 1
    fi
  fi
  log "Запуск goreleaser (публикация на GitHub)"
  goreleaser release --clean --config "$DST_CFG" "${EXTRA_FLAGS[@]}"
fi

popd >/dev/null

# Откатываем временный коммит, если он был
if [[ "${TEMP_COMMIT_DONE:-false}" == "true" && -n "$PREV_HEAD" ]]; then
  if git reset --soft "$PREV_HEAD" >/dev/null 2>&1; then
    log "Откатил временный коммит"
  else
    log "Не удалось откатить временный коммит — проверьте историю вручную"
  fi
fi

log "Готово."


