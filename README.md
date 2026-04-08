# EnvGuard

A local-first CLI tool that encrypts your `.env` file for team sharing and prevents hardcoded secrets from reaching your Git history.

**100% offline. No telemetry. No external requests. Single binary.**

---

## Install

**Using Go:**

```bash
go install github.com/m-hammad-faisal/envguard@latest
```

**Download binary (no Go required):**
Download from [Releases](https://github.com/m-hammad-faisal/envguard/releases/latest)

**Or build from source:**

```bash
git clone https://github.com/m-hammad-faisal/envguard
cd envguard
go build -o envguard .
```

---

## Quickstart

```bash
# 1. Initialize in your project root
envguard init

# 2. Fill in your .env values, then encrypt and share with the team
envguard push

# 3. Any team member can pull the secrets (requires the shared passphrase)
envguard pull
```

The pre-commit hook runs `envguard scan` automatically on every `git commit`.

---

## Commands

### `envguard init`

- Creates `.envguard/` directory
- Installs or appends to `.git/hooks/pre-commit` (safe with Husky/Lefthook)
- Detects your framework (Node.js, Go, Python) and writes the correct env var reference syntax to `.envguard/config.json`
- Generates a starter `.env` with commented keys based on detected dependencies
- Adds `.env` to `.gitignore`

### `envguard push`

- Reads your local `.env`
- Prompts for the team passphrase (hidden input)
- Encrypts with AES-256-GCM + Argon2id key derivation
- Writes `.envguard/secrets.enc`
- Commit `secrets.enc` + `config.json` to share with the team

### `envguard pull`

- Reads `.envguard/secrets.enc` from the repo
- Prompts for the team passphrase
- Decrypts and writes your local `.env`
- Installs the pre-commit hook if not present (useful after fresh clones)

### `envguard scan`

- Called automatically by the pre-commit hook
- Parses your local `.env` to build a value → key reverse map
- Scans all staged files for exact secret value matches
- If a match is found: shows a colored diff preview and offers to auto-replace the hardcoded value with the correct env var reference (e.g., `process.env.STRIPE_KEY`)
- Aborts the commit if secrets are found and you decline the fix

---

## Security model

- **Encryption:** AES-256-GCM
- **Key derivation:** Argon2id (OWASP recommended parameters: time=1, memory=64MB, threads=4)
- **Binary format:** `[version][salt][argon2 params][nonce][ciphertext]` — params are stored with the ciphertext so future parameter changes never break existing files
- **Zero network access:** everything happens locally

### Known limitations

**Key rotation:** This tool uses symmetric encryption with a shared passphrase. When a team member leaves, you must:

1. Generate a new passphrase
2. Run `envguard push` with the new passphrase
3. Share the new passphrase out-of-band with remaining team members
4. Each member runs `envguard pull` with the new passphrase

**False negatives:** `envguard scan` uses exact string matching. It will **not** catch secrets that are:

- Dynamically interpolated: `` `prefix_${secret}` ``
- Split across concatenation: `"sk_live" + "_abc123"`
- Base64 encoded
- Stored in environment variables at scan time (only `.env` file values are checked)

This is intentional — the alternative (regex/AI-based heuristics) introduces false positives that break CI and erode developer trust in the tool.

---

## Supported frameworks

| Framework | Detected by                                     | Replacement template    |
| --------- | ----------------------------------------------- | ----------------------- |
| Next.js   | `package.json` containing `"next"`              | `process.env.{{KEY}}`   |
| Node.js   | `package.json`                                  | `process.env.{{KEY}}`   |
| Go        | `go.mod`                                        | `os.Getenv("{{KEY}}")`  |
| Python    | `requirements.txt`, `Pipfile`, `pyproject.toml` | `os.environ["{{KEY}}"]` |
| Other     | Prompts you at `init`                           | Custom                  |

---

## Running tests

```bash
go test ./...
```

---

## Roadmap

- **Phase 2:** Local web UI (`localhost:8080`) for visualizing active env vars and branch status
- **Phase 3:** Enterprise cloud sync — hosted KMS, asymmetric keypairs per user, RBAC, audit logs

---

## Contributing

PRs welcome. Please run `go test ./...` before submitting.
