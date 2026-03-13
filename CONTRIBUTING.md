# Contributing to AnchorChain

Thanks for helping improve AnchorChain! This project builds on the open-source Factom Protocol and welcomes community contributions that make it easier to run, extend, and integrate the network.

## Getting Started
1. Fork the repository and clone your fork locally.
2. Install Go 1.22 or newer.
3. Run `make build` to compile `anchorchaind` and `anchor-cli`.
4. Launch the devnet (`./bin/anchorchaind devnet`) and exercise `anchor-cli` to verify your environment.

## Pull Request Guidelines
- Keep changes scoped and well-described. Avoid touching the vendored `node/` Factom sources unless absolutely necessary.
- Run `make build` plus any relevant tests before submitting.
- Include documentation updates (README/docs) when you add or change user-facing behavior.
- Note any breaking changes clearly in the PR description.

## Code Style
- Go code should follow the standard `gofmt`/`goimports` formatting.
- Favor small, composable packages and clear error messages.
- Public APIs should include comments that explain intent and usage.

## Reporting Issues
- Use GitHub issues for bugs, doc requests, and feature proposals.
- When reporting a bug, include reproduction steps, CLI output, and environment info (OS, Go version, commit hash).

## Community Expectations
- Follow the [Code of Conduct](CODE_OF_CONDUCT.md).
- Be respectful and constructive in reviews and discussions.

Thank you for helping keep AnchorChain reliable, transparent, and welcoming!
