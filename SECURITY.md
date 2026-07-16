# Security Policy

`offline` is a sandboxing tool — bugs in its isolation logic (namespace
setup, capability dropping, seccomp filter) can mean network access it
claims to block actually leaks through. Please report those with priority.

## Supported Versions

| Version | Supported |
| ------- | --------- |
| latest  | ✅        |
| < latest| ❌        |

## Reporting a Vulnerability

**Do not open a public issue for security problems**, especially sandbox
escapes or network-isolation bypasses.

Report privately via [GitHub Security Advisories](https://github.com/fabiocicerchia/offline/security/advisories/new)
(preferred) or email **info@fabiocicerchia.it**.

Please include a description, reproduction steps, and impact. We aim to
acknowledge within 48 hours and to ship a fix or mitigation as soon as
practical, keeping you updated along the way.
