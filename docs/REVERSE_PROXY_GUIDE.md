# 🔄 Гайд по настройке LCG за Reverse Proxy

## 📋 Переменные окружения для Reverse Proxy

### 🔧 **Основные настройки:**

```bash
# Включить аутентификацию
LCG_SERVER_REQUIRE_AUTH=true

# Настроить домен для cookies (опционально)
LCG_DOMAIN=.yourdomain.com

# Настроить путь для cookies (для префикса пути)
LCG_COOKIE_PATH=/lcg

# Управление Secure флагом cookies
LCG_COOKIE_SECURE=false

# Разрешить HTTP (для работы за reverse proxy)
LCG_SERVER_ALLOW_HTTP=true

# Настроить хост и порт
LCG_SERVER_HOST=0.0.0.0
LCG_SERVER_PORT=8080

# Пароль для входа (по умолчанию: admin#123456)
LCG_SERVER_PASSWORD=your_secure_password
```

## 🚀 **Запуск за Reverse Proxy**

### **1. Nginx конфигурация:**

```nginx
server {
    listen 443 ssl;
    server_name yourdomain.com;
    
    # SSL настройки
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Важно для работы cookies
        proxy_cookie_domain localhost yourdomain.com;
    }
}
```

### **2. Apache конфигурация:**

```apache
<VirtualHost *:443>
    ServerName yourdomain.com
    SSLEngine on
    SSLCertificateFile /path/to/cert.pem
    SSLCertificateKeyFile /path/to/key.pem
    
    ProxyPreserveHost On
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/
    
    # Настройки для cookies
    ProxyPassReverseCookieDomain localhost yourdomain.com
</VirtualHost>
```

### **3. Caddy конфигурация:**

```caddy
yourdomain.com {
    reverse_proxy localhost:8080 {
        header_up Host {host}
        header_up X-Real-IP {remote}
        header_up X-Forwarded-For {remote}
        header_up X-Forwarded-Proto {scheme}
    }
}
```

## 🏃‍♂️ **Команды запуска**

### **Базовый запуск:**

```bash
LCG_SERVER_REQUIRE_AUTH=true LCG_SERVER_ALLOW_HTTP=true ./lcg serve -H 0.0.0.0 -p 8080
```

### **С настройкой домена:**

```bash
LCG_SERVER_REQUIRE_AUTH=true \
LCG_SERVER_ALLOW_HTTP=true \
LCG_DOMAIN=.yourdomain.com \
LCG_COOKIE_SECURE=false \
./lcg serve -H 0.0.0.0 -p 8080
```

### **С кастомным паролем:**

```bash
LCG_SERVER_REQUIRE_AUTH=true \
LCG_SERVER_ALLOW_HTTP=true \
LCG_SERVER_PASSWORD=my_secure_password \
LCG_DOMAIN=.yourdomain.com \
./lcg serve -H 0.0.0.0 -p 8080
```

## 🔒 **Безопасность**

### **Рекомендуемые настройки:**

- ✅ `LCG_SERVER_REQUIRE_AUTH=true` - всегда включайте аутентификацию
- ✅ `LCG_COOKIE_SECURE=false` - для HTTP за reverse proxy
- ✅ `LCG_DOMAIN=.yourdomain.com` - для правильной работы cookies
- ✅ Сильный пароль в `LCG_SERVER_PASSWORD`

### **Настройки Reverse Proxy:**

- ✅ Передавайте заголовки `X-Forwarded-*`
- ✅ Настройте `proxy_cookie_domain` в Nginx
- ✅ Используйте HTTPS на уровне reverse proxy

## 🐳 **Docker Compose пример**

```yaml
version: '3.8'
services:
  lcg:
    image: your-lcg-image
    environment:
      - LCG_SERVER_REQUIRE_AUTH=true
      - LCG_SERVER_ALLOW_HTTP=true
      - LCG_DOMAIN=.yourdomain.com
      - LCG_COOKIE_SECURE=false
      - LCG_SERVER_PASSWORD=secure_password
    ports:
      - "8080:8080"
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    ports:
      - "443:443"
    depends_on:
      - lcg
```

## 🔍 **Диагностика проблем**

### **Проверка cookies:**

```bash
# Проверить установку cookies
curl -I https://yourdomain.com/login

# Проверить домен cookies
curl -v https://yourdomain.com/login 2>&1 | grep -i cookie
```

### **Логи приложения:**

```bash
# Запуск с debug режимом
LCG_SERVER_REQUIRE_AUTH=true \
LCG_SERVER_ALLOW_HTTP=true \
./lcg -d serve -H 0.0.0.0 -p 8080
```

## 📝 **Примечания**

- **SameSite=Lax** - более мягкий режим для reverse proxy
- **Domain cookies** - работают только с указанным доменом
- **Secure=false** - обязательно для HTTP за reverse proxy
- **X-Forwarded-* заголовки** - важны для правильной работы

## 🆘 **Частые проблемы**

1. **Cookies не работают** → Проверьте `LCG_DOMAIN` и настройки reverse proxy
2. **Ошибка 403 CSRF** → Проверьте передачу cookies через reverse proxy
3. **Не работает аутентификация** → Убедитесь что `LCG_SERVER_REQUIRE_AUTH=true`
4. **Проблемы с HTTPS** → Настройте `LCG_COOKIE_SECURE=false` для HTTP за reverse proxy

## 🛣️ **Конфигурация с префиксом пути**

### **Пример: example.com/lcg**

#### **Переменные окружения для префикса:**

```bash
LCG_SERVER_REQUIRE_AUTH=true \
LCG_SERVER_ALLOW_HTTP=true \
LCG_DOMAIN=.example.com \
LCG_COOKIE_PATH=/lcg \
LCG_COOKIE_SECURE=false \
./lcg serve -H 0.0.0.0 -p 8080
```

#### **Nginx с префиксом:**

```nginx
server {
    listen 443 ssl;
    server_name example.com;
    
    location /lcg/ {
        proxy_pass http://localhost:8080/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Важно для работы cookies с префиксом
        proxy_cookie_domain localhost example.com;
        proxy_cookie_path / /lcg/;
    }
}
```

#### **Apache с префиксом:**

```apache
<VirtualHost *:443>
    ServerName example.com
    SSLEngine on
    
    ProxyPreserveHost On
    ProxyPass /lcg/ http://localhost:8080/
    ProxyPassReverse /lcg/ http://localhost:8080/
    
    # Настройки для cookies с префиксом
    ProxyPassReverseCookieDomain localhost example.com
    ProxyPassReverseCookiePath / /lcg/
</VirtualHost>
```

#### **Caddy с префиксом:**

```caddy
example.com {
    reverse_proxy /lcg/* localhost:8080 {
        header_up Host {host}
        header_up X-Real-IP {remote}
        header_up X-Forwarded-For {remote}
        header_up X-Forwarded-Proto {scheme}
    }
}
```
