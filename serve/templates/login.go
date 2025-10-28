package templates

// LoginPageTemplate шаблон страницы авторизации
const LoginPageTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(-45deg, #667eea, #764ba2, #f093fb, #f5576c, #4facfe, #00f2fe);
            background-size: 400% 400%;
            animation: gradientShift 15s ease infinite;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            overflow: hidden;
            position: relative;
        }
        
        @keyframes gradientShift {
            0% { background-position: 0% 50%; }
            50% { background-position: 100% 50%; }
            100% { background-position: 0% 50%; }
        }
        
        /* Плавающие элементы */
        .floating-elements {
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            overflow: hidden;
            z-index: 1;
        }
        
        .floating-element {
            position: absolute;
            opacity: 0.1;
            animation: float 20s infinite linear;
        }
        
        .floating-element:nth-child(1) { left: 10%; animation-delay: 0s; animation-duration: 25s; }
        .floating-element:nth-child(2) { left: 20%; animation-delay: 2s; animation-duration: 30s; }
        .floating-element:nth-child(3) { left: 30%; animation-delay: 4s; animation-duration: 20s; }
        .floating-element:nth-child(4) { left: 40%; animation-delay: 6s; animation-duration: 35s; }
        .floating-element:nth-child(5) { left: 50%; animation-delay: 8s; animation-duration: 28s; }
        .floating-element:nth-child(6) { left: 60%; animation-delay: 10s; animation-duration: 22s; }
        .floating-element:nth-child(7) { left: 70%; animation-delay: 12s; animation-duration: 32s; }
        .floating-element:nth-child(8) { left: 80%; animation-delay: 14s; animation-duration: 26s; }
        .floating-element:nth-child(9) { left: 90%; animation-delay: 16s; animation-duration: 24s; }
        
        @keyframes float {
            0% { transform: translateY(100vh) rotate(0deg); opacity: 0; }
            10% { opacity: 0.1; }
            90% { opacity: 0.1; }
            100% { transform: translateY(-100px) rotate(360deg); opacity: 0; }
        }
        
        .lock-icon {
            font-size: 2rem;
            color: rgba(255, 255, 255, 0.3);
        }
        
        .key-icon {
            font-size: 1.5rem;
            color: rgba(255, 255, 255, 0.3);
        }
        
        .shield-icon {
            font-size: 1.8rem;
            color: rgba(255, 255, 255, 0.3);
        }
        
        .star-icon {
            font-size: 1.2rem;
            color: rgba(255, 255, 255, 0.3);
        }
        
        .login-container {
            background: rgba(255, 255, 255, 0.95);
            backdrop-filter: blur(10px);
            padding: 2rem;
            border-radius: 20px;
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.2);
            width: 100%;
            max-width: 400px;
            position: relative;
            z-index: 10;
            border: 1px solid rgba(255, 255, 255, 0.3);
        }
        
        .login-header {
            text-align: center;
            margin-bottom: 2rem;
        }
        
        .login-header h1 {
            color: #333;
            font-size: 1.8rem;
            margin-bottom: 0.5rem;
        }
        
        .login-header p {
            color: #666;
            font-size: 0.9rem;
        }
        
        .form-group {
            margin-bottom: 1.5rem;
        }
        
        .form-group label {
            display: block;
            margin-bottom: 0.5rem;
            color: #333;
            font-weight: 500;
        }
        
        .form-group input {
            width: 100%;
            padding: 0.75rem;
            border: 2px solid #e1e5e9;
            border-radius: 10px;
            font-size: 1rem;
            transition: all 0.3s ease;
            background: rgba(255, 255, 255, 0.9);
        }
        
        .form-group input:focus {
            outline: none;
            border-color: #667eea;
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.2);
            background: rgba(255, 255, 255, 1);
        }
        
        .login-button {
            width: 100%;
            padding: 0.75rem;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 10px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.3s ease;
            position: relative;
            overflow: hidden;
        }
        
        .login-button::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
            transition: left 0.5s;
        }
        
        .login-button:hover {
            transform: translateY(-3px);
            box-shadow: 0 10px 25px rgba(102, 126, 234, 0.4);
        }
        
        .login-button:hover::before {
            left: 100%;
        }
        
        .login-button:active {
            transform: translateY(-1px);
        }
        
        .message {
            margin-top: 1rem;
            padding: 0.75rem;
            border-radius: 5px;
            text-align: center;
        }
        
        .message.success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        
        .message.error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
        
        .loading {
            display: none;
            text-align: center;
            margin-top: 1rem;
        }
        
        .spinner {
            border: 2px solid #f3f3f3;
            border-top: 2px solid #667eea;
            border-radius: 50%;
            width: 20px;
            height: 20px;
            animation: spin 1s linear infinite;
            margin: 0 auto;
        }
        
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <!-- Плавающие элементы фона -->
    <div class="floating-elements">
        <div class="floating-element lock-icon">🔒</div>
        <div class="floating-element key-icon">🔑</div>
        <div class="floating-element shield-icon">🛡️</div>
        <div class="floating-element star-icon">⭐</div>
        <div class="floating-element lock-icon">🔐</div>
        <div class="floating-element key-icon">🗝️</div>
        <div class="floating-element shield-icon">🔒</div>
        <div class="floating-element star-icon">✨</div>
        <div class="floating-element lock-icon">🔒</div>
    </div>
    
    <div class="login-container">
        <div class="login-header">
            <h1>🔐 Авторизация</h1>
            <p>Войдите в систему для доступа к LCG</p>
        </div>
        
        <form id="loginForm">
            <input type="hidden" id="csrf_token" name="csrf_token" value="{{.CSRFToken}}">
            
            <div class="form-group">
                <label for="username">Имя пользователя:</label>
                <input type="text" id="username" name="username" required placeholder="Введите имя пользователя">
            </div>
            
            <div class="form-group">
                <label for="password">Пароль:</label>
                <input type="password" id="password" name="password" required placeholder="Введите пароль">
            </div>
            
            <button type="submit" class="login-button">Войти</button>
        </form>
        
        <div class="loading" id="loading">
            <div class="spinner"></div>
            <p>Проверка авторизации...</p>
        </div>
        
        <div id="message"></div>
    </div>

    <script>
        document.getElementById('loginForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const form = e.target;
            const formData = new FormData(form);
            const username = formData.get('username');
            const password = formData.get('password');
            
            // Показываем загрузку
            document.getElementById('loading').style.display = 'block';
            document.getElementById('message').innerHTML = '';
            
            try {
                const csrfToken = document.getElementById('csrf_token').value;
                const response = await fetch('{{.BasePath}}/api/login', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify({
                        username: username,
                        password: password,
                        csrf_token: csrfToken
                    })
                });
                
                const data = await response.json();
                
                if (data.success) {
                    // Успешная авторизация, перенаправляем на главную страницу
                    window.location.href = '{{.BasePath}}/';
                } else {
                    // Ошибка авторизации
                    showMessage(data.error || 'Ошибка авторизации', 'error');
                }
            } catch (error) {
                showMessage('Ошибка соединения с сервером', 'error');
            } finally {
                document.getElementById('loading').style.display = 'none';
            }
        });
        
        function showMessage(text, type) {
            const messageDiv = document.getElementById('message');
            messageDiv.innerHTML = '<div class="message ' + type + '">' + text + '</div>';
        }
    </script>
</body>
</html>`
