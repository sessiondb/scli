# Contributing Guidelines

Thank you for your interest in contributing to SessionDB CLI (scli)! We greatly value feedback and contributions from our community. This document will guide you through the contribution process.

## How can I contribute?

### Finding Issues to Work On

- Check our existing [open issues](https://github.com/sessiondb/scli/issues)
- Look for **good first issue** labels to start with
- Review recently closed issues to avoid duplicates

### Types of Contributions

- **Report Bugs:** Use our Bug Report template (or open an issue with clear reproduction steps)
- **Request Features:** Submit using a Feature Request template or issue with use case and requirements (e.g. new commands, install/deploy improvements)
- **Improve Documentation:** Create an issue with a documentation label; update README or help text for commands
- **Report Performance Issues:** Describe the scenario, environment, and impact
- **Report Security Issues:** Follow our [Security Policy](SECURITY.md)
- **Join Discussions:** Participate in project discussions and issue threads

### Creating Helpful Issues

When creating issues, include:

**For Feature Requests:**

- Clear use case and requirements (e.g. which command or workflow)
- Proposed solution or improvement
- Any open questions or considerations

**For Bug Reports:**

- Step-by-step reproduction steps (exact `scli` commands and options)
- Version information (scli version, OS, Go version if built from source)
- Relevant environment details (install root, config.toml presence, etc.)
- Any modifications you've made
- Expected vs actual behavior

## Submitting Pull Requests

### Development

- Set up your development environment: clone the repo, use `go build -o scli .` (see [README](README.md))
- Work against the latest `main` branch
- Focus on specific changes; avoid unrelated edits
- Ensure all tests pass locally (`go test ./...`)
- Follow our [commit convention](#commit-convention)

### Submit PR

- Ensure your branch can be auto-merged (rebase on `main` if needed)
- Address any CI failures
- Respond to review comments promptly

For substantial changes, consider splitting into multiple PRs:

1. **First PR:** Structure, flags, or config changes
2. **Second PR:** Core implementation (e.g. new command or subcommand)
3. **Final PR:** Documentation updates and tests

## Commit Convention

We follow **Conventional Commits**. All commits and PRs should include type specifiers:

- `feat:` new feature (e.g. new command or flag)
- `fix:` bug fix
- `docs:` documentation only
- `chore:` maintenance (deps, tooling, etc.)
- `refactor:` code change that neither fixes a bug nor adds a feature
- `test:` adding or updating tests

Example: `feat(deploy): add --platform docker option`

## Project-Specific Notes

- **Stack:** Go CLI; commands are in `cmd_*.go`; keep changes focused on install, migrate, deploy, and server lifecycle.
- **Compatibility:** Consider impact on different install roots, config.toml vs .env, and systemd/bare metal flows when changing behavior.

## How can I get help?

- Open a [Discussion](https://github.com/sessiondb/scli/discussions) or comment on relevant issues
- Tag maintainers in issues when you need guidance

## Where do I go from here?

- Build and run locally: `go build -o scli .` and try `scli init`, `scli install`, etc. (see [README](README.md))
- Run the test suite: `go test ./...`
- Refer to the main [SessionDB](https://github.com/sessiondb/sessiondb) project for server-side contribution guidelines
