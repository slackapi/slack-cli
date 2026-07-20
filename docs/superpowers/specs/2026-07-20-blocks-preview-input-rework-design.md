# `slack blocks preview` — input rework and review fixes

**Date:** 2026-07-20
**Status:** Approved, pending implementation plan
**Branch:** bill-bkb-open-command

## Background

The new `slack blocks preview` command (in `cmd/blocks/`, gated behind the
`block-kit-builder` experiment) opens a set of Block Kit blocks in the Block Kit
Builder in a browser. A code review of the initial implementation surfaced
seven findings. This design reworks how the command accepts input and folds in
the remaining fixes.

### Findings addressed

1. **#1 Interactive hang** — bare `slack blocks preview` (no arg, no pipe) calls
   `io.ReadAll(stdin)` with no TTY guard and blocks forever on an interactive
   terminal.
2. **#2 Wrong enterprise ID** — `teamOrEnterpriseID` keys off
   `EnterpriseID != ""` instead of `IsEnterpriseInstall`, so an org-grid
   *workspace* install opens the Builder in the org (`E…`) context instead of the
   workspace (`T…`) context.
3. **#3 Stdin/team-prompt conflict** — after consuming piped stdin for the
   blocks, the interactive "Select a team" prompt reads the now-exhausted stdin
   pipe and fails on EOF (only masked when the user has exactly one auth).
4. **#4 Experimental command not hidden** — the `blocks` command is listed in
   `slack --help` for all users despite being experimental and non-functional
   without the flag.
5. **#5 Unfriendly empty-arg error** — an empty input produces a raw
   `ErrUnableToParseJSON` instead of the friendly "No blocks were provided".
6. **#6 Malformed URL for bad host** — `buildBlockKitBuilderURL` silently emits
   garbage (`//app./…`) for an empty or scheme-less host.
7. **#7 Unwrapped `url.Parse` error** — the raw stdlib error is returned across a
   package boundary without a structured `slackerror` code (CLAUDE.md rule).

## Design

### 1. Input model: `--blocks` flag with `-` stdin sentinel

Replace the positional `[blocks]` argument with a `--blocks` string flag,
following the `slack api --json '<body>'` precedent. The command becomes
`Args: cobra.NoArgs`.

Behavior matrix:

| Invocation | Source | Notes |
|---|---|---|
| `preview --blocks '[...]'` | literal flag value | stdin untouched |
| `cat f \| preview --blocks -` | stdin (explicit `-` sentinel) | |
| `cat f \| preview` | stdin (auto-detected pipe) | no flag needed |
| `preview` on a TTY, no flag | friendly `ErrMissingInput` error | **no hang** (fixes #1) |

A new `resolveBlocksInput` function resolves the source in this order:

1. `--blocks` set and value `!= "-"` → use the literal value. An empty string
   returns a friendly `ErrMissingInput` "No blocks were provided" (fixes #5).
2. `--blocks -`, **or** flag unset but stdin is a pipe → `io.ReadAll(stdin)`.
3. Flag unset and stdin is an interactive terminal → friendly `ErrMissingInput`
   with remediation. Critically, **no `ReadAll` runs against a TTY**, so the
   command cannot block (fixes #1).

`resolveBlocksInput` also returns a boolean indicating whether the blocks came
from stdin; Section 2 consumes it.

#### Pipe detection (infrastructure)

For `cat f | preview`, stdout is still a TTY — only *stdin* is the pipe — so the
existing `IsTTY()` (which stats stdout) cannot gate stdin reads. We generalize
the existing TTY mechanism rather than special-casing stdin:

- Add `Stdin() types.File` to the `types.Os` interface, the `Os` implementation
  (`return os.Stdin`), and `OsMock` (`AddDefaultMocks` gains
  `m.On("Stdin").Return(os.Stdin)`), mirroring the existing `Stdout()`.
- Add `IsStdinTTY() bool` to the `IOStreamer` interface, implemented exactly like
  `IsTTY()` but statting stdin:
  `(mode & os.ModeCharDevice) == os.ModeCharDevice`. A pipe is therefore
  "not a stdin TTY".

This keeps the fix at the right altitude and remains mockable, so tests stay
hermetic.

### 2. Team selection when stdin is consumed (fix #3)

When blocks are read from stdin, the interactive team picker would read the
exhausted pipe and fail. We rely on what `PromptTeamSlackAuth` already does:

- returns the sole auth directly when exactly one exists (no prompt — works even
  with consumed stdin), and
- honors the `--team` flag via `SelectPromptConfig{Flag: teamFlag}` before
  touching stdin.

The only broken case is **blocks-from-stdin + ≥2 auths + no `--team`**. We detect
exactly that case *before* calling the picker and return a clean, actionable
error instead of proceeding into a survey form that fails on dead stdin:

```
Error: the team could not be determined
Suggestion: Select a team with the --team flag when piping blocks
```

When blocks come from `--blocks '<literal>'`, stdin is untouched and the picker
works interactively as before — no change. The "came from stdin" boolean from
`resolveBlocksInput` drives this; no extra plumbing.

### 3. Remaining fixes

**#2 — Enterprise ID selection.** Change the discriminator to
`auth.IsEnterpriseInstall`, matching `SlackAuth.AuthLevel()`:

```go
func teamOrEnterpriseID(auth *types.SlackAuth) string {
	if auth.IsEnterpriseInstall {
		return auth.EnterpriseID
	}
	return auth.TeamID
}
```

An org-grid workspace install (`EnterpriseID` set, `IsEnterpriseInstall=false`)
now correctly opens the Builder in the workspace `T…` context.

**#4 — Hide the experimental command.** Set `Hidden: true` on the `blocks` parent
command in `blocks.go`, with a comment to remove it when the `block-kit-builder`
experiment graduates. The `preview` `PreRunE` experiment gate stays as-is.

**#6 + #7 — URL hardening in `buildBlockKitBuilderURL`.**

- Wrap the `url.Parse` error:
  `return "", slackerror.Wrap(err, slackerror.ErrInvalidArguments)`.
- After parsing, if `parsed.Host == ""` (the result for `""` or a scheme-less
  `"app.slack.com"`), return `slackerror.New(slackerror.ErrInvalidArguments)`
  instead of silently building `//app./…`.

## Testing

### `cmd/blocks/preview_test.go`

- `--blocks '[...]'` literal → opens builder (replaces the old positional case).
- `cat | preview --blocks -` → explicit stdin sentinel.
- `cat | preview` (piped, no flag) → auto-detected stdin; mock stdin stat reports
  a pipe (not `ModeCharDevice`).
- bare `preview` on a TTY, no flag → `ErrMissingInput`; assert neither
  `OpenURL` nor a blocking `ReadAll` is reached (no hang).
- `--blocks ""` → friendly "No blocks were provided".
- blocks-from-stdin + ≥2 auths + no `--team` → clean error pointing to `--team`
  (fix #3).
- blocks-from-stdin + `--team` set → succeeds.
- org-grid workspace install (`EnterpriseID` set, `IsEnterpriseInstall=false`) →
  uses `T…` (fix #2).
- `buildBlockKitBuilderURL` empty-host and scheme-less-host → error (#6/#7).

### `cmd/blocks/blocks_test.go`

- assert the `blocks` command has `Hidden == true` (#4).

### Infrastructure tests

- `IsStdinTTY()` and `Os.Stdin()` follow the existing `IsTTY()`/`Stdout()` mock
  patterns.

## Docs

Rerun `slack docgen ./docs/reference` to regenerate `slack_blocks.md` and
`slack_blocks_preview.md` (the `--blocks` flag, `NoArgs`, and updated examples
flow from the command definitions).

Because `blocks` becomes `Hidden`, confirm whether docgen still emits hidden
commands. If hidden commands are excluded from generated docs, revert the
`slack.md` index line and the two generated pages so the published docs match
reality until the experiment graduates.

## Verification

- `make test testdir=cmd/blocks`
- `make lint`
- `gofmt -w` on all changed Go files
- Manual smoke test:
  `./bin/slack blocks preview --experiment block-kit-builder --blocks '[{"type":"divider"}]'`

## Out of scope

- Any change to the Block Kit Builder URL format beyond the empty-host guard.
- Removing the `block-kit-builder` experiment gate (graduation is a later
  change).
