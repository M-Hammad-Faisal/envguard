# Changelog

All notable changes to EnvGuard will be documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Versions follow [Semantic Versioning](https://semver.org/).

---

## [0.1.6] - 2026-04-08

### Fixed

- Add `contents: write` permission to release workflow
- Remove unused `promptPassphraseViaTTY` function
- Check all error return values in tests (errcheck linter)

## [0.1.5] - 2026-04-08

### Added

- GoReleaser with pre-built binaries for Linux, macOS, Windows (amd64 + arm64)
- golangci-lint CI workflow
- PR checks workflow
- Issue templates, PR template, CODEOWNERS, CONTRIBUTING, SECURITY

## [0.1.0] - 2026-04-08

### Added

- Initial release
- `envguard init` with framework detection (Node.js, Go, Python)
- `envguard push` — AES-256-GCM encryption with Argon2id key derivation
- `envguard pull` — decryption and local .env hydration
- `envguard scan` — pre-commit hook with colored diff and auto-fix
- Versioned binary format for forward compatibility
- Binary file detection to prevent scan crashes
- Full test suite across crypto, envparse, and cmd packages
