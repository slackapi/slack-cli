---
sidebar_label: Using environment variables
slug: /tools/slack-cli/guides/using-environment-variables-with-the-slack-cli
---

# Using environment variables with the Slack CLI

You can store and use environment variables with your Slack app by using a collection of Slack CLI commands and features. You can even access some pre-set environment variables!

:::note[You may be looking for the [_Using environment variables with the Deno Slack SDK_](/tools/deno-slack-sdk/guides/using-environment-variables) guide.]
:::
   
## Using the Slack CLI `env` commands

There are three Slack CLI subcommands that can be used to modify environment variables:

## `slack env set`

Use this to set an environment variable for the project. You can set the environment variable within the command, or run `slack env set` alone to go through an interactive interface.

```
slack env set MAGIC_PASSWORD abracadbra
```

## `slack env unset`

Use this to remove variables for the project. You can unset environment variables within the command, or run `slack env unset` alone to view all environment variables and select which one to unset. 

```
slack env unset MAGIC_PASSWORD
```

## `slack env list`

Use this to view the variables set for this project

```
slack env list
```

## Using CLI-provided variables

The Slack CLI provides an envelope of environment variables set automatically.

| Variable | Origin | Use | Bolt frameworks | Deno Slack SDK |
|----------|--------|-----|-----------------|-----------------|
| `SLACK_APP_TOKEN` | Set from the API response after successful app installation via `slack app install` or `slack app link`. Preserved if already exists (with warning if it differs). | Authenticate with Slack API as the app. Required for Socket Mode connections and API calls. | ✅ Supported | ❌ Not used |
| `SLACK_BOT_TOKEN` | Set from the API response after successful app installation via `slack app install` or `slack app link`. Preserved if already exists (with warning if it differs). | Authenticate with Slack API as the bot user. Used for making API calls on behalf of the app. | ✅ Supported | ❌ Not used |
| `SLACK_CLI_XAPP` | Built from the app token and passed to the `start` hook via environment map during `slack run`. | Used by Bolt frameworks for Socket Mode connection. | ✅ Supported | ❌ Not used |
| `SLACK_CLI_XOXB` | Built from the bot access token and passed to the `start` hook via environment map during `slack run`. | Used by Bolt frameworks for API calls. | ✅ Supported | ❌ Not used |
| `SLACK_APP_PATH` | Set when a custom start path is provided via `slack run` | Used to run from a non-root directory. | ✅ Supported | ✅ Supported |
| `SLACK_CLI_CUSTOM_FILE_PATH` | Set to the same value as `SLACK_APP_PATH` when a custom start path is provided via `slack run` | Used to run from a non-root directory. | ✅ Supported | ✅ Supported |
