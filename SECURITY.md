# Security Policy

## Supported Versions

Only the latest release of `wt` is supported with security updates.

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

**Email:** <ungkientrung@gmail.com>

Please include:

- Description of the vulnerability
- Steps to reproduce (if possible)
- Impact assessment
- Suggested fix (if known)

## Response Expectations

- We will acknowledge receipt within 48 hours
- We will provide a detailed response within 7 days
- We will work on a fix and coordinate disclosure timeline

## Security Considerations

`wt` is designed with security in mind:

- **No automatic network calls**: `wt` only hits the network when you explicitly use `--fetch`
- **Config trust model**: Configuration file is repo-local and may execute commands (treat it like code)
- **Secret safety**: `wt` does not print `.env` file contents in normal operation
- **Debug mode**: Use `WT_DEBUG=1` for detailed logging (use with caution in production)

## Best Practices

- Review `.wt.config.json` before committing it
- Be cautious with `postCreateCmd` - these commands run automatically
- Use `--force` flags only when necessary (they bypass dirty checks)
- Keep `wt` updated to get security patches

## Private Disclosure

We prefer private disclosure of security vulnerabilities. Please do not create public GitHub issues for security bugs.
