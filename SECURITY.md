# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| Latest  | Yes       |
| Others  | No        |

Always run the latest version. We do not backport fixes.

---

## Reporting a vulnerability

Do not open a public GitHub issue for security vulnerabilities.

Use [GitHub Private Vulnerability Reporting](https://github.com/M-Hammad-Faisal/envguard/security/advisories/new).

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Your suggested fix (optional)

You will receive a response within 72 hours. If the vulnerability is confirmed, a patch will be released and you will be credited in the release notes unless you prefer otherwise.

---

## Known limitations

These are documented design constraints, not vulnerabilities:

- **Shared passphrase model:** EnvGuard uses symmetric encryption. Offboarding a team member requires manual key rotation. This is a known MVP constraint.
- **Verbatim scan only:** The pre-commit scanner catches exact string matches. It will not catch secrets in template literals, string concatenations, or base64-encoded values.
- **Partial staging:** Auto-fix re-stages the entire file, not just the changed lines.