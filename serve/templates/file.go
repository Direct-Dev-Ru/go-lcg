package templates

// FileViewTemplate шаблон для просмотра файла результата
const FileViewTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Filename}} - LCG Results</title>
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
        .content h1 {
            color: #2d5016;
            border-bottom: 2px solid #2d5016;
            padding-bottom: 10px;
        }
        .content h2 {
            color: #4a7c59;
            margin-top: 30px;
        }
        .content h3 {
            color: #2d5016;
        }
        .content code {
            background: #f0f8f0;
            padding: 2px 6px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            color: #2d5016;
            border: 1px solid #a8e6cf;
        }
        .content pre {
            background: #f0f8f0;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #2d5016;
            overflow-x: auto;
        }
        .content pre code {
            background: none;
            padding: 0;
            border: none;
            color: #2d5016;
        }
        .content blockquote {
            border-left: 4px solid #4a7c59;
            margin: 20px 0;
            padding: 10px 20px;
            background: #f0f8f0;
            border-radius: 0 8px 8px 0;
        }
        .content ul, .content ol {
            padding-left: 20px;
        }
        .content li {
            margin: 5px 0;
        }
        .content strong {
            color: #2d5016;
        }
        .content em {
            color: #4a7c59;
        }
        
        /* Мобильная адаптация */
        @media (max-width: 768px) {
            body { padding: 10px; }
            .container { margin: 0; border-radius: 8px; box-shadow: 0 10px 20px rgba(0,0,0,0.1); }
            .header { padding: 16px; }
            .header h1 { font-size: 1.2em; }
            .back-btn { padding: 6px 12px; font-size: 0.9em; }
            .content { padding: 20px; }
            .content pre { font-size: 14px; }
        }
        @media (max-width: 480px) {
            .header h1 { font-size: 1em; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📄 {{.Filename}}</h1>
            <a href="{{.BasePath}}/" class="back-btn">← Назад к списку</a>
        </div>
        <div class="content">
            {{.Content}}
        </div>
    </div>
</body>
</html>`
