package templates

import "html/template"

// ExecutePageCSSTemplate - CSS стили для страницы выполнения запросов
var ExecutePageCSSTemplate = template.Must(template.New("execute_css").Parse(`
* {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            /* Динамичный плавный градиент (современный стиль) */
            background: linear-gradient(135deg, #5b86e5, #36d1dc, #4a7c59, #764ba2);
            background-size: 400% 400%;
            animation: gradientShift 18s ease infinite;
            min-height: 100vh;
            padding: 20px;
        }
        
        /* Анимация плавного перелива фона */
        @keyframes gradientShift {
            0% { background-position: 0% 50%; }
            50% { background-position: 100% 50%; }
            100% { background-position: 0% 50%; }
        }
        
        /* Учитываем системные настройки доступности */
        @media (prefers-reduced-motion: reduce) {
            body { animation: none; }
        }
        
        /* Улучшения для touch-устройств */
        .nav-btn, .submit-btn, .reset-btn, .verbose-btn, .action-btn {
            -webkit-tap-highlight-color: transparent;
            touch-action: manipulation;
        }
        
        /* Оптимизация производительности */
        .container {
            will-change: transform;
        }
        
        /* Улучшение читаемости на мобильных */
        @media (max-width: 768px) {
            .command-result code {
                font-size: 0.9em;
                padding: 1px 4px;
            }
            .command-result pre {
                font-size: 14px;
                padding: 12px;
            }
            .explanation-content {
                font-size: 15px;
                line-height: 1.6;
            }
        }
        .container {
            max-width: 800px;
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
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        .header p {
            opacity: 0.9;
            font-size: 1.1em;
        }
        .content {
            padding: 30px;
        }
        .nav-buttons {
            display: flex;
            gap: 10px;
            margin-bottom: 30px;
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
        }
        .nav-btn:hover {
            background: #2980b9;
        }
        .form-section {
            background: #f8f9fa;
            padding: 25px;
            border-radius: 8px;
            margin-bottom: 20px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        .form-group label {
            display: block;
            margin-bottom: 8px;
            font-weight: 600;
            color: #2d5016;
        }
        .form-group select,
        .form-group textarea {
            width: 100%;
            padding: 12px;
            border: 1px solid #ddd;
            border-radius: 6px;
            font-size: 16px;
            transition: border-color 0.3s ease;
        }
        .form-group select:focus,
        .form-group textarea:focus {
            outline: none;
            border-color: #2d5016;
            box-shadow: 0 0 0 3px rgba(45, 80, 22, 0.1);
        }
        .form-group textarea {
            resize: vertical;
            min-height: 120px;
        }
        .submit-btn {
            background: linear-gradient(135deg, #2d5016 0%, #4a7c59 100%);
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            width: 100%;
        }
        .submit-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(45, 80, 22, 0.3);
        }
        .submit-btn:disabled {
            background: #6c757d;
            cursor: not-allowed;
            transform: none;
            box-shadow: none;
        }
        .form-buttons {
            display: flex;
            gap: 15px;
            margin-top: 20px;
        }
        .reset-btn {
            background: linear-gradient(135deg, #e74c3c 0%, #c0392b 100%);
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            flex: 1;
        }
        .reset-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(231, 76, 60, 0.3);
        }
        .reset-btn:disabled {
            background: #6c757d;
            cursor: not-allowed;
            transform: none;
            box-shadow: none;
        }
        .result-section {
            margin-top: 30px;
        }
        .command-result {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #2d5016;
            margin-bottom: 20px;            
        }
        .command-result h3 {
            color: #2d5016;
            margin-bottom: 15px;
        }
        /* Заголовки внутри результата команды */
        .command-result h1,
        .command-result h2,
        .command-result h3,
        .command-result h4,
        .command-result h5,
        .command-result h6 {
            margin-top: 18px;   /* отделяем сверху */
            margin-bottom: 10px;/* и немного снизу */
            line-height: 1.25;
        }
        /* Ритм текста внутри markdown-блока команды */
        .command-result .command-md { line-height: 1.7; }
        .command-result p { margin: 10px 0 14px; line-height: 1.7; }
        .command-result ul,
        .command-result ol { margin: 10px 0 14px 24px; line-height: 1.7; }
        .command-result li { margin: 6px 0; }
        .command-result hr { margin: 18px 0; border: 0; border-top: 1px solid #e1e5e9; }
        /* Подсветка code внутри результата команды */
        .command-result code {
            background: #e6f4ea; /* светло-зеленый фон */
            color: #2e7d32;      /* зеленый текст */
            border: 1px solid #b7dfb9;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 0.98em;  /* немного крупнее базового */
        }
        .command-result pre {
            background: #eaf7ee;               /* мягкий зеленоватый фон */
            border-left: 4px solid #2e7d32;    /* зеленая полоса слева */
            padding: 14px;
            border-radius: 8px;
            overflow-x: auto;
            margin: 12px 0 16px;               /* вертикальные отступы вокруг кода */
        }
        .command-result pre code {
            background: none;
            border: none;
            color: #2e7d32;
            font-size: 16px; /* увеличить размер шрифта в блоках */
        }
        .command-code {
            background: #2d5016;
            color: white;
            padding: 15px;
            border-radius: 6px;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 16px;
            margin-bottom: 15px;
            word-break: break-all;
        }
        .result-meta {
            display: flex;
            gap: 20px;
            color: #6c757d;
            font-size: 14px;
        }
        .explanation-section {
            background: #f0f8f0;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #4a7c59;
            margin-top: 20px;
        }
        .explanation-section h3 {
            color: #2d5016;
            margin-bottom: 15px;
        }
        .explanation-content {
            color: #333;
            line-height: 1.6;
        }
        .explanation-content h1,
        .explanation-content h2,
        .explanation-content h3,
        .explanation-content h4,
        .explanation-content h5,
        .explanation-content h6 {
            color: #2d5016;
            margin-top: 20px;
            margin-bottom: 10px;
        }
        .explanation-content h1 {
            border-bottom: 2px solid #2d5016;
            padding-bottom: 5px;
        }
        .explanation-content h2 {
            border-bottom: 1px solid #4a7c59;
            padding-bottom: 3px;
        }
        .explanation-content code {
            background: #f0f8f0;
            padding: 2px 6px;
            border-radius: 4px;
            font-family: 'Monaco', 'Menlo', monospace;
            color: #2d5016;
            border: 1px solid #a8e6cf;
        }
        .explanation-content pre {
            background: #f0f8f0;
            padding: 15px;
            border-radius: 8px;
            border-left: 4px solid #2d5016;
            overflow-x: auto;
        }
        .explanation-content pre code {
            background: none;
            padding: 0;
            border: none;
            color: #2d5016;
        }
        .explanation-content blockquote {
            border-left: 4px solid #4a7c59;
            margin: 15px 0;
            padding: 10px 20px;
            background: #f0f8f0;
            border-radius: 0 8px 8px 0;
        }
        .explanation-content ul,
        .explanation-content ol {
            padding-left: 20px;
        }
        .explanation-content li {
            margin: 5px 0;
        }
        .explanation-content strong {
            color: #2d5016;
        }
        .explanation-content em {
            color: #4a7c59;
        }
        .verbose-buttons {
            display: flex;
            gap: 10px;
            margin-top: 20px;
            flex-wrap: wrap;
            justify-content: center;
        }
        .verbose-btn {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            color: #495057;
            padding: 10px 15px;
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.3s ease;
            font-size: 14px;
        }
        .v-btn {
            background: #e3f2fd;
            border: 1px solid #bbdefb;
            color: #1976d2;
        }
        .v-btn:hover {
            background: #bbdefb;
            border-color: #90caf9;
        }
        .v-btn:disabled {
            background: #f5f5f5;
            border-color: #e0e0e0;
            color: #9e9e9e;
            cursor: not-allowed;
        }
        .vv-btn {
            background: #e1f5fe;
            border: 1px solid #b3e5fc;
            color: #0277bd;
        }
        .vv-btn:hover {
            background: #b3e5fc;
            border-color: #81d4fa;
        }
        .vv-btn:disabled {
            background: #f5f5f5;
            border-color: #e0e0e0;
            color: #9e9e9e;
            cursor: not-allowed;
        }
        .vvv-btn {
            background: #e8eaf6;
            border: 1px solid #c5cae9;
            color: #3f51b5;
        }
        .vvv-btn:hover {
            background: #c5cae9;
            border-color: #9fa8da;
        }
        .vvv-btn:disabled {
            background: #f5f5f5;
            border-color: #e0e0e0;
            color: #9e9e9e;
            cursor: not-allowed;
        }
        .action-buttons {
            display: flex;
            gap: 10px;
            margin-top: 20px;
            justify-content: center;
            flex-wrap: wrap;
        }
        .action-btn {
            background: #e8f5e8;
            border: 1px solid #c8e6c9;
            color: #2e7d32;
            padding: 10px 20px;
            border-radius: 6px;
            cursor: pointer;
            transition: all 0.3s ease;
            font-size: 14px;
            text-decoration: none;
            display: inline-block;
        }
        .action-btn:hover {
            background: #c8e6c9;
            border-color: #a5d6a7;
            color: #1b5e20;
        }
        .error-message {
            background: #f8d7da;
            color: #721c24;
            padding: 20px;
            border-radius: 8px;
            border-left: 4px solid #dc3545;
        }
        .error-message h3 {
            color: #721c24;
            margin-bottom: 10px;
        }
        .loading {
            display: none;
            text-align: center;
            padding: 20px;
        }
        .loading.show {
            display: block;
        }
        .spinner {
            border: 3px solid #f3f3f3;
            border-top: 3px solid #2d5016;
            border-radius: 50%;
            width: 30px;
            height: 30px;
            animation: spin 1s linear infinite;
            margin: 0 auto 10px;
        }
        .verbose-loading {
            display: none;
            text-align: center;
            padding: 10px;
            margin-top: 10px;
        }
        .verbose-loading.show {
            display: block;
        }
        .verbose-spinner {
            border: 2px solid #f3f3f3;
            border-top: 2px solid #1976d2;
            border-radius: 50%;
            width: 20px;
            height: 20px;
            animation: spin 1s linear infinite;
            margin: 0 auto 5px;
        }
        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
        .scroll-to-top {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: #3498db;
            color: white;
            border: none;
            border-radius: 50%;
            width: 50px;
            height: 50px;
            font-size: 20px;
            cursor: pointer;
            transition: all 0.3s ease;
            box-shadow: 0 4px 8px rgba(0,0,0,0.2);
            z-index: 1000;
        }
        .scroll-to-top:hover {
            background: #2980b9;
            transform: translateY(-2px);
            box-shadow: 0 6px 12px rgba(0,0,0,0.3);
        }
        
        /* Мобильная оптимизация */
        @media (max-width: 768px) {
            body {
                padding: 10px;
            }
            .container {
                margin: 0;
                border-radius: 8px;
                box-shadow: 0 10px 20px rgba(0,0,0,0.1);
            }
            .header {
                padding: 20px;
            }
            .header h1 {
                font-size: 2em;
            }
            .content {
                padding: 20px;
            }
            .nav-buttons {
                flex-direction: column;
                gap: 8px;
            }
            .nav-btn {
                width: 100%;
                text-align: center;
                padding: 12px 16px;
            }
            .form-buttons {
                flex-direction: column;
                gap: 10px;
            }
            .submit-btn, .reset-btn {
                width: 100%;
                padding: 16px 20px;
                font-size: 16px;
            }
            .verbose-buttons {
                flex-direction: column;
                gap: 8px;
            }
            .verbose-btn {
                width: 100%;
                padding: 12px 16px;
                font-size: 14px;
            }
            .action-buttons {
                flex-direction: column;
                gap: 8px;
            }
            .action-btn {
                width: 100%;
                padding: 12px 16px;
                font-size: 14px;
            }
            .command-result {
                padding: 15px;
                margin-bottom: 15px;
            }
            .command-code {
                font-size: 14px;
                padding: 12px;
                word-break: break-word;
            }
            .explanation-section {
                padding: 15px;
            }
            .result-meta {
                flex-direction: column;
                gap: 8px;
                font-size: 12px;
            }
            .scroll-to-top {
                bottom: 15px;
                right: 15px;
                width: 45px;
                height: 45px;
                font-size: 18px;
            }
        }
        
        /* Очень маленькие экраны */
        @media (max-width: 480px) {
            .header h1 {
                font-size: 1.8em;
            }
            .header p {
                font-size: 1em;
            }
            .content {
                padding: 15px;
            }
            .form-group textarea {
                min-height: 100px;
                font-size: 16px; /* Предотвращает зум на iOS */
            }
            .form-group select {
                font-size: 16px; /* Предотвращает зум на iOS */
            }
            .command-result h3 {
                font-size: 1.2em;
            }
            .explanation-content h1,
            .explanation-content h2,
            .explanation-content h3 {
                font-size: 1.3em;
            }
        }
`))