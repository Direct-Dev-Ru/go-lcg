# 🚀 Гайд по созданию релизов с помощью GoReleaser

Этот документ описывает процесс создания релизов для проекта `linux-command-gpt` с использованием GoReleaser.

## 📋 Содержание

- [Установка GoReleaser](#установка-goreleaser)
- [Конфигурация](#конфигурация)
- [Процесс создания релиза](#процесс-создания-релиза)
- [Автоматизация](#автоматизация)
- [Устранение проблем](#устранение-проблем)

## 🔧 Установка GoReleaser

### Linux/macOS

```bash
# Скачать и установить последнюю версию
curl -sL https://git.io/goreleaser | bash

# Или через Homebrew (macOS)
brew install goreleaser

# Или через Go
go install github.com/goreleaser/goreleaser@latest
```

### Windows

```powershell
# Через Chocolatey
choco install goreleaser

# Или скачать с GitHub Releases
# https://github.com/goreleaser/goreleaser/releases
```

## ⚙️ Конфигурация

### Файл `.goreleaser.yaml`

В проекте используется следующая конфигурация GoReleaser:

```yaml
version: 2

before:
  hooks:
    - go mod tidy
    - go generate ./...

builds:
  - binary: lcg
    env:
      - CGO_ENABLED=0
    goarch:
      - amd64
      - arm64
      - arm
    goos:
      - linux
      - darwin

archives:
  - formats: [tar.gz]
    name_template: >-
      {{ .ProjectName }}_
      {{- title .Os }}_
      {{- if eq .Arch "amd64" }}x86_64
      {{- else if eq .Arch "386" }}i386
      {{- else }}{{ .Arch }}{{ end }}
      {{- if .Arm }}v{{ .Arm }}{{ end }}
    format_overrides:
      - goos: windows
        formats: [zip]

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"

release:
  footer: >-
    ---
    Released by [GoReleaser](https://github.com/goreleaser/goreleaser).
```

### Ключевые настройки

- **builds**: Сборка для Linux, macOS, Windows (amd64, arm64, arm)
- **archives**: Создание архивов tar.gz для Unix и zip для Windows
- **changelog**: Автоматическое создание changelog из git commits
- **release**: Настройки GitHub релиза

## 🚀 Процесс создания релиза

### 1. Подготовка

```bash
# Убедитесь, что все изменения закоммичены
git status

# Обновите версию в VERSION.txt
echo "v2.0.2" > VERSION.txt

# Создайте тег
git tag v2.0.2
git push origin v2.0.2
```

### 2. Настройка переменных окружения

```bash
# Установите GitHub токен
export GITHUB_TOKEN="your_github_token_here"

# Или создайте файл .env
echo "GITHUB_TOKEN=your_github_token_here" > .env
```

### 3. Создание релиза

#### Полный релиз

```bash
# Создать релиз с загрузкой на GitHub
goreleaser release

# Создать релиз без загрузки (только локально)
goreleaser release --clean
```

#### Тестовый релиз (snapshot)

```bash
# Создать тестовую сборку
goreleaser release --snapshot

# Тестовая сборка без загрузки
goreleaser release --snapshot --clean
```

### 4. Проверка результатов

После выполнения команды GoReleaser создаст:

- **Архивы**: `dist/` - готовые архивы для всех платформ
- **Чексуммы**: `dist/checksums.txt` - контрольные суммы файлов
- **GitHub релиз**: Автоматически созданный релиз на GitHub

## 🤖 Автоматизация

### GitHub Actions

Создайте файл `.github/workflows/release.yml`:

```yaml
name: Release
on:
  push:
    tags:
      - 'v*'
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v3
        with:
          go-version: '1.21'
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v4
        with:
          distribution: goreleaser
          version: latest
          args: release
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### Локальные скрипты

В проекте есть готовые скрипты:

```bash
# Предварительная подготовка
./shell-code/pre-release.sh

# Создание релиза
./shell-code/release.sh
```

## 📁 Структура релиза

После создания релиза в директории `dist/` будут созданы:

```
dist/
├── artifacts.json          # Метаданные артефактов
├── CHANGELOG.md            # Автоматически созданный changelog
├── config.yaml             # Конфигурация GoReleaser
├── digests.txt             # Хеши файлов
├── go-lcg_2.0.1_checksums.txt
├── go-lcg_Darwin_arm64.tar.gz
├── go-lcg_Darwin_x86_64.tar.gz
├── go-lcg_Linux_arm64.tar.gz
├── go-lcg_Linux_i386.tar.gz
├── go-lcg_Linux_x86_64.tar.gz
├── go-lcg_Windows_arm64.zip
├── go-lcg_Windows_i386.zip
├── go-lcg_Windows_x86_64.zip
└── metadata.json           # Метаданные релиза
```

## 🔍 Устранение проблем

### Правильные флаги GoReleaser

**Важно**: В современных версиях GoReleaser флаг `--skip-publish` больше не поддерживается. Используйте:

- `--clean` - очищает директорию `dist/` перед сборкой
- `--snapshot` - создает тестовую сборку без создания тега
- `--debug` - подробный вывод для отладки
- `--skip-validate` - пропускает валидацию конфигурации

### Частые ошибки

#### 1. Ошибка аутентификации GitHub

```
Error: failed to get GitHub token: missing github token
```

**Решение**: Установите `GITHUB_TOKEN` в переменные окружения.

#### 2. Ошибка создания тега

```
Error: git tag v1.0.0 already exists
```

**Решение**: Удалите существующий тег или используйте другую версию.

#### 3. Ошибка сборки

```
Error: failed to build for linux/amd64
```

**Решение**: Проверьте, что код компилируется локально:

```bash
go build -o lcg .
```

### Отладка

```bash
# Подробный вывод
goreleaser release --debug

# Проверка конфигурации
goreleaser check

# Только сборка без релиза
goreleaser build

# Создание релиза без публикации (только локальная сборка)
goreleaser release --clean

# Создание snapshot релиза без публикации
goreleaser release --snapshot --clean
```

## 📝 Лучшие практики

### 1. Версионирование

- Используйте семантическое версионирование (SemVer)
- Обновляйте `VERSION.txt` перед созданием релиза
- Создавайте теги в формате `v1.0.0`

### 2. Changelog

- Пишите понятные commit messages
- Используйте conventional commits для автоматического changelog
- Исключайте технические коммиты из changelog

### 3. Тестирование

- Всегда тестируйте snapshot релизы перед полным релизом
- Проверяйте сборки на разных платформах
- Тестируйте установку из релиза

### 4. Безопасность

- Никогда не коммитьте токены в репозиторий
- Используйте GitHub Secrets для CI/CD
- Регулярно обновляйте токены доступа

## 🎯 Пример полного процесса

```bash
# 1. Подготовка
git checkout main
git pull origin main

# 2. Обновление версии
echo "v2.0.2" > VERSION.txt
git add VERSION.txt
git commit -m "chore: bump version to v2.0.2"

# 3. Создание тега
git tag v2.0.2
git push origin v2.0.2

# 4. Создание релиза
export GITHUB_TOKEN="your_token"
goreleaser release

# 5. Проверка
ls -la dist/
```

## 📚 Дополнительные ресурсы

- [Официальная документация GoReleaser](https://goreleaser.com/)
- [Примеры конфигураций](https://github.com/goreleaser/goreleaser/tree/main/examples)
- [GitHub Actions для GoReleaser](https://github.com/goreleaser/goreleaser-action)

---

**Примечание**: Этот гайд создан специально для проекта `linux-command-gpt`. Для других проектов может потребоваться адаптация конфигурации.
