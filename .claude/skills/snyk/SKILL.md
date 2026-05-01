---
name: snyk
description: Run Snyk security scans to find dependency vulnerabilities and source code issues. Use for monthly security reviews or when checking for new vulnerabilities.
---

Run a Snyk security scan on this project.

## 1. Check prerequisites

Run `which snyk` to verify Snyk is installed. If not found, tell the user to install it with `brew install snyk` or `npm install -g snyk`.

Run `snyk auth check` or `snyk whoami` to verify authentication. If not authenticated, tell the user to run `! snyk auth` to log in interactively.

## 2. Run `snyk test` (dependency vulnerabilities — primary scan)

Run `snyk test` to scan Go module dependencies for known vulnerabilities.

**This is the most important scan.** Summarize the results:

- Group vulnerabilities by severity: **Critical > High > Medium > Low**
- For each vulnerability, note:
  - The affected package and version
  - Whether a fix is available (upgrade path exists) or requires waiting on the upstream maintainer
- For fixable issues, propose the specific `go get` upgrade commands
- For unfixable issues, note them as "waiting on upstream" — these are deferred

## 3. Run `snyk code test` (source code analysis — secondary scan)

Run `snyk code test` to scan the project's own Go source code for security issues.

**This scan is optional and secondary.** Summarize the results:

- Group findings by severity
- Identify which issues are simple/quick to fix vs. complex
- Focus on simple fixes that can be resolved quickly

## 4. Present a prioritized action plan

Combine both scan results into a single prioritized plan:

1. **Fix now** — dependency upgrades with available fixes (propose commands)
2. **Fix now** — simple source code issues from `snyk code test`
3. **Defer** — dependency vulnerabilities waiting on upstream fixes
4. **Defer** — complex source code issues that need more investigation

Ask the user which items they'd like to tackle, then help resolve them.
