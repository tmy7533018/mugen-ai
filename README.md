# mugen-ai

Local AI assistant for [mugen-shell](https://github.com/tmy7533018/mugen-shell), powered by [Ollama](https://ollama.com).

## Requirements

- [Ollama](https://ollama.com) installed and running
- A model pulled, e.g. `ollama pull gemma3:4b`
- [mugen-shell](https://github.com/tmy7533018/mugen-shell) (for desktop integration)

## Install

```sh
go install github.com/tmy7533018/mugen-ai@latest
```

### Autostart with systemd (user service)

A user service unit is provided in [`contrib/systemd/mugen-ai.service`](contrib/systemd/mugen-ai.service):

```sh
mkdir -p ~/.config/systemd/user
cp contrib/systemd/mugen-ai.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable --now mugen-ai.service
```

The unit assumes the binary is at `~/go/bin/mugen-ai` (default for `go install`).

## Configuration

On first run, a config file is created at `~/.config/mugen-ai/config.toml`:

```toml
[personality]
system_prompt = "You are a helpful desktop assistant. Be concise."

[context]
locale = "en"
city = ""
```

- **`system_prompt`** — customize the AI's personality
- **`locale`** — `"en"` or `"ja"` for date formatting
- **`city`** — set a city name (e.g. `"Tokyo"`) to inject live weather via [wttr.in](https://wttr.in). Leave empty to disable.

Date and time are always injected into the system prompt. CLI flags (`--model`, `--system`) override config values.

## Usage

Start the server (used by mugen-shell):

```sh
mugen-ai serve
```

Chat in the terminal:

```sh
mugen-ai chat
```

### Options

```
--model        Ollama model to use (default: gemma3:4b, overrides config)
--port         Server port (default: 11435)
--ollama-host  Ollama host URL (default: http://localhost:11434)
--system       System prompt (overrides config)
```

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/chat` | Send a message, receive SSE stream |
| DELETE | `/history` | Clear conversation history |
| GET | `/health` | Health check (also verifies Ollama connectivity) |
| GET | `/models` | List available Ollama models |
| PUT | `/model` | Switch the active model (`{"model": "name"}`) |
