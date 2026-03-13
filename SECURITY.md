# Security Policy

AnchorChain builds on the Factom Protocol and takes security seriously. Please follow the guidelines below when reporting potential vulnerabilities.

## Supported Versions
- `main` (development) — actively supported
- Tagged releases — supported until superseded by the next minor release

## Reporting a Vulnerability
1. **Do not open a public issue**. Instead, email security@anchorchain.dev (or the maintainer listed in the repository) with:
   - A detailed description of the issue
   - Steps to reproduce
   - Impact assessment / severity
   - Suggested mitigation if known
2. Encrypt messages using PGP if possible.
3. We aim to acknowledge reports within 3 business days.
4. Once confirmed, we will coordinate a disclosure timeline with you and credit reporters who wish to be acknowledged.

## Scope
- `anchorchaind`, `anchor-cli`, and the public HTTP API facade
- Devnet tooling and config shipped in this repository
- Documentation errors that could lead to insecure defaults

Issues in the vendored upstream Factom protocol should additionally be reported to the original project when appropriate.

Thank you for keeping AnchorChain safe.
