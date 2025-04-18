---
slug: /slack-cli/guides/authorizing-the-slack-cli
---

# Authorizing the Slack CLI {#authorize-cli}

Once you have the Slack CLI installed for either [Windows](/slack-cli/guides/installing-the-slack-cli-for-windows) or [Mac/Linux](/slack-cli/guides/installing-the-slack-cli-for-mac-and-linux), authorize the Slack CLI in your workspace with the following command:

```zsh
slack login
```

## Authorization ticket {#ticket}

In your terminal window, you should see an authorization ticket in the form of a
slash command, and a prompt to enter a challenge code:

```zsh
$ slack login

📋 Run the following slash command in any Slack channel or DM
   This will open a modal with user permissions for you to approve
   Once approved, a challenge code will be generated in Slack

/slackauthticket ABC123defABC123defABC123defABC123defXYZ

? Enter challenge code
```

Copy the slash command and paste it into any Slack conversation in the workspace you will be developing in.

When you send the message containing the slash command, a modal will pop up, prompting you to grant certain permissions to the Slack CLI. Click **Confirm** in the modal to continue to the next step.

A new modal with a challenge code will appear. Copy that challenge code, and paste it back into your terminal:

```zsh
? Enter challenge code eXaMpLeCoDe

✅ You've successfully authenticated! 🎉
   Authorization data was saved to ~/.slack/credentials.json

💡 Get started by creating a new app with slack create my-app
   Explore the details of available commands with slack help
```

Verify that your Slack CLI is set up by running `slack auth list` in your terminal
window:

```zsh
$ slack auth list

myworkspace (Team ID: T123ABC456)
User ID: U123ABC456
Last updated: 2023-01-01 12:00:00 -07:00
Authorization Level: Workspace
```

You should see an entry for the workspace you just authorized. If you don't, get a new authorization ticket with `slack login` to try
again.

You're now ready to begin building workflow apps!

### Version update notifications {#version-updates}

Once a day, the Slack CLI checks for updates after running any command. When an update is available, a notification will be displayed with a link where you can find and download the new version.

Update notifications can be disabled using a command-line flag or an environment variable. When running any command, you can append the `--skip-update` or `-s` flag. Alternatively, you can set the `SLACK_SKIP_UPDATE` environment variable and assign it any value.

## CI/CD authorization {#ci-cd}

Setting up a CI/CD pipeline requires authorization using a service token. Service tokens are **long-lived, non-rotatable user tokens** &mdash; they won't expire, so they can be used to perform any Slack CLI action without the need to refresh tokens.

To get a service token, you'll use the `slack auth token` command to get a `slackauthticket`, which you'll copy and paste into your workspace to exchange for the service token. The service token will not be saved to your `credentials.json` file; instead, it is presented in the terminal for you to copy and paste for use in your CI/CD pipeline. Once copied, you'll use the `slack login --auth <your-service-token>` command to authorize your Slack CLI. Detailed instructions are below.

:::info

The service token will not conflict with your regular authentication token; you can continue using your regular authentication token within the Slack CLI while using the service token for your CI/CD pipeline.

:::

### Best practices for service tokens {#best-practices-tokens}

Since a service token obtained via the Slack CLI is tied to the developer who requested it, it is recommended &mdash; especially for those who are part of large enterprises &mdash; to create a "Slack service account" within your workspace for the purpose of obtaining service tokens.

This "Slack service account" will be identical to other user accounts, but service tokens can now be associated with this account rather than an individual organization member. This can reduce the security risk of an individual developer's token being compromised, as well as lessening the dependence on a single individual for service token access and management.

### Obtaining a service token {#obtain-token}

Run the following command to get a `slackauthticket`:

```
slack auth token
```

Then, copy and paste the `/slackauthticket <your-auth-ticket>` slash command into the message composer anywhere within your Slack workspace.

Copy the given challenge code, and paste it into the Slack CLI prompt.

Securely store the given service token, as this authorization information will _not_ be saved to your local `credentials.json` file.

### Using a service token {#use-token}

Run the following command to authorize your Slack CLI with the service token:

```
slack --token <your-service-token>
```

The Slack CLI will attempt to verify the token and use it to log in.

The `--token` global flag allows you to pass the service token with any Slack CLI command; for example:

```
slack deploy --token <your-service-token>
```

### Revoking a service token {#revoke-token}

Run the following command to revoke a service token:

```
slack auth revoke --token <your-service-token>
```
