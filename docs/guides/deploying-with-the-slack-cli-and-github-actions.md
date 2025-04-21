---
sidebar_label: Deploying with GitHub Actions
slug: /slack-cli/guides/deploying-the-slack-cli-with-github-actions
---

# Deploying with the Slack CLI & GitHub Actions

This tutorial demonstrates how to use CI/CD to facilitate automatic deployments when code changes are pushed to GitHub.

Before we begin, you'll need to do the following:

* Create a new GitHub repository &mdash; any name will do.
* [Install](/slack-cli/guides/installing-the-slack-cli-for-mac-and-linux) the Slack CLI on your machine.
* [Authorize](/slack-cli/guides/authorizing-the-slack-cli) the Slack CLI to your workspace.

Once those steps have been completed, we're ready to move on to building our automated deployment app.

## Create a new app {#create}

Run `slack create` to create a new Slack app project. Select a template to build from; in this case, hit **Enter** to choose the default `Issue submission` template.

Refer to [Create or remove an app](/deno-slack-sdk/guides/creating-an-app) for more details.

## Initial deploy {#initial-deploy}

Run `slack deploy` to manually deploy the new app to your workspace for the first time. During the process, you'll be prompted to create a link trigger. Refer to [Link triggers](/deno-slack-sdk/guides/creating-link-triggers) for more details.

Once created, copy the link and share it in a Slack channel. You'll see a button appear to start the workflow; click it to verify that the default workflow is functioning properly.

## Obtain a service token {#obtain-service-token}

To automate subsequent deployments, we'll need to obtain a [service token](/slack-cli/guides/authorizing-the-slack-cli#ci-cd) for the app.

Navigate to the root directory of your project and run the `slack auth token` command. You'll be prompted to run a slash command called `slackauthticket`, similar to the way you authorized your app earlier. Copy that command and run it in Slack, then copy the challenge code you receive and enter it into your terminal. You'll see a service token that starts with `xoxp-` &mdash; make sure you save this token. Note that this token is connected to a specific workspace and organization.

## Add your service token to GitHub as a secret {#add-service-token}

In your GitHub repository, navigate to **Settings** > **Secrets and variables** > **Actions** and click **New repository secret** to add a new secret as follows:

* Name: SLACK_SERVICE_TOKEN
* Secret: the `xoxp-` prefixed service token obtained above

Be sure to save your changes!

## Add a new deployment file with your GitHub actions workflow {#add-deployment-file}

Create a folder called `.github/workflows/` and add a new file called `deployment.yml` with the following content:

```yml
name: Slack App Deployment

on:
  push:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
    - uses: actions/checkout@v4
    - name: Install Deno runtime
      uses: denoland/setup-deno@v1
      with:
        deno-version: v1.x

    - name: Install Slack CLI
      if: steps.cache-slack.outputs.cache-hit != 'true'
      run: |
        curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash

    - name: Deploy the app
      env:
        SLACK_SERVICE_TOKEN: ${{ secrets.SLACK_SERVICE_TOKEN }}
      run: |
        cd gh-actions-demo/
        slack deploy -s --token $SLACK_SERVICE_TOKEN
```

This file instructs the Slack CLI to deploy the latest revision of your repository to the Slack platform when any changes are pushed to the main branch.

You can go with any other Linux option you prefer for the GitHub-hosted runner; refer to [About GitHub-hosted runners](https://docs.github.com/en/actions/using-github-hosted-runners/about-github-hosted-runners/about-github-hosted-runners#viewing-available-runners-for-a-repository) for more details.

## Test it out {#test}

To test it out, make a PR and push some changes to the main branch. If the GitHub Actions job successfully completes, your app should now be available in your workspace.

Your GitHub repository is now set up for team collaboration! Your team members can review code changes by submitting PRs, and once these are merged into the main branch, the changes will be deployed automatically. Should you need to reverse a deployment automatically, you can create a reverting PR and quickly merge that.

## Further reading {#read}

Check out these articles to expand your knowledge and skills of automated deployments and the Slack CLI:

➡️ [CI/CD overview and setup](https://tools.slack.dev/slack-cli/guides/setting-up-ci-cd-with-the-slack-cli)

➡️ [CI/CD authorization](/slack-cli/guides/authorizing-the-slack-cli#ci-cd)
