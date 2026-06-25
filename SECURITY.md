# Security Policy

Slack takes the security of its software and services seriously, including all open-source repositories managed through the [slackapi](https://github.com/slackapi) GitHub organization.

## Reporting a Vulnerability

**Do NOT report security vulnerabilities through public GitHub issues, pull requests, or discussions.**

If you believe you have found a security vulnerability in the Slack CLI, please report it through the Slack bug bounty program on HackerOne:

**<https://hackerone.com/slack>**

Even if the Slack CLI is not explicitly listed as an in-scope asset on the HackerOne program page, reports for vulnerabilities in this tool should still be submitted there. The Slack security team triages reports for all `slackapi` open-source repositories through this program.

If HackerOne is inaccessible, you may alternatively report the issue to [security@salesforce.com](mailto:security@salesforce.com).

Please do not discuss potential vulnerabilities in public without first coordinating with the security team.

## What to Include

To help us triage and respond quickly, please include:

- Type of vulnerability (e.g., credential exposure, authentication bypass, arbitrary code execution)
- Affected version(s) of the Slack CLI (`slack version`)
- Step-by-step reproduction instructions
- Proof-of-concept code, commands, or payloads, if available
- Impact assessment: what an attacker could achieve
- Any specific configuration or environment required to trigger the vulnerability
- Affected source file paths, if known

## Threat Model

The Slack CLI is a command-line tool that authenticates developers to Slack, stores credentials locally, scaffolds and runs app projects, and communicates with Slack's APIs. The security boundary covers the safe handling of credentials on the local machine, the integrity of authentication flows, and the confidentiality of data in transit.

### In Scope

The following are considered CLI vulnerabilities:

- Leakage of authentication tokens or credentials through logs, error messages, telemetry, or world-readable files on disk
- Authentication or session-handling flaws in the `slack login` / device-flow authorization process
- Failure to enforce TLS for connections to Slack's APIs
- Arbitrary command or code execution triggered by malicious project files, manifests, hook responses, or API responses
- Path traversal or file overwrite outside the project directory during project create/init or other file operations
- Insecure default file permissions on stored credentials or configuration
- Injection vulnerabilities where CLI internals unsafely interpolate untrusted data into commands or requests

### Out of Scope

The following are NOT CLI vulnerabilities:

- Vulnerabilities in the Go runtime, operating system, terminal emulator, or hosting infrastructure
- Security issues in Slack's server-side platform infrastructure (report those directly under Slack's main HackerOne scope)
- Vulnerabilities in third-party Go modules chosen and vendored outside of the CLI's direct dependencies
- Security issues in developer application code created or run with the CLI (e.g., insecure logic in a generated app)
- Attacks that require possession of a valid, already-authenticated session or stolen credentials
- Social-engineering attacks that depend on a developer running an untrusted command or installing a malicious SDK
- Issues that only affect end-of-life versions with no reproduction on supported versions

## Disclosure Policy

This project follows coordinated disclosure:

- Allow a reasonable timeframe for the team to investigate, develop, and release a fix before any public disclosure.
- Researchers who follow responsible disclosure practices are eligible for recognition and bounty consideration through the Slack HackerOne program.
