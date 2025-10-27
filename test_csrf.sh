#!/bin/bash

# 🛡️ CSRF Protection Test Script
# Тестирует CSRF защиту LCG приложения

echo "🛡️ Тестирование CSRF защиты LCG"
echo "=================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода результатов
print_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅ $test_name: PASS${NC} - $message"
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}❌ $test_name: FAIL${NC} - $message"
    else
        echo -e "${YELLOW}⚠️  $test_name: $status${NC} - $message"
    fi
}

# Проверяем, запущен ли сервер
echo -e "${BLUE}🔍 Проверяем доступность сервера...${NC}"
if ! curl -s http://localhost:8080/login > /dev/null 2>&1; then
    echo -e "${RED}❌ Сервер не доступен на localhost:8080${NC}"
    echo "Запустите сервер командой: LCG_SERVER_REQUIRE_AUTH=true ./lcg serve -p 8080"
    exit 1
fi

echo -e "${GREEN}✅ Сервер доступен${NC}"

# Тест 1: Попытка выполнения команды без CSRF токена
echo -e "\n${BLUE}🧪 Тест 1: Выполнение команды без CSRF токена${NC}"
response=$(curl -s -w "%{http_code}" -X POST http://localhost:8080/api/execute \
    -H "Content-Type: application/json" \
    -d '{"prompt":"whoami","system_id":"1"}' \
    -o /dev/null)

if [ "$response" = "403" ]; then
    print_result "CSRF защита /api/execute" "PASS" "Запрос заблокирован (403 Forbidden)"
else
    print_result "CSRF защита /api/execute" "FAIL" "Запрос прошел (HTTP $response)"
fi

# Тест 2: Попытка сохранения результата без CSRF токена
echo -e "\n${BLUE}🧪 Тест 2: Сохранение результата без CSRF токена${NC}"
response=$(curl -s -w "%{http_code}" -X POST http://localhost:8080/api/save-result \
    -H "Content-Type: application/json" \
    -d '{"result":"test result","command":"test command"}' \
    -o /dev/null)

if [ "$response" = "403" ]; then
    print_result "CSRF защита /api/save-result" "PASS" "Запрос заблокирован (403 Forbidden)"
else
    print_result "CSRF защита /api/save-result" "FAIL" "Запрос прошел (HTTP $response)"
fi

# Тест 3: Попытка добавления в историю без CSRF токена
echo -e "\n${BLUE}🧪 Тест 3: Добавление в историю без CSRF токена${NC}"
response=$(curl -s -w "%{http_code}" -X POST http://localhost:8080/api/add-to-history \
    -H "Content-Type: application/json" \
    -d '{"prompt":"test prompt","result":"test result"}' \
    -o /dev/null)

if [ "$response" = "403" ]; then
    print_result "CSRF защита /api/add-to-history" "PASS" "Запрос заблокирован (403 Forbidden)"
else
    print_result "CSRF защита /api/add-to-history" "FAIL" "Запрос прошел (HTTP $response)"
fi

# Тест 4: Проверка GET запросов (должны работать)
echo -e "\n${BLUE}🧪 Тест 4: GET запросы (должны работать)${NC}"
response=$(curl -s -w "%{http_code}" http://localhost:8080/login -o /dev/null)

if [ "$response" = "200" ]; then
    print_result "GET запросы" "PASS" "GET запросы работают (HTTP $response)"
else
    print_result "GET запросы" "FAIL" "GET запросы не работают (HTTP $response)"
fi

# Тест 5: Проверка наличия CSRF токена на странице входа
echo -e "\n${BLUE}🧪 Тест 5: Наличие CSRF токена на странице входа${NC}"
csrf_token=$(curl -s http://localhost:8080/login | grep -o 'name="csrf_token"[^>]*value="[^"]*"' | sed 's/.*value="\([^"]*\)".*/\1/')

if [ -n "$csrf_token" ]; then
    print_result "CSRF токен на странице входа" "PASS" "Токен найден: ${csrf_token:0:20}..."
else
    print_result "CSRF токен на странице входа" "FAIL" "Токен не найден"
fi

# Тест 6: Попытка атаки с поддельным CSRF токеном
echo -e "\n${BLUE}🧪 Тест 6: Атака с поддельным CSRF токеном${NC}"
response=$(curl -s -w "%{http_code}" -X POST http://localhost:8080/api/execute \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: fake_token" \
    -d '{"prompt":"whoami","system_id":"1"}' \
    -o /dev/null)

if [ "$response" = "403" ]; then
    print_result "CSRF защита от поддельного токена" "PASS" "Поддельный токен заблокирован (403 Forbidden)"
else
    print_result "CSRF защита от поддельного токена" "FAIL" "Поддельный токен принят (HTTP $response)"
fi

# Итоговый отчет
echo -e "\n${BLUE}📊 Итоговый отчет:${NC}"
echo "=================================="

# Подсчитываем результаты
total_tests=6
passed_tests=0

# Проверяем каждый тест
if curl -s -w "%{http_code}" -X POST http://localhost:8080/api/execute -H "Content-Type: application/json" -d '{"prompt":"test"}' -o /dev/null | grep -q "403"; then
    ((passed_tests++))
fi

if curl -s -w "%{http_code}" -X POST http://localhost:8080/api/save-result -H "Content-Type: application/json" -d '{"result":"test"}' -o /dev/null | grep -q "403"; then
    ((passed_tests++))
fi

if curl -s -w "%{http_code}" -X POST http://localhost:8080/api/add-to-history -H "Content-Type: application/json" -d '{"prompt":"test"}' -o /dev/null | grep -q "403"; then
    ((passed_tests++))
fi

if curl -s -w "%{http_code}" http://localhost:8080/login -o /dev/null | grep -q "200"; then
    ((passed_tests++))
fi

if curl -s http://localhost:8080/login | grep -q 'name="csrf_token"'; then
    ((passed_tests++))
fi

if curl -s -w "%{http_code}" -X POST http://localhost:8080/api/execute -H "Content-Type: application/json" -H "X-CSRF-Token: fake" -d '{"prompt":"test"}' -o /dev/null | grep -q "403"; then
    ((passed_tests++))
fi

echo -e "Пройдено тестов: ${GREEN}$passed_tests${NC} из ${BLUE}$total_tests${NC}"

if [ $passed_tests -eq $total_tests ]; then
    echo -e "${GREEN}🎉 Все тесты пройдены! CSRF защита работает корректно.${NC}"
    exit 0
elif [ $passed_tests -ge 4 ]; then
    echo -e "${YELLOW}⚠️  Большинство тестов пройдено, но есть проблемы с CSRF защитой.${NC}"
    exit 1
else
    echo -e "${RED}❌ Критические проблемы с CSRF защитой!${NC}"
    exit 2
fi
