#!/usr/bin/env python3
"""
Скрипт для тестирования GitHub API
Использование: GITHUB_TOKEN=your_token python3 test_api.py
"""

import os
import sys
import requests

# Цвета
class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    NC = '\033[0m'

def log(message):
    print(f"{Colors.GREEN}[INFO]{Colors.NC} {message}")

def error(message):
    print(f"{Colors.RED}[ERROR]{Colors.NC} {message}")

def main():
    REPO = "direct-dev-ru/go-lcg"
    
    token = os.getenv('GITHUB_TOKEN')
    if not token:
        error("GITHUB_TOKEN не установлен")
        sys.exit(1)
    
    session = requests.Session()
    session.headers.update({
        'Authorization': f'token {token}',
        'Accept': 'application/vnd.github.v3+json'
    })
    
    print("=== ТЕСТИРОВАНИЕ GITHUB API ===")
    
    # Тест 1: Проверка доступа к репозиторию
    print("1. Проверка доступа к репозиторию...")
    response = session.get(f"https://api.github.com/repos/{REPO}")
    
    if response.status_code == 200:
        repo_data = response.json()
        print(f"✅ Доступ к репозиторию есть")
        print(f"   Репозиторий: {repo_data['full_name']}")
        print(f"   Описание: {repo_data.get('description', 'Нет описания')}")
    else:
        print(f"❌ Ошибка доступа: {response.status_code}")
        print(f"   Ответ: {response.text}")
    
    # Тест 2: Проверка прав
    print("\n2. Проверка прав...")
    if response.status_code == 200:
        permissions = repo_data.get('permissions', {})
        if permissions.get('admin'):
            print("✅ Есть права администратора")
        elif permissions.get('push'):
            print("✅ Есть права на запись")
        else:
            print("❌ Недостаточно прав для создания релизов")
    
    # Тест 3: Последние релизы
    print("\n3. Последние релизы:")
    releases_response = session.get(f"https://api.github.com/repos/{REPO}/releases")
    
    if releases_response.status_code == 200:
        releases = releases_response.json()
        if releases:
            for release in releases[:5]:
                print(f"   - {release['tag_name']} ({release['name']})")
        else:
            print("   Релизов пока нет")
    else:
        print(f"   Ошибка получения релизов: {releases_response.status_code}")
    
    print("\n=== ТЕСТ ЗАВЕРШЕН ===")

if __name__ == "__main__":
    main()
