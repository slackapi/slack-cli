# VHS Tape DSL Reference

## Output

- `Output <path>.gif` - Create a GIF output at the given path

## Require

- `Require <string>` - Ensure a program is on the $PATH to proceed

## Settings

- `Set FontSize <number>` - Set the font size of the terminal
- `Set FontFamily <string>` - Set the font family of the terminal
- `Set Height <number>` - Set the height of the terminal
- `Set Width <number>` - Set the width of the terminal
- `Set LetterSpacing <float>` - Set the font letter spacing (tracking)
- `Set LineHeight <float>` - Set the font line height
- `Set LoopOffset <float>%` - Set the starting frame offset for the GIF loop
- `Set Theme <json|string>` - Set the theme of the terminal
- `Set Padding <number>` - Set the padding of the terminal
- `Set Framerate <number>` - Set the framerate of the recording
- `Set PlaybackSpeed <float>` - Set the playback speed of the recording
- `Set MarginFill <file|#000000>` - Set the file or color the margin will be filled with
- `Set Margin <number>` - Set the size of the margin (no effect without MarginFill)
- `Set BorderRadius <number>` - Set terminal border radius, in pixels
- `Set WindowBar <string>` - Set window bar type (Rings, RingsRight, Colorful, ColorfulRight)
- `Set WindowBarSize <number>` - Set window bar size, in pixels (default: 40)
- `Set TypingSpeed <time>` - Set the typing speed of the terminal (default: 50ms)

## Sleep

- `Sleep <time>` - Sleep for a set amount of time in seconds

## Type

- `Type "<characters>"` - Type characters into the terminal
- `Type@<time> "<characters>"` - Type characters with a custom delay between each character

## Keys

All keys accept an optional `@<time>` delay and an optional repeat `[number]`.

- `Escape` - Press the Escape key
- `Backspace` - Press the Backspace key
- `Delete` - Press the Delete key
- `Insert` - Press the Insert key
- `Enter` - Press the Enter key
- `Space` - Press the Space key
- `Tab` - Press the Tab key
- `Up` / `Down` / `Left` / `Right` - Press arrow keys
- `PageUp` / `PageDown` - Press Page Up/Down keys
- `Ctrl+<key>` - Press Control + key (e.g. `Ctrl+C`)

## Display

- `Hide` - Hide subsequent commands from the output
- `Show` - Show subsequent commands in the output
