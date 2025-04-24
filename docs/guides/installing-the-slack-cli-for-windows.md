---
sidebar_label: Installing for Windows
slug: /slack-cli/guides/installing-the-slack-cli-for-windows
---

# Installing the Slack CLI for Windows


The Slack CLI is a set of tools critical to building workflow apps. This is your one-stop shop for those tools.

✨ **If you've not used the Slack CLI before, we recommend following our [Deno Slack SDK getting started guide](/deno-slack-sdk/guides/getting-started) instead**. We'll still get your wagon loaded up before you depart for the trail, but we'll also give you some additional guidance.

⤵️ **If you need to authorize the Slack CLI, [go here](/slack-cli/guides/authorizing-the-slack-cli)**.

:::info

The minimum required Slack CLI version for Enterprise Grid as of September 19th, 2023 is `v2.9.0`. If you attempt to log in with an older version, you'll receive a `cli_update_required` error from the Slack API. Run `slack upgrade` to get the latest version.

:::

:::warning

**PowerShell is required for installing the Slack CLI on Windows machines.** An alternative shell will not work.

:::

<Tabs groupId="installation">
<TabItem value="Automated" label="Automated Installation">

**Run the automated installer from Windows PowerShell:**

```zsh
irm https://downloads.slack-edge.com/slack-cli/install-windows.ps1 | iex
```

:::warning

PowerShell is required for installing the Slack CLI on Windows machines; an alternative shell will not work.

:::

This will install the Slack CLI and all required dependencies, including [Deno](/deno-slack-sdk/guides/installing-deno),
the runtime environment for workflow apps. If you have VSCode installed,
the [VSCode Deno
extension](https://marketplace.visualstudio.com/items?itemName=denoland.vscode-deno)
will be installed.

<details>
<summary>Optional: Use an alias for the Slack CLI binary</summary>

If you have another CLI tool in your path called `slack`, you can rename the slack binary to a different name before you add it to your path.

To do this, copy the Slack CLI into any folder that is already in your path, or add a new folder to your path by listing the folder you installed the Slack CLI to in your Environment Variables. You may not have access to edit System variables, so you might need to add it to your account's User variables. You can open the Environment Variables dialog by pressing the `Win`+`R` keys to open the Run window, and then entering the following command:

```pwsh
rundll32.exe sysdm.cpl,EditEnvironmentVariables
```

You can also use the `-Alias` flag as described within **Optional: customize installation using flags**.

</details>

<details>
<summary>Optional: customize installation using flags</summary>

There are several flags available to customize the installation. Since flags
cannot be passed to remote scripts, you must first download the installation
script to a local file:

```zsh
irm https://downloads.slack-edge.com/slack-cli/install-windows.ps1 -outfile 'install-windows.ps1'
```

The available flags are:

| Flag | What it does | Example |
| :--  | :--          | :--     |
| `-Alias` | Installs the Slack CLI as the provided alias | `-Alias slackcli` will create a binary named `slackcli.exe` and add it to your path |
| `-Version` | Installs a specific version of the Slack CLI | `-Version 2.1.0` installs version `2.1.0` of the Slack CLI |
| `-SkipGit` | If true, will not attempt to install Git when Git is not present | `-SkipGit $true` |
| `-SkipDeno` | If true, will not attempt to install Deno when Deno is not present | `-SkipDeno $true` |

You can also see all available flags by passing `-?` to the installation script:

```zsh
.\install-windows.ps1 -?
```

Here's an example invocation using every flag:

```zsh
.\install-windows.ps1 -Version 2.1.0 -Alias slackcli -SkipGit $true -SkipDeno $true
```

</details>

<details>
<summary>Troubleshooting</summary>

#### Errors

Error: _Not working? You may need to update your session's Language Mode._

Solution: For the installer to work correctly, your PowerShell session's [language mode](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_language_modes?view=powershell-7.3#what-is-a-language-mode) will need to be set to `FullLanguage`. To check your session's language mode, run the following in your PowerShell window: `ps $ExecutionContext.SessionState.LanguageMode`. To run the installer, your session's language mode will need to be `FullLanguage`. If it's not, you can set your session's language mode to `FullLanguage` with the following command: `ps $ExecutionContext.SessionState.LanguageMode = "FullLanguage"`

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

<ts-icon class="ts_icon_windows"></ts-icon> &nbsp; <a href="https://downloads.slack-edge.com/slack-cli/slack_cli_5_windows_64-bit.zip"><strong>Windows (.zip)</strong></a>

**5\. Add the** `slack` **CLI to your path.**

:::info
  <p><strong>Existing</strong> <code>slack</code> <strong>binary in path?</strong></p>
  <p>If you have another CLI tool in your path called <code>slack</code>, we recommend renaming our slack binary to a different name before adding it to your path. See the <strong>Automated installation</strong> tab for more details.</p>

:::

**6\. Verify that** `slack` **is installed and in your path:**
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

## Installing PowerShell {#powershell}

Run the following command to install PowerShell 7 on your machine:

```pwsh
iex "& { $(irm https://aka.ms/install-powershell.ps1) } -UseMSI"
```

The following articles may also be helpful should you run into any issues:

* [Installing PowerShell on Windows](https://learn.microsoft.com/en-us/powershell/scripting/install/installing-powershell-on-windows?view=powershell-7.4)
* [How to install and update PowerShell 6](https://www.thomasmaurer.ch/2019/03/how-to-install-and-update-powershell-6/)
