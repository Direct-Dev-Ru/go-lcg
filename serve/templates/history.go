package templates

// HistoryPageTemplate шаблон страницы истории запросов
const HistoryPageTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>История запросов - LCG Results</title>
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
        .clear-btn {
            background: #e74c3c;
        }
        .clear-btn:hover {
            background: #c0392b;
        }
        .history-item {
            background: #f0f8f0;
            border: 1px solid #a8e6cf;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 15px;
            position: relative;
            cursor: pointer;
            transition: all 0.3s ease;
        }
        .history-item:hover {
            border-color: #2d5016;
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(45,80,22,0.2);
        }
        .history-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .history-index {
            background: #2d5016;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-weight: bold;
        }
        .history-timestamp {
            color: #666;
            font-size: 0.9em;
        }
        .history-command {
            font-weight: 600;
            color: #333;
            margin-bottom: 8px;
        }
        .history-response {
            background: #f8f9fa;
            padding: 10px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9em;
            color: #2d5016;
            border-left: 3px solid #2d5016;
            max-height: 72px; /* ~4 строки */
            overflow: hidden;
            display: -webkit-box;
            -webkit-line-clamp: 4;
            -webkit-box-orient: vertical;
        }
        .delete-btn {
            background: #e74c3c;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.8em;
            transition: background 0.3s ease;
        }
        .delete-btn:hover {
            background: #c0392b;
        }
        .empty-state {
            text-align: center;
            padding: 60px 20px;
            color: #666;
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
            .history-header { flex-direction: column; align-items: flex-start; gap: 8px; }
            .history-item { padding: 15px; }
            .history-response { font-size: 0.85em; }
            .search-container input { font-size: 16px; width: 96% !important; }
        }
        
        @media (max-width: 480px) {
            .header h1 { font-size: 1.8em; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📝 История запросов</h1>
            <p>Управление историей запросов {{.AppName}}</p>
        </div>
        <div class="content">
            <div class="nav-buttons">
                <a href="{{.BasePath}}/" class="nav-btn">🏠 Главная</a>
                <a href="{{.BasePath}}/run" class="nav-btn">🚀 Выполнение</a>
                <a href="{{.BasePath}}/prompts" class="nav-btn">⚙️ Промпты</a>
                <button class="nav-btn clear-btn" onclick="clearHistory()">🗑️ Очистить всю историю</button>
            </div>
            
            <!-- Поиск -->
            <div class="search-container" style="margin: 20px 0;">
                <input type="text" id="searchInput" placeholder="🔍 Поиск по командам, ответам и объяснениям..." 
                       style="width: 100%; padding: 12px; border: 1px solid #ddd; border-radius: 6px; font-size: 16px;">
                <div id="searchResults" style="margin-top: 10px; color: #666; font-size: 14px;"></div>
            </div>

            {{if .Entries}}
            {{range .Entries}}
            <div class="history-item" onclick="viewHistoryEntry({{.Index}})">
                <div class="history-header">
                    <div>
                        <span class="history-index">#{{.Index}}</span>
                        <span class="history-timestamp">{{.Timestamp}}</span>
                    </div>
                    <button class="delete-btn" onclick="event.stopPropagation(); deleteHistoryEntry({{.Index}})">🗑️ Удалить</button>
                </div>
                <div class="history-command">{{.Command}}</div>
                <div class="history-response">{{.Response}}</div>
            </div>
            {{end}}
            {{else}}
            <div class="empty-state">
                <h3>📝 История пуста</h3>
                <p>Здесь будут отображаться запросы после использования команды lcg</p>
            </div>
            {{end}}
        </div>
    </div>
    
    <script>
        function viewHistoryEntry(index) {
            window.location.href = '{{.BasePath}}/history/view/' + index;
        }
        
        function deleteHistoryEntry(index) {
            if (confirm('Вы уверены, что хотите удалить запись #' + index + '?')) {
                fetch('{{.BasePath}}/history/delete/' + index, {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при удалении записи');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при удалении записи');
                });
            }
        }
        
        function clearHistory() {
            if (confirm('Вы уверены, что хотите очистить всю историю?\\n\\nЭто действие нельзя отменить.')) {
                fetch('{{.BasePath}}/history/clear', {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        location.reload();
                    } else {
                        alert('Ошибка при очистке истории');
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Ошибка при очистке истории');
                });
            }
        }
        
        // Поиск по истории
        function performSearch() {
            const searchTerm = document.getElementById('searchInput').value.trim();
            const searchResults = document.getElementById('searchResults');
            const historyItems = document.querySelectorAll('.history-item');
            
            if (searchTerm === '') {
                // Показать все записи
                historyItems.forEach(item => {
                    item.style.display = 'block';
                });
                searchResults.textContent = '';
                return;
            }
            
            let visibleCount = 0;
            let totalCount = historyItems.length;
            
            historyItems.forEach(item => {
                const command = item.querySelector('.history-command').textContent.toLowerCase();
                const response = item.querySelector('.history-response').textContent.toLowerCase();
                
                // Объединяем команду и ответ для поиска
                const searchContent = command + ' ' + response;
                
                let matches = false;
                
                // Проверяем, есть ли фраза в кавычках
                if (searchTerm.startsWith("'") && searchTerm.endsWith("'")) {
                    // Поиск точной фразы
                    const phrase = searchTerm.slice(1, -1).toLowerCase();
                    matches = searchContent.includes(phrase);
                } else {
                    // Поиск по отдельным словам
                    const words = searchTerm.toLowerCase().split(/\s+/);
                    matches = words.every(word => searchContent.includes(word));
                }
                
                if (matches) {
                    item.style.display = 'block';
                    visibleCount++;
                } else {
                    item.style.display = 'none';
                }
            });
            
            // Обновляем информацию о результатах
            if (visibleCount === 0) {
                searchResults.textContent = '🔍 Ничего не найдено';
                searchResults.style.color = '#e74c3c';
            } else if (visibleCount === totalCount) {
                searchResults.textContent = '';
            } else {
                searchResults.textContent = '🔍 Найдено: ' + visibleCount + ' из ' + totalCount + ' записей';
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
