# Experiments

The Slack CLI has an experiment flag, behind which we put features under development and may cause breaking changes. These features are fleeting and many will not work without a development instance, but are available for use "at your own risk."

## Available experiments

The following is a list of currently available experiments. We may remove an experiment once the feature is released.

* `bolt-install`: enables creating, installing, and running Bolt projects that manage their app manifest on app settings (remote manifest).
    * `slack create` and `slack init` now set manifest source to "app settings" (remote) for Bolt JS & Bolt Python projects.
* `read-only-collaborators`: enables creating and modifying collaborator permissions via the `slack collaborator` commands.
