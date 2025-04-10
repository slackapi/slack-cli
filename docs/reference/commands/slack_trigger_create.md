## slack trigger create

Create a trigger for a workflow

### Synopsis

Create a trigger to start a workflow

```
slack trigger create [flags]
```

### Examples

```
# Create a trigger by selecting an app and trigger definition
$ slack trigger create

# Create a trigger from a definition file
$ slack trigger create --trigger-def "triggers/shortcut_trigger.ts"

# Create a trigger for a workflow
$ slack trigger create --workflow "#/workflows/my_workflow"
```

### Options

```
      --description string           the description of this trigger
  -h, --help                         help for create
      --interactivity                when used with --workflow, adds a
                                       "slack#/types/interactivity" parameter
                                       to the trigger with the name specified
                                       by --interactivity-name
      --interactivity-name string    when used with --interactivity, specifies
                                       the name of the interactivity parameter
                                       to use (default "interactivity")
      --org-workspace-grant string   grant access to a specific org workspace ID
                                       (or 'all' for all workspaces in the org)
      --title string                 the title of this trigger
                                       (default "My Trigger")
      --trigger-def string           path to a JSON file containing the trigger
                                       definition. Overrides other flags setting
                                       trigger properties.
      --workflow string              a reference to the workflow to execute
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

