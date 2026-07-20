# Blocks Preview Input Rework Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework `slack blocks preview` to take blocks via a `--blocks` flag (with a `-` stdin sentinel and auto pipe-detection) and fold in the seven code-review findings.

**Architecture:** Add a small, testable stdin-TTY detection capability to the shared `Os`/`IOStreams` infrastructure, then rework `cmd/blocks/preview.go` input resolution on top of it. The remaining fixes (enterprise ID, team-prompt guard, URL hardening, hidden command, docs) are localized edits to `cmd/blocks/`.

**Tech Stack:** Go, Cobra, testify/mock, afero. Test harness: `test/testutil.TableTestCommand`.

## Global Constraints

- Wrap errors crossing package boundaries with `slackerror.Wrap(err, slackerror.ErrCode)` (repo CLAUDE.md).
- Register new error codes in `internal/slackerror/errors.go`, alphabetically, in both the const block and `ErrorCodeMap`. (No new codes are needed in this plan — `ErrInvalidBlocks`, `ErrMissingInput`, `ErrMissingFlag`, `ErrInvalidArguments` all already exist.)
- Use `clients.Fs` / the `Os` abstraction, never direct `os` calls, so code stays testable.
- Test naming: `Test_StructName_FunctionName` or `Test_FunctionName`. Constructors first, then alphabetical.
- Table-driven tests use the map pattern with `tc` as the case variable.
- Run `gofmt -w` on every changed Go file. Run `make lint` before finishing.
- Commit messages use conventional-commit prefixes and end with:
  `Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>`
- The entire `cmd/blocks/` feature is currently uncommitted working-tree state on this branch. "Modify" below means editing an existing working-tree file even though it is not yet committed.

---

### Task 1: Stdin-TTY detection infrastructure

Adds `Os.Stdin()` and `IOStreams.IsStdinTTY()`, mirroring the existing `Os.Stdout()` / `IsTTY()`. `IsTTY()` stats **stdout**; for `cat f | preview` stdout is still a terminal while only stdin is a pipe, so a separate stdin check is required.

**Files:**
- Modify: `internal/shared/types/slackdeps.go` (add `Stdin() File` to the `Os` interface, after `Stdout() File` at line 67)
- Modify: `internal/slackdeps/os.go` (add `Stdin()` impl after `Stdout()` at line 100)
- Modify: `internal/slackdeps/os_mock.go` (add `Stdin()` mock method + default in `AddDefaultMocks`)
- Modify: `internal/iostreams/iostreams.go` (add `IsStdinTTY()` to `IOStreamer` interface + impl)
- Modify: `internal/iostreams/iostreams_mock.go` (add `IsStdinTTY()` mock method + default)
- Test: `internal/iostreams/iostreams_test.go`

**Interfaces:**
- Produces:
  - `types.Os` interface gains `Stdin() File`
  - `(*slackdeps.Os).Stdin() types.File`
  - `(*slackdeps.OsMock).Stdin() types.File`
  - `IOStreamer` interface gains `IsStdinTTY() bool`
  - `(*iostreams.IOStreams).IsStdinTTY() bool` — returns true when stdin is a character device (interactive terminal), false when piped/redirected or on stat error
  - `(*iostreams.IOStreamsMock).IsStdinTTY() bool`

- [ ] **Step 1: Write the failing test**

Add to `internal/iostreams/iostreams_test.go` (immediately after `Test_IOStreams_IsTTY`, keeping alphabetical-ish grouping with the other `IsTTY` test):

```go
func Test_IOStreams_IsStdinTTY(t *testing.T) {
	tests := map[string]struct {
		fileInfo os.FileInfo
		expected bool
	}{
		"interactive when stdin is a char device": {
			fileInfo: &slackdeps.FileInfoCharDevice{},
			expected: true,
		},
		"not interactive when stdin is a named pipe": {
			fileInfo: &slackdeps.FileInfoNamedPipe{},
			expected: false,
		},
		"not interactive when the stat check errors": {
			expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fsMock := slackdeps.NewFsMock()
			osMock := slackdeps.NewOsMock()
			config := config.NewConfig(fsMock, osMock)
			osMock.On("Stdin").Return(&slackdeps.FileMock{FileInfo: tc.fileInfo})
			io := NewIOStreams(config, fsMock, osMock)

			isStdinTTY := io.IsStdinTTY()
			assert.Equal(t, tc.expected, isStdinTTY)
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `make test testdir=internal/iostreams testname=Test_IOStreams_IsStdinTTY`
Expected: FAIL — compile error, `io.IsStdinTTY` undefined and `osMock.On("Stdin")` has no matching method.

- [ ] **Step 3: Add `Stdin()` to the `Os` interface**

In `internal/shared/types/slackdeps.go`, add to the `Os` interface right after the `Stdout() File` line (line 67):

```go
	// Stdout returns the file descriptor for stdout
	Stdout() File

	// Stdin returns the file descriptor for stdin
	Stdin() File
```

- [ ] **Step 4: Implement `Os.Stdin()`**

In `internal/slackdeps/os.go`, add after the `Stdout()` method (after line 100):

```go
// Stdin returns the file descriptor for stdin
func (c *Os) Stdin() types.File {
	return os.Stdin
}
```

- [ ] **Step 5: Add the `OsMock.Stdin()` method and default**

In `internal/slackdeps/os_mock.go`, add to `AddDefaultMocks` right after `m.On("Stdout").Return(os.Stdout)` (line 58):

```go
	m.On("Stdin").Return(os.Stdin)
```

Then add the method after the `Stdout()` mock method (after line 138):

```go
// Stdin mocks the stdin with a file that can be adjusted
func (m *OsMock) Stdin() types.File {
	args := m.Called()
	return args.Get(0).(types.File)
}
```

- [ ] **Step 6: Add `IsStdinTTY()` to the `IOStreamer` interface + implement it**

In `internal/iostreams/iostreams.go`, add to the `IOStreamer` interface right after the `IsTTY() bool` declaration (line 83):

```go
	// IsTTY returns true if the device is an interactive terminal
	IsTTY() bool

	// IsStdinTTY returns true if stdin is an interactive terminal (not a pipe
	// or redirected file)
	IsStdinTTY() bool
```

Then add the implementation right after the `IsTTY()` method (after its closing brace near line 131):

```go
// IsStdinTTY returns true if stdin is an interactive terminal
//
// Unlike IsTTY, which inspects stdout, this inspects stdin so that piped or
// redirected input (e.g. `cat blocks.json | slack ...`) is detected even when
// stdout is still attached to a terminal.
func (io *IOStreams) IsStdinTTY() bool {
	if o, err := io.os.Stdin().Stat(); o == nil || err != nil {
		return false
	} else {
		return (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice
	}
}
```

- [ ] **Step 7: Add the `IOStreamsMock.IsStdinTTY()` method and default**

In `internal/iostreams/iostreams_mock.go`, add to `AddDefaultMocks` right after `m.On("IsTTY").Return(false)` (line 73):

```go
	m.On("IsStdinTTY").Return(false)
```

Then add the method right after the `IsTTY()` mock method (after line 98):

```go
func (m *IOStreamsMock) IsStdinTTY() bool {
	args := m.Called()
	return args.Bool(0)
}
```

- [ ] **Step 8: Run the test to verify it passes**

Run: `make test testdir=internal/iostreams testname=Test_IOStreams_IsStdinTTY`
Expected: PASS (all three sub-cases).

- [ ] **Step 9: Run the broader packages to catch interface-conformance breaks**

Run: `make test testdir=internal/iostreams` then `make test testdir=internal/slackdeps`
Expected: PASS. (If any other `IOStreamer` implementer exists it would fail to compile here; there is only `IOStreams` and `IOStreamsMock`.)

- [ ] **Step 10: Format and commit**

```bash
gofmt -w internal/shared/types/slackdeps.go internal/slackdeps/os.go internal/slackdeps/os_mock.go internal/iostreams/iostreams.go internal/iostreams/iostreams_mock.go internal/iostreams/iostreams_test.go
git add internal/shared/types/slackdeps.go internal/slackdeps/os.go internal/slackdeps/os_mock.go internal/iostreams/iostreams.go internal/iostreams/iostreams_mock.go internal/iostreams/iostreams_test.go
git commit -m "$(cat <<'EOF'
feat: add IsStdinTTY for detecting piped stdin

Adds Os.Stdin() and IOStreams.IsStdinTTY(), mirroring Stdout()/IsTTY(),
so callers can distinguish piped/redirected stdin from an interactive
terminal even when stdout is still a TTY.

Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>
EOF
)"
```

---

### Task 2: Rework preview input model (`--blocks` flag, `-` sentinel, no-hang)

Replaces the positional `[blocks]` arg with a `--blocks` string flag, adds `-` and auto-pipe stdin handling, and makes bare `preview` on a terminal error instead of hanging. Fixes findings #1 and #5.

**Files:**
- Modify: `cmd/blocks/preview.go`
- Test: `cmd/blocks/preview_test.go`

**Interfaces:**
- Consumes: `clients.IO.IsStdinTTY() bool` (Task 1); `clients.IO.ReadIn() io.Reader`; `promptTeamSlackAuthFunc` (existing package var).
- Produces:
  - `resolveBlocksInput(clients *shared.ClientFactory, flagValue string, flagChanged bool) (input string, fromStdin bool, err error)`
  - `previewCommandRunE(clients *shared.ClientFactory, cmd *cobra.Command, blocksFlag string, blocksFlagChanged bool) error` (signature changes — no longer takes `args`)
  - `--blocks` flag on the preview command; command is now `Args: cobra.NoArgs`
  - `fromStdin` boolean consumed by Task 3

- [ ] **Step 1: Write the failing tests**

Replace the body of `Test_Blocks_PreviewCommand` in `cmd/blocks/preview_test.go` with the cases below (the argument-based case becomes flag-based; add the stdin-sentinel, auto-pipe, and no-hang cases). Keep the existing `enableExperiment` and `stubTeamAuth` helpers.

```go
func Test_Blocks_PreviewCommand(t *testing.T) {
	var restore func()
	testutil.TableTestCommand(t, testutil.CommandTests{
		"opens the builder with blocks from the --blocks flag": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				enableExperiment(ctx, cm)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
				cm.IO.AssertCalled(t, "PrintTrace", mock.Anything, slacktrace.BlocksPreviewSuccess, []string{expectedURL})
			},
			Teardown: func() { restore() },
		},
		"opens the builder with blocks from stdin via the - sentinel": {
			CmdArgs: []string{"--blocks", "-"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				enableExperiment(ctx, cm)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			Teardown: func() { restore() },
		},
		"opens the builder with auto-detected piped stdin": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				// default IsStdinTTY() is false (piped)
				enableExperiment(ctx, cm)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			Teardown: func() { restore() },
		},
		"accepts a blocks object payload": {
			CmdArgs: []string{"--blocks", `{"blocks":[{"type":"divider"}]}`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				enableExperiment(ctx, cm)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				expectedURL := `https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%7B%22type%22:%22divider%22%7D%5D%7D`
				cm.Browser.AssertCalled(t, "OpenURL", expectedURL)
			},
			Teardown: func() { restore() },
		},
		"errors when the experiment is not enabled": {
			CmdArgs:              []string{"--blocks", `[{"type":"divider"}]`},
			ExpectedErrorStrings: []string{slackerror.ErrMissingExperiment, "experimental"},
		},
		"errors without hanging when no blocks are provided on a terminal": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.On("IsStdinTTY").Return(true)
				enableExperiment(ctx, cm)
			},
			ExpectedErrorStrings: []string{slackerror.ErrMissingInput, "No blocks were provided"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"errors when the --blocks flag is empty": {
			CmdArgs: []string{"--blocks", ""},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableExperiment(ctx, cm)
			},
			ExpectedErrorStrings: []string{slackerror.ErrMissingInput, "No blocks were provided"},
		},
		"errors when the blocks are not valid json": {
			CmdArgs: []string{"--blocks", `not json`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableExperiment(ctx, cm)
			},
			ExpectedErrorStrings: []string{slackerror.ErrUnableToParseJSON},
		},
		"errors when the json is not a blocks payload": {
			CmdArgs: []string{"--blocks", `{"foo":"bar"}`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				enableExperiment(ctx, cm)
			},
			ExpectedErrorStrings: []string{slackerror.ErrInvalidBlocks},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewPreviewCommand(cf)
	})
}
```

Note: the `Test_buildBlockKitBuilderURL` and `Test_normalizeBlocksPayload` functions in this file are unchanged by this task.

- [ ] **Step 2: Run tests to verify they fail**

Run: `make test testdir=cmd/blocks testname=Test_Blocks_PreviewCommand`
Expected: FAIL — the command still uses a positional arg, so `--blocks` is an unknown flag and the argument-based expectations do not match.

- [ ] **Step 3: Rework the command constructor and input resolution**

In `cmd/blocks/preview.go`, replace `NewPreviewCommand`, `previewCommandRunE`, and `readBlocksInput` with the following. Keep `previewCommandPreRunE`, `normalizeBlocksPayload`, `teamOrEnterpriseID`, and `buildBlockKitBuilderURL` as they are (later tasks touch two of them). The `io` import stays (used by `readStdinBlocks`).

```go
func NewPreviewCommand(clients *shared.ClientFactory) *cobra.Command {
	var blocksFlag string
	cmd := &cobra.Command{
		Use:   "preview [flags]",
		Short: "Preview blocks in the Block Kit Builder",
		Long: strings.Join([]string{
			"Open a set of Block Kit blocks in the Block Kit Builder in a web browser.",
			"",
			"Provide blocks with the --blocks flag or pipe them through standard input.",
			"The input is a JSON array of blocks or a JSON object with a \"blocks\" array.",
			"Pass - to --blocks to read explicitly from standard input.",
		}, "\n"),
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Preview blocks passed with a flag",
				Command: "blocks preview --blocks '[{\"type\":\"divider\"}]'",
			},
			{
				Meaning: "Preview blocks piped from a file",
				Command: "blocks preview < blocks.json",
			},
		}),
		Args: cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return previewCommandPreRunE(clients)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return previewCommandRunE(clients, cmd, blocksFlag, cmd.Flags().Changed("blocks"))
		},
	}
	cmd.Flags().StringVar(&blocksFlag, "blocks", "", "blocks to preview as a JSON array or object\n  (use - to read from standard input)")
	return cmd
}

// previewCommandRunE resolves blocks from the flag or standard input and opens
// them in the Block Kit Builder for the selected team
func previewCommandRunE(clients *shared.ClientFactory, cmd *cobra.Command, blocksFlag string, blocksFlagChanged bool) error {
	ctx := cmd.Context()
	clients.IO.PrintTrace(ctx, slacktrace.BlocksPreviewStart)

	blocksInput, _, err := resolveBlocksInput(clients, blocksFlag, blocksFlagChanged)
	if err != nil {
		return err
	}

	blocksJSON, err := normalizeBlocksPayload(blocksInput)
	if err != nil {
		return err
	}

	auth, err := promptTeamSlackAuthFunc(ctx, clients, "Select a team to preview blocks for", nil)
	if err != nil {
		return err
	}

	builderURL, err := buildBlockKitBuilderURL(clients.API().Host(), teamOrEnterpriseID(auth), blocksJSON)
	if err != nil {
		return err
	}

	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji: "eyes",
		Text:  "Block Kit Builder",
		Secondary: []string{
			builderURL,
		},
	}))
	clients.Browser().OpenURL(builderURL)

	clients.IO.PrintTrace(ctx, slacktrace.BlocksPreviewSuccess, builderURL)
	return nil
}

// resolveBlocksInput returns the blocks to preview and whether they were read
// from standard input. Resolution order: an explicit --blocks value, the -
// sentinel or an auto-detected stdin pipe, otherwise a friendly error. Reading
// stdin is never attempted against an interactive terminal, so a bare command
// on a TTY errors instead of blocking on io.ReadAll.
func resolveBlocksInput(clients *shared.ClientFactory, flagValue string, flagChanged bool) (string, bool, error) {
	switch {
	case flagChanged && flagValue == "-":
		return readStdinBlocks(clients)
	case flagChanged:
		input := strings.TrimSpace(flagValue)
		if input == "" {
			return "", false, missingBlocksError()
		}
		return input, false, nil
	case !clients.IO.IsStdinTTY():
		return readStdinBlocks(clients)
	default:
		return "", false, missingBlocksError()
	}
}

// readStdinBlocks reads and trims blocks from standard input
func readStdinBlocks(clients *shared.ClientFactory) (string, bool, error) {
	piped, err := io.ReadAll(clients.IO.ReadIn())
	if err != nil {
		return "", true, slackerror.Wrap(err, slackerror.ErrMissingInput)
	}
	input := strings.TrimSpace(string(piped))
	if input == "" {
		return "", true, missingBlocksError()
	}
	return input, true, nil
}

// missingBlocksError is the friendly error returned when no blocks are supplied
func missingBlocksError() error {
	return slackerror.New(slackerror.ErrMissingInput).
		WithMessage("No blocks were provided").
		WithRemediation("Provide blocks with the %s flag or pipe them through standard input", style.Highlight("--blocks"))
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `make test testdir=cmd/blocks testname=Test_Blocks_PreviewCommand`
Expected: PASS (all cases, including the no-hang and empty-flag cases).

- [ ] **Step 5: Format and commit**

```bash
gofmt -w cmd/blocks/preview.go cmd/blocks/preview_test.go
git add cmd/blocks/preview.go cmd/blocks/preview_test.go
git commit -m "$(cat <<'EOF'
feat: accept blocks via --blocks flag with stdin sentinel

Replaces the positional argument with a --blocks flag, supports the -
stdin sentinel and auto-detected piped stdin, and errors on a bare
interactive invocation instead of blocking forever on stdin.

Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>
EOF
)"
```

---

### Task 3: Guard team selection when blocks come from stdin

When blocks are read from stdin, the interactive team picker would read the exhausted stdin pipe and fail on EOF. Detect the one broken case — stdin input + more than one auth + no `--team` — and return a clean, actionable error before prompting. Fixes finding #3.

**Files:**
- Modify: `cmd/blocks/preview.go`
- Test: `cmd/blocks/preview_test.go`

**Interfaces:**
- Consumes: `fromStdin` from `resolveBlocksInput` (Task 2); `clients.Config.TeamFlag string`; `clients.Auth().Auths(ctx) ([]types.SlackAuth, error)`; `promptTeamSlackAuthFunc`.
- Produces: `selectTeamAuth(ctx context.Context, clients *shared.ClientFactory, fromStdin bool) (*types.SlackAuth, error)`

- [ ] **Step 1: Write the failing tests**

Add these two cases inside the `testutil.CommandTests{...}` map in `Test_Blocks_PreviewCommand` (in `cmd/blocks/preview_test.go`):

```go
		"errors when piping blocks with multiple teams and no --team flag": {
			CmdArgs: []string{"--blocks", "-"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				cm.Auth.On("Auths", mock.Anything).Return([]types.SlackAuth{
					{TeamID: "T123", TeamDomain: "team-a"},
					{TeamID: "T456", TeamDomain: "team-b"},
				}, nil)
				enableExperiment(ctx, cm)
			},
			ExpectedErrorStrings: []string{slackerror.ErrMissingFlag, "--team"},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertNotCalled(t, "OpenURL", mock.Anything)
			},
		},
		"opens the builder when piping blocks with the --team flag set": {
			CmdArgs: []string{"--blocks", "-", "--team", "T123"},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				cm.IO.Stdin = bytes.NewBufferString(`[{"type":"divider"}]`)
				enableExperiment(ctx, cm)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123"})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/T123/builder")
				}))
			},
			Teardown: func() { restore() },
		},
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `make test testdir=cmd/blocks testname=Test_Blocks_PreviewCommand`
Expected: FAIL — the multi-team stdin case currently proceeds to the stubbed/real prompt instead of returning `ErrMissingFlag`.

- [ ] **Step 3: Add the team-selection guard**

In `cmd/blocks/preview.go`, add the `context` import if it is not already imported (it is needed for the new function signature), then add this function directly below `previewCommandRunE`:

```go
// selectTeamAuth chooses the team whose Block Kit Builder to open. When blocks
// were read from stdin the interactive picker cannot prompt (stdin is spent),
// so with more than one authorization and no --team flag we return an
// actionable error instead of failing on EOF inside the prompt.
func selectTeamAuth(ctx context.Context, clients *shared.ClientFactory, fromStdin bool) (*types.SlackAuth, error) {
	if fromStdin && clients.Config.TeamFlag == "" {
		auths, err := clients.Auth().Auths(ctx)
		if err != nil {
			return nil, err
		}
		if len(auths) > 1 {
			return nil, slackerror.New(slackerror.ErrMissingFlag).
				WithMessage("The team could not be determined").
				WithRemediation("Select a team with the %s flag when piping blocks", style.Highlight("--team"))
		}
	}
	return promptTeamSlackAuthFunc(ctx, clients, "Select a team to preview blocks for", nil)
}
```

Then, in `previewCommandRunE`, capture the `fromStdin` return value and route team selection through `selectTeamAuth`. Change:

```go
	blocksInput, _, err := resolveBlocksInput(clients, blocksFlag, blocksFlagChanged)
```
to:
```go
	blocksInput, fromStdin, err := resolveBlocksInput(clients, blocksFlag, blocksFlagChanged)
```

and change:

```go
	auth, err := promptTeamSlackAuthFunc(ctx, clients, "Select a team to preview blocks for", nil)
```
to:
```go
	auth, err := selectTeamAuth(ctx, clients, fromStdin)
```

Add `"context"` to the import block if absent.

- [ ] **Step 4: Run tests to verify they pass**

Run: `make test testdir=cmd/blocks testname=Test_Blocks_PreviewCommand`
Expected: PASS. In particular the multi-team stdin case returns `ErrMissingFlag` and does not open a browser, and the single-team stdin cases from Task 2 still pass (the stub returns before the `len(auths) > 1` check matters).

- [ ] **Step 5: Format and commit**

```bash
gofmt -w cmd/blocks/preview.go cmd/blocks/preview_test.go
git add cmd/blocks/preview.go cmd/blocks/preview_test.go
git commit -m "$(cat <<'EOF'
fix: require --team when piping blocks with multiple auths

Reading blocks from stdin consumes the pipe, so the interactive team
picker cannot prompt. Detect stdin input with more than one auth and no
--team flag and return an actionable error instead of failing on EOF.

Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>
EOF
)"
```

---

### Task 4: Use the enterprise ID only for enterprise installs

`teamOrEnterpriseID` currently keys off `EnterpriseID != ""`, so an org-grid workspace install (enterprise ID present but not an enterprise install) opens the wrong Builder context. Key off `IsEnterpriseInstall` instead, matching `SlackAuth.AuthLevel()`. Fixes finding #2.

**Files:**
- Modify: `cmd/blocks/preview.go:170-175`
- Test: `cmd/blocks/preview_test.go`

**Interfaces:**
- Produces: `teamOrEnterpriseID(auth *types.SlackAuth) string` — unchanged signature, corrected behavior.

- [ ] **Step 1: Write the failing tests**

Add these two cases to the `testutil.CommandTests{...}` map in `Test_Blocks_PreviewCommand`:

```go
		"uses the enterprise id for enterprise installs": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				enableExperiment(ctx, cm)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123", EnterpriseID: "E456", IsEnterpriseInstall: true})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/E456/builder")
				}))
			},
			Teardown: func() { restore() },
		},
		"uses the team id for org-grid workspace installs": {
			CmdArgs: []string{"--blocks", `[{"type":"divider"}]`},
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
				cm.API.On("Host").Return("https://slack.com")
				enableExperiment(ctx, cm)
				restore = stubTeamAuth(&types.SlackAuth{TeamID: "T123", EnterpriseID: "E456", IsEnterpriseInstall: false})
			},
			ExpectedAsserts: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock) {
				cm.Browser.AssertCalled(t, "OpenURL", mock.MatchedBy(func(url string) bool {
					return assert.Contains(t, url, "/block-kit-builder/T123/builder")
				}))
			},
			Teardown: func() { restore() },
		},
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `make test testdir=cmd/blocks testname=Test_Blocks_PreviewCommand`
Expected: FAIL — the org-grid workspace case currently produces `/block-kit-builder/E456/builder` instead of `T123`.

- [ ] **Step 3: Fix the discriminator**

In `cmd/blocks/preview.go`, replace `teamOrEnterpriseID`:

```go
// teamOrEnterpriseID returns the enterprise ID for enterprise installs and the
// team ID otherwise, matching the identifier used in Block Kit Builder URLs
func teamOrEnterpriseID(auth *types.SlackAuth) string {
	if auth.IsEnterpriseInstall {
		return auth.EnterpriseID
	}
	return auth.TeamID
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `make test testdir=cmd/blocks testname=Test_Blocks_PreviewCommand`
Expected: PASS (both the enterprise-install and org-grid-workspace cases).

- [ ] **Step 5: Format and commit**

```bash
gofmt -w cmd/blocks/preview.go cmd/blocks/preview_test.go
git add cmd/blocks/preview.go cmd/blocks/preview_test.go
git commit -m "$(cat <<'EOF'
fix: use enterprise id only for enterprise installs in blocks preview

An org-grid workspace install has an enterprise ID but is not an
enterprise install; key off IsEnterpriseInstall so those installs open
the Builder in the workspace context.

Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>
EOF
)"
```

---

### Task 5: Harden `buildBlockKitBuilderURL` against bad hosts

Guard against an empty or scheme-less host (which otherwise silently yields a malformed URL like `//app./...`) and wrap the `url.Parse` error with a structured code. Fixes findings #6 and #7.

**Files:**
- Modify: `cmd/blocks/preview.go:180-189`
- Test: `cmd/blocks/preview_test.go`

**Interfaces:**
- Produces: `buildBlockKitBuilderURL(apiHost string, id string, blocksJSON string) (string, error)` — unchanged signature; now errors on invalid hosts.

- [ ] **Step 1: Write the failing tests**

In `cmd/blocks/preview_test.go`, extend `Test_buildBlockKitBuilderURL` to cover error cases. Replace the test with:

```go
func Test_buildBlockKitBuilderURL(t *testing.T) {
	tests := map[string]struct {
		apiHost     string
		id          string
		blocksJSON  string
		expected    string
		expectedErr string
	}{
		"production host": {
			apiHost:    "https://slack.com",
			id:         "T123",
			blocksJSON: `{"blocks":[]}`,
			expected:   "https://app.slack.com/block-kit-builder/T123/builder#%7B%22blocks%22:%5B%5D%7D",
		},
		"developer host": {
			apiHost:    "https://dev1234.slack.com",
			id:         "E456",
			blocksJSON: `{"blocks":[]}`,
			expected:   "https://app.dev1234.slack.com/block-kit-builder/E456/builder#%7B%22blocks%22:%5B%5D%7D",
		},
		"empty host": {
			apiHost:     "",
			id:          "T123",
			blocksJSON:  `{"blocks":[]}`,
			expectedErr: slackerror.ErrInvalidArguments,
		},
		"scheme-less host": {
			apiHost:     "app.slack.com",
			id:          "T123",
			blocksJSON:  `{"blocks":[]}`,
			expectedErr: slackerror.ErrInvalidArguments,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := buildBlockKitBuilderURL(tc.apiHost, tc.id, tc.blocksJSON)
			if tc.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr, slackerror.ToSlackError(err).Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `make test testdir=cmd/blocks testname=Test_buildBlockKitBuilderURL`
Expected: FAIL — the empty and scheme-less host cases currently return no error.

- [ ] **Step 3: Add the host guard and wrap the parse error**

In `cmd/blocks/preview.go`, replace `buildBlockKitBuilderURL`:

```go
// buildBlockKitBuilderURL constructs the Block Kit Builder URL for the given
// API host, team or enterprise ID, and compact blocks JSON. The blocks JSON is
// placed in the URL fragment, which url.URL.String percent-encodes.
func buildBlockKitBuilderURL(apiHost string, id string, blocksJSON string) (string, error) {
	parsed, err := url.Parse(apiHost)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrInvalidArguments)
	}
	if parsed.Host == "" {
		return "", slackerror.New(slackerror.ErrInvalidArguments).
			WithMessage("The API host %q is not a valid URL", apiHost)
	}
	parsed.Host = "app." + parsed.Host
	parsed.Path = fmt.Sprintf("/block-kit-builder/%s/builder", id)
	parsed.Fragment = blocksJSON
	return parsed.String(), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `make test testdir=cmd/blocks testname=Test_buildBlockKitBuilderURL`
Expected: PASS (all four cases).

- [ ] **Step 5: Format and commit**

```bash
gofmt -w cmd/blocks/preview.go cmd/blocks/preview_test.go
git add cmd/blocks/preview.go cmd/blocks/preview_test.go
git commit -m "$(cat <<'EOF'
fix: reject invalid API hosts when building Block Kit Builder URL

Guard against empty or scheme-less hosts that would otherwise yield a
malformed URL, and wrap the url.Parse error with a structured code.

Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>
EOF
)"
```

---

### Task 6: Hide the experimental command and remove its generated docs

Hide the `blocks` command until the `block-kit-builder` experiment graduates, and remove the generated doc pages/index entry (docgen skips hidden commands via `IsAvailableCommand()`). Fixes finding #4.

**Files:**
- Modify: `cmd/blocks/blocks.go`
- Test: `cmd/blocks/blocks_test.go`
- Delete: `docs/reference/commands/slack_blocks.md` (currently untracked)
- Delete: `docs/reference/commands/slack_blocks_preview.md` (currently untracked)
- Modify: `docs/reference/commands/slack.md` (remove the added `slack blocks` index line)

**Interfaces:**
- Consumes: nothing new.
- Produces: `blocks` command has `Hidden: true`; example text updated to `--blocks` form.

- [ ] **Step 1: Write the failing test**

Replace `Test_Blocks_Command` in `cmd/blocks/blocks_test.go` with a version that asserts the command is hidden. Because a hidden command still prints its own help when invoked directly, the existing output expectations remain valid; add the hidden assertion via `ExpectedAsserts` is not suitable here (it inspects mocks, not the command), so assert on the constructed command directly in a small standalone test:

```go
func Test_Blocks_Command(t *testing.T) {
	testutil.TableTestCommand(t, testutil.CommandTests{
		"prints help without a subcommand": {
			Setup: func(t *testing.T, ctx context.Context, cm *shared.ClientsMock, cf *shared.ClientFactory) {
			},
			ExpectedOutputs: []string{
				"Work with Block Kit blocks",
				"preview",
			},
		},
	}, func(cf *shared.ClientFactory) *cobra.Command {
		return NewCommand(cf)
	})
}

func Test_Blocks_Command_Hidden(t *testing.T) {
	clientsMock := shared.NewClientsMock()
	clients := shared.NewClientFactory(clientsMock.MockClientFactory())
	cmd := NewCommand(clients)
	assert.True(t, cmd.Hidden, "blocks command should be hidden while experimental")
}
```

Add the import `"github.com/stretchr/testify/assert"` to `cmd/blocks/blocks_test.go`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `make test testdir=cmd/blocks testname=Test_Blocks_Command_Hidden`
Expected: FAIL — `cmd.Hidden` is false.

- [ ] **Step 3: Hide the command and update the example**

In `cmd/blocks/blocks.go`, add `Hidden: true` with an explanatory comment and update the example to the `--blocks` form:

```go
	cmd := &cobra.Command{
		Use:   "blocks <subcommand> [flags]",
		Short: "Work with Block Kit blocks",
		Long:  "Work with Block Kit blocks, such as previewing them in the Block Kit Builder.",
		// Hidden while gated behind the block-kit-builder experiment. Remove
		// when the experiment graduates.
		Hidden: true,
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{
				Meaning: "Preview blocks in the Block Kit Builder",
				Command: "blocks preview --blocks '[{\"type\":\"divider\"}]'",
			},
		}),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `make test testdir=cmd/blocks`
Expected: PASS (both `Test_Blocks_Command` and `Test_Blocks_Command_Hidden`).

- [ ] **Step 5: Remove the generated docs and index entry**

The two `slack_blocks*.md` pages are untracked; remove them from disk. Edit `slack.md` to drop the index line.

```bash
rm docs/reference/commands/slack_blocks.md docs/reference/commands/slack_blocks_preview.md
```

In `docs/reference/commands/slack.md`, remove this line:

```
* [slack blocks](slack_blocks)	 - Work with Block Kit blocks
```

Optional verification that docgen agrees (regenerates docs without the hidden command; requires a build):

```bash
make build-ci && ./bin/slack docgen ./docs/reference && git status docs/reference/commands
```
Expected: no `slack_blocks*.md` files reappear and `slack.md` has no `slack blocks` line. If docgen produces unrelated diffs, discard them (`git checkout -- <file>`); this task only owns the blocks-related doc changes.

- [ ] **Step 6: Format and commit**

```bash
gofmt -w cmd/blocks/blocks.go cmd/blocks/blocks_test.go
git add cmd/blocks/blocks.go cmd/blocks/blocks_test.go docs/reference/commands/slack.md
git rm --ignore-unmatch docs/reference/commands/slack_blocks.md docs/reference/commands/slack_blocks_preview.md
git commit -m "$(cat <<'EOF'
chore: hide experimental blocks command from help and docs

Hide the blocks command while gated behind the block-kit-builder
experiment and drop its generated doc pages, which docgen omits for
hidden commands.

Co-Authored-By: Claude <svc-devxp-claude@slack-corp.com>
EOF
)"
```

---

### Task 7: Full verification sweep

Confirm the whole feature builds, lints, and passes tests together, plus a manual smoke test.

**Files:** none (verification only).

- [ ] **Step 1: Run the full affected test packages**

Run:
```bash
make test testdir=cmd/blocks
make test testdir=internal/iostreams
make test testdir=internal/slackdeps
```
Expected: PASS for all.

- [ ] **Step 2: Lint**

Run: `make lint`
Expected: no findings in `cmd/blocks/`, `internal/iostreams/`, `internal/slackdeps/`, `internal/shared/types/`.

- [ ] **Step 3: Build**

Run: `make build-ci`
Expected: builds `./bin/slack` with no errors.

- [ ] **Step 4: Manual smoke test — flag path**

Run: `./bin/slack blocks preview --experiment block-kit-builder --blocks '[{"type":"divider"}]' --team <your-team-id>`
Expected: prints the Block Kit Builder section with an `https://app.<host>/block-kit-builder/<id>/builder#...` URL and attempts to open the browser.

- [ ] **Step 5: Manual smoke test — no hang on bare invocation**

Run: `./bin/slack blocks preview --experiment block-kit-builder` in an interactive terminal (no pipe).
Expected: returns immediately with "No blocks were provided" and remediation mentioning `--blocks` — does NOT hang waiting on stdin.

- [ ] **Step 6: Manual smoke test — piped stdin**

Run: `echo '[{"type":"divider"}]' | ./bin/slack blocks preview --experiment block-kit-builder --team <your-team-id>`
Expected: same successful output as Step 4.

- [ ] **Step 7: Confirm the command is hidden**

Run: `./bin/slack --help` and `./bin/slack --help --verbose`
Expected: `blocks` is absent from the standard command list.

---

## Notes for the implementer

- Findings mapped to tasks: #1 → Task 2; #2 → Task 4; #3 → Task 3; #4 → Task 6; #5 → Task 2; #6 → Task 5; #7 → Task 5. The stdin-TTY infrastructure (Task 1) is the enabler for #1.
- `normalizeBlocksPayload` is intentionally unchanged — its existing tests in `Test_normalizeBlocksPayload` stay green throughout.
- Tasks 2–6 all edit `cmd/blocks/preview.go` (except Task 6, which edits `blocks.go`); implement them in order so each commit is coherent.
