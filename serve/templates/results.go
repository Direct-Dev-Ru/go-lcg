package templates

// ResultsPageTemplate шаблон главной страницы со списком файлов
const ResultsPageTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.AppAbbreviation}} Результаты - {{.AppName}}</title>
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
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
            font-size: 1.1em;
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
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: #f0f8f0;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
            border-left: 4px solid #2d5016;
        }
        .stat-number {
            font-size: 2em;
            font-weight: bold;
            color: #2d5016;
        }
        .stat-label {
            color: #666;
            margin-top: 5px;
        }
        .files-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 20px;
            align-items: stretch;
            grid-auto-rows: auto;
        }
        .file-card {
            background: white;
            border: 1px solid #e1e5e9;
            border-radius: 8px;
            padding: 20px;
            transition: all 0.3s ease;
            position: relative;
        }
        .file-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(45,80,22,0.2);
            border-color: #2d5016;
        }
        .file-card-content {
            cursor: pointer;
            padding-left: 28px;
        }
        .file-actions {
            position: absolute;
            top: 10px;
            left: 10px;
            display: flex;
            gap: 8px;
        }
        .delete-btn {
            background: transparent;
            color: #ef9a9a; /* бледно-красный */
            border: none;
            padding: 4px 8px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 18px;
            line-height: 1;
            transition: color 0.2s ease, transform 0.2s ease;
        }
        .delete-btn:hover {
            color:rgb(171, 27, 24); /* чуть ярче при ховере */
            transform: translateY(-1px);
        }
        .file-name {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
            font-size: 1.1em;
            padding-right: 10px;
        }
        .file-info {
            color: #666;
            font-size: 0.9em;
            margin-bottom: 10px;
        }
        .file-preview {
            background: #f0f8f0;
            padding: 10px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.85em;
            color: #2d5016;
            max-height: 100px;
            overflow: hidden;
            border-left: 3px solid #2d5016;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
        }
        .empty-state h3 {
            color: #333;
            margin-bottom: 10px;
        }
        .nav-btn, .nav-button {
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
        .nav-btn:hover, .nav-button:hover {
            background: #2980b9;
        }
        
        /* Мобильная адаптация */
        @media (max-width: 768px) {
            body { padding: 10px; }
            .container { margin: 0; border-radius: 8px; box-shadow: 0 10px 20px rgba(0,0,0,0.1); }
            .header { padding: 20px; }
            .header h1 { font-size: 1.9em; }
            .content { padding: 20px; }
            .files-grid { dummy-attr: none; }
            /* Стили карточек как в истории */
            .file-card { background: #f0f8f0; border: 1px solid #a8e6cf; padding: 15px; }
            .file-card:hover { border-color: #2d5016; box-shadow: 0 8px 25px rgba(45,80,22,0.2); transform: translateY(-2px); }
            .file-name { color: #333; margin-bottom: 8px; }
            .file-info { color: #666; font-size: 0.9em; }
            .file-preview { background: #f8f9fa; border-left: 3px solid #2d5016; font-size: 0.85em; }
            .file-actions { top: 8px; left: 8px; }
            .delete-btn { padding: 2px 6px; font-size: 16px; }
            .stats { grid-template-columns: 1fr 1fr; }
            .nav-buttons { flex-direction: column; gap: 8px; }
            .nav-btn, .nav-button { text-align: center; padding: 12px 16px; font-size: 14px; }
            .search-container input { font-size: 16px; width: 96% !important; }
        }
        @media (max-width: 480px) {
            .header h1 { font-size: 1.6em; }
            .content { padding: 16px; }
            .stats { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 {{.AppAbbreviation}} - {{.AppName}}</h1>
            <p>Просмотр сохраненных результатов {{.AppName}}</p>
        </div>
        <div class="content">
            <div class="nav-buttons">
                <button class="nav-btn" onclick="location.reload()">🔄 Обновить</button>
                <a href="{{.BasePath}}/run" class="nav-btn">🚀 Выполнение</a>
                <a href="{{.BasePath}}/history" class="nav-btn">📝 История</a>
                <a href="{{.BasePath}}/prompts" class="nav-btn">⚙️ Промпты</a>
            </div>
            
            <!-- Поиск -->
            <div class="search-container">
                <input type="text" id="searchInput" placeholder="🔍 Поиск по содержимому файлов..." 
                       style="width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 16px;">
                <div id="searchResults" style="margin-top: 10px; color: #666; font-size: 14px;"></div>
            </div>
            
            <div class="stats">
                <div class="stat-card">
                    <div class="stat-number">{{.TotalFiles}}</div>
                    <div class="stat-label">Всего файлов</div>
                </div>
                <div class="stat-card">
                    <div class="stat-number">{{.RecentFiles}}</div>
                    <div class="stat-label">За последние 7 дней</div>
                </div>
            </div>

            {{if .Files}}
            <div class="files-grid">
                {{range .Files}}
                <div class="file-card" data-content="{{.Content}}">
                    <div class="file-actions">
                        <button class="delete-btn" onclick="deleteFile('{{.Name}}')" title="Удалить файл">✖</button>
                    </div>
                    <div class="file-card-content" onclick="window.location.href='{{$.BasePath}}/file/{{.Name}}'">
                        <div class="file-name">{{.DisplayName}}</div>
                        <div class="file-info">
                            📅 {{.ModTime}} | 📏 {{.Size}}
                        </div>
                        <div class="file-preview">{{.Preview}}</div>
                    </div>
                </div>
                {{end}}
            </div>
            {{else}}
            <div class="empty-state">
                <h3>📁 Папка пуста</h3>
                <p>Здесь будут отображаться сохраненные результаты после использования команды lcg</p>
            </div>
            {{end}}
        </div>
    </div>
    
    <script>
        function deleteFile(filename) {
            if (confirm('Вы уверены, что хотите удалить файл "' + filename + '"?\\n\\nЭто действие нельзя отменить.')) {
                fetch('{{.BasePath}}/delete/' + encodeURIComponent(filename), {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при удалении файла');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при удалении файла');
                });
            }
        }
        
        // Поиск по содержимому файлов
        function performSearch() {
            const searchTerm = document.getElementById('searchInput').value.trim();
            const searchResults = document.getElementById('searchResults');
            const fileCards = document.querySelectorAll('.file-card');
            
            if (searchTerm === '') {
                // Показать все файлы
                fileCards.forEach(card => {
                    card.style.display = 'block';
                });
                searchResults.textContent = '';
                return;
            }
            
            let visibleCount = 0;
            let totalCount = fileCards.length;
            
            fileCards.forEach(card => {
                const fileName = card.querySelector('.file-name').textContent.toLowerCase();
                const fullContent = card.getAttribute('data-content').toLowerCase();
                
                // Проверяем поиск по полному содержимому файла
                const fileContent = fileName + ' ' + fullContent;
                
                let matches = false;
                
                // Проверяем, есть ли фраза в кавычках
                if (searchTerm.startsWith("'") && searchTerm.endsWith("'")) {
                    // Поиск точной фразы
                    const phrase = searchTerm.slice(1, -1).toLowerCase();
                    matches = fileContent.includes(phrase);
                } else {
                    // Поиск по отдельным словам
                    const words = searchTerm.toLowerCase().split(/\s+/);
                    matches = words.every(word => fileContent.includes(word));
                }
                
                if (matches) {
                    card.style.display = 'block';
                    visibleCount++;
                } else {
                    card.style.display = 'none';
                }
            });
            
            // Обновляем информацию о результатах
            if (visibleCount === 0) {
                searchResults.textContent = '🔍 Ничего не найдено';
                searchResults.style.color = '#e74c3c';
            } else if (visibleCount === totalCount) {
                searchResults.textContent = '';
            } else {
                searchResults.textContent = '🔍 Найдено: ' + visibleCount + ' из ' + totalCount + ' файлов';
                searchResults.style.color = '#27ae60';
            }
        }
        
        // Обработчик ввода в поле поиска
        document.getElementById('searchInput').addEventListener('input', performSearch);
        
        // Обработчик Enter в поле поиска
        document.getElementById('searchInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                performSearch();
            }
        });
    </script>
</body>
</html>`
