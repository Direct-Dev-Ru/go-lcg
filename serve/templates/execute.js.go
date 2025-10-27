package templates

import "html/template"

var ExecutePageScriptsTemplate = template.Must(template.New("execute_scripts").Parse(`
<script>
    // Обработка отправки формы (блокируем все кнопки кроме навигации)
    document.getElementById('executeForm').addEventListener('submit', function(e) {
        // Предотвращаем множественные отправки на мобильных устройствах
        if (this.dataset.submitting === 'true') {
            e.preventDefault();
            return false;
        }
        
        // Валидация длины полей
        const prompt = document.getElementById('prompt').value;
        const maxUserMessageLength = {{.MaxUserMessageLength}};
        if (prompt.length > maxUserMessageLength) {
            alert('Пользовательское сообщение слишком длинное: максимум ' + maxUserMessageLength + ' символов');
            e.preventDefault();
            return false;
        }
        this.dataset.submitting = 'true';
        const submitBtn = document.getElementById('submitBtn');
        const loading = document.getElementById('loading');
        const actionButtons = document.querySelectorAll('.action-btn');
        const verboseButtons = document.querySelectorAll('.verbose-btn');
        const scrollBtn = document.getElementById('scrollToTop');
        
        submitBtn.disabled = true;
        submitBtn.textContent = '⏳ Выполняется...';
        loading.classList.add('show');
        
        // Блокируем кнопки действий (сохранение/история)
        actionButtons.forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.5';
        });
        
        // Блокируем кнопки подробностей (v/vv/vvv)
        verboseButtons.forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.5';
        });
        
        // Прячем кнопку "Наверх" до получения нового ответа
        if (scrollBtn) {
            scrollBtn.style.display = 'none';
        }
    });
    
    // Запрос подробного объяснения
    function requestExplanation(verbose) {
        const form = document.getElementById('executeForm');
        const prompt = document.getElementById('prompt').value;
        const systemId = document.getElementById('system_id').value;
        const verboseLoading = document.getElementById('verboseLoading');
        const verboseButtons = document.querySelectorAll('.verbose-btn');
        const actionButtons = document.querySelectorAll('.action-btn');
        
        if (!prompt.trim()) {
            alert('Сначала выполните основной запрос');
            return;
        }
        
        // Показываем лоадер и блокируем ВСЕ кнопки кроме навигации
        verboseLoading.classList.add('show');
        verboseButtons.forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.5';
        });
        actionButtons.forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.5';
        });
        
        // Создаем скрытое поле для verbose
        const verboseInput = document.createElement('input');
        verboseInput.type = 'hidden';
        verboseInput.name = 'verbose';
        verboseInput.value = verbose;
        form.appendChild(verboseInput);
        
        // Отправляем форму
        form.submit();
    }
    
    // Сохранение результата
    function saveResult() {
        const resultDataField = document.getElementById('resultData');
        const prompt = document.getElementById('prompt').value;
        const csrfToken = document.querySelector('input[name="csrf_token"]').value;
        
        if (!resultDataField.value || !prompt.trim()) {
            alert('Нет данных для сохранения');
            return;
        }
        
        try {
            const resultData = JSON.parse(resultDataField.value);
            const requestData = {
                prompt: prompt,
                command: resultData.command,
                explanation: resultData.explanation || '',
                model: resultData.model || 'Unknown'
            };
            
            fetch('{{.BasePath}}/api/save-result', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken,
                },
                body: JSON.stringify(requestData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('✅ Результат сохранен: ' + data.file);
                } else {
                    alert('❌ Ошибка сохранения: ' + data.error);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('❌ Ошибка при сохранении результата');
            });
        } catch (error) {
            console.error('Error parsing result data:', error);
            alert('❌ Ошибка при чтении данных результата');
        }
    }
    
    // Добавление в историю
    function addToHistory() {
        const resultDataField = document.getElementById('resultData');
        const prompt = document.getElementById('prompt').value;
        const systemId = document.getElementById('system_id').value;
        const csrfToken = document.querySelector('input[name="csrf_token"]').value;
        
        if (!resultDataField.value || !prompt.trim()) {
            alert('Нет данных для сохранения в историю');
            return;
        }
        
        try {
            const resultData = JSON.parse(resultDataField.value);
            const systemName = document.querySelector('option[value="' + systemId + '"]')?.textContent || 'Unknown';
            
            const requestData = {
                prompt: prompt,
                command: resultData.command,
                response: resultData.command,
                explanation: resultData.explanation || '',
                system: systemName
            };
            
            fetch('{{.BasePath}}/api/add-to-history', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken,
                },
                body: JSON.stringify(requestData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('✅ ' + data.message);
                } else {
                    alert('❌ Ошибка: ' + data.error);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('❌ Ошибка при добавлении в историю');
            });
        } catch (error) {
            console.error('Error parsing result data:', error);
            alert('❌ Ошибка при чтении данных результата');
        }
    }
    
    // Функция прокрутки наверх
    function scrollToTop() {
        window.scrollTo({
            top: 0,
            behavior: 'smooth'
        });
    }
    
    // Показываем кнопку "Наверх" при появлении объяснения
    function showScrollToTopButton() {
        const scrollBtn = document.getElementById('scrollToTop');
        if (scrollBtn) {
            scrollBtn.style.display = 'block';
        }
    }
    
    // Скрываем кнопку "Наверх" при загрузке страницы
    function hideScrollToTopButton() {
        const scrollBtn = document.getElementById('scrollToTop');
        if (scrollBtn) {
            scrollBtn.style.display = 'none';
        }
    }
    
    // Сброс формы к начальному состоянию
    function resetForm() {
        // Очищаем поля формы
        document.getElementById('prompt').value = '';
        document.getElementById('resultData').value = '';
        
        // Сбрасываем флаг отправки формы
        const form = document.getElementById('executeForm');
        if (form) {
            form.dataset.submitting = 'false';
        }
        
        // Скрываем все секции результатов
        const resultSection = document.querySelector('.result-section');
        const verboseButtons = document.querySelector('.verbose-buttons');
        const actionButtons = document.querySelector('.action-buttons');
        const explanationSection = document.querySelector('.explanation-section');
        const loading = document.getElementById('loading');
        const verboseLoading = document.getElementById('verboseLoading');
        const scrollBtn = document.getElementById('scrollToTop');
        
        if (resultSection) resultSection.style.display = 'none';
        if (verboseButtons) verboseButtons.style.display = 'none';
        if (actionButtons) actionButtons.style.display = 'none';
        if (explanationSection) explanationSection.style.display = 'none';
        if (loading) loading.classList.remove('show');
        if (verboseLoading) verboseLoading.classList.remove('show');
        if (scrollBtn) scrollBtn.style.display = 'none';
        
        // Разблокируем кнопки
        const submitBtn = document.getElementById('submitBtn');
        const resetBtn = document.getElementById('resetBtn');
        const allButtons = document.querySelectorAll('.action-btn, .verbose-btn');
        
        if (submitBtn) {
            submitBtn.disabled = false;
            submitBtn.textContent = '🚀 Выполнить запрос';
        }
        if (resetBtn) resetBtn.disabled = false;
        
        allButtons.forEach(btn => {
            btn.disabled = false;
            btn.style.opacity = '1';
        });
        
        // Прокручиваем к началу формы (с проверкой поддержки smooth)
        const formSection = document.querySelector('.form-section');
        if (formSection) {
            if ('scrollBehavior' in document.documentElement.style) {
                formSection.scrollIntoView({ behavior: 'smooth' });
            } else {
                formSection.scrollIntoView();
            }
        }
    }
    
    // Сохранение результатов в скрытое поле
    function saveResultToHiddenField() {
        const resultDataField = document.getElementById('resultData');
        const commandElement = document.querySelector('.command-code, .command-md');
        const explanationElement = document.querySelector('.explanation-content');
        const modelElement = document.querySelector('.result-meta span:first-child');
        
        if (commandElement) {
            const command = commandElement.textContent.trim();
            const explanation = explanationElement ? explanationElement.innerHTML.trim() : '';
            const model = modelElement ? modelElement.textContent.replace('Модель: ', '') : 'Unknown';
            
            const resultData = {
                command: command,
                explanation: explanation,
                model: model
            };
            
            resultDataField.value = JSON.stringify(resultData);
        }
    }
    
    // Показываем кнопку при появлении объяснения и сохраняем результаты
    document.addEventListener('DOMContentLoaded', function() {
        const explanationSection = document.querySelector('.explanation-section');
        if (explanationSection) {
            showScrollToTopButton();
        }
        
        // Сохраняем результаты в скрытое поле при загрузке страницы
        saveResultToHiddenField();
    });
</script>
`))