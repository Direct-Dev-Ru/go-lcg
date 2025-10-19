# Linux Command GPT (lcg)

This repo is forked from <https://github.com/asrul10/linux-command-gpt.git>

Generate Linux commands from natural language. Supports Ollama and Proxy backends, system prompts, different explanation levels (v/vv/vvv), and JSON history.

## Installation

Build from source:

```bash
git clone --depth 1 https://github.com/Direct-Dev-Ru/go-lcg.git ~/.linux-command-gpt
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

- `LCG_PROVIDER` (ollama|proxy), `LCG_HOST`, `LCG_MODEL`, `LCG_PROMPT`
- `LCG_TIMEOUT` (default 120), `LCG_RESULT_FOLDER` (default ./gpt_results)
- `LCG_RESULT_HISTORY` (default $(LCG_RESULT_FOLDER)/lcg_history.json)
- `LCG_JWT_TOKEN` (for proxy)

## Flags

- `--file, -f` read part of prompt from file
- `--sys, -s` system prompt content or ID
- `--prompt-id, --pid` choose built-in prompt (1–5)
- `--timeout, -t` request timeout (sec)
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

## Saving results

Files are saved to `LCG_RESULT_FOLDER`.

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
