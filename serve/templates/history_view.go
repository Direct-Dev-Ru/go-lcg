package templates

// HistoryViewTemplate шаблон для просмотра записи истории
const HistoryViewTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Запись #{{.Index}} - LCG History</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #56ab2f 0%, #a8e6cf 100%);
            min-height: 100vh;
        }
        .container {
            max-width: 1000px;
            margin: 0 auto;
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #2d5016 0%, #4a7c59 100%);
            color: white;
            padding: 20px 30px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            margin: 0;
            font-size: 1.5em;
            font-weight: 300;
        }
        .back-btn {
            background: rgba(255,255,255,0.2);
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            text-decoration: none;
            transition: background 0.3s ease;
        }
        .back-btn:hover {
            background: rgba(255,255,255,0.3);
        }
        .content {
            padding: 30px;
            line-height: 1.6;
        }
        .history-meta {
            background: #f0f8f0;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            border-left: 4px solid #2d5016;
        }
        .history-meta-item {
            margin: 5px 0;
            color: #666;
        }
        .history-meta-label {
            font-weight: 600;
            color: #2d5016;
        }
        .history-command {
            background: #f8f9fa;
            padding: 15px;
            border-radius: 8px;
            margin-bottom: 20px;
            border-left: 4px solid #4a7c59;
        }
        .history-command h3 {
            margin: 0 0 10px 0;
            color: #2d5016;
        }
        .history-command-text {
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 1.1em;
            color: #333;
            white-space: pre-wrap;
        }
        .history-response {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #2d5016;
        }
        .history-response h3 {
            margin: 0 0 15px 0;
            color: #2d5016;
        }
        .history-response-content {
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.95em;
            color: #333;
            white-space: pre-wrap;
            line-height: 1.5;
        }
        .history-explanation {
            background: #f0f8f0;
            padding: 20px;
            border-radius: 8px;
            margin-top: 20px;
            border-left: 4px solid #4a7c59;
        }
        .history-explanation h3 {
            margin: 0 0 15px 0;
            color: #2d5016;
        }
        .history-explanation-content {
            color: #333;
            line-height: 1.6;
        }
        .history-explanation-content h1,
        .history-explanation-content h2,
        .history-explanation-content h3,
        .history-explanation-content h4,
        .history-explanation-content h5,
        .history-explanation-content h6 {
            color: #2d5016;
            margin-top: 20px;
            margin-bottom: 10px;
        }
        .history-explanation-content h1 {
            border-bottom: 2px solid #2d5016;
            padding-bottom: 5px;
        }
        .history-explanation-content h2 {
            border-bottom: 1px solid #4a7c59;
            padding-bottom: 3px;
        }
        .history-explanation-content code {
            background: #f0f8f0;
            padding: 2px 6px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            color: #2d5016;
            border: 1px solid #a8e6cf;
        }
        .history-explanation-content pre {
            background: #f0f8f0;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #2d5016;
            overflow-x: auto;
        }
        .history-explanation-content pre code {
            background: none;
            padding: 0;
            border: none;
            color: #2d5016;
        }
        .history-explanation-content blockquote {
            border-left: 4px solid #4a7c59;
            margin: 15px 0;
            padding: 10px 20px;
            background: #f0f8f0;
            border-radius: 0 8px 8px 0;
        }
        .history-explanation-content ul,
        .history-explanation-content ol {
            padding-left: 20px;
        }
        .history-explanation-content li {
            margin: 5px 0;
        }
        .history-explanation-content strong {
            color: #2d5016;
        }
        .history-explanation-content em {
            color: #4a7c59;
        }
        .actions {
            margin-top: 20px;
            display: flex;
            gap: 10px;
        }
        .action-btn {
            background: #3498db;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            text-decoration: none;
            transition: background 0.3s ease;
            display: inline-block;
        }
        .action-btn:hover {
            background: #2980b9;
        }
        .delete-btn {
            background: #e74c3c;
        }
        .delete-btn:hover {
            background: #c0392b;
        }
        
        /* Мобильная адаптация */
        @media (max-width: 768px) {
            body { padding: 10px; }
            .container { margin: 0; border-radius: 8px; box-shadow: 0 10px 20px rgba(0,0,0,0.1); }
            .header { padding: 16px; }
            .header h1 { font-size: 1.2em; }
            .back-btn { padding: 6px 12px; font-size: 0.9em; }
            .content { padding: 20px; }
            .actions { flex-direction: column; }
            .action-btn { text-align: center; }
            .history-response-content { font-size: 0.9em; }
        }
        @media (max-width: 480px) {
            .header h1 { font-size: 1em; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📝 Запись #{{.Index}}</h1>
            <a href="{{.BasePath}}/history" class="back-btn">← Назад к истории</a>
        </div>
        <div class="content">
            <div class="history-meta">
                <div class="history-meta-item">
                    <span class="history-meta-label">📅 Время:</span> {{.Timestamp}}
                </div>
                <div class="history-meta-item">
                    <span class="history-meta-label">🔢 Индекс:</span> #{{.Index}}
                </div>
            </div>
            
            <div class="history-command">
                <h3>💬 Запрос пользователя:</h3>
                <div class="history-command-text">{{.Command}}</div>
            </div>
            
            <div class="history-response">
                <h3>🤖 Ответ Модели:</h3>
                <div class="history-response-content">{{.Response}}</div>
            </div>
            
            {{.ExplanationHTML}}
            
            <div class="actions">
                <a href="{{.BasePath}}/history" class="action-btn">📝 К истории</a>
                <button class="action-btn delete-btn" onclick="deleteHistoryEntry({{.Index}})">🗑️ Удалить запись</button>
            </div>
        </div>
    </div>
    
    <script>
        function deleteHistoryEntry(index) {
            if (confirm('Вы уверены, что хотите удалить запись #' + index + '?')) {
                fetch('{{.BasePath}}/history/delete/' + index, {
                    method: 'DELETE'
                })
                .then(response => {
                    if (response.ok) {
                        window.location.href = '{{.BasePath}}/history';
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
    </script>
</body>
</html>`
