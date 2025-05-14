// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

# Adding command options

This guide explains how command options (flags) are used in the Slack CLI project and provides step-by-step instructions for adding a new option to an existing command.

## Table of Contents
1. @Understanding Command Options in Slack CLI
2. @How Command Options are Defined
3. @Step-by-Step Guide for Adding a New Option
4. @Testing Your New Option
5. @Best Practices

## Understanding Command Options in Slack CLI

The Slack CLI uses the @Cobra library for command-line functionality. Command options (or flags) provide a way to modify the behavior of a command. For example, the `platform run` command includes options like `--activity-level` to specify the logging level, or `--cleanup` to uninstall the local app after exiting.

There are two main types of flags in Cobra:

1. **Persistent Flags**: Available to the command they're assigned to, as well as all its sub-commands.
   ```go
   rootCmd.PersistentFlags().StringVar(&config.APIHostFlag, "apihost", "", "Slack API host")
   ```

2. **Local Flags**: Only available to the specific command they're assigned to.
   ```go
   cmd.Flags().BoolVar(&runFlags.cleanup, "cleanup", false, "uninstall the local app after exiting")
   ```

## How Command Options are Defined

In the Slack CLI project, command options follow a consistent pattern:

1. **Flag Storage**: Each command package defines a struct to store flag values.
   ```go
   type runCmdFlags struct {
       activityLevel       string
       noActivity          bool
       cleanup             bool
       hideTriggers        bool
       orgGrantWorkspaceID string
   }
   
   var runFlags runCmdFlags
   ```

2. **Flag Definition**: Options are defined in the command's constructor function.
   ```go
   cmd.Flags().BoolVar(&runFlags.cleanup, "cleanup", false, "uninstall the local app after exiting")
   ```

3. **Flag Usage**: The flag values are accessed in the command's run function through the struct variables.
   ```go
   runArgs := platform.RunArgs{
       Activity:            !runFlags.noActivity,
       ActivityLevel:       runFlags.activityLevel,
       Cleanup:             runFlags.cleanup,
       // ...
   }
   ```

4. **Flag Helpers**: The `cmdutil` package provides helper functions for working with flags.
   ```go
   cmdutil.IsFlagChanged(cmd, "flag-name")
   ```

## Step-by-Step Guide for Adding a New Option

Let's add a new flag called `--watch-ignore` to the `platform run` command to specify patterns to ignore while watching for changes.

### Step 1: Update the Flag Storage Struct

Locate the command's flag struct in the command file: @run.go

```go
type runCmdFlags struct {
    activityLevel       string
    noActivity          bool
    cleanup             bool
    hideTriggers        bool
    orgGrantWorkspaceID string
    watchIgnore         []string // New flag for patterns to ignore
}
```

### Step 2: Define the Flag in the Command Constructor

Add the flag definition in the `NewRunCommand` function:

```go
func NewRunCommand(clients *shared.ClientFactory) *cobra.Command {
    // ... existing code
    
    // Add flags
    cmd.Flags().StringVar(&runFlags.activityLevel, "activity-level", platform.ActivityMinLevelDefault, "activity level to display")
    cmd.Flags().BoolVar(&runFlags.noActivity, "no-activity", false, "hide Slack Platform log activity")
    cmd.Flags().BoolVar(&runFlags.cleanup, "cleanup", false, "uninstall the local app after exiting")
    cmd.Flags().StringVar(&runFlags.orgGrantWorkspaceID, cmdutil.OrgGrantWorkspaceFlag, "", cmdutil.OrgGrantWorkspaceDescription())
    cmd.Flags().BoolVar(&runFlags.hideTriggers, "hide-triggers", false, "do not list triggers and skip trigger creation prompts")
    
    // Add the new flag
    cmd.Flags().StringSliceVar(&runFlags.watchIgnore, "watch-ignore", nil, "patterns to ignore while watching for changes")
    
    // ... rest of the function
}
```

### Step 3: Update the Command's Run Function

Modify the `RunRunCommand` function to use the new flag:

```go
func RunRunCommand(clients *shared.ClientFactory, cmd *cobra.Command, args []string) error {
    // ... existing code
    
    runArgs := platform.RunArgs{
        Activity:            !runFlags.noActivity,
        ActivityLevel:       runFlags.activityLevel,
        App:                 selection.App,
        Auth:                selection.Auth,
        Cleanup:             runFlags.cleanup,
        ShowTriggers:        triggers.ShowTriggers(clients, runFlags.hideTriggers),
        OrgGrantWorkspaceID: runFlags.orgGrantWorkspaceID,
        WatchIgnore:         runFlags.watchIgnore, // Pass the new flag value
    }
    
    // ... rest of the function
}
```

### Step 4: Update the RunArgs Struct

Update the `RunArgs` struct in the internal platform package: @run.go

```go
type RunArgs struct {
    Activity            bool
    ActivityLevel       string
    App                 types.App
    Auth                types.SlackAuth
    Cleanup             bool
    ShowTriggers        bool
    OrgGrantWorkspaceID string
    WatchIgnore         []string // New field for ignore patterns
}
```

### Step 5: Use the New Flag Value in the Run Function

Update the `Run` function in the platform package to use the new flag value:

```go
func Run(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, runArgs RunArgs) (*logger.LogEvent, types.InstallState, error) {
    // ... existing code
    
    watchOptions := &watcher.Options{
        IgnorePatterns: runArgs.WatchIgnore, // Use the new flag value
    }
    
    // ... update the watcher setup to use the options
}
```

### Step 6: Update Command Examples

Add an example for the new flag in the command constructor:

```go
Example: style.ExampleCommandsf([]style.ExampleCommand{
    {Command: "platform run", Meaning: "Start a local development server"},
    {Command: "platform run --activity-level debug", Meaning: "Run a local development server with debug activity"},
    {Command: "platform run --cleanup", Meaning: "Run a local development server with cleanup"},
    {Command: "platform run --watch-ignore '**/node_modules/**'", Meaning: "Ignore node_modules while watching for changes"},
}),
```

## Testing Your New Option

For proper test coverage of your new flag, you need to:

1. Update existing tests
2. Add new test cases

### Step 1: Update Existing Test Cases

In `cmd/platform/run_test.go`, update the test cases to include the new flag:

```go
func TestRunCommand_Flags(t *testing.T) {
    tests := map[string]struct {
        cmdArgs         []string
        appFlag         string
        tokenFlag       string
        selectedAppAuth prompts.SelectedApp
        selectedAppErr  error
        expectedRunArgs platform.RunArgs
        expectedErr     error
    }{
        // ... existing test cases
        
        "Run with watch-ignore flag": {
            cmdArgs: []string{"--watch-ignore", "**/node_modules/**,**/dist/**"},
            selectedAppAuth: prompts.SelectedApp{
                App:  types.NewApp(),
                Auth: types.SlackAuth{},
            },
            expectedRunArgs: platform.RunArgs{
                Activity:      true,
                ActivityLevel: "info",
                Auth:          types.SlackAuth{},
                App:           types.NewApp(),
                Cleanup:       false,
                ShowTriggers:  true,
                WatchIgnore:   []string{"**/node_modules/**", "**/dist/**"}, // Check the flag is passed through
            },
            expectedErr: nil,
        },
    }
    
    // ... test implementation
}
```

### Step 2: Add Tests for the Platform Package

Update the tests in the platform package (`internal/pkg/platform/run_test.go`) to test that the flag is used correctly:

```go
func TestRun_WatchIgnore(t *testing.T) {
    ctx := slackcontext.MockContext(context.Background())
    clientsMock := shared.NewClientsMock()
    clientsMock.AddDefaultMocks()
    
    // Create test instance
    clients := shared.NewClientFactory(clientsMock.MockClientFactory())
    logger := logger.New(func(event *logger.LogEvent) {})
    
    // Test with ignore patterns
    runArgs := platform.RunArgs{
        App:         types.NewApp(),
        Auth:        types.SlackAuth{},
        WatchIgnore: []string{"**/node_modules/**", "**/dist/**"},
    }
    
    // Run the function (may need to adapt to your testing approach)
    _, _, err := platform.Run(ctx, clients, logger, runArgs)
    
    // Assert that the ignore patterns were used correctly
    // (how exactly depends on your implementation)
    require.NoError(t, err)
    // Add specific assertions about how the patterns should have been used
}
```

### Step 3: Test Help Text

Also test that the help text for the command includes the new flag:

```go
func TestRunCommand_Help(t *testing.T) {
    clients := shared.NewClientFactory()
    cmd := NewRunCommand(clients)
    
    var buf bytes.Buffer
    cmd.SetOut(&buf)
    err := cmd.Help()
    
    require.NoError(t, err)
    helpText := buf.String()
    
    assert.Contains(t, helpText, "--watch-ignore")
    assert.Contains(t, helpText, "patterns to ignore while watching for changes")
}
```

## Best Practices

When adding new command options, follow these best practices:

1. **Meaningful Names**: Choose clear, descriptive flag names.
   - Good: `--watch-ignore`
   - Avoid: `--wignore` or `--wi`

2. **Consistent Naming**: Follow existing naming patterns.
   - Use kebab-case for flag names (e.g., `--org-workspace-grant`).
   - Use camelCase for flag variables (e.g., `orgGrantWorkspaceID`).

3. **Good Descriptions**: Write clear, concise descriptions.
   - Use sentence fragments without ending periods.
   - If needed, use `\n` to add line breaks for complex descriptions.

4. **Appropriate Flag Types**: Choose the right type for each flag.
   - For simple on/off settings, use `BoolVar`.
   - For text values, use `StringVar`.
   - For lists, use `StringSliceVar`.
   - For numbers, use `IntVar` or `Float64Var`.

5. **Default Values**: Set sensible default values if applicable.
   - For optional flags, consider what happens when the flag is not provided.
   - Document default values in the help text.

6. **Examples**: Update command examples to showcase the new flag.
   - Include realistic examples of how the flag might be used.

7. **Thorough Testing**: Test all combinations and edge cases.
   - Test without the flag (default behavior).
   - Test with the flag set.
   - Test with invalid values, if applicable. 

8. **Updating documentation**: Ensure you update any related documentation in  `/docs` 
   - Check for any guides, tutorials, reference docs, or other documentation that may use the command.
   - Ensure it's updated to include behavioral changes as well as any API changes.
   - Follow existing docs patterns and best practices.


By following these steps and best practices, you can successfully add a new command option to the Slack CLI that integrates well with the existing codebase and provides value to users.


