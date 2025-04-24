---
sidebar_label: Installing for MacOS & Linux
slug: /slack-cli/guides/installing-the-slack-cli-for-mac-and-linux
---

# Installing the Slack CLI for Mac & Linux


The Slack CLI is a set of tools critical to building workflow apps. This is your one-stop shop for those tools.

✨ **If you've not used the Slack CLI before, we recommend following our [Deno Slack SDK getting started guide](/deno-slack-sdk/guides/getting-started) instead**. We'll still get your wagon loaded up before you depart for the trail, but we'll also give you some additional guidance.

⤵️ **If you need to authorize the Slack CLI, [go here](/slack-cli/guides/authorizing-the-slack-cli)**.

:::info

The minimum required Slack CLI version for Enterprise Grid as of September 19th, 2023 is `v2.9.0`. If you attempt to log in with an older version, you'll receive a `cli_update_required` error from the Slack API. Run `slack upgrade` to get the latest version.

:::

<Tabs groupId="installation">
<TabItem value="Automated" label="Automated Installation">

**Run the automated installer from your terminal window:**

```zsh
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash
```

This will install the Slack CLI and all required dependencies, including [Deno](/deno-slack-sdk/guides/installing-deno),
the runtime environment for workflow apps. If you have VSCode installed,
the [VSCode Deno
extension](https://marketplace.visualstudio.com/items?itemName=denoland.vscode-deno)
will be installed.
<details>
<summary>Optional: Use an alias for the Slack CLI binary</summary>

If you have another CLI tool in your path called `slack`, you can rename the slack binary to a different name before you add it to your path.

To do this, pass the `-s` argument to the installer script:

```zsh
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s <your-preferred-alias>
```

The alias you use should come after any flags used in the installation script. For example, if you use both flags noted below to pass a version and skip the Deno installation, your install script might look like this:

```
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s -- -v 2.1.0 -d <your-preferred-alias>
```

You can also copy the Slack CLI into any folder that is already in your path (such as `/usr/local/bin`&mdash;you can use`echo $PATH` to find these), or add a new folder to your path by listing the folder you installed the Slack CLI to in `/etc/paths`.

If you don't rename the slack binary to a different name, the installation script will detect existing binaries named `slack` and bail if it finds one&mdash;it will not overwrite your existing `slack` binary.

</details>

<details>
<summary>Optional: customize installation using flags</summary>

There are two optional flags available to customize the installation.

1. Specify a version you'd like to install using the version flag, `-v`. The absence of this flag will ensure the latest Slack CLI version is installed.
```
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s -- -v 2.1.0
```

2. Skip the Deno installation by using the `-d` flag, like this:
```
curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash -s -- -d
```
</details>

<details>
<summary>Troubleshooting</summary>

#### Errors

Error: _Failed to create a symbolic link! The installer doesn't have write access to /usr/local/bin. Please check permission and try again..._

Solution: Sudo actions within the scripts were removed so as not to create any security concerns. The `$HOME` env var is updated to `/root` &mdash; however, the installer is using `$HOME` for both Deno and the SDK install, which causes the whole install to be placed under `/root`, making both Deno and the SDK unusable for users without root permissions.
* For users who do not have root permissions, run the sudo actions manually as follows: `sudo mkdir -p -m 775 /usr/local/bin`, then `sudo ln -sf "$slack_cli_bin_path" "/usr/local/bin/$SLACK_CLI_NAME"` where `$slack_cli_bin_path` is typically `$HOME/.slack/bin/slack` and `$SLACK_CLI_NAME` is typically the alias (by default it’s `slack`).
* For users who do have root permissions, you can run the installation script as `sudo curl -fsSL https://downloads.slack-edge.com/slack-cli/install.sh | bash`. In this case, the script is executed as root.

</details>
</TabItem>
<TabItem value="Manual" label="Manual Installation">

**1\. Download and install [Deno](https://deno.land).** Refer to [Install Deno](/deno-slack-sdk/guides/installing-deno) for more details.

**2\. Verify that Deno is installed and in your path.**

The minimum version of Deno runtime required is currently version 1.37.0.

```bash
$ deno --version
deno 1.46.2* (release, x86_64-apple-darwin)
v8 10.*
typescript 4.*
```

**3\. Download and install
   [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git), a
   dependency of the** `slack` **CLI.**

**4\. Download the** `slack` **CLI installer for your environment.**

<ts-icon class="ts_icon_apple"></ts-icon> &nbsp; <a href="https://downloads.slack-edge.com/slack-cli/slack_cli_3.0.5_macOS_64-bit.tar.gz"><strong>Download for macOS (.tar.gz)</strong></a>

<ts-icon class="ts_icon_plug"></ts-icon> &nbsp; <a href="https://downloads.slack-edge.com/slack-cli/slack_cli_3.0.5_linux_64-bit.tar.gz"><strong>Download for Linux (.tar.gz)</strong></a>

**5\. Add the** `slack` **CLI to your path.**

:::info
  <p><strong>Existing</strong> <code>slack</code> <strong>binary in path?</strong></p>
  <p>If you have another CLI tool in your path called <code>slack</code>, we recommend renaming our slack binary to a different name before adding it to your path. See the <strong>Automated installation</strong> tab for more details.</p>
:::

**6\. Verify that** `slack` **is installed and in your path.**

```
$ slack version
Using slack v3.0.5
```

**7\. Verify that all dependencies have been installed.**

Run the following command:

```
$ slack doctor
```

**A few notes about hooks**

If you have upgraded your CLI version but your `deno-slack-hooks` version is less than `v1.3.0`, when running `slack doctor`, you will see the following near the end of the output:

```
✔ Configurations (your project's CLI settings)
        Project ID: 1a2b3c4d-ef5g-67hi-8j9k1l2m3n4o

    ✘ Runtime (foundations for the application)
        Error: The `doctor` hook was not found (sdk_hook_not_found)
        Suggestion: Ensure this hook is implemented in your `slack.json`

    ✔ Dependencies (requisites for development)
        deno_slack_hooks: 1.2.3 → 1.3.0 (supported version)
```

In addition, if you attempt to run the `slack run` command without this dependency installed, you will see a similar error in your console:

```
🚫 The `start` script was not found (sdk_hook_not_found)

    💡 Suggestion
        Hook scripts are defined in the Slack configuration file ('slack.json').
        Every app requires a 'slack.json' file and you can find a working example at:
        https://github.com/slack-samples/deno-starter-template/blob/main/slack.json

```

Ensure that `deno-slack-hooks` is installed at the project level and that the version is not less than `v1.3.0`.

**8\. [Install the VSCode extension for
   Deno](/deno-slack-sdk/guides/installing-deno#vscode) (recommended).**

</TabItem>
</Tabs>
