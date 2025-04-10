---
sidebar_label: Troubleshooting errors
slug: /slack-cli/guides/troubleshooting-slack-cli-errors
---

# Troubleshooting Slack CLI errors

Troubleshooting errors can be tricky. There's a lot going on between your development environment, the Slack CLI, and your code!

View the [full list of of errors in the reference](/slack-cli/reference/errors).

## VSCode and the Deno plugin {#vscode-deno}

If you are using VSCode with the Deno plugin and you run into an error where Deno isn't being honored properly by VSCode, this is because VSCode treats the folder that's opened as the workspace root by default

If you open a parent folder, the `.vscode/settings.json` file must be moved to the root of _that_ folder. VSCode requires that `deno.enable: true` is set in that `.vscode/settings.json` file, and VSCode only honors this setting if it's in the root of the project.

Other common errors you may run into are static errors when opening a parent directory that contains one or many apps inside. These include:

* _An import path cannot end with a '.ts' extension. Consider importing './workflows/greeting_workflow.js' instead_.
    * This error is due to Deno not being set up correctly.
* _Relative import path "deno-slack-sdk/mod.ts" not prefixed with / or ./ or ../deno(import-prefix-missing)_.
    * This error is due to an invalid import map.

These errors occur because of that first one we covered &mdash; VSCode treats the folder that’s opened as the workspace root by default, and looks for the `.vscode/settings.json` and `deno.jsonc` files there. To resolve this, open the app folder directly, or set up your own workspace config in VSCode.

