# Linux Command GPT (lcg)

This repo is forked from <https://github.com/asrul10/linux-command-gpt.git>

Generate Linux commands from natural language. Supports Ollama and Proxy backends, system prompts, different explanation levels (v/vv/vvv), and JSON history.

## Installation

Build from source:

```bash
git clone --depth 1 https://github.com/Direct-Dev-Ru/linux-command-gpt.git ~/.linux-command-gpt
cd ~/.linux-command-gpt
go build -o lcg

# Add to your PATH
ln -s ~/.linux-command-gpt/lcg ~/.local/bin
```

## Quick start

```bash
lcg "I want to extract linux-command-gpt.tar.gz file"
```

After generation you will see a CAPS warning that the answer is from AI and must be verified, the command, and the action menu:

```text
ACTIONS: (c)opy, (s)ave, (r)egenerate, (e)xecute, (v|vv|vvv)explain, (n)othing
```

Explanations:

- `v` — short; `vv` — medium; `vvv` — detailed with alternatives.

Clipboard support requires `xclip` or `xsel`.

## Environment

- `LCG_PROVIDER` (default `ollama`) — provider type: `ollama` or `proxy`
- `LCG_HOST` (default `http://192.168.87.108:11434/`) — base API URL
- `LCG_MODEL` (default `hf.co/yandex/YandexGPT-5-Lite-8B-instruct-GGUF:Q4_K_M`)
- `LCG_PROMPT` — default system prompt content
- `LCG_PROXY_URL` (default `/api/v1/protected/sberchat/chat`) — proxy chat endpoint
- `LCG_COMPLETIONS_PATH` (default `api/chat`) — Ollama chat endpoint (relative)
- `LCG_TIMEOUT` (default `300`) — request timeout in seconds
- `LCG_RESULT_FOLDER` (default `~/.config/lcg/gpt_results`) — folder for saved results
- `LCG_RESULT_HISTORY` (default `$(LCG_RESULT_FOLDER)/lcg_history.json`) — JSON history path
- `LCG_PROMPT_FOLDER` (default `~/.config/lcg/gpt_sys_prompts`) — folder for system prompts
- `LCG_PROMPT_ID` (default `1`) — default system prompt ID
- `LCG_JWT_TOKEN` — JWT token for proxy provider
- `LCG_NO_HISTORY` — if `1`/`true`, disables history writes for the process
- `LCG_SERVER_PORT` (default `8080`), `LCG_SERVER_HOST` (default `localhost`) — HTTP server settings

## Flags

- `--file, -f` read part of prompt from file
- `--sys, -s` system prompt content or ID
- `--prompt-id, --pid` choose built-in prompt (1–5)
- `--timeout, -t` request timeout (sec)
- `--no-history, --nh` disable writing/updating JSON history for this run
- `--debug, -d` show debug information (request parameters and prompts)
- `--version, -v` print version; `--help, -h` help

## Commands

- `models`, `health`, `config`
- `prompts list|add|delete`
- `test-prompt <prompt-id> <command>`
- `update-jwt`, `delete-jwt` (proxy)
- `update-key`, `delete-key` (not needed for ollama/proxy)
- `history list` — list history from JSON
- `history view <index>` — view by index
- `history delete <index>` — delete by index (re-numbering)
- `serve-result` — start HTTP server to browse saved results (`--port`, `--host`)

## Saving results

Files are saved to `LCG_RESULT_FOLDER` (default `~/.config/lcg/gpt_results`).

- Command result: `gpt_request_<MODEL>_YYYY-MM-DD_HH-MM-SS.md`
  - `# <title>` — H1 with original request (trimmed to 120 chars: first 116 + `...`)
  - `## Prompt`
  - `## Response`

- Detailed explanation: `gpt_explanation_<MODEL>_YYYY-MM-DD_HH-MM-SS.md`
  - `# <title>`
  - `## Prompt`
  - `## Command`
  - `## Explanation and Alternatives (model: <MODEL>)`

## History

- Stored as JSON array in `LCG_RESULT_HISTORY`.
- On new request, if the same command exists, you will be prompted to view or overwrite.
- Showing from history does not call the API; the standard action menu is shown.

For full guide in Russian, see `USAGE_GUIDE.md`.
