# Experiments

The Slack CLI has an experiment (`-e`) flag behind which we put features currently under development. These features may be fleeting, may not be perfectly polished, and many will not work without a development instance - but we have made them available for use "at your own risk."

## Available experiments

The following is a list of currently available experiments. We'll remove experiments from this page if we decide they are no longer needed or once they are released, in which case we'll make an announcement about the feature's general availability in the [developer changelog](https://docs.slack.dev/changelog).

* `bolt-install`: enables creating, installing, and running Bolt projects that manage their app manifest on app settings (remote manifest).
    * `slack create` and `slack init` now set manifest source to "app settings" (remote) for Bolt JS & Bolt Python projects ([PR#96](https://github.com/slackapi/slack-cli/pull/96)).
    * `slack run` supports creating and installing Bolt Framework apps that have the manifest source set to "app settings (remote)" ([PR#111](https://github.com/slackapi/slack-cli/pull/111)).
    * `--experiment bolt-install` is now enabled by default.
* `read-only-collaborators`: enables creating and modifying collaborator permissions via the `slack collaborator` commands.

## Experiments changelog

Below is a list of updates related to experiments.

* **June 2025**: Enabled the `bolt-install` experiment by default to support using `slack run` with apps that have a manifest source set to "app settings" (remote).
* **May 2025**: Added the experiment `bolt-install` to enable creating, installing, and running Bolt projects that manage their app manifest on app settings (remote manifest).
* **February 2025**: Added full Bolt framework support to the Slack CLI and removed the features from behind the experiment flag. See the changelog announcement [here](https://docs.slack.dev/changelog/2025/02/27/slack-cli-release).
* **August 2024**: Added the `bolt` experiment for the `slack create` command.
* **January 2024**: Added the experiment `read-only-collaborators`.

## Feedback

We love feedback from our community, so we encourage you to explore and interact with the [GitHub repo](https://github.com/slackapi/slack-cli). Contributions, bug reports, and any feedback are all helpful; let us nurture the Slack CLI together to help make building Slack apps more pleasant for everyone.
