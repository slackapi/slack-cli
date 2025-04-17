---
sidebar_label: Setting up CI/CD
slug: /slack-cli/guides/setting-up-ci-cd-with-the-slack-cli

---

# Setting up CI/CD with the Slack CLI

CI/CD is an acronym for Continuous Integration and Continuous Delivery. Also referred to as _CI/CD pipeline_, it is a common term in the world of DevOps.

DevOps is another acronym of sorts that stands for Development and Operations, a combination of software development (typically encompassing planning, building, coding, and testing) and operations (including software releases, deployment, and status monitoring).

## What is CI/CD? {#what-cicd}

CI/CD describes the set of tools and practices that allow development teams to automate the process of building, testing, and deploying code. Automating testing and deployment enables developers to find and address defects earlier in the development process; therefore, software applications can be delivered more quickly.

From your first tiny commit to deploying your code to production, your code passes through various validation steps, some or all of which are performed automatically. This is the deployment pipeline in action.

### Continuous Integration {#what-integration}

To break it down further, Continuous Integration refers to the building and testing side of the process, and can include things like the following:

* integrating code commits into your main repository branch,
* automatically running static (linting), unit, integration, regression, or acceptance tests on each commit; these are typically configured using a [provider](#providers), or
* automatically kicking off builds after code is successfully merged.

An underlying philosophy of Continuous Integration is that by merging, testing, and validating code frequently, you can identify defects, conflicts, and other issues more quickly and fix them more easily.

#### Providers {#providers}

CI providers are tools for automating the code changes that are part of your pipeline. CI/CD processes are set up at the repository level, and since these setups are specific to each provider, yours will be slightly different depending on which one you choose.

Some popular providers include:

* [GitHub Actions](https://github.com/features/actions)
* [CircleCI](https://circleci.com/)
* [GitLab](https://about.gitlab.com/)
* [Jenkins](https://www.jenkins.io/)
* [Azure Pipelines](https://azure.microsoft.com/en-us/products/devops/pipelines)
* [Bamboo](https://www.atlassian.com/software/bamboo)

### Continuous Delivery {#what-delivery}

Continuous Delivery refers to the deployment side of the process, and can include things like the following:

* automatically provisioning infrastructure; for example, user provisioning (granting users permissions to services or applications) or service provisioning (setting up credentials and system privileges)
* manually deploying code to testing or production environments
* automatically deploying code to testing or production environments (also referred to as _continuous deployment_)
* leveraging feature toggles for code not yet slated for release to production

An underlying philosophy of Continuous Delivery is that once code is tested as part of the CI process, your code can be packaged with everything it needs and released to production at any time &mdash; so everything should be in order and ready to go at a moment's notice.

## Continuous Integration setup {#ci-pipeline}

Before setting up a CI/CD pipeline, you should first familiarize yourself with the [Deno](/deno-slack-sdk/guides/installing-deno) environment if you haven't already. On the CI side of things, we'll be using [Deno's built-in tools](https://deno.land/manual/tools), which allow developers to leverage tools such as GitHub Actions to add steps for testing, linting, and formatting our code.

You'll also need to accommodate requests from your network to a variety of hosts. Refer to [Which hosts are involved in the creation and execution of apps created with the Slack CLI?](https://docs.slack.dev/faq#hosts) for more details.

In addition, you'll need to obtain a service token to authorize your CI/CD setup. Refer to [CI/CD authorization](/slack-cli/guides/authorizing-the-slack-cli#ci-cd) for more details about obtaining, using, and revoking service tokens.

Once you've done those things, you're ready to get started! Let's walk through an example.

Let's take a look at the [Virtual Running Buddies sample app](https://github.com/slack-samples/deno-virtual-running-buddies).

First, we'll open the `deno.jsonc` file located at the root of the project:

```json
// deno.jsonc

{
  "$schema": "https://deno.land/x/deno/cli/schemas/config-file.v1.json",
  "fmt": {
    "files": {
      "include": [
        "README.md",
        "datastores",
        "external_auth",
        "functions",
        "manifest.ts",
        "triggers",
        "types",
        "views",
        "workflows"
        ]
    }
  },
  "importMap": "import_map.json",
  "lint": {
    "files": {
      "include": [
        "datastores",
        "external_auth",
        "functions",
        "manifest.ts",
        "triggers",
        "types",
        "views",
        "workflows"
        ]
    }
  },
  "lock": false,
  "tasks": {
    "test": "deno fmt --check && deno lint && deno test --allow-read --allow-none"
  }
}
```

This file is your [configuration file](https://deno.land/manual/getting_started/configuration_file). It allows you to customize Deno's built-in TypeScript compiler, formatter, and linter.

We'll also point to our import map here (`import_map.json`), which allows you to manage what versions of modules or the standard library are included with your project:

```json
// import_map.json

{
  "imports": {
    "deno-slack-sdk/": "https://deno.land/x/deno_slack_sdk@2.1.5/",
    "deno-slack-api/": "https://deno.land/x/deno_slack_api@2.1.1/",
    "mock-fetch/": "https://deno.land/x/mock_fetch@0.3.0/"
  }
}
```

Next, we'll look inside the `deno.yml` file, which is located in the `.github/workflows` folder for this sample. Its contents are as follows:

```yaml
# deno.yml

name: Deno app build and testing

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  deno:
    runs-on: ubuntu-latest
    timeout-minutes: 5

    steps:
      - name: Set up repo
        uses: actions/checkout@v3

      - name: Install Deno
        uses: denoland/setup-deno@v1
        with:
          deno-version: v1.x

      - name: Verify formatting
        run: deno fmt --check

      - name: Run linter
        run: deno lint

      - name: Run tests
        run: deno task test

      - name: Run type check
        run: deno check *.ts && deno check **/*.ts
```

This is the meat and potatoes of our CI setup. The pipeline is kicked off by a push to or pull request from the main branch, and then we run through all the subcommands we want to complete within the _jobs/steps_ section, including setting up our repository and installing Deno.

This also includes calling Deno's [task runner](https://deno.land/manual/tools/task_runner) to run any unit tests we have created for our custom functions. This allows us to run all of our unit tests automatically, rather than running each one manually from the command line. In this sample, this means all of the files located in the sample app's `functions` folder ending in _test.ts_.

✨  **For more information about creating unit tests**, refer to [Testing custom functions](/deno-slack-sdk/guides/creating-custom-functions#testing).

:::info

If you've created your project by [cloning one of our sample apps](/deno-slack-sdk/guides/creating-an-app), note that the `.github` folder will not be included. You'll need to create it yourself, but you can use the handy dandy **Copy** button next to the code samples on this page to get started!

:::

While not part of this sample app, you can also generate test coverage reports from your `deno.yml` file. For more information, refer to [Test coverage](https://deno.land/manual/basics/testing/coverage).

## Continuous Delivery setup {#cd-pipeline}

On the CD side of things, there are various ways you can deploy Deno projects to the cloud. Your setup will differ based on which platform you choose.

Let's look at an example `deploy.yml` file, which you would also place in the `.github/workflows` folder along with your `deno.yml` file. The steps below need to run within the app folder and are for already-deployed apps only. Its contents are as follows:

```yaml

# deploy.yml

name: Deploy to Slack Cloud

on:
  push:
    tags: [ '*.*.*' ]

jobs:
  deploy:
    runs-on: macos-latest

    steps:
      - name: Set up repo
        uses: actions/checkout@v3

      - name: Install CLI
        run: curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash

      - name: Deploy
        run: slack deploy --app ${{ secrets.APP }} --workspace ${{ secrets.WORKSPACE }} --token ${{ secrets.SLACK_SERVICE_TOKEN }}
```

Central to this file is calling the `slack deploy` command to deploy your app to Slack's managed infrastructure. Using this command, the latest changes to your app will be deployed to a workspace once pushed/pulled/merged/tagged/etc. as specified in your workflow.

## Onward {#onward}

Want to learn more about how to use the Slack CLI? [Start here](/slack-cli/guides/installing-the-slack-cli-for-mac-and-linux)!

✨  **For more information about deploying to Slack's managed infrastructure**, refer to [Deploy to Slack](/deno-slack-sdk/guides/deploying-to-slack).

✨  **For more information specific to different platforms**, refer to [Deploying Deno](https://deno.land/manual/advanced/deploying_deno).

✨  **Just want to write some unit tests?** Refer to [Testing custom functions](/deno-slack-sdk/guides/creating-custom-functions#testing).
