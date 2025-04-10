## slack trigger update

Updates an existing trigger

### Synopsis

Updates an existing trigger with the provided definition. Only supports full replacement, no partial update.

```
slack trigger update --trigger-id <id> [flags]
```

### Examples

```
# Update a trigger definition with a selected file
$ slack trigger update --trigger-id Ft01234ABCD

# Update a trigger with a workflow id and title
$ slack trigger update --trigger-id Ft01234ABCD \
    --workflow "#/workflows/my_workflow" --title "Updated trigger"
```

### Options

```
      --description string          the description of this trigger
  -h, --help                        help for update
      --interactivity               when used with --workflow, adds a
                                      "slack#/types/interactivity" parameter
                                      to the trigger with the name specified
                                      by --interactivity-name
      --interactivity-name string   when used with --interactivity, specifies
                                      the name of the interactivity parameter
                                      to use (default "interactivity")
      --title string                the title of this trigger
                                       (default "My Trigger")
      --trigger-def string          path to a JSON file containing the trigger
                                      definition. Overrides other flags setting
                                      trigger properties.
      --trigger-id string           the ID of the trigger to update
      --workflow string             a reference to the workflow to execute
                                      formatted as:
                                      "#/workflows/<workflow_callback_id>"
```

### Options inherited from parent commands

```
  -a, --app string           use a specific app ID or environment
      --config-dir string    use a custom path for system config directory
  -e, --experiment strings   use the experiment(s) in the command
  -f, --force                ignore warnings and continue executing command
      --no-color             remove styles and formatting from outputs
  -s, --skip-update          skip checking for latest version of CLI
  -w, --team string          select workspace or organization by team name or ID
      --token string         set the access token associated with a team
  -v, --verbose              print debug logging and additional info
```

### SEE ALSO

* [slack trigger](slack_trigger)	 - List details of existing triggers

