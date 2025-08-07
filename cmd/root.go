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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/slackapi/slack-cli/cmd/app"
	"github.com/slackapi/slack-cli/cmd/auth"
	"github.com/slackapi/slack-cli/cmd/collaborators"
	"github.com/slackapi/slack-cli/cmd/datastore"
	"github.com/slackapi/slack-cli/cmd/docgen"
	"github.com/slackapi/slack-cli/cmd/doctor"
	"github.com/slackapi/slack-cli/cmd/env"
	"github.com/slackapi/slack-cli/cmd/externalauth"
	"github.com/slackapi/slack-cli/cmd/feedback"
	"github.com/slackapi/slack-cli/cmd/fingerprint"
	"github.com/slackapi/slack-cli/cmd/function"
	"github.com/slackapi/slack-cli/cmd/help"
	"github.com/slackapi/slack-cli/cmd/manifest"
	"github.com/slackapi/slack-cli/cmd/openformresponse"
	"github.com/slackapi/slack-cli/cmd/platform"
	"github.com/slackapi/slack-cli/cmd/project"
	"github.com/slackapi/slack-cli/cmd/triggers"
	"github.com/slackapi/slack-cli/cmd/upgrade"
	versioncmd "github.com/slackapi/slack-cli/cmd/version"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/pkg/version"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackcontext"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/slackapi/slack-cli/internal/update"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type AliasInfo struct {
	CanonicalName  string
	CommandName    string
	ParentName     string
	CommandFactory func(*shared.ClientFactory) *cobra.Command
}

func (a *AliasInfo) SetCommandName(name string) {
	a.CommandName = name
}

var AliasMap = map[string]*AliasInfo{
	"activity":  {CommandFactory: platform.NewActivityCommand, CanonicalName: "platform activity", ParentName: "platform"},
	"create":    {CommandFactory: project.NewCreateCommand, CanonicalName: "project create", ParentName: "project"},
	"delete":    {CommandFactory: app.NewDeleteCommand, CanonicalName: "app delete", ParentName: "app"},
	"deploy":    {CommandFactory: platform.NewDeployCommand, CanonicalName: "platform deploy", ParentName: "platform"},
	"init":      {CommandFactory: project.NewInitCommand, CanonicalName: "project init", ParentName: "project"},
	"install":   {CommandFactory: app.NewAddCommand, CanonicalName: "app install", ParentName: "app"},
	"list":      {CommandFactory: auth.NewListCommand, CanonicalName: "auth list", ParentName: "auth"},
	"login":     {CommandFactory: auth.NewLoginCommand, CanonicalName: "auth login", ParentName: "auth"},
	"logout":    {CommandFactory: auth.NewLogoutCommand, CanonicalName: "auth logout", ParentName: "auth"},
	"run":       {CommandFactory: platform.NewRunCommand, CanonicalName: "platform run", ParentName: "platform"},
	"samples":   {CommandFactory: project.NewSamplesCommand, CanonicalName: "project samples", ParentName: "project"},
	"uninstall": {CommandFactory: app.NewUninstallCommand, CanonicalName: "app uninstall", ParentName: "app"},
}
var processName = cmdutil.GetProcessName()

func NewRootCommand(clients *shared.ClientFactory, updateNotification *update.UpdateNotification) *cobra.Command {
	// rootCmd is the base Cobra command and all subcommands, that are part of the cmd package,
	// will be are added to the root command.
	return &cobra.Command{
		Use:           fmt.Sprintf("%v <command> <subcommand> [flags]", processName),
		Short:         "Slack command-line tool",
		SilenceErrors: true, // we will bubble up and handle errors ourselves.
		SilenceUsage:  true, // don't post help/usage every time a command raises an error. Can be overridden by child commands.
		Example: style.ExampleCommandsf([]style.ExampleCommand{
			{Command: "login", Meaning: "Log in to your Slack account"},
			{Command: "create", Meaning: "Create a new Slack app"},
			{Command: "init", Meaning: "Initialize an existing Slack app"},
			{Command: "run", Meaning: "Start a local development server"},
			{Command: "deploy", Meaning: "Deploy to the Slack Platform"},
		}),
		Long: strings.Join([]string{
			`{{Emoji "sparkles"}}CLI to create, run, and deploy Slack apps`,
			"",
			`{{Emoji "books"}}Get started by reading the docs: {{LinkText "https://docs.slack.dev/tools/slack-cli"}}`,
		}, "\n"),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clients.IO.SetCmdIO(cmd)

			// Set user-invoked command (for metrics)
			clients.Config.Command = strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], " ")
			clients.Config.CommandCanonical = clients.Config.Command
			aliasedCommand, isAlias := AliasMap[clients.Config.Command]
			if isAlias {
				clients.Config.CommandCanonical = aliasedCommand.CanonicalName
			}

			// Set flag names provided by user (includes both root and subcommand flags, used for metrics)
			flagset := cmd.Flags()
			flagset.VisitAll(func(flag *pflag.Flag) {
				if flag.Changed {
					clients.Config.RawFlags = append(clients.Config.RawFlags, flag.Name)
				}
			})

			// Check for an CLI update in the background while the command runs
			updateNotification = update.New(clients, version.Get(), "SLACK_SKIP_UPDATE")
			updateNotification.CheckForUpdateInBackground(ctx, false)
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// TODO: since commands are moving to `*E` cobra lifecycle methods, this method may not be invoked if those earlier lifecycle methods return an error. Maybe move this to the cleanup() method below? but maybe this is OK, no need to prompt users to update if they encounter an error?
			// when the command is `slack update`
			return updateNotification.PrintAndPromptUpdates(cmd, version.Get())
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

}

// Init bootstraps the CLI process. Put things that do not rely on specific arguments or flags passed to the CLI in here. If you need flag/argument values, InitConfig below.
func Init(ctx context.Context) (*cobra.Command, *shared.ClientFactory) {
	// clients stores shared clients and configurations used across the commands and handlers
	var clients *shared.ClientFactory
	// updateNotification will check for an update in the background and print a message after the command runs
	var updateNotification *update.UpdateNotification

	clients = shared.NewClientFactory(shared.SetVersion(version.Raw()))
	rootCmd := NewRootCommand(clients, updateNotification)

	// Support `--version` by setting root command's `Version` and custom template.
	// Add a newline to `SetVersionTemplate` to display correctly on terminals.
	rootCmd.Version = version.Get()
	rootCmd.SetVersionTemplate(versioncmd.Template() + "\n")

	// Add subcommands (each subcommand may add their own child subcommands)
	// Please keep these sorted
	subCommands := []*cobra.Command{
		app.NewCommand(clients),
		auth.NewCommand(clients),
		collaborators.NewCommand(clients),
		datastore.NewCommand(clients),
		docgen.NewCommand(clients),
		env.NewCommand(clients),
		externalauth.NewCommand(clients),
		fingerprint.NewCommand(clients),
		function.NewCommand(clients),
		manifest.NewCommand(clients),
		openformresponse.NewCommand(clients),
		platform.NewCommand(clients),
		project.NewCommand(clients),
		triggers.NewCommand(clients),
		upgrade.NewCommand(clients),
		versioncmd.NewCommand(clients),
	}

	for _, subCommand := range subCommands {
		rootCmd.AddCommand(subCommand)
	}
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	topLevelCommands := []*cobra.Command{
		doctor.NewDoctorCommand(clients),
		feedback.NewFeedbackCommand(clients),
	}

	for _, topLevelCommand := range topLevelCommands {
		rootCmd.AddCommand(topLevelCommand)
	}

	// Shortcuts / "Aliases"
	// Is there a way to programmatically fetch the parent/package name?
	aliases := map[string]string{}
	for name, alias := range AliasMap {
		cmd := alias.CommandFactory(clients)
		rootCmd.AddCommand(cmd)
		AliasMap[name].SetCommandName(cmd.Name())
		aliases[alias.CommandName] = alias.ParentName
	}

	clients.Config.InitializeGlobalFlags(rootCmd)
	clients.Config.SetFlags(rootCmd)

	rootCmd.SetHelpFunc(help.HelpFunc(clients, aliases))

	// OnInitialize will execute before any root or child commands' Pre* methods.
	// This is a good place to house CLI bootup routines.
	cobra.OnInitialize(func() {
		err := InitConfig(ctx, clients, rootCmd)
		if err != nil {
			clients.IO.PrintError(ctx, err.Error())
			clients.Os.Exit(int(iostreams.ExitError))
		}
	})
	// Since we use the *E cobra lifecycle methods, OnFinalize is one of the few ways we can ensure something _always_ runs at the end of any command invocation, regardless if an error is raised or not during execution.
	cobra.OnFinalize(func() {
		cleanup(ctx, clients)
	})
	return rootCmd, clients
}

// InitConfig reads in config files and ENV variables if set and sets up the CLI for functioning. Executes _before_ any Pre* methods from the root or child commands, but after Cobra parses flags and command arguments.
// Put global CLI initialization routines that rely on flag and argument parsing in here please!
// TODO: consider using arguments for this function for certain dependencies, like working directory and other OS-specific strings, that OnInitialize above can provide during actual execution, but that we can override with test values for easier testing.
func InitConfig(ctx context.Context, clients *shared.ClientFactory, rootCmd *cobra.Command) error {
	// Get the current working directory (usually, but not always the project)
	workingDirPath, err := clients.Os.Getwd()
	if err != nil {
		workingDirPath = "."
	}

	// Set the execution directory path
	clients.Os.SetExecutionDir(workingDirPath)

	// Load environment variables
	if err := clients.Config.LoadEnvironmentVariables(); err != nil {
		return err
	}

	// Set custom system config directory
	if clients.Config.ConfigDirFlag != "" {
		clients.Config.SystemConfig.SetCustomConfigDirPath(clients.Config.ConfigDirFlag)
	}

	// Init color and formatting
	style.ToggleStyles(clients.IO.IsTTY() && !clients.Config.NoColor)
	style.ToggleSpinner(clients.IO.IsTTY() && !clients.Config.NoColor && !clients.Config.DebugEnabled)

	// Find and replace deprecated flags
	if err := clients.Config.DeprecatedFlagSubstitutions(rootCmd); err != nil {
		return err
	}

	// Set the preference
	trustSources, err := clients.Config.SystemConfig.GetTrustUnknownSources(ctx)
	if err != nil {
		return err
	}
	clients.Config.TrustUnknownSources = trustSources

	// Init clients that use flags
	clients.Config.APIHostResolved = clients.Auth().ResolveAPIHost(ctx, clients.Config.APIHostFlag, nil)
	clients.Config.LogstashHostResolved = clients.Auth().ResolveLogstashHost(ctx, clients.Config.APIHostResolved, clients.CLIVersion)

	// Init System ID
	if systemID, err := clients.Config.SystemConfig.InitSystemID(ctx); err != nil {
		clients.IO.PrintDebug(ctx, "Error initializing user-level config system_id: %s", err.Error())
	} else {
		// Used by Logstash
		// TODO(slackcontext) Consolidate storing SystemID to slackcontext
		clients.Config.SystemID = systemID
		// Used by OpenTracing
		ctx = slackcontext.SetSystemID(ctx, systemID)
		rootCmd.SetContext(ctx)
		// Debug logging
		clients.IO.PrintDebug(ctx, "system_id: %s", clients.Config.SystemID)
	}

	// Init Project ID, if current directory is a project
	if projectID, _ := clients.Config.ProjectConfig.InitProjectID(ctx, false); projectID != "" {
		// Used by Logstash
		// TODO(slackcontext) Consolidate storing ProjectID to slackcontext
		clients.Config.ProjectID = projectID
		// Used by OpenTracing
		ctx = slackcontext.SetProjectID(ctx, projectID)
		rootCmd.SetContext(ctx)
		// Debug logging
		clients.IO.PrintDebug(ctx, "project_id: %s", clients.Config.ProjectID)
	}

	// Init configurations
	clients.Config.LoadExperiments(ctx, clients.IO.PrintDebug)
	// TODO(slackcontext) Consolidate storing CLI version to slackcontext
	clients.Config.Version = clients.CLIVersion

	// The domain auths (token->domain) shouldn't change for the execution of the CLI so preload them into config!
	clients.Config.DomainAuthTokens = clients.Auth().MapAuthTokensToDomains(ctx)

	// Load the project CLI/SDK Configuration file
	if err = clients.InitSDKConfig(ctx, workingDirPath); err != nil {
		if !clients.Os.IsNotExist(err) {
			clients.IO.PrintDebug(ctx,
				"failed to initialize hook configurations: %s",
				strings.TrimSpace(err.Error()))
		}
	}

	// Initialize the project runtime
	err = clients.InitRuntime(ctx, workingDirPath)
	if err != nil {
		if err.(*slackerror.Error).Code != slackerror.ErrRuntimeNotSupported {
			return err
		}
	} else {
		clients.Config.RuntimeName = clients.Runtime.Name()
		clients.Config.RuntimeVersion = clients.Runtime.Version()
	}

	// Initialize .slackignore contents
	config.InitSlackIgnore()

	// Init debug log file with CLI Version, OS, SessionID, TraceID, SystemID, ProjectID, etc
	return clients.IO.InitLogFile(ctx)
}

// ExecuteContext sets up a cancellable context for use with IOStreams' interrupt channel.
// It listens for process interrupts and sends to IOStreams' GetInterruptChannel() for use in
// in communicating process interrupts elsewhere in the code.
func ExecuteContext(ctx context.Context, rootCmd *cobra.Command, clients *shared.ClientFactory) {
	// Derive a cancel context that is cancelled when the main execution is interrupted or cleaned up.
	// Sub-commands can register for the cleanup wait group with clients.CleanupWaitGroup.Add(1)
	// and listen for <-ctx.Done() to be notified when the main execution is interrupted, in order
	// to have a chance to cleanup. This is useful for long running processes and background goroutines,
	// such as the activity and upgrade commands.
	ctx, cancel := context.WithCancel(ctx)

	completedChan := make(chan bool, 1)      // completed is used for signalling an end to command
	exitChan := make(chan bool, 1)           // exit blocks the command from exiting until completed
	interruptChan := make(chan os.Signal, 1) // interrupt catches signals to avoid abrupt exits

	signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
	defer func() {
		signal.Stop(interruptChan)
		cancel()
	}()

	// Wait for either an interrupt signal (ctrl+c) or an explicit call to ctx.cancel()
	go func() {
		select {
		// Received interrupt signal, so cancel the context and cleanup
		case <-interruptChan:
			clients.IO.PrintInfo(ctx, false, "\n") // flush the CTRL + C character
			clients.IO.PrintDebug(ctx, "Got process interrupt signal, cancelling context")
			cancel()
			go func() {
				// Explicitly call cleanup as process interrupt handling below will break normal command execution/error handling
				clients.IO.SetExitCode(iostreams.ExitCancel)
				_ = clients.EventTracker.FlushToLogstash(ctx, clients.Config, clients.IO, iostreams.ExitCancel)
				clients.IO.PrintDebug(ctx, "Root waiting for cleanup waitgroup...")
				clients.CleanupWaitGroup.Wait()
				clients.IO.PrintDebug(ctx, "... root cleanup waitgroup done.")
				cleanup(ctx, clients)
				clients.IO.PrintDebug(ctx, "Exiting with cancel exit code.")
				clients.Os.Exit(int(iostreams.ExitCancel))
			}()
		// Received completed execution, so exit the process successfully
		case <-completedChan:
			exitChan <- true
		}

		// If we get a second interrupt, no matter what exit the process
		<-interruptChan
		clients.IO.PrintDebug(ctx, "Got second process interrupt signal, exiting the process")
		clients.Os.Exit(int(iostreams.ExitCancel))
	}()

	// The cleanup() method in the root command will invoke via `defer` from within Execute.
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		errMsg := err.Error()
		clients.EventTracker.SetErrorMessage(errMsg)
		if slackErr, ok := err.(*slackerror.Error); ok {
			clients.EventTracker.SetErrorCode(slackErr.Code)
		}
		if slackerror.Is(err, slackerror.ErrProcessInterrupted) {
			clients.IO.SetExitCode(iostreams.ExitCancel)
			clients.IO.PrintDebug(ctx, errMsg)
		} else {
			switch clients.IO.GetExitCode() {
			case iostreams.ExitOK:
				clients.IO.SetExitCode(iostreams.ExitError)
			}
			clients.IO.PrintError(ctx, errMsg)
		}
		defer clients.Os.Exit(int(clients.IO.GetExitCode()))
		completedChan <- true
	} else {
		completedChan <- true
	}

	<-exitChan
	_ = clients.EventTracker.FlushToLogstash(ctx, clients.Config, clients.IO, clients.IO.GetExitCode())
}

// cleanup is attached to the root command via cobra's OnFinalize event subscriber.
// It is invoked by cobra via `defer` when the Execute method above finishes - regardless of an error occurring or not.
func cleanup(ctx context.Context, clients *shared.ClientFactory) {
	clients.IO.PrintDebug(ctx, "Starting root command cleanup routine")
	// clean up any json in project .slack folder if needed
	if sdkConfigExists, _ := clients.SDKConfig.Exists(); sdkConfigExists {
		clients.AppClient().CleanUp()
	}
}
