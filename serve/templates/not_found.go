package templates

// NotFoundTemplate современная страница 404
const NotFoundTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Страница не найдена — 404</title>
  <style>
    :root {
      --bg: #1a0b0b;           /* глубокий темно-красный фон */
      --bg2: #2a0f0f;          /* второй оттенок фона */
      --fg: #ffeaea;           /* светлый текст с красным оттенком */
      --accent: #ff3b30;       /* основной красный (iOS red) */
      --accent2: #ff6f61;      /* дополнительный коралловый */
      --accentGlow: rgba(255,59,48,0.35);
      --accentGlow2: rgba(255,111,97,0.30);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      display: grid;
      place-items: center;
      background:
        radial-gradient(1200px 600px at 10% 10%, rgba(255,59,48,0.12), transparent),
        radial-gradient(1200px 600px at 90% 90%, rgba(255,111,97,0.12), transparent),
        linear-gradient(135deg, var(--bg), var(--bg2));
      color: var(--fg);
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Inter, sans-serif;
      overflow: hidden;
    }
    .glow {
      position: absolute;
      inset: -20%;
      background:
        radial-gradient(700px 340px at 20% 30%, rgba(255,59,48,0.22), transparent 60%),
        radial-gradient(700px 340px at 80% 70%, rgba(255,111,97,0.20), transparent 60%);
      filter: blur(40px);
      z-index: 0;
    }
    .card {
      position: relative;
      z-index: 1;
      width: min(720px, 92vw);
      padding: 32px;
      border-radius: 20px;
      background: rgba(255,255,255,0.03);
      border: 1px solid rgba(255,255,255,0.08);
      box-shadow: 0 10px 40px rgba(80,0,0,0.45), inset 0 0 0 1px rgba(255,255,255,0.03);
      backdrop-filter: blur(10px);
      text-align: center;
    }
    @keyframes pulse {
      0%, 100% {
        transform: scale(1);
        text-shadow: 0 8px 40px var(--accentGlow);
      }
      50% {
        transform: scale(1.15);
        text-shadow: 0 12px 60px var(--accentGlow), 0 0 30px var(--accentGlow2);
      }
    }
    .code {
      font-size: clamp(48px, 12vw, 120px);
      line-height: 0.9;
      font-weight: 800;
      letter-spacing: -2px;
      background: linear-gradient(90deg, var(--accent), var(--accent2));
      -webkit-background-clip: text;
      background-clip: text;
      color: transparent;
      margin: 8px 0 12px 0;
      text-shadow: 0 8px 40px var(--accentGlow);
      animation: pulse 2.5s ease-in-out infinite;
      transform-origin: center;
    }
    .title {
      font-size: clamp(18px, 3.2vw, 28px);
      font-weight: 600;
      opacity: 0.95;
      margin-bottom: 8px;
    }
    .desc {
      font-size: 15px;
      opacity: 0.75;
      margin: 0 auto 20px auto;
      max-width: 60ch;
    }
    .btns {
      display: flex;
      flex-wrap: wrap;
      gap: 12px;
      justify-content: center;
      margin-top: 8px;
    }
    .btn {
      appearance: none;
      border: none;
      cursor: pointer;
      padding: 12px 18px;
      border-radius: 12px;
      font-weight: 600;
      color: white;
      background: linear-gradient(135deg, var(--accent), #c62828);
      box-shadow: 0 6px 18px var(--accentGlow);
      transition: transform .2s ease, box-shadow .2s ease, filter .2s ease;
      text-decoration: none;
      display: inline-flex;
      align-items: center;
      gap: 8px;
    }
    .btn:hover { transform: translateY(-2px); filter: brightness(1.05); }
    .btn.secondary { background: linear-gradient(135deg, #e65100, var(--accent2)); box-shadow: 0 6px 18px var(--accentGlow2); }
    .hint { margin-top: 16px; font-size: 13px; opacity: 0.6; }
  </style>
  <script>
    function goHome() {
      window.location.href = '{{.BasePath}}/';
    }
    function bindEsc() {
      const handler = (e) => { if (e.key === 'Escape' || e.key === 'Esc') { e.preventDefault(); goHome(); } };
      window.addEventListener('keydown', handler);
      document.addEventListener('keydown', handler);
      // фокус на body для гарантии получения клавиш
      if (document && document.body) {
        document.body.setAttribute('tabindex', '-1');
        document.body.focus({ preventScroll: true });
      }
    }
    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', bindEsc);
    } else {
      bindEsc();
    }
  </script>
  <meta http-equiv="refresh" content="30">
  <meta name="robots" content="noindex">
  <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'><text y='50%' x='50%' dominant-baseline='middle' text-anchor='middle' font-size='42'>🚫</text></svg>">
</head>
<body>
  <div class="glow"></div>
  <div class="card">
    <div class="code">404</div>
    <div class="title">Страница не найдена</div>
    <p class="desc">Такой страницы не существует. Вы можете вернуться на главную страницу или выполнить команду.</p>
    <div class="btns">
      <a class="btn" href="{{.BasePath}}/">🏠 На главную</a>
      <a class="btn secondary" href="{{.BasePath}}/run">🚀 К выполнению</a>
    </div>
    <div class="hint">Нажмите Esc, чтобы вернуться на главную</div>
  </div>
</body>
</html>
`
