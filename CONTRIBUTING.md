# Contributing to EnvGuard

Thanks for your interest. This document covers everything you need to contribute effectively.

---

## Before you start

- Check [existing issues](https://github.com/M-Hammad-Faisal/envguard/issues) before opening a new one.
- For large changes, open an issue first to discuss the approach. Don't spend days on a PR that won't be merged.
- For small fixes (typos, docs, obvious bugs), just open a PR directly.

---

## Setup

```bash
git clone https://github.com/M-Hammad-Faisal/envguard
cd envguard
go mod tidy
go test ./...
```

Requirements:

- Go 1.21+
- golangci-lint (`brew install golangci-lint` or see [install guide](https://golangci-lint.run/usage/install/))

---

## Development workflow

```bash
# Make your changes
# Run tests
go test ./...

# Run linter
golangci-lint run ./...

# Format code
gofmt -w .

# Build and test manually
go build -o envguard .
./envguard --help
```

All three must pass before submitting a PR.

---

## Project structure

```text
envguard/
├── main.go                      # Entry point
├── cmd/
│   ├── root.go                  # Cobra root command
│   ├── init.go                  # envguard init
│   ├── push.go                  # envguard push
│   ├── pull.go                  # envguard pull
│   ├── scan.go                  # envguard scan (git hook)
│   ├── tty_unix.go              # TTY handling for Unix/macOS
│   └── tty_windows.go           # TTY handling for Windows
├── crypto/
│   └── crypto.go                # AES-256-GCM + Argon2id
└── internal/
└── envparse/
└── parse.go             # .env file parser
```

---

## What we want

- Bug fixes with a failing test that proves the bug
- Performance improvements with a benchmark
- New framework detection rules in `detectFramework()`
- Better error messages
- Windows compatibility improvements

## What we do not want

- External network requests — Core Rule 1 is non-negotiable
- New dependencies without prior discussion
- AI-generated code submitted without review or understanding
- Breaking changes to the binary format of `secrets.enc` without a migration path

---

## Commit messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat: add Ruby on Rails framework detection
fix: handle empty .env file in scan command
docs: update README install instructions
test: add edge case for binary file detection
chore: bump golang.org/x/crypto to 0.45.0
```

---

## Security vulnerabilities

Do not open a public issue for security vulnerabilities.
Email directly or use [GitHub Private Vulnerability Reporting](https://github.com/M-Hammad-Faisal/envguard/security/advisories/new).

---

## License

By contributing, you agree your contributions will be licensed under the [MIT License](LICENSE).
