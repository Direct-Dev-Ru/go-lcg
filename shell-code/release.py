#!/usr/bin/env python3
"""
Скрипт для создания релиза на GitHub
Использование: GITHUB_TOKEN=your_token python3 release.py
"""

import os
import sys
import json
import requests
from pathlib import Path

# Цвета для вывода
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

def log(message):
    print(f"{Colors.GREEN}[INFO]{Colors.NC} {message}")

def error(message):
    print(f"{Colors.RED}[ERROR]{Colors.NC} {message}", file=sys.stderr)

def warn(message):
    print(f"{Colors.YELLOW}[WARN]{Colors.NC} {message}")

def debug(message):
    print(f"{Colors.BLUE}[DEBUG]{Colors.NC} {message}")

# Конфигурация
REPO = "direct-dev-ru/go-lcg"
VERSION_FILE = "VERSION.txt"
BINARIES_DIR = "binaries-for-upload"

def check_environment():
    """Проверка переменных окружения"""
    token = os.getenv('GITHUB_TOKEN')
    if not token:
        error("GITHUB_TOKEN не установлен")
        sys.exit(1)
    log(f"GITHUB_TOKEN установлен (длина: {len(token)} символов)")
    return token

def get_version():
    """Получение версии из файла"""
    version_file = Path(VERSION_FILE)
    if not version_file.exists():
        error(f"Файл {VERSION_FILE} не найден")
        sys.exit(1)
    
    version = version_file.read_text().strip()
    tag = f"lcg.{version}"
    log(f"Версия: {version}")
    log(f"Тег: {tag}")
    return tag

def check_files():
    """Проверка файлов для загрузки"""
    binaries_path = Path(BINARIES_DIR)
    if not binaries_path.exists():
        error(f"Директория {BINARIES_DIR} не найдена")
        sys.exit(1)
    
    files = list(binaries_path.glob("*"))
    files = [f for f in files if f.is_file()]
    
    if not files:
        error(f"В директории {BINARIES_DIR} нет файлов")
        sys.exit(1)
    
    log(f"Найдено файлов: {len(files)}")
    for file in files:
        log(f"  - {file.name} ({file.stat().st_size} байт)")
    
    return files

def create_github_session(token):
    """Создание сессии для GitHub API"""
    session = requests.Session()
    session.headers.update({
        'Authorization': f'token {token}',
        'Accept': 'application/vnd.github.v3+json',
        'User-Agent': 'release-script'
    })
    return session

def check_existing_release(session, tag):
    """Проверка существующего релиза"""
    log("Проверяем существующий релиз...")
    url = f"https://api.github.com/repos/{REPO}/releases/tags/{tag}"
    
    response = session.get(url)
    if response.status_code == 200:
        release_data = response.json()
        log(f"Реліз {tag} уже существует")
        return release_data
    elif response.status_code == 404:
        log(f"Реліз {tag} не найден, создаем новый")
        return None
    else:
        error(f"Ошибка проверки релиза: {response.status_code}")
        debug(f"Ответ: {response.text}")
        sys.exit(1)

def create_release(session, tag):
    """Создание нового релиза"""
    log(f"Создаем новый релиз {tag}...")
    
    data = {
        "tag_name": tag,
        "name": tag,
        "body": f"Release {tag}",
        "draft": False,
        "prerelease": False
    }
    
    url = f"https://api.github.com/repos/{REPO}/releases"
    response = session.post(url, json=data)
    
    if response.status_code == 201:
        release_data = response.json()
        log("Реліз создан успешно")
        return release_data
    else:
        error(f"Ошибка создания релиза: {response.status_code}")
        debug(f"Ответ: {response.text}")
        sys.exit(1)

def upload_file(session, upload_url, file_path):
    """Загрузка файла в релиз"""
    filename = file_path.name
    log(f"Загружаем: {filename}")
    
    # Убираем {?name,label} из URL
    upload_url = upload_url.replace("{?name,label}", "")
    
    with open(file_path, 'rb') as f:
        headers = {'Content-Type': 'application/octet-stream'}
        params = {'name': filename}
        
        response = session.post(
            upload_url,
            data=f,
            headers=headers,
            params=params
        )
    
    if response.status_code == 201:
        log(f"✓ {filename} загружен")
        return True
    else:
        error(f"Ошибка загрузки {filename}: {response.status_code}")
        debug(f"Ответ: {response.text}")
        return False

def main():
    """Основная функция"""
    log("=== НАЧАЛО РАБОТЫ СКРИПТА ===")
    
    # Проверки
    token = check_environment()
    tag = get_version()
    files = check_files()
    
    # Создание сессии
    session = create_github_session(token)
    
    # Проверка/создание релиза
    release = check_existing_release(session, tag)
    if not release:
        release = create_release(session, tag)
    
    # Получение URL для загрузки
    upload_url = release['upload_url']
    log(f"Upload URL: {upload_url}")
    
    # Загрузка файлов
    log("=== ЗАГРУЗКА ФАЙЛОВ ===")
    uploaded = 0
    failed = 0
    
    for file_path in files:
        if upload_file(session, upload_url, file_path):
            uploaded += 1
        else:
            failed += 1
    
    # Результат
    log("=== РЕЗУЛЬТАТ ===")
    log(f"Успешно загружено: {uploaded}")
    if failed > 0:
        warn(f"Ошибок: {failed}")
    else:
        log("Все файлы загружены успешно!")
    
    log(f"Реліз доступен: https://github.com/{REPO}/releases/tag/{tag}")
    log("=== СКРИПТ ЗАВЕРШЕН ===")

if __name__ == "__main__":
    main()
