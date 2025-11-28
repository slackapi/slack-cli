---
sidebar_label: Installing for Windows
slug: /tools/slack-cli/guides/installing-the-slack-cli-for-windows
---

# Installing the Slack CLI for Windows

The Slack CLI is a useful tool for building Slack apps. This is your one-stop shop for installing this tool.

:::warning[PowerShell is required for installing the Slack CLI on Windows machines; an alternative shell will not work.]

:::

<Tabs groupId="installation">
<TabItem value="Automated" label="Automated Installation">

**Run the automated installer from Windows PowerShell:**

```pwsh
irm https://downloads.slack-edge.com/slack-cli/install-windows.ps1 | iex
```

This will install the Slack CLI and configure the command.

Runtime installations are left to the developer and depend on the app being built. For more information and next steps, review the quickstart guides:

- [Bolt for JavaScript](/tools/bolt-js/getting-started)
- [Bolt for Python](/tools/bolt-python/getting-started)
- [Deno Slack SDK](/tools/deno-slack-sdk/guides/getting-started)

<details>
<summary>Optional: use an alias for the Slack CLI binary</summary>

If you have another CLI tool in your path called `slack`, you can rename this `slack` binary to a different name to avoid errors during installation. The Slack CLI won't overwrite the existing one!

To do this, use the `-Alias` flag as described within the **Optional: Customize installation using flags** section.

</details>

<details>
<summary>Optional: customize installation using flags</summary>

There are several flags available to customize the installation. Since flags cannot be passed to remote scripts, you must first download the automated installer to a local file:

```pwsh
irm https://downloads.slack-edge.com/slack-cli/install-windows.ps1 -outfile 'install-windows.ps1'
```

The available flags are:

| Flag       | Description                                                      | Example                                                                             |
| :--------- | :--------------------------------------------------------------- | :---------------------------------------------------------------------------------- |
| `-Alias`   | Installs the Slack CLI as the provided alias                     | `-Alias slackcli` will create a binary named `slackcli.exe` and add it to your path |
| `-Version` | Installs a specific version of the Slack CLI                     | `-Version 2.1.0` installs version `2.1.0` of the Slack CLI                          |
| `-SkipGit` | If true, will not attempt to install Git when Git is not present | `-SkipGit $true` skips installing `git` if Git is not found                         |

You can also see all available flags by passing `-?` to the automated installer:

```pwsh
.\install-windows.ps1 -?
```

Here's an example invocation using every flag:

```pwsh
.\install-windows.ps1 -Version 2.1.0 -Alias slackcli -SkipGit $true
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

Manual installation allows you to customize certain paths used when installing the Slack CLI. Runtime installations are omitted from these steps but are still required to run an app.

**1\. Download and install [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git), a dependency of the** `slack` **CLI.**

**2\. Download the** `slack` **CLI installer for your environment.**

<ts-icon class="ts_icon_windows"></ts-icon> &nbsp; <a href="https://downloads.slack-edge.com/slack-cli/slack_cli_3.10.0_windows_64-bit.zip"><strong>Windows (.zip)</strong></a>

**3\. Add the** `slack` **CLI to your path.**

Copy the Slack CLI into any folder that is already in your path, or add a new folder to your path by adding the folder you installed the Slack CLI to in your Environment Variables. You may not have access to edit System variables, so you might need to add it to your account's User variables. You can open the Environment Variables dialog by pressing the `Win`+`R` keys to open the Run window, and then entering the following command:

```pwsh
rundll32.exe sysdm.cpl,EditEnvironmentVariables
```

We recommend using an alias if another `slack` binary exists. To do this, change the alias used at the end of the symbolic link to something else that makes sense.

**4\. Verify that** `slack` **is installed and in your path:**

```pwsh
$ slack version
Using slack v3.10.0
```

</TabItem>
</Tabs>

## Installing PowerShell {#powershell}

Run the following command to install PowerShell 7 on your machine:

```pwsh
iex "& { $(irm https://aka.ms/install-powershell.ps1) } -UseMSI"
```

The following articles may also be helpful should you run into any issues:

- [Installing PowerShell on Windows](https://learn.microsoft.com/en-us/powershell/scripting/install/installing-powershell-on-windows?view=powershell-7.4)
- [How to install and update PowerShell 6](https://www.thomasmaurer.ch/2019/03/how-to-install-and-update-powershell-6/)
