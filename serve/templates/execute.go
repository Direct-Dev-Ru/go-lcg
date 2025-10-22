package templates

import "html/template"

// ExecutePageTemplate - шаблон страницы выполнения запросов
var ExecutePageTemplate = template.Must(template.New("execute").Parse(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Linux Command GPT</title>
    <style>
        {{template "execute_css" .}}        
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Header}}</h1>
            <p>Выполнение запросов к Linux Command GPT через веб-интерфейс</p>
        </div>
        <div class="content">
            <div class="nav-buttons">
                <a href="/" class="nav-btn">🏠 Главная</a>
                <a href="/history" class="nav-btn">📝 История</a>
                <a href="/prompts" class="nav-btn">⚙️ Промпты</a>
            </div>
            
            <form method="POST" id="executeForm">
                <div class="form-section">
                    <div class="form-group">
                        <label for="system_id">🤖 Системный промпт:</label>
                        <select name="system_id" id="system_id" required>
                            {{range .SystemOptions}}
                            <option value="{{.ID}}">{{.ID}}. {{.Name}}</option>
                            {{end}}
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="prompt">💬 Ваш запрос:</label>
                        <textarea name="prompt" id="prompt" placeholder="Опишите, что вы хотите сделать..." required>{{.CurrentPrompt}}</textarea>
                    </div>
                    
                    <!-- Скрытое поле для хранения результатов -->
                    <input type="hidden" id="resultData" name="resultData" value="">
                    
                    <div class="form-buttons">
                        <button type="submit" class="submit-btn" id="submitBtn">
                            🚀 Выполнить запрос
                        </button>
                        <button type="button" class="reset-btn" id="resetBtn" onclick="resetForm()">
                            🔄 Сброс
                        </button>
                    </div>
                </div>
            </form>
            
            <div class="loading" id="loading">
                <div class="spinner"></div>
                <p>Обрабатываю запрос...</p>
            </div>
            
            {{.ResultSection}}
            
            {{.VerboseButtons}}
            
            <div class="verbose-loading" id="verboseLoading">
                <div class="verbose-spinner"></div>
                <p>Получаю подробное объяснение...</p>
            </div>
            
            {{.ActionButtons}}
        </div>
    </div>
    
    <!-- Кнопка "Наверх" -->
    <button class="scroll-to-top" id="scrollToTop" onclick="scrollToTop()" style="display: none;">↑</button>
    
    {{template "execute_scripts" .}}
</body>
</html>`))

// Объединяем шаблоны
func init() {
    template.Must(ExecutePageTemplate.AddParseTree("execute_css", ExecutePageCSSTemplate.Tree))
    template.Must(ExecutePageTemplate.AddParseTree("execute_scripts", ExecutePageScriptsTemplate.Tree))
}
