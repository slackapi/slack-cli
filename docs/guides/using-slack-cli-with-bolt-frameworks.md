---
sidebar_label: Using with Bolt frameworks
---

# Using the Slack CLI with Bolt frameworks

You can use the Slack CLI to streamline development of apps using [Bolt for JavaScript](/tools.slack.dev/bolt-js) and [Bolt for Python](/tools.slack.dev/bolt-python).

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

:::info

For Bolt for Python projects, automatic project dependency installation is currently unsupported, and will need to be done manually. For more information, refer to [Getting started with Bolt for Python](https://tools.slack.dev/bolt-python/getting-started).

:::

To run your new app, use the `slack run` command with the experiment flag as follows:

```
slack run
```

You'll be prompted to choose your team/workspace, and then your app should let you know that it's up and running. 🎉