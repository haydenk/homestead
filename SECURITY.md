# Security Policy

## Supported Versions

Only the latest release is actively supported with security fixes.

| Version | Supported |
|---------|-----------|
| Latest  | ✅ |
| Older   | ❌ |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Report vulnerabilities privately using [GitHub's private vulnerability reporting](https://github.com/haydenk/homestead/security/advisories/new). This ensures the issue can be reviewed and a fix prepared before public disclosure.

Please include as much of the following as possible:

- A description of the vulnerability and its potential impact
- The affected version(s)
- Steps to reproduce or a proof of concept
- Any suggested mitigations

You can expect an acknowledgement within **72 hours** and a status update within **7 days**.

## Disclosure Policy

Once a fix is available, the vulnerability will be disclosed via a GitHub Security Advisory along with the patched release. Credit will be given to the reporter unless anonymity is requested.

## Scope

Homestead is a self-hosted personal dashboard intended to run on a private network. Please keep this context in mind when evaluating severity:

- Vulnerabilities that require existing network access to the dashboard are **lower severity**
- Vulnerabilities that allow unauthenticated remote code execution or data exfiltration are **high severity** regardless of deployment context
