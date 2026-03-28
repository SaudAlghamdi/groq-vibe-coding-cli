# groqcode

AI-powered coding agent in your terminal, powered by Groq's fast LLM inference.

## Quick Start

```bash
# Set your API key
export GROQ_API_KEY="gsk_..."

# Or run init to create a config file
groqcode init

# Start interactive chat
groqcode

# One-shot question
groqcode ask "what does this project do?"

# List available models
groqcode models
```

## Installation

```bash
go install github.com/saudalghamdi/groq-vibe-coding-cli@latest
```

Or build from source:

```bash
git clone https://github.com/saudalghamdi/groq-vibe-coding-cli.git
cd groq-vibe-coding-cli
go build -o groqcode .
```

## Features

- **Interactive TUI** — Chat with an AI coding agent in your terminal
- **Tool Use** — The agent can read, write, edit files, run commands, and search your codebase
- **Streaming** — Responses stream in real-time via SSE
- **Safety** — Write/edit/command operations require approval (`y/n/always`)
- **Fast** — Powered by Groq's inference engine for low-latency responses

## Commands

| Command | Description |
|---------|-------------|
| `groqcode` | Launch interactive chat agent |
| `groqcode ask "question"` | One-shot question, print answer, exit |
| `groqcode init` | Create config file with API key prompt |
| `groqcode models` | List available Groq models |

## Interactive Commands

| Command | Description |
|---------|-------------|
| `/clear` | Clear conversation history |
| `/model` | Show current model |
| `/model <name>` | Switch to a different model |
| `/help` | Show help |
| `/exit` | Exit groqcode |

## Agent Tools

The AI agent has access to these tools:

| Tool | Description | Requires Approval |
|------|-------------|-------------------|
| ReadFile | Read file contents with line numbers | No |
| WriteFile | Write/create files | Yes |
| EditFile | String replacement edits | Yes |
| RunCommand | Execute shell commands (30s timeout) | Yes |
| SearchFiles | Recursive regex search | No |
| ListDirectory | List directory contents | No |
| GetFileInfo | File metadata and language detection | No |

## Configuration

Config file: `~/.groqcode/config.yaml`

```yaml
api_key: "gsk_..."
model: "llama-3.3-70b-versatile"
max_iterations: 25
auto_approve_reads: true
custom_system_prompt: ""
theme: "dark"
```

Environment variable `GROQ_API_KEY` overrides the config file API key.

## Flags

| Flag | Description |
|------|-------------|
| `--model` | Override the model for this session |
| `--yolo` | Auto-approve all tool operations (use with caution!) |

## License

MIT
