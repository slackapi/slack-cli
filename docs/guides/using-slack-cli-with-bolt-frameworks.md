---
sidebar_label: Using with Bolt frameworks
---

# Using the Slack CLI with Bolt frameworks

You can use the Slack CLI to streamline development of apps using [Bolt for JavaScript](/tools/bolt-js) and [Bolt for Python](/tools/bolt-python).

:::info[Feeling adventurous?]

To create a Bolt app using features currently under development, refer to the [experiments](/tools/slack-cli/reference/experiments) page.

:::

## Getting started

Creating a Bolt app via the Slack CLI is similar to creating other apps with the Slack CLI. Run the following command to begin:

```
slack create
```

Select an option from the following list. For this example, choose **Starter app**:

```zsh
// highlight-next-line
> Starter app - Getting started Slack app
Automation app - Custom steps and workflows
AI app - Slack agents & assistants
View more samples
```

You will then be prompted to choose between **Bolt for JavaScript** or **Bolt for Python**. Choose your favorite flavor.

Your app will be cloned from the respective [JavaScript](https://github.com/slack-samples/bolt-js-starter-template) or [Python](https://github.com/slack-samples/bolt-python-starter-template) project template on our Slack Platform Sample Code repository, and its project dependencies will be installed. Then, `cd` into your project folder. 

:::info[For Bolt for Python projects, automatic project dependency installation is currently unsupported, and will need to be done manually.] 

For more information, refer to [Getting started with Bolt for Python](/bolt-python/getting-started).

:::

To run your new app, use the `slack run` command with the experiment flag as follows:

```
slack run
```

You'll be prompted to choose your team/workspace, and then your app should let you know that it's up and running. 🎉

## App manifest

The Slack app manifest is the configuration of the app. The `manifest.json` file included with selected templates and samples reflects the features and permissions of your app. When you create an app with the CLI, the corresponding app and matching manifest can be found on [app settings](https://api.slack.com/apps).

For Bolt apps created through the CLI, by default, the manifest source set in the `config.json` file is `remote`. This means that the manifest in your [app settings](https://api.slack.com/apps) is the source of truth. To modify the manifest (add new features, scopes, etc.), do so in the app settings. If you change the `config.json` to reflect a `local` manifest source and modify the local `manifest.json` file, the CLI will ask for confirmation before overriding the settings upstream on reinstall (run). This prompt appears if the app manifest on app settings differs from a known state saved in `.slack/cache`. There is not currently a dedicated manifest update command.

In contrast, Deno apps created with the CLI have the manifest source configuration of `local` because those apps are not managed in the [app settings page](https://api.slack.com/apps).