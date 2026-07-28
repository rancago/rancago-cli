# Contributing to Rancago CLI

Thanks for your interest in contributing to **Rancago CLI**. Contributions are welcome via issues and pull requests.

## Quick Rules

- Keep changes focused and small when possible.
- Prefer improvements that keep the CLI commands predictable and user-friendly.
- Avoid introducing new dependencies unless there is a clear benefit.
- Do not commit secrets (keys, tokens, credentials). Use env vars and `.env.example` patterns instead.

## Development Setup

```bash
git clone https://github.com/rancago/rancago-cli.git
cd rancago-cli
go mod tidy
```

## Suggested Workflow

1. Create a branch from `main`
2. Make your changes
3. Run formatting and tests
4. Open a pull request

## Testing

```bash
go test ./...
```

## Pull Request Checklist

- Code builds and tests pass locally
- CLI output remains consistent (help text, flags, exit codes)
- No generated binaries or large artifacts committed
