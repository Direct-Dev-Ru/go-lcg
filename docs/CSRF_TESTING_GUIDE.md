# 🛡️ Руководство по тестированию CSRF защиты

## 📋 Обзор

Это руководство поможет вам протестировать CSRF защиту в LCG приложении и понять, как работают CSRF атаки.

## 🚀 Быстрый старт

### 1. Запуск сервера с CSRF защитой

```bash
# Запуск с аутентификацией и CSRF защитой
LCG_SERVER_REQUIRE_AUTH=true ./lcg serve -p 8080
```

### 2. Автоматическое тестирование

```bash
# Запуск автоматических тестов
./test_csrf.sh
```

### 3. Ручное тестирование

```bash
# Откройте в браузере
open csrf_test.html
```

## 🧪 Типы тестов

### ✅ **Тест 1: Защищенные запросы**

- **Цель**: Проверить, что POST запросы без CSRF токена блокируются
- **Ожидаемый результат**: 403 Forbidden
- **Endpoints**: `/api/execute`, `/api/save-result`, `/api/add-to-history`

### ✅ **Тест 2: Разрешенные запросы**

- **Цель**: Проверить, что GET запросы работают
- **Ожидаемый результат**: 200 OK
- **Endpoints**: `/login`, `/`, `/history`

### ✅ **Тест 3: CSRF токены**

- **Цель**: Проверить наличие CSRF токенов в формах
- **Ожидаемый результат**: Токены присутствуют в HTML

### ✅ **Тест 4: Поддельные токены**

- **Цель**: Проверить защиту от поддельных токенов
- **Ожидаемый результат**: 403 Forbidden

## 🎯 Сценарии атак

### **Сценарий 1: Выполнение команд**

```html
<!-- Злонамеренная форма -->
<form action="http://localhost:8080/api/execute" method="POST">
    <input type="hidden" name="prompt" value="rm -rf /">
    <input type="hidden" name="system_id" value="1">
    <button type="submit">Нажми меня!</button>
</form>
```

### **Сценарий 2: Сохранение данных**

```html
<!-- Злонамеренная форма -->
<form action="http://localhost:8080/api/save-result" method="POST">
    <input type="hidden" name="result" value="Вредоносные данные">
    <input type="hidden" name="command" value="malicious_command">
    <button type="submit">Сохранить</button>
</form>
```

### **Сценарий 3: JavaScript атака**

```javascript
// Злонамеренный JavaScript
fetch('http://localhost:8080/api/execute', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({prompt: 'whoami', system_id: '1'})
});
```

## 🔍 Анализ результатов

### **✅ Защита работает, если:**

- Все POST запросы возвращают 403 Forbidden
- В ответах есть "CSRF token required"
- GET запросы работают нормально
- CSRF токены присутствуют в формах

### **❌ Уязвимость есть, если:**

- POST запросы выполняются успешно (200 OK)
- Команды выполняются на сервере
- Данные сохраняются без CSRF токенов
- Нет проверки Origin/Referer заголовков

## 🛠️ Инструменты тестирования

### **1. Автоматический скрипт**

```bash
./test_csrf.sh
```

- Тестирует все основные endpoints
- Проверяет CSRF токены
- Выводит подробный отчет

### **2. HTML тестовая страница**

```bash
open csrf_test.html
```

- Интерактивное тестирование
- Визуальная проверка результатов
- Тестирование в браузере

### **3. Демонстрационная атака**

```bash
open csrf_demo.html
```

- Показывает, как работают CSRF атаки
- Демонстрирует уязвимости
- Образовательные цели

## 🔧 Настройка тестов

### **Переменные окружения для тестирования:**

```bash
# Включить аутентификацию
LCG_SERVER_REQUIRE_AUTH=true

# Настроить CSRF защиту
LCG_COOKIE_SECURE=false
LCG_DOMAIN=.localhost
LCG_COOKIE_PATH=/

# Запуск сервера
./lcg serve -H 0.0.0.0 -p 8080
```

### **Настройка reverse proxy для тестирования:**

```bash
# Для тестирования за reverse proxy
LCG_SERVER_REQUIRE_AUTH=true \
LCG_SERVER_ALLOW_HTTP=true \
LCG_DOMAIN=.yourdomain.com \
LCG_COOKIE_PATH=/lcg \
LCG_COOKIE_SECURE=false \
./lcg serve -H 0.0.0.0 -p 8080
```

## 📊 Интерпретация результатов

### **Успешные тесты:**

``` text
✅ CSRF защита /api/execute: PASS - Запрос заблокирован (403 Forbidden)
✅ CSRF защита /api/save-result: PASS - Запрос заблокирован (403 Forbidden)
✅ CSRF защита /api/add-to-history: PASS - Запрос заблокирован (403 Forbidden)
✅ GET запросы: PASS - GET запросы работают (HTTP 200)
✅ CSRF токен на странице входа: PASS - Токен найден
✅ CSRF защита от поддельного токена: PASS - Поддельный токен заблокирован (403 Forbidden)
```

### **Проблемные тесты:**

``` text
❌ CSRF защита /api/execute: FAIL - Запрос прошел (HTTP 200)
❌ CSRF защита /api/save-result: FAIL - Запрос прошел (HTTP 200)
❌ CSRF токен на странице входа: FAIL - Токен не найден
```

## 🚨 Частые проблемы

### **1. Cookies не работают**

- Проверьте настройки `LCG_DOMAIN`
- Убедитесь, что `LCG_COOKIE_PATH` правильный
- Проверьте настройки reverse proxy

### **2. CSRF токены не генерируются**

- Убедитесь, что `LCG_SERVER_REQUIRE_AUTH=true`
- Проверьте инициализацию CSRF менеджера
- Проверьте логи сервера

### **3. Запросы проходят без токенов**

- Проверьте middleware в `serve/middleware.go`
- Убедитесь, что CSRF middleware применяется
- Проверьте исключения в middleware

## 📝 Рекомендации

### **Для разработчиков:**

1. Всегда тестируйте CSRF защиту
2. Используйте автоматические тесты
3. Проверяйте все POST endpoints
4. Валидируйте CSRF токены

### **Для администраторов:**

1. Регулярно запускайте тесты
2. Мониторьте логи на подозрительную активность
3. Настройте правильные заголовки в reverse proxy
4. Используйте HTTPS в продакшене

## 🎓 Образовательные материалы

- **OWASP CSRF Prevention Cheat Sheet**: <https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html>
- **CSRF атаки**: <https://owasp.org/www-community/attacks/csrf>
- **SameSite cookies**: <https://web.dev/samesite-cookies-explained/>

---

**⚠️ ВНИМАНИЕ**: Эти тесты предназначены только для проверки безопасности вашего собственного приложения. Не используйте их для атак на чужие системы!
