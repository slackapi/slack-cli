---
name: record-demo
description: Record terminal demos as GIFs using VHS tape files. Use when asked to record a demo, create a terminal recording, or generate a GIF of CLI usage.
argument-hint: <description>
---

Record a demo of `$ARGUMENTS` using VHS.

1. **Check VHS is installed**: Run `which vhs`. If not found, tell the user to install it with `brew install vhs`.

2. **Create a `.tape` file** in `demos/` (create the directory if needed). Use this template:

   ```tape
   # Demo: <description>
   Output demos/<name>.gif    # Always use .gif format

   Set Shell "zsh"
   Set FontSize 16
   Set Width 1200
   Set Height 600
   Set Theme "Catppuccin Mocha"
   Set TypingSpeed 75ms
   Set Padding 20

   Type "<command>"
   Sleep 500ms
   Enter
   Sleep 3s
   ```

3. **VHS Tape DSL reference**: For the full command reference, see [vhs-reference.md](vhs-reference.md).

4. **Tips for good demos**:
   - Add `Sleep` after commands to let output render and be readable
   - Use `Hide`/`Show` to run setup commands invisibly
   - Keep demos short and focused (under 30 seconds)
   - Add a `Sleep 3s` at the end so viewers can see final output
   - Always add a `Sleep 500ms` pause after each `Type` line (before `Enter`) to mimic a human pause
   - For this project, use `./bin/slack` directly as the CLI binary path (do NOT export PATH)
   - When demoing from an app directory, use `Hide` to `cd` into the app, then use `../bin/slack` as the binary path
   - Prefer passing app IDs directly (e.g. `--app A12345`) to avoid interactive prompts when the ID is known
5. **Run the recording**: Execute `vhs <filename>.tape` to generate the output file.

6. **Review**: Open the `demos/` directory with `open demos/` so the user can review the output.
