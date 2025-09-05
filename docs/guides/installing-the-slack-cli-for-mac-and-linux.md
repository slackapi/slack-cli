---
sidebar_label: Installing for MacOS & Linux
slug: /tools/slack-cli/guides/installing-the-slack-cli-for-mac-and-linux
---

# Installing the Slack CLI for Mac & Linux

The Slack CLI is a useful tool for building Slack apps. This is your one-stop shop for installing this tool.

<Tabs groupId="installation">
<TabItem value="Automated" label="Automated Installation">

**Run the automated installer from your terminal window:**

```sh
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash
```

This will install the Slack CLI and configure the command.

Runtime installations are left to the developer and depend on the app being built. For more information and next steps, review the quickstart guides:

- [Bolt for JavaScript](/tools/bolt-js/getting-started)
- [Bolt for Python](/tools/bolt-python/getting-started)
- [Deno Slack SDK](/tools/deno-slack-sdk/guides/getting-started)

<details>
<summary>Optional: Use an alias for the Slack CLI binary</summary>

If you have another CLI tool in your path called `slack`, you can rename this `slack` binary to a different name to avoid errors during installation. The Slack CLI won't overwrite the existing one!

To do this, pass the `-s` argument and an alias to the automated installer:

```sh
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s <your-preferred-alias>
```

The alias you use should come after any flags used in the installer. For example, if you use the version flag your installation script might look like this:

```sh
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s -- -v 2.1.0 <your-preferred-alias>
```

</details>

<details>
<summary>Optional: Download a specific version</summary>

The latest Slack CLI version is installed by default, but a particular version can be pinned using the `-v` flag:

```sh
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s -- -v 2.1.0
```

</details>

<details>
<summary>Troubleshooting: Failed to create a symbolic link</summary>

The automated installer attempts to add the Slack CLI to a known directory in the path: `/usr/local/bin`.

Some machines protect this directory and elevated access might be needed for updated permissions. We do not recommend using `sudo` with the automated installer since unexpected side effects might place downloads in the wrong spot.

If root access is available, create a symbolic link to the downloaded Slack CLI with the following command:

```sh
sudo ln -s "$HOME/.slack/bin/slack" "/usr/local/bin/slack"
```

We recommend using an alias if another `slack` binary exists. To do this, change the alias used at the end of the symbolic link to something else that makes sense.

For users without root permissions, please follow the **Manual Installation** steps.

</details>
</TabItem>
<TabItem value="Manual" label="Manual Installation">

Manual installation allows you to customize certain paths used when installing the Slack CLI. Runtime installations are omitted from these steps but are still required to run an app.

**1\. Download and install [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git), a dependency of the** `slack` **CLI.**

**2\. Download the** `slack` **CLI installer for your environment.**

🍎 ⚡️ [**Download for macOS Apple Silicon (.tar.gz)**](https://downloads.slack-edge.com/slack-cli/slack_cli_3.6.1_macOS_arm64.tar.gz)

🍏 🪨 [**Download for macOS Intel (.tar.gz)**](https://downloads.slack-edge.com/slack-cli/slack_cli_3.6.1_macOS_amd64.tar.gz)

🐧 💾 [**Download for Linux (.tar.gz)**](https://downloads.slack-edge.com/slack-cli/slack_cli_3.6.1_linux_64-bit.tar.gz)

**3\. Add the** `slack` **CLI to your path.**

Create a symbolic link to the Slack CLI download from (or move the downloaded binary to) any folder that is already in your path.

In the following example we download the Slack CLI to the `.slack` directory and create a symbolic link to `.local` path:

```sh
ln -s "$HOME/.slack/bin/slack" "$HOME/.local/bin/slack"
```

We recommend using an alias if another `slack` binary exists. To do this, change the alias used at the end of the symbolic link to something else that makes sense.

**4\. Verify that** `slack` **is installed and in your path.**

```sh
$ slack version
Using slack v3.6.1
```

</TabItem>
</Tabs>
