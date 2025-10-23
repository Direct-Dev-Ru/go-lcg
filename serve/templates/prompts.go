package templates

// PromptsPageTemplate шаблон страницы управления промптами
const PromptsPageTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Системные промпты - LCG Results</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #56ab2f 0%, #a8e6cf 100%);
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2d5016 0%, #4a7c59 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        .content {
            padding: 30px;
        }
        .nav-buttons {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            flex-wrap: wrap;
        }
        .nav-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 1em;
            text-decoration: none;
            transition: background 0.3s ease;
            display: inline-block;
            text-align: center;
        }
        .nav-btn:hover {
            background: #2980b9;
        }
        .add-btn {
            background: #27ae60;
        }
        .add-btn:hover {
            background: #229954;
        }
        .prompt-item {
            background: #f0f8f0;
            border: 1px solid #a8e6cf;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 15px;
            position: relative;
        }
        .prompt-item:hover {
            border-color: #2d5016;
        }
        .prompt-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .prompt-id {
            background: #2d5016;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-weight: bold;
        }
        .prompt-name {
            font-weight: 600;
            color: #333;
            font-size: 1.2em;
        }
        .prompt-description {
            color: #666;
            margin-bottom: 10px;
        }
        .prompt-content {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9em;
            color: #2d5016;
            border-left: 3px solid #2d5016;
            white-space: pre-wrap;
        }
        .prompt-actions {
            position: absolute;
            top: 10px;
            right: 10px;
            display: flex;
            gap: 8px;
        }
        .action-btn {
            background: #4a7c59;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            transition: background 0.3s ease;
        }
        .action-btn:hover {
            background: #2d5016;
        }
        .delete-btn {
            background: #e74c3c;
        }
        .delete-btn:hover {
            background: #c0392b;
        }
        .restore-btn {
            background: #3498db;
        }
        .restore-btn:hover {
            background: #2980b9;
        }
        .default-badge {
            background: #28a745;
            color: white;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 0.7em;
            margin-left: 8px;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }
        .lang-switcher {
            display: flex;
            gap: 5px;
            margin-left: auto;
        }
        .lang-btn {
            background: #6c757d;
            color: white;
            border: none;
            padding: 8px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9em;
            transition: background 0.3s ease;
        }
        .lang-btn:hover {
            background: #5a6268;
        }
        .lang-btn.active {
            background: #3498db;
        }
        .lang-btn.active:hover {
            background: #2980b9;
        }
        .tabs {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            border-bottom: 2px solid #e9ecef;
        }
        .tab-btn {
            background: #f8f9fa;
            color: #6c757d;
            border: none;
            padding: 12px 20px;
            border-radius: 6px 6px 0 0;
            cursor: pointer;
            font-size: 1em;
            transition: all 0.3s ease;
            border-bottom: 3px solid transparent;
        }
        .tab-btn:hover {
            background: #e9ecef;
            color: #495057;
        }
        .tab-btn.active {
            background: #3498db;
            color: white;
            border-bottom-color: #2980b9;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        
        /* Мобильная адаптация */
        @media (max-width: 768px) {
            body { padding: 10px; }
            .container { margin: 0; border-radius: 8px; box-shadow: 0 10px 20px rgba(0,0,0,0.1); }
            .header { padding: 20px; }
            .header h1 { font-size: 2em; }
            .content { padding: 20px; }
            .nav-buttons { flex-direction: column; gap: 8px; }
            .nav-btn { text-align: center; padding: 12px 16px; font-size: 14px; }
            .lang-switcher { margin-left: 0; }
            .tabs { flex-direction: column; gap: 8px; }
            .tab-btn { text-align: center; }
            .prompt-item { padding: 15px; }
            .prompt-content { font-size: 0.85em; }
        }
        @media (max-width: 480px) {
            .header h1 { font-size: 1.8em; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>⚙️ Системные промпты</h1>
            <p>Управление системными промптами Linux Command GPT</p>
        </div>
        <div class="content">
            <div class="nav-buttons">
                <a href="/" class="nav-btn">🏠 Главная</a>
                <a href="/run" class="nav-btn">🚀 Выполнение</a>
                <a href="/history" class="nav-btn">📝 История</a>
                <button class="nav-btn add-btn" onclick="showAddForm()">➕ Добавить промпт</button>
                <div class="lang-switcher">
                    <button class="lang-btn {{if eq .Lang "ru"}}active{{end}}" onclick="switchLang('ru')">🇷🇺 RU</button>
                    <button class="lang-btn {{if eq .Lang "en"}}active{{end}}" onclick="switchLang('en')">🇺🇸 EN</button>
                </div>
            </div>
            
            <!-- Вкладки -->
            <div class="tabs">
                <button class="tab-btn active" onclick="switchTab('system')">⚙️ Системные промпты</button>
                <button class="tab-btn" onclick="switchTab('verbose')">📝 Промпты подробности (v/vv/vvv)</button>
            </div>

            <!-- Вкладка системных промптов -->
            <div id="system-tab" class="tab-content active">
                {{if .Prompts}}
                {{range .Prompts}}
                <div class="prompt-item">
                    <div class="prompt-actions">
                        <button class="action-btn" onclick="editPrompt({{.ID}}, '{{.Name}}', '{{.Description}}', '{{.Content}}')">✏️</button>
                        <button class="action-btn restore-btn" onclick="restorePrompt({{.ID}})" title="Восстановить к значению по умолчанию">🔄</button>
                        <button class="action-btn delete-btn" onclick="deletePrompt({{.ID}})">🗑️</button>
                    </div>
                    <div class="prompt-header">
                        <div>
                            <span class="prompt-id">#{{.ID}}</span>
                            <span class="prompt-name">{{.Name}}</span>
                            {{if .IsDefault}}<span class="default-badge">Встроенный</span>{{end}}
                        </div>
                    </div>
                    <div class="prompt-description">{{.Description}}</div>
                    <div class="prompt-content">{{.Content}}</div>
                </div>
                {{end}}
                {{else}}
                <div class="empty-state">
                    <h3>⚙️ Промпты не найдены</h3>
                    <p>Добавьте пользовательские промпты для настройки поведения системы</p>
                </div>
                {{end}}
            </div>
            
            <!-- Вкладка промптов подробности -->
            <div id="verbose-tab" class="tab-content">
                {{if .VerbosePrompts}}
                {{range .VerbosePrompts}}
                <div class="prompt-item">
                    <div class="prompt-actions">
                        <button class="action-btn" onclick="editVerbosePrompt('{{.Mode}}', '{{.Content}}')">✏️</button>
                        <button class="action-btn restore-btn" onclick="restoreVerbosePrompt('{{.Mode}}')" title="Восстановить к значению по умолчанию">🔄</button>
                    </div>
                    <div class="prompt-header">
                        <div>
                            <span class="prompt-id">#{{.Mode}}</span>
                            <span class="prompt-name">{{.Name}}</span>
                            {{if .IsDefault}}<span class="default-badge">Встроенный</span>{{end}}
                        </div>
                    </div>
                    <div class="prompt-description">{{.Description}}</div>
                    <div class="prompt-content">{{.Content}}</div>
                </div>
                {{end}}
                {{else}}
                <div class="empty-state">
                    <h3>📝 Промпты подробности</h3>
                    <p>Промпты для режимов v, vv, vvv</p>
                </div>
                {{end}}
            </div>
        </div>
    </div>
    
    <!-- Форма добавления/редактирования -->
    <div id="promptForm" style="display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000;">
        <div style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); background: white; padding: 30px; border-radius: 12px; max-width: 600px; width: 90%;">
            <h3 id="formTitle">Добавить промпт</h3>
            <form id="promptFormData">
                <input type="hidden" id="promptId" name="id">
                <div style="margin-bottom: 15px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 600;">Название:</label>
                    <input type="text" id="promptName" name="name" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;" required>
                </div>
                <div style="margin-bottom: 15px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 600;">Описание:</label>
                    <input type="text" id="promptDescription" name="description" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px;" required>
                </div>
                <div style="margin-bottom: 20px;">
                    <label style="display: block; margin-bottom: 5px; font-weight: 600;">Содержание:</label>
                    <textarea id="promptContent" name="content" rows="6" style="width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; font-family: monospace;" required></textarea>
                </div>
                <div style="text-align: right;">
                    <button type="button" onclick="hideForm()" style="background: #6c757d; color: white; border: none; padding: 8px 16px; border-radius: 4px; margin-right: 10px; cursor: pointer;">Отмена</button>
                    <button type="submit" style="background: #2d5016; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer;">Сохранить</button>
                </div>
            </form>
        </div>
    </div>
    
    <script>
        function showAddForm() {
            document.getElementById('formTitle').textContent = 'Добавить промпт';
            document.getElementById('promptFormData').reset();
            document.getElementById('promptId').value = '';
            document.getElementById('promptForm').style.display = 'block';
        }
        
        function editPrompt(id, name, description, content) {
            document.getElementById('formTitle').textContent = 'Редактировать промпт';
            document.getElementById('promptId').value = id;
            document.getElementById('promptName').value = name;
            document.getElementById('promptDescription').value = description;
            document.getElementById('promptContent').value = content;
            document.getElementById('promptForm').style.display = 'block';
        }
        
        function hideForm() {
            document.getElementById('promptForm').style.display = 'none';
        }
        
        function switchTab(tabName) {
            // Скрываем все вкладки
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            
            // Убираем активный класс с кнопок
            document.querySelectorAll('.tab-btn').forEach(btn => {
                btn.classList.remove('active');
            });
            
            // Показываем нужную вкладку
            document.getElementById(tabName + '-tab').classList.add('active');
            
            // Активируем нужную кнопку
            event.target.classList.add('active');
        }
        
        function switchLang(lang) {
            // Сохраняем текущие промпты перед переключением языка
            saveCurrentPrompts(lang);
            
            // Перезагружаем страницу с новым языком
            const url = new URL(window.location);
            url.searchParams.set('lang', lang);
            window.location.href = url.toString();
        }
        
        function saveCurrentPrompts(lang) {
            // Отправляем запрос для сохранения текущих промптов с новым языком
            fetch('/prompts/save-lang', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    lang: lang
                })
            })
            .catch(error => {
                console.error('Error saving prompts:', error);
            });
        }
        
        function editVerbosePrompt(mode, content) {
            // Редактирование промпта подробности
            document.getElementById('formTitle').textContent = 'Редактировать промпт подробности (' + mode + ')';
            document.getElementById('promptId').value = mode;
            document.getElementById('promptName').value = mode;
            document.getElementById('promptDescription').value = 'Промпт для режима ' + mode;
            document.getElementById('promptContent').value = content;
            document.getElementById('promptForm').style.display = 'block';
        }
        
        function deletePrompt(id) {
            if (confirm('Вы уверены, что хотите удалить промпт #' + id + '?')) {
                fetch('/prompts/delete/' + id, {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при удалении промпта');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при удалении промпта');
                });
            }
        }
        
        document.getElementById('promptFormData').addEventListener('submit', function(e) {
            e.preventDefault();
            
            // Валидация длины полей
            const name = document.getElementById('promptName').value;
            const description = document.getElementById('promptDescription').value;
            const content = document.getElementById('promptContent').value;
            
            const maxContentLength = {{.MaxSystemPromptLength}};
            const maxNameLength = {{.MaxPromptNameLength}};
            const maxDescLength = {{.MaxPromptDescLength}};
            
            if (content.length > maxContentLength) {
                alert('Содержимое промпта слишком длинное: максимум ' + maxContentLength + ' символов');
                return;
            }
            if (name.length > maxNameLength) {
                alert('Название промпта слишком длинное: максимум ' + maxNameLength + ' символов');
                return;
            }
            if (description.length > maxDescLength) {
                alert('Описание промпта слишком длинное: максимум ' + maxDescLength + ' символов');
                return;
            }
            
            const formData = new FormData(this);
            const id = formData.get('id');
            
            // Определяем, это системный промпт или промпт подробности
            const isVerbosePrompt = ['v', 'vv', 'vvv'].includes(id);
            
            let url, method;
            if (isVerbosePrompt) {
                url = '/prompts/edit-verbose/' + id;
                method = 'PUT';
            } else {
                url = id ? '/prompts/edit/' + id : '/prompts/add';
                method = id ? 'PUT' : 'POST';
            }
            
            fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    name: formData.get('name'),
                    description: formData.get('description'),
                    content: formData.get('content')
                })
            })
            .then(response => {
                if (response.ok) {
                    location.reload();
                } else {
                    alert('Ошибка при сохранении промпта');
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Ошибка при сохранении промпта');
            });
        });

        // Функция восстановления системного промпта
        function restorePrompt(id) {
            if (confirm('Восстановить промпт к значению по умолчанию?')) {
                fetch('/prompts/restore/' + id, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert('Промпт восстановлен');
                        location.reload();
                    } else {
                        alert('Ошибка: ' + data.error);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при восстановлении промпта');
                });
            }
        }

        // Функция восстановления verbose промпта
        function restoreVerbosePrompt(mode) {
            if (confirm('Восстановить промпт к значению по умолчанию?')) {
                fetch('/prompts/restore-verbose/' + mode, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    }
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert('Промпт восстановлен');
                        location.reload();
                    } else {
                        alert('Ошибка: ' + data.error);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при восстановлении промпта');
                });
            }
        }
    </script>
</body>
</html>`
