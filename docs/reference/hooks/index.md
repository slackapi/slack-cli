# Hooks

Communication between the CLI and the application SDK is managed by a project-level configuration file called `hooks.json`. This file is included in our app templates and defines script _hooks_.

Hooks are small scripts that are _executed_ by the CLI and _implemented_ by the SDK. These commands perform actions on the project.

The `hooks.json` file allows the CLI and SDK a standard way to communicate while remaining decoupled and abstracted. This interface is a key design of the Slack CLI: many application and project level tasks are delegated from the CLI to the SDK. This delegation, decoupling and abstraction allows for language-agnostic SDK implementations.

## Core concepts {#core-concepts}

### Hooks: How the CLI and the SDK communicate {#communication}

Hooks are scripts that execute when a specific event happens or a Slack CLI command is invoked.

A hook script may be triggered when:

- generating the app manifest
- bundling function code before deployment to Slack infrastructure
- handling an application event during local development runs

When an event occurs, the CLI will execute a hook script by spawning a separate process, possibly passing a JSON object through `STDIN` and/or other parameters via command line flags to the hook script process, and waiting for a JSON response via the spawned process’ `STDOUT`. This system is heavily inspired by git hooks.

Since communication over hooks involves inter-process communication (one process running the CLI communicating with another process managed by the SDK ie. the hook), the exit code of the hook process signals success or failure to the CLI process. The CLI will assume an exit code of 0 from a hook process signals success, while any other exit code signals failure.

Some hooks may return data as part of their functionality. The CLI will use the `STDOUT` and `STDERR` of the hook process to transmit its response. For details on how a hook process can shape its response, and delineate diagnostic data from response data, see the section on [protocol negotiation](#protocol).

### Discover hook scripts and default configuration with `get-hooks` {#discover}

In order for the CLI to reliably discover the hooks for the [Deno SDK](https://github.com/slackapi/deno-slack-sdk), [Bolt Frameworks](https://docs.slack.dev/tools/), and future community-driven SDKs, the CLI employs a service-discovery-like approach to querying the SDK for what functionality it supports.

A project includes a `hooks.json` file in its `.slack` directory which by default contains a single hook (`get-hooks`). The SDK is responsible for implementing this hook and returning to the CLI a JSON object with all hook definitions and their relevant default configuration (one hook to rule them all 💍).

App developers do not need to edit or change their `hooks.json` file when upgrading their SDK because the interface contents are dictated by the SDK. If an App Developer would like to override specific hooks, they can do so by overriding the relevant hook section itself within `hooks.json` with their own implementation. More details on overriding behaviour can be found in the [Hook resolution](#hook-resolution) section.

Refer to the [CLI-SDK JSON interface](#interface-format) section for other examples.

### CLI-SDK protocol negotiation {#protocol}

As the needs of app developers evolve, so will the interface and the rules of communication between the CLI and the SDK. These rules are negotiated via the initial `get-hooks` handshake and are specified via the `protocol-version` field returned by the SDK.

At the time of writing, only two protocol versions are supported:

1. `v1`: The base and default protocol
2. `v2`: The second revision of the protocol, implementing `message boundaries`. This enables delineating responses to hook invocations from diagnostic/additional data such as logging.

If at any point protocol negotiation fails or does not adhere to the rules of communication, the CLI will fall back to using the default protocol.

#### Working implementations of protocol negotiation

- In the CLI:
  - [List of protocols supported by the CLI](https://github.com/slackapi/slack-cli/blob/d2349b6328820d2dcb01312abd4d8b3694f5137e/internal/hooks/protocol.go#L21-L22)
  - [CLI protocol negotiation](https://github.com/slackapi/slack-cli/blob/d2349b6328820d2dcb01312abd4d8b3694f5137e/internal/hooks/sdk_config.go#L56-L68)
  - [CLI implementation of v2 protocol 'message-boundaries'](https://github.com/slackapi/slack-cli/blob/d2349b6328820d2dcb01312abd4d8b3694f5137e/internal/hooks/hook_executor_v2.go#L41)
  - [CLI implementation of the default/v1 protocol](https://github.com/slackapi/slack-cli/blob/d2349b6328820d2dcb01312abd4d8b3694f5137e/internal/hooks/hook_executor_default.go#L34)
- SDK hooks:
  - [deno-slack-sdk’s implementation](https://github.com/slackapi/deno-slack-protocols/blob/main/src/mod.ts)
  - [node-slack-sdk’s implementation](https://github.com/slackapi/node-slack-sdk/blob/main/packages/cli-hooks/src/protocols.js)
  - [python-slack-sdk’s implementation](https://github.com/slackapi/python-slack-hooks/blob/main/slack_cli_hooks/protocol/__init__.py)

### Ensuring backwards compatibility {#compatibility}

A hook’s name space (CLI) and its associated script implementation (SDK) will change over time. This can break backwards compatibility and require App Developers to juggle different CLI versions and SDK versions in order to maintain compatibility. It’s a frustrating situation that can ruin the developer experience.

An additive approach to hook names or configuration settings allows us to keep hooks backwards-compatible for as long as possible and allows for a smoother upgrade experience. This approach also allows for tools to provide generous timeframes for supporting old hooks vs. new ones, allowing for deprecation windows and gradual rollouts. For configuration settings, an additive approach is accomplished by adding new configuration values that are not Golang defaults (e.g. bool defaults to false).

For example, the hook name `run-v2` may be the successor to the hook named `run`. The SDK can implement either hook and the CLI will trigger the latest version, possibly falling back to earlier versions of the hook where applicable. The CLI can also warn of impending removal of older hooks, providing hints to the developer when tooling behaviour changes.

## Hook specification {#specification}

Hooks are entry points for the CLI to initiate inter-process communication with the SDK. SDKs should implement one hook: `get-hooks`. It is the recommended approach for SDKs to enable communication with the CLI (for more details, see the [Hook Resolution](#hook-resolution) section).

A hook encapsulates an isolated piece of functionality that the SDK provides to app developers; the CLI delegates execution of the functionality to the SDK via hook invocation in a separate process. SDKs are responsible for ensuring that hook processes exit with a status code of 0; otherwise the CLI will surface the `sdk_hook_invocation_failed` error to the end-user and report the hook process’ `STDOUT` and `STDERR` to the CLI process’ `STDOUT`.

## Hooks list {#hooks-list}

The following Slack CLI commands invoke hooks:

| Command                | Hook(s)                                                                 |
| ---------------------- | ----------------------------------------------------------------------- |
| `slack deploy`         | [`get-manifest`](#get-manifest), [`build`](#build), [`deploy`](#deploy) |
| `slack doctor`         | [`check-update`](#check-update), [`doctor`](#doctor)                    |
| `slack manifest info`  | [`get-manifest`](#get-manifest)                                         |
| `slack run`            | [`get-manifest`](#get-manifest), [`start`](#start)                      |
| `slack trigger create` | [`get-trigger`](#get-trigger)                                           |
| `slack upgrade`        | [`check-update`](#check-update), [`install-update`](#install-update)    |

More details on these hooks can be found in their dedication sections below.

### `get-hooks` (required) {#get-hooks}

Implementing this hook allows for the SDK to control what hooks, and therefore features, it implements and allows for interoperability with the CLI.

For more details on how this hook is resolved, looked up, and how it interacts with resolution of other hooks, please see the [Hook resolution](#hook-resolution) section.

This hook acts as the very first interaction between the CLI and SDK, and thus special rules apply. It is invoked before any other hook and therefore its implementation should be performant. It does not support sending diagnostic information.

#### Output

This hook should return the CLI-SDK interface in [JSON](#interface-format) format via `STDOUT`.

#### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ✅ Supported        | ✅ Supported    | ❌ Unsupported |

### `get-manifest` (optional) {#get-manifest}

This hook allows for the application under development to be created on `api.slack.com/apps` as well as installed to workspaces that the CLI has been authorized to.

Implementing this hook signals to the CLI that the SDK manages the [application manifest](https://docs.slack.dev/app-manifests/).

#### Output

The [app manifest](https://docs.slack.dev/reference/app-manifest) in JSON format to `STDOUT`.

#### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ✅ Supported        | ✅ Supported    | ❌ Unsupported |

### `build` (optional) {#build}

Implementing this hook allows for the CLI to [deploy function code to Slack's managed infrastructure](https://docs.slack.dev/tools/deno-slack-sdk/guides/deploying-to-slack/). The work of assembling the application bundle according to Slack's application bundle format is delegated to the SDK via this hook.

The application bundle format has few restrictions, but the critical ones are:

- It must have a `manifest.json` at the root of the package
- This manifest must contain a `functions` key which is an object, with the keys of this object being each custom developer-defined function's `callback_id`. The contents of this object follow the standard app manifest format.
- Each custom function's source code should be bundled into a single `.js` file with its `callback_id` as its filename, and exist inside a `functions/` subdirectory within the app bundle. For example, a custom function with a callback_id of `greeting_function` should bundle all of its code, including dependencies, into a `greeting_function.js` file that would exist at `functions/greeting_function.js` within the application bundle.

The above requirements come from the [deno-slack-runtime](https://github.com/slackapi/deno-slack-runtime) project, which implements the expected Slack deployment bundle format. It contains a hard-coded [reference](https://github.com/slackapi/deno-slack-runtime/blob/main/src/mod.ts#L73) to the above-mentioned `functions/` sub-directory, and combines it with the [specific custom function `callback_id`](https://github.com/slackapi/deno-slack-runtime/blob/main/src/mod.ts#L17-L19) to resolve an import path for userland function source code.

:::warning[This hook should only be implemented by official Slack SDKs and is only relevant to apps [deployed to Slack's managed infrastructure](https://docs.slack.dev/tools/deno-slack-sdk/guides/deploying-to-slack/).]
:::

#### Input

Two command-line flags are provided as part of the hook invocation:

- `--source`: the path to the root of the application project directory
- `--output`: the path to a temporary directory that the SDK should output the application bundle to, as per the application bundle format

#### Output

No further output over STDOUT required from the SDK other than writing the application bundle to the filesystem location specified by the `--output` flag.

#### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ❌ Unsupported      | ❌ Unsupported  | ❌ Unsupported |

### `start` (optional) {#start}

The `slack run` CLI command will invoke the `start` hook to initiate a local-run development mode allowing for quick iteration during app development. This hook is responsible for actually running the app, but has two operating modes: one where the hook manages the connection to Slack via [Socket Mode](https://docs.slack.dev/apis/events-api/using-socket-mode/), and one where it does not. Which mode should be employed by the CLI when invoking this hook is dictated by the `config.sdk-managed-connection-enabled` property in the response from [`get-hooks`](#get-hooks).

#### Non-SDK-managed connection {#non-sdk-managed-connection}

This section applies when `config.sdk-managed-connection-enabled` is undefined or set to `false` in `hooks.json` or in the [`get-hooks`](#get-hooks) hook response.

When the app developer wants to initiate a development-mode local run of their application via the `slack run` CLI command, by default the CLI will create a [Socket Mode](https://docs.slack.dev/apis/events-api/using-socket-mode/) connection to the Slack backend and start listening for events on behalf of the app.

Any events coming down the wire from Slack will be fed over `STDIN` to this hook for the SDK to process. _Each event incoming from Slack will invoke this hook independently_, meaning one `start` hook process will be spawned per incoming event. The SDK should process each incoming event and output a JSON object to `STDOUT` representing the response to send back to Slack over the socket connection managed by the CLI.

##### Input

Several parameters are passed to the SDK over `STDIN` as JSON. The format of this JSON payload is as follows:

```json
{
  "body": {
    "type": "app_deleted",
    ...
  },
  "context": {
    "bot_access_token": "xoxb-1234",
    "app_id": "A1234",
    "team_id": "T1234",
    "variables": {}
  }
}
```

| Field                    | Description                                                                                                                                                                                                                             | Required |
| ------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| body                     | Object whose keys represent the incoming Slack event payload as described in [Events API Event Types](https://docs.slack.dev/reference/events?APIs=Events). The particular content of this key is dependent on the incoming event type. | Required |
| context                  | Object representing variables relevant for the locally-running application.                                                                                                                                                             | Required |
| context.bot_access_token | String; a bot access token that the SDK can provide to developer functions for issuing calls to the Slack API.                                                                                                                          | Required |
| context.app_id           | String; the current application ID.                                                                                                                                                                                                     | Required |
| context.team_id          | String; the ID of the team or workspace the locally-running app is installed to.                                                                                                                                                        | Required |
| context.variables        | Object containing environment variables that may or may not be defined by the application via a `.env` file in the root of the application directory.                                                                                   | Required |

Note: The CLI provides the `SLACK_APP_TOKEN` and `SLACK_BOT_TOKEN` environment variables if the respective scopes are requested.

##### Output

Each incoming event from the socket connection will invoke this hook separately. As such, this hook's response to `STDOUT` should be the JSON response to the Slack event sent to the app. The CLI will handle [acknowledging events](https://docs.slack.dev/apis/events-api/using-socket-mode/#acknowledge) by sending back the proper `envelope_id` attribute to the Slack backend. Therefore, the SDK’s `start` hook response `STDOUT` should be the `payload` of the response and nothing else. It is recommended to use the `v2` (`message-boundaries`) [protocol](#protocol) to more easily delineate logging/diagnostics from event responses to be sent to Slack.

##### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ❌ Unsupported      | ❌ Unsupported  | ❌ Unsupported |

#### SDK-managed connection {#sdk-managed-connection}

This section applies when `config.sdk-managed-connection-enabled` is set to `true` in `hooks.json` or in the [`get-hooks`](#get-hooks) hook response.

Implementing the `start` hook with `config.sdk-managed-connection-enabled` set to `true` will instruct the CLI to delegate connection management to the hook implementation as defined in our [Implementing Socket Mode documentation](https://docs.slack.dev/apis/events-api/using-socket-mode/#implementing). Because establishing a network connection and handling incoming events is assumed to be a long-running process, invoking this hook will block the CLI process.

##### Input

The application's app-level token and bot access token will be provided as environment variables to the hook process (`SLACK_CLI_XAPP` and `SLACK_CLI_XOXB` respectively, as well as `SLACK_APP_TOKEN` and `SLACK_BOT_TOKEN`). The SDK should use the app token to [create a socket connection](https://docs.slack.dev/apis/events-api/using-socket-mode/#call) with the Slack backend on behalf of the app. Additionally, the SDK may use the provided bot token to facilitate API calls to the Slack API.

All Bolt SDKs leverage this `start` hook operating mode.

A custom start path can be set with the `SLACK_CLI_CUSTOM_FILE_PATH` variable.

##### Output

Any `STDOUT` received from the this hook would be immediately streamed to the CLI process’ `STDOUT`.

##### Support

| Deno           | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| -------------- | ------------------- | --------------- | -------------- |
| ❌ Unsupported | ✅ Supported        | ✅ Supported    | ❌ Unsupported |

### `check-update` (optional) {#check-update}

Check that an update for the SDK is available.

For Deno, the SDK would check `deno.land` for the latest version. Likewise for npm. The [`install-update`](#install-update) hook would then update any files required for managing dependencies.

#### Output

The format for the output JSON is as follows:

```json
{
  "name": "",
  "releases": [
    {
      "name": "",
      "current": "",
      "latest": "",
      "update": true,
      "breaking": false,
      "message": "",
      "url": "",
      "error": ""
    },
    ...
  ],
  "message": "",
  "url": "",
  "error": ""
}
```

| Field             | Description                                                                                                                                                                               | Required |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| name              | String containing the name corresponding to the overall package/library, in which individual component releases are bundled.                                                              | Required |
| releases          | Array of objects representing individual releases.                                                                                                                                        | Required |
| releases.name     | String containing the name of the dependency.                                                                                                                                             | Required |
| releases.current  | String containing the current version.                                                                                                                                                    | Optional |
| releases.latest   | String containing the latest version available.                                                                                                                                           | Optional |
| releases.update   | Boolean indicating whether there is an update available.                                                                                                                                  | Optional |
| releases.breaking | Boolean indicating whether the update is breaking.                                                                                                                                        | Optional |
| releases.message  | String containing a message about the dependency.                                                                                                                                         | Optional |
| releases.url      | String containing a URL with update information.                                                                                                                                          | Optional |
| releases.error    | Object with a single key, "message", which has a string value containing error information.                                                                                               | Optional |
| message           | String containing any additional details that should be surfaced to the user. For example, this could include how to manually update or a warning that certain files will be overwritten. | Optional |
| url               | String containing a URL where one can learn more about the release.                                                                                                                       | Optional |
| error             | Object with a single key, "message", which has a string value containing error information.                                                                                               | Optional |

#### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ✅ Supported        | ✅ Supported    | ❌ Unsupported |

### `install-update` (optional) {#install-update}

Update the SDK version and/or any dependencies required by the app.

#### Output

The format for the output JSON is as follows:

```json
{
  "name": "",
  "releases": [
    {
      "name": "",
      "current": "",
      "latest": "",
      "update": true,
      "breaking": false,
      "message": "",
      "url": "",
      "error": ""
    },
    ...
  ],
  "message": "",
  "url": "",
  "error": ""
}
```

| Field             | Description                                                                                                                                                                               | Required |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| name              | String containing the name corresponding to the overall package/library, in which individual component releases are bundled.                                                              | Required |
| releases          | Array of objects representing individual releases.                                                                                                                                        | Required |
| releases.name     | String containing the name of the dependency.                                                                                                                                             | Required |
| releases.current  | String containing the current version.                                                                                                                                                    | Optional |
| releases.latest   | String containing the latest version available.                                                                                                                                           | Optional |
| releases.update   | Boolean indicating whether there is an update available.                                                                                                                                  | Optional |
| releases.breaking | Boolean indicating whether the update is breaking.                                                                                                                                        | Optional |
| releases.message  | String containing a message about the dependency.                                                                                                                                         | Optional |
| releases.url      | String containing a URL with update information.                                                                                                                                          | Optional |
| releases.error    | Object with a single key, "message", which has a string value containing error information.                                                                                               | Optional |
| message           | String containing any additional details that should be surfaced to the user. For example, this could include how to manually update or a warning that certain files will be overwritten. | Optional |
| url               | String containing a URL where one can learn more about the release.                                                                                                                       | Optional |
| error             | Object with a single key, "message", which has a string value containing error information.                                                                                               | Optional |

#### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ❌ Unsupported      | ❌ Unsupported  | ❌ Unsupported |

### `get-trigger` (optional) {#get-trigger}

Used by the `triggers create` and `triggers update` CLI commands when the `--trigger-def` argument is passed. This hook should convert the app developer's code into a valid JSON blob that the CLI can upload to the trigger endpoints. The CLI handles adding the `trigger_id` when using the `trigger update` command.

#### Output

Below is a sample JSON blob representing the trigger:

```json
{
  "type": "shortcut",
  "name": "Submit an issue",
  "description": "Submit an issue to the channel",
  "workflow": "#/workflows/submit_issue",
  "workflow_app_id": "A0168GS8ZFV",
  "inputs": {
    "channel": {
      "value": "{{data.channel_id}}"
    },
    "interactivity": {
      "value": "{{data.interactivity}}"
    }
  }
}
```

#### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ❌ Unsupported      | ❌ Unsupported  | ❌ Unsupported |

### `doctor` (optional) {#doctor}

Used as part of the `doctor` CLI command to check against lower-level App / SDK requirements and ensure that the app developer has everything in place on their system to use the SDK.

For example, `deno-slack-sdk` should ensure that the Deno runtime is in place, `bolt-js` should ensure that Node is in place, `bolt-python` should ensure that Python is in place, and so on.

#### Output

The format for the output JSON is as follows:

```json
{
  "versions": [
    {
      "name": "",
      "current": "",
      "message": "",
      "error": ""
    },
    ...
  ]
}
```

| Field            | Description                                               | Required |
| ---------------- | --------------------------------------------------------- | -------- |
| versions         | Array of objects containing runtime details.              | Required |
| versions.name    | String containing the name of the runtime dependency.     | Required |
| versions.current | String containing the current system version.             | Required |
| versions.message | String containing a message about the runtime dependency. | Optional |
| versions.error   | String containing error information for the runtime.      | Optional |

#### Support

| Deno         | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| ------------ | ------------------- | --------------- | -------------- |
| ✅ Supported | ✅ Supported        | ✅ Supported    | ❌ Unsupported |

### `deploy` (optional) {#deploy}

This script is invoked by the `slack deploy` CLI command. It takes as input the app and bot token as environment variables to manage connections and authentication.

This script is provided by the developer and is meant to simplify app management. There is no restriction on the script.

If this script isn't provided, an attempt is made to deploy the app to Slack's managed infrastructure. This is expected for Deno apps but will fail for Bolt apps.

#### Input

Access to `STDIN` is necessary for certain scripts.

#### Output

Any `STDOUT` or `STDERR` received from the this hook is immediately streamed to the CLI process’ `STDOUT` and `STDERR`.

#### Support

| Deno           | Bolt for JavaScript | Bolt for Python | Bolt for Java  |
| -------------- | ------------------- | --------------- | -------------- |
| ❌ Unsupported | ✅ Supported        | ✅ Supported    | ❌ Unsupported |

## CLI-SDK JSON interface format {#interface-format}

The format for the JSON representing the CLI-SDK interface is as follows:

```json
{
  "runtime": "deno",
  "hooks": {
    "get-hooks": "command to be invoked",
    "get-manifest": "...",
    "build": "...",
    "start": "...",
    "check-update": "...",
    "install-update": "...",
    "get-trigger": "...",
    "doctor": "...",
    "deploy": "..."
  },
  "config": {
    "protocol-version": ["message-boundaries"],
    "sdk-managed-connection-enabled": false,
    "watch": {
      "manifest": {
        "paths": ["manifest.json"]
      },
      "app": {
        "paths": ["app.js", "listeners/"],
        "filter-regex": "\\.(ts|js)$"
      }
    },
    "trigger-paths": ["triggers/"]
  }
}
```

| Field                                 | Description                                                                                                                                                                                                                                                                                                                                                                                                     | Required |
| ------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| runtime                               | String denoting the target runtime for app functions to execute in. For apps deployed to Slack's managed infrastructure, the only accepted value today is `deno`.                                                                                                                                                                                                                                               | Required |
| hooks                                 | Object whose keys must match the hook names outlined in the above [Hooks Specification](#specification). Arguments can be provided within this string by separating them with spaces.                                                                                                                                                                                                                           | Required |
| config                                | Object of key-value settings.                                                                                                                                                                                                                                                                                                                                                                                   | Optional |
| config.protocol-version               | Array of strings representing the named CLI-SDK protocols supported by the SDK, in descending order of support, as in the first element in the array defines the preferred protocol for use by the SDK, the second element defines the next-preferred protocol, and so on. The only supported named protocol currently is `message-boundaries`. The CLI will use the v1 protocol if this field is not provided. | Optional |
| config.watch                          | Object with configuration settings for file-watching during `slack run`. Supports updating the `manifest` on change and reloading the `app` server. Read [Watch configurations](#watch-configurations) for details.                                                                                                                                                                                             | Optional |
| config.sdk-managed-connection-enabled | Boolean specifying whether the WebSocket connection between the CLI and Slack should be managed by the CLI or by the SDK during `slack run` executions. If `true`, the SDK will manage this connection. If `false` or not provided, the CLI will manage this connection.                                                                                                                                        | Optional |
| config.trigger-paths                  | Array of strings that are paths to files of trigger definitions.                                                                                                                                                                                                                                                                                                                                                | Optional |

This format must be adhered to, in order of preference, either:

1. As the response to `get-hooks`, or
2. Comprising the contents of the `hooks.json` file

### Watch configurations {#watch-configurations}

The `config.watch` setting looks for file changes during local development with the `slack run` command. The CLI supports separate file watchers for **manifest** changes and changes to **application code** as options for reinstalling the app or reloading the server.

```json
{
  "config": {
    "watch": {
      "manifest": {
        "paths": ["manifest.json"]
      },
      "app": {
        "paths": ["app.js", "listeners/"],
        "filter-regex": "\\.(ts|js)$"
      }
    }
  }
}
```

| Field                       | Description                                                                        | Required |
| --------------------------- | ---------------------------------------------------------------------------------- | -------- |
| watch.manifest              | Object configuring the manifest watcher for reinstalling the app.                  | Optional |
| watch.manifest.paths        | Array of file paths or directories to watch for manifest changes.                  | Required |
| watch.manifest.filter-regex | Regex pattern to filter which files trigger manifest reinstall (e.g., `\\.json$`). | Optional |
| watch.app                   | Object configuring the app watcher for restarting the app server.                  | Optional |
| watch.app.paths             | Array of file paths or directories to watch for app/code changes.                  | Required |
| watch.app.filter-regex      | Regex pattern to filter which files trigger server reload (e.g., `\\.(ts\|js)$`).  | Optional |

**Note:** For backward compatibility, top-level `paths` and `filter-regex` fields are treated as manifest watching configuration only. No server reloading will occur with the legacy structure.

## Hook resolution {#hook-resolution}

The CLI will employ the following algorithm in order to resolve the command to be executed for a particular hook:

1. If the `hooks.json` file contains a key for the desired hook, then the CLI will use the command defined under this key directly and end resolution. This might occur in the case where a user has defined an override of an existing hook.
2. If a key for the desired hook does not exist in `hooks.json`, the CLI will look for a key in the `get-hooks` response’s `hooks` key sub-object. If it exists, the CLI will execute the command for that hook and end resolution.
3. If neither of the prior attempts succeed, then the CLI should error out with an `sdk_hook_not_found` error.

### Examples

#### Simple example with only get-hooks

```json
{
  "hooks": {
    // If the SDK implements the below hook, then the hook is responsible for
    // returning as output any required vs. optional hooks.
    // Alternatively, you can skip implementing get-hooks and just use the below
    // contents in your hooks.json. Beware, though! This makes upgrading the SDK a
    // more error-prone process because developers need to edit this file directly.
    "get-hooks": "deno run -q --unstable --allow-read --allow-net https://deno.land/x/deno_slack_hooks@0.0.4/mod.ts"
  }
}
```

#### Overriding a specific hook with a custom command

```json
{
  "hooks": {
    "get-hooks": "deno run -q --unstable --allow-read --allow-net https://deno.land/x/deno_slack_hooks@0.0.4/mod.ts",
    // This is a user-defined custom hook that overrides the default "get-manifest"
    "get-manifest": "deno run -q --unstable --config=deno.jsonc --allow-read --allow-net https://deno.land/x/deno_slack_builder@0.0.8/mod.ts --manifest",
    // This is a user-defined custom hook that adds new functionality
    "custom-hook": "deno run my-custom-hook.ts"
  }
}
```

#### Complete example returned by SDK from get-hooks script implemented by Bolt (this is in-memory)

```json
{
  "hooks": {
    "get-manifest": "deno run -q --unstable --config=deno.jsonc --allow-read --allow-net https://deno.land/x/deno_slack_builder@0.0.8/mod.ts --manifest",
    "build": "deno run -q --unstable --config=deno.jsonc --allow-read --allow-write --allow-net https://deno.land/x/deno_slack_builder@0.0.8/mod.ts",
    "start": "deno run -q --unstable --config=deno.jsonc --allow-read --allow-net https://deno.land/x/deno_slack_runtime@0.0.5/local-run.ts"
  },
  "config": {
    "watch": {
      "manifest": {
        "paths": ["manifest.json"]
      },
      "app": {
        "paths": ["app.js", "listeners/"],
        "filter-regex": "\\.(ts|js)$"
      }
    },
    "sdk-managed-connection-enabled": "true"
  }
}
```

#### Complete example returned by SDK from the get-hooks script implemented by Deno SDK (this is in-memory)

```json
{
  "runtime": "deno",
  "hooks": {
    "get-manifest": "deno run -q --unstable --config=deno.jsonc --allow-read --allow-net https://deno.land/x/deno_slack_builder@0.0.8/mod.ts --manifest",
    "build": "deno run -q --unstable --config=deno.jsonc --allow-read --allow-write --allow-net https://deno.land/x/deno_slack_builder@0.0.8/mod.ts",
    "start": "deno run -q --unstable --config=deno.jsonc --allow-read --allow-net https://deno.land/x/deno_slack_runtime@0.0.5/local-run.ts"
  },
  "config": {
    "watch": {
      "filter-regex": "^manifest\\.(ts|js|json)$",
      "paths": ["."]
    }
  }
}
```

**Note:** The legacy format (top-level `paths` and `filter-regex`) is treated as manifest watching only. No server reloading will occur with this configuration.

## Terms {#terms}

### Types of developers

1. App Developers are using the CLI and SDK to create and build a project
2. SDK Developers are building and maintaining the CLI and/or SDKs (controlled by Slack or community-driven)

### Types of SDKs

1. Slack Deno SDK - Run on Slack
2. Bolt Frameworks (JavaScript, Python, and Java) - Remote using Slack's Bolt Framework
3. No SDK at all - Run on Slack, Remote Self-Hosted
4. Community SDKs, frameworks, and custom apps (e.g. Ruby, Golang, etc), if they exist

### Other definitions

- hooks.json - The CLI-SDK Interface implemented as a JSON object (`{...}`) or JSON file (hooks.json).
- command - This refers to a CLI command, eg. `slack doctor`
- hook - This refers to some CLI functionality delegated to the SDK and is defined at the level of the hooks.json file
