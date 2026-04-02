# Style Guide

A current suggestion for how Slack CLI inputs are handled and outputs are formatted.

- **Input**
  - [Prompts are Flags with Forms](#prompts-are-flags-with-forms)
- **Output**
  - [Help Descriptions find Complete Sentences](#help-descriptions-find-complete-sentences)
  - [Section Formats with Command Headings](#section-formats-with-command-headings)

## Input

Customizations to commands are made through arguments, flags, environment variables, saved files, details from the Slack API itself, or sometimes just kept as "sensible" defaults.

### Prompts are Flags with Forms

When information is needed we can prompt for text, confirmation, or a selection.

These decisions can be made in an interactive terminal (TTY) or not, such as in a scripting environment.

A flag option should exist for each prompt with a form fallback. Either default values should be used if forms are attempted in a non-TTY setup or an error and remmediation to use a flag should be returned.

## Output

Results of a command go toward informing current happenings and suggesting next steps.

### Help Descriptions find Complete Sentences

The output of extended help descriptions should be complete sentences:

```txt
$ slack docs search --help
Search the Slack developer docs and return results in text, JSON, or browser
format.
```

This example uses punctuation and breaks lines at or before the 80 character count.

### Section Formats with Command Headings

A command often prints information and details about the process happenings. We format this as a section:

```txt
📚 App Install
   Installing "focused-lamb-99" app to "devrelsandbox"
   Finished in 2.0s
```

This example highlights some recommendations:

- An emoji is used with the section header.
- The section header text is the command name, with "Title Case" letters.
- Following details reveal progress of the process.
