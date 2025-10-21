---
sidebar_label: Removing an app
slug: /tools/slack-cli/guides/removing-an-app
---

# Removing an app

<PaidPlanBanner />

All good things must come to an end. You can `uninstall` your app if you need to remove an app from a workspace, change app permissions, or `delete` the app in its entirety.

## Uninstall an app from your team {#uninstall-app}

Removing an app from a workspace doesn't have to be a permanent decision. Sometimes uninstalling the app to remove it's active presence in channels is sufficient! This option has the added benefit of reinstallation at a later time without recreating the entire app.

To uninstall an app using the CLI, use the [`uninstall`](/tools/slack-cli/reference/commands/slack_app_uninstall/) command. Then, choose the workspace you want to remove the app from:

```bash
slack uninstall -a A123ABC456 -t T123ABC456
```

```bash
⚠️ Warning
   App (A123ABC456) will be uninstalled from my-workspace (T123ABC456)
   All triggers, workflows, and functions will be deleted
   Datastore records will be persisted

❓ Are you sure you want to uninstall? Yes

🏠 Workspace uninstall
   Uninstalled the app "my-app" from workspace "my-workspace"
```

## Delete an app from your team {#delete-app}

:::danger[Deleting your app _permanently_ deletes all of its data]

Your app's related workflows, functions, and datastores will also be deleted. This decision is final and cannot be undone.

:::

To delete an app using the CLI, use the `slack delete` command:

```bash
slack delete -a A123ABC456 -t T123ABC456
```

```bash
⚠️ Danger zone
   App (A123ABC456) will be permanently deleted
   All triggers, workflows, and functions will be deleted
   All datastores for this app will be deleted
   Once you delete this app, there is no going back

❓ Are you sure you want to delete the app? Yes

🏠 App Uninstall
   Uninstalled the app "my-app" from "my-workspace"

📚 App Manifest
   Deleted the app manifest for "my-app" from "my-workspace"

🏘️ Apps
   This project has no apps
```