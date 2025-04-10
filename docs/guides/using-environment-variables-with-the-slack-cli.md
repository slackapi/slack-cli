---
sidebar_label: Using environment variables
slug: /slack-cli/guides/using-environment-variables-with-the-slack-cli
---

# Using environment variables with the Slack CLI

Storing and using environment variables in an application allows for certain variables to be maintained outside of the code of the application. You can use environment variables from within Slack [functions](/deno-slack-sdk/guides/creating-custom-functions), [triggers](/deno-slack-sdk/guides/using-triggers), and [manifests](/deno-slack-sdk/guides/using-the-app-manifest).

## Using environment variables with a custom function {#custom-function}

When accessing environment variables from within a [custom function](/deno-slack-sdk/guides/creating-custom-functions), where you store them differs when the app is local versus deployed.

### Storing local environment variables {#local-env-vars}

Local environment variables are stored in a `.env` file at the root of the project and made available for use in [custom functions](/deno-slack-sdk/guides/creating-custom-functions) via the `env` [context property](/deno-slack-sdk/guides/creating-custom-functions#context).

A local `.env` file might look like this:
```env
MY_ENV_VAR=asdf1234
```

Note that changes to your `.env` file will be reflected when you restart your local development server.

While the `.env` file should **never** be committed to source control for security reasons, you can see a sample `.env` file we've included in the [Timesheet approval sample app](https://github.com/slack-samples/deno-timesheet-approval) and the [Incident management sample app](https://github.com/slack-samples/deno-incident-management). 

### Storing deployed environment variables {#deployed-env-vars}

When your app is [deployed](/deno-slack-sdk/guides/deploying-to-slack), it will no longer use the `.env` file. Instead, you will have to add the environment variables using the [`env add`](/slack-cli/reference/commands/slack_env_add) command. Environment variables added with `env add` will be made available to your deployed app's [custom functions](/deno-slack-sdk/guides/creating-custom-functions) just as they are locally; see examples in the next section.

For the above example, we could run the following command before deploying our app:

```zsh
slack env add MY_ENV_VAR asdf1234
```

If your token contains non-alphanumeric characters, wrap it in quotes like this:

```zsh
slack env add SLACK_API_URL "https://dev<yournumber>.slack.com/api/"
```

Your environment variables are always encrypted before being stored on our servers and will be automatically decrypted when you use them&mdash;including when listing environment variables with `slack env list`. 

### Access variables from within function {#access-function}

We can retrieve the `MY_ENV_VAR` environment variable from within a [custom Slack function](/deno-slack-sdk/guides/creating-custom-functions) via the `env` [context property](/deno-slack-sdk/guides/creating-custom-functions#context) like this:

```javascript
// functions/my_function.ts

import { DefineFunction, SlackFunction } from "deno-slack-sdk/mod.ts";
import { MyFunctionDefinition } from "functions/myfunction.ts"


export default SlackFunction(MyFunctionDefinition, ({ env }) => {
  const myEnvVar = env["MY_ENV_VAR"];
  // ...
  return { outputs: {} };
});
```

Environment variables also play an important part in making calls to a third-party API. Learn more about how to do that in the [FAQ](https://docs.slack.dev/faq#third-party).

## Using environment variables with a trigger or manifest {#using-trigger-manifest}

Accessing environment variables from within a [trigger](/deno-slack-sdk/guides/using-triggers) definition or when constructing the [manifest](/deno-slack-sdk/guides/using-the-app-manifest) differs slightly from custom functions.

Whether your app is being run locally or already deployed, constructing these definitions happens entirely on your machine and so the environment variables stored on your machine are used.

### Storing environment variables {#trigger-manifest-env-vars}

Environment variables used in trigger or manifest definitions should be saved in the local `.env` file for your project [as shown above](#local-env-vars). The values from this file are collected and used when generating these definitions.

Regardless of whether you're working with a local or deployed app, the same values from this file will be used. Read on to learn how to access these stored variables in code.

### Accessing variables from a trigger or manifest {#accessing-trigger-manifest}

The Deno runtime provides a helpful `load` function to autoload environment variables as part of the `dotenv` module of the standard library. We'll leverage this to easily access our environment variables.

Including this module in code will automatically import local environment variables for immediate use! Start by adding [the latest version](https://deno.land/std/dotenv/mod.ts) of this module to your `import_map.json`:

```json title="test"
{
  "imports": {
    "deno-slack-sdk/": "https://deno.land/x/deno_slack_sdk@a.b.c/",
    "deno-slack-api/": "https://deno.land/x/deno_slack_api@x.y.z/",
    "std/": "https://deno.land/std@0.202.0/"
  }
}
```

Then, you can import the module into any file that makes use of environment variables and start accessing the environment with [`Deno.env.get("VARIABLE_NAME")`](https://examples.deno.land/environment-variables) like so:

```javascript
// manifest.ts
import { Manifest } from "deno-slack-sdk/mod.ts";
import ExampleWorkflow from "./workflows/example_workflow.ts";

import "std/dotenv/load.ts";

export default Manifest({
  name: "Chatbot4000",
  displayName: Deno.env.get("CHATBOT_DISPLAY_NAME"),
  description: "Workflows for communicating with an imagined chatbot",
  icon: "assets/icon.png",
  workflows: [ExampleWorkflow],
  outgoingDomains: [
    Deno.env.get("CHATBOT_API_URL")!,
  ],
  botScopes: ["commands", "chat:write", "chat:write.public"],
});
```

After including this new module, you may have to run [`deno cache manifest.ts`](https://docs.deno.com/runtime/manual/getting_started/command_line_interface#cache-and-compilation-flags) to refresh your local dependency cache.

Variable values such as these are commonly used to specify [outgoing domains](/deno-slack-sdk/guides/using-the-app-manifest#manifest-properties) used by functions, channel IDs for [event triggers](/deno-slack-sdk/guides/creating-event-triggers#event-object), or client IDs of an [external authentication](/deno-slack-sdk/guides/integrating-with-services-requiring-external-authentication#define) provider. But, don't let that limit you — environment variables can be used in so many other places!

#### Requiring environment variables values {#required-manifest-variable-values}

Setting values for environment variables can sometimes be forgotten, which can cause problems at runtime. Catching errors for these missing values early is often better than waiting for that runtime problem.

Including a `!` with your call to `Deno.env.get()` will ensure this value is defined at the time of building a definition and will throw an error otherwise.

The previous example uses this pattern to ensure an outgoing domain is always set:

```javascript
  outgoingDomains: [
    Deno.env.get("CHATBOT_API_URL")!,
  ],
```

With this addition, running `slack deploy` without defining a value for `CHATBOT_API_URL` in the `.env` file will throw an error to give you a chance to set it before actually deploying!

## Enabling debug mode {#debug}

The included environment variable `SLACK_DEBUG` can enable a basic debug mode. Set `SLACK_DEBUG` to `true` to have all function-related payloads logged. 

For local apps, add the following to your `.env` file:

```
SLACK_DEBUG=true
```

For deployed apps, run the following command before deployment:

```
slack env add SLACK_DEBUG true
```

## Included local and deployed variables {#included}

Slack provides two environment variables by default, `SLACK_WORKSPACE` and `SLACK_ENV`. The workspace name is specified by `SLACK_WORKSPACE` and `SLACK_ENV` provides a distinction between the `local` and `deployed` app. Use these values if you want to have different values based on the workspace or environment that the app is installed in.

These variables are automatically included when generating the manifest or triggers only. For access from within a custom function, these variables can be set from the `.env` file or with the [`env add`](/slack-cli/reference/commands/slack_env_add) command.

A custom `WorkspaceMapSchema` can be created and used with these variables to decide which values to use for certain instances of an app. This can be used as an alternative to a local `.env` file or in conjunction with it. The following snippet works well for inclusion in your app manifest, or for triggers (for example, to change event trigger channel IDs):

```javascript
// Custom schemas can be defined for workspace values
type WorkspaceSchema = { channel_id: string };
type WorkspaceMapSchema = {
  [workspace: string]: {
    [environment: string]: WorkspaceSchema;
  };
};

// Custom values can be set for each known workspace
export const workspaceValues: WorkspaceMapSchema = {
  beagoodhost: {
    deployed: {
      channel_id: "C123ABC456",
    },
    local: {
      channel_id: "C123ABC456",
    },
  },
  sandbox: {
    deployed: {
      channel_id: "C222BBB222",
    },
    local: {
      channel_id: "C222BBB222",
    },
  },
};

// Fallback options can also be defined
export const defaultValues: WorkspaceSchema = {
  channel_id: "{{data.channel_id}}",
};

// Included environment variables will determine which value is used
const environment = Deno.env.get("SLACK_ENV") || "";
const workspace = Deno.env.get("SLACK_WORKSPACE") || "";
const { channel_id } = workspaceValues[workspace]?.[environment] ??
  defaultValues;
```
