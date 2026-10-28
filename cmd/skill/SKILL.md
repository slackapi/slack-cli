---
name: slack-cli
description: "Use when a developer works with the Slack CLI (slack command) to create, run, or manage a Slack app from the terminal: logging in or authenticating (slack login), adding a team or switching workspaces, running locally (slack run), deploying, editing the app manifest, calling Web API methods (slack api), or searching Slack docs (slack docs search)."
---

# Slack CLI

Use the Slack CLI to create, run, and manage Slack apps — including calling Web API methods directly and searching Slack developer documentation from the terminal.

For initial setup (sandbox creation, project scaffolding from templates), use the `slack:create-slack-app` skill instead.

---

## Step 1: Detect the Slack CLI

Resolve the path to the **public** Slack CLI before running any other command. We refer to the resolved value as `SLACK_CMD` — substitute it everywhere below. **Try 1a → 1b → 1c in order; do not skip ahead.**

### 1a. Standard install path (preferred — no verification needed)

The official install scripts place the binary at a fixed per-user path. If the file exists, use it directly: only the install script writes to that path.

- macOS / Linux: `~/.slack/bin/slack`
- Windows: `%USERPROFILE%\AppData\Local\slack-cli\bin\slack.exe` (or `%USERPROFILE%\.slack-cli\bin\slack.exe` as a fallback)

POSIX shell:

```bash
[ -x "$HOME/.slack/bin/slack" ] && echo "$HOME/.slack/bin/slack"
```

PowerShell:

```powershell
$p = "$env:USERPROFILE\AppData\Local\slack-cli\bin\slack.exe"
if (Test-Path $p) { $p } elseif (Test-Path "$env:USERPROFILE\.slack-cli\bin\slack.exe") { "$env:USERPROFILE\.slack-cli\bin\slack.exe" }
```

If a path is returned, set `SLACK_CMD` to that **absolute path** and proceed to Step 2. Do not also run `_fingerprint`.

### 1b. `slack` on PATH (verify with fingerprint)

If 1a found nothing, try `slack` on PATH:

```bash
slack _fingerprint 2>/dev/null
```

If it outputs `d41d8cd98f00b204e9800998ecf8427e`, set `SLACK_CMD=slack` and proceed.

### 1c. Ask about an alias, or install

If 1b fails or returns a different value, ask the developer using AskUserQuestion:

- "The `slack` command on your system doesn't appear to be the public Slack CLI. Do you have it installed under a different name or alias?"
- Options: "Yes, it's aliased as..." (let them provide the alias), "No, I need to install it"
- If they provide an alias, verify it with `<alias> _fingerprint 2>/dev/null` and set `SLACK_CMD=<alias>`.
- If they need to install it, run:

  ```text
  curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash
  ```

  Then re-run **1a** — the install script will have written `~/.slack/bin/slack`.

**Common mistakes:** Don't use `which slack` to discover the binary — `which` resolves any shell alias and defeats the point of 1a. In Git Bash on Windows, use the POSIX probe form, not PowerShell.

---

## Step 2: Command Discovery via Help

**Always run `SLACK_CMD <command> --help` before constructing a command you have not used in the current session.** Do not guess at flags — the help output is the source of truth.

- `SLACK_CMD help` — lists all available command groups
- `SLACK_CMD <command> --help` — shows subcommands, flags, and usage examples

### Resolving `--app` and `--team` values

When a command requires `--app` or `--team`:

- **App ID**: Run `SLACK_CMD app list` from the project directory to see installed apps and their IDs.
- **Team ID**: Run `SLACK_CMD auth list` to see authenticated workspaces and their team IDs.

---

## Step 3: Searching Documentation (`slack docs search`)

Search Slack's developer documentation directly from the terminal.

```bash
SLACK_CMD docs search "<query>" --output=text --limit=5
```

Use `--output=text` for concise terminal-readable results. Use this when you are already running CLI commands and want to:

- Check a Slack feature before implementing it
- Verify API behavior or required scopes
- Look up Block Kit elements, event types, or method parameters

---

## Step 4: Calling Web API Methods (`slack api`)

Call any Slack Web API method directly. Run `SLACK_CMD api --help` for full details and examples.

Parameters are passed as positional `key=value` pairs (NOT `--key=value` flags):

```bash
SLACK_CMD api chat.postMessage channel=C0123456789 text="Hello from the CLI"
```

**Important distinction**: `--team`, `--token`, `--json`, and `--data` are meta-flags (prefixed with `--`). API method parameters use positional `key=value` syntax without dashes.

**Reference**: Full method list at <https://docs.slack.dev/reference/methods.md>.

---

## Step 5: Authentication (`slack auth`)

### When to run this flow

Slack auth is **per-team**, not a single boolean. Run the seamless login flow below whenever any of these is true:

- The developer is not authenticated to any team yet.
- The developer wants to **add a new team / workspace / sandbox**, even if `SLACK_CMD auth list` already shows other teams.
- The developer asks to log in, switch teams, or re-authenticate.

`SLACK_CMD auth list` showing other teams is **not** a reason to skip login — those are different teams. Ask the developer which team they want, then run the flow.

The flow below is the same for every kind of workspace. `/slackauthticket` works in any workspace the developer belongs to, whether that is a developer sandbox, a Free Team, or a production workspace, so there is no sandbox-specific login. Which target to prefer is the `slack:create-slack-app` skill's call (**Step 3: Choose Where to Install the App**).

### Inspect existing auth (optional)

```bash
SLACK_CMD auth list
```

Use this to show the developer which teams are already authenticated, or to confirm a successful login. Do not treat a non-empty list as "auth complete" when the developer asked to log in to a new team.

### Run the seamless login flow

The agent drives this end-to-end — **no separate terminal window, no browser confirmation**.

`SLACK_CMD login --no-prompt` makes the CLI emit a single-use ticket and exit immediately instead of waiting on stdin. Slack itself renders the challenge code **inside a workspace modal** when the developer sends the `/slackauthticket` slash command — there is no browser step. The agent submits the challenge back to the CLI to complete login.

**1. Start login and capture the ticket**

```bash
SLACK_CMD login --no-prompt
```

The CLI prints a `/slackauthticket <ticket>` slash command and exits. Capture the ticket — you will need it in step 4. Sample output:

```text
📋 Run the following slash command from any Slack channel in the workspace
   you'd like to authenticate

   /slackauthticket eyJ0eXAiOiJKV1QiLCJh…

? Slack will then show you a challenge code. Submit it via:
   slack login --ticket <ticket> --challenge <code>
```

**2. Hand the slash command to the developer**

Show the full `/slackauthticket …` line and ask the developer to paste it into the message box of the Slack workspace they want to authenticate, then send it. Slack responds with a modal containing a short challenge code (e.g. `JDt1IK7X`).

**3. Collect the challenge code**

Use `AskUserQuestion` to ask the developer for the challenge code shown in the Slack modal. Wait for their answer — do not guess or default.

**4. Complete login**

```bash
SLACK_CMD login --ticket <ticket> --challenge <code>
```

On success the CLI returns the team name and ID. Verify with `SLACK_CMD auth list` and report the team back to the developer.

**Troubleshooting**: tickets are single-use and time-limited. If step 4 fails with an invalid/expired ticket or wrong challenge, restart from step 1 with a fresh `SLACK_CMD login --no-prompt` — do not retry the same ticket.

Use `--team <team_id>` on individual commands to target a specific workspace without switching globally.

### Red flags — STOP

If you catch yourself thinking any of these, you are about to regress to the old broken flow:

| Rationalization | Reality |
|---|---|
| "`auth list` already shows teams, so login isn't needed." | Auth is per-team. The developer asked for a _new_ team — drive the flow. |
| "`slack login` needs browser confirmation, so I can't drive it." | False with `--no-prompt`. The challenge code appears in Slack's modal, not a browser. The agent runs both `slack login` invocations itself. |
| "I should tell the developer to run `slack login` in a separate terminal." | Never. Step 5 is the agent's job from start to finish. |

**All of these mean: run `SLACK_CMD login --no-prompt` yourself and follow the four numbered steps above.**

---

## Step 6: Running an App Locally (`slack run`)

Run `SLACK_CMD run --help` for all available flags.

### Resolve the app or team target

Run `SLACK_CMD app list` from the project directory to check for installed apps:

- **If an app exists**: Use `--app=<app_id>` in the run command.
- **If no apps are listed**: Use `--team=<team_id>` instead (get the team ID from `SLACK_CMD auth list`).

### Start the dev server in the background

Run this command as a **background process** so the developer can continue working:

**If an app ID was found:**

```bash
cd <project-dir> && SLACK_CMD run --org-workspace-grant=all --app=<app_id>
```

**If no apps — use the team ID:**

```bash
cd <project-dir> && SLACK_CMD run --org-workspace-grant=all --team=<team_id>
```

Tell the developer the app is running in the background. They can ask:

- "What's the status of the dev server?" — to check on it
- "Show me the output from slack run" — to see activity logs
- "Stop the dev server" — to terminate the process

---

## Step 7: Managing the Manifest

Run `SLACK_CMD manifest --help` for subcommands (`validate`, `info`) and the `--source local|remote` flag.

Use `SLACK_CMD manifest validate` before deploying or when something seems wrong with the app configuration.

---

## Step 8: Other Commands

For any other command group (e.g., `trigger`, `datastore`, `env`, `collaborator`, `external-auth`, `deploy`), run `SLACK_CMD <command> --help` to discover subcommands and flags. Run `SLACK_CMD help` to see all available command groups.

---

## Notes

- `SLACK_CMD` is a placeholder — always substitute the actual command name resolved in Step 1.
- **Always run `--help`** before constructing a command you have not used in the current session.
- Interactive commands (e.g., `slack trigger create` without `--trigger-def`) cannot be run in the background. Tell the developer to run these in a **new terminal window**. `slack login` is **not** in this category — drive it inline using the `--no-prompt` / `--ticket` / `--challenge` flow in Step 5.
- `slack run` runs locally for development. `slack deploy` deploys to Slack's hosted infrastructure, unless `.slack/hooks.json` defines a `deploy` hook, in which case it runs that hook instead (see the `slack:deploy-slack-app` skill). These are different operations, so do not confuse them.
