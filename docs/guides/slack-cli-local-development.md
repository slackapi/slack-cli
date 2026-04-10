---
sidebar_label: Developing locally
slug: /tools/slack-cli/guides/slack-cli-local-development
---

# Developing locally using the Slack CLI

The Slack CLI provides some features, and customization of said features, to streamline developing apps locally. This guide details those features.

## Configuring file watching and auto-restart

The Slack CLI automatically watches your app files and restarts the server when they change. It also watches `manifest.json` and reinstalls the app when the manifest changes.

For Bolt for Python projects, the Slack CLI watches all `.py` files in the root directory.

For Bolt for JavaScript projects, the Slack CLI watches all `.js` files in the root directory.

View the [hooks reference](/tools/slack-cli/reference/hooks#watch-configurations) for detailed configuration options.

### Customizing watch paths

You can override the default watch paths in `.slack/hooks.json` if so desired. 

For example, if you're working in TypeScript you'd want to only watch the `dist` build output directory:

```json
{
  "config": {
    "watch": {
      "manifest": {
        "paths": ["manifest.json"]
      },
      "app": {
        "paths": ["dist/"]
      }
    }
  }
}
```

A restart _clears application state_. This is usually desired! Just be aware of this if you're testing workflows that require data to persist across multiple user interactions.

### Using a remote manifest

By default, the Slack CLI uses your local manifest and reinstalls the app if it changes. If you've created your manifest in App Settings, however, you'll have `manifest.source` set to `remote` in `.slack/config.json` and will need to manually reinstall the app if you change the manifest. 