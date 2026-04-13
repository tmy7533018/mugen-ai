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

## Usage

Start the server (used by mugen-shell):

```sh
mugen-ai serve --model gemma3:4b
```

Chat in the terminal:

```sh
mugen-ai chat --model gemma3:4b
```

### Options

```
--model        Ollama model to use (default: gemma3:4b)
--port         Server port (default: 11435)
--ollama-host  Ollama host URL (default: http://localhost:11434)
--system       System prompt
```

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/chat` | Send a message, receive SSE stream |
| DELETE | `/history` | Clear conversation history |
| GET | `/health` | Health check |
| GET | `/models` | List available Ollama models |
