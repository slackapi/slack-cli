---
slug: /slack-cli/guides/uninstalling-the-slack-cli
---

# Uninstalling the Slack CLI

All good things come to an end! If you need to uninstall the Slack CLI, run the commands below. Note that these instructions will uninstall the Slack CLI, but not its dependencies. Follow [these instructions](https://docs.deno.com/runtime/manual/tools/script_installer#uninstall) to uninstall Deno.

✨  **Just need to uninstall an app?** Refer to [uninstall an app from your workspace](/deno-slack-sdk/guides/creating-an-app#uninstall-app).


<Tabs groupId="operating-systems">
<TabItem value="nix" label="MacOS & Linux uninstallation">


Run the following commands in your terminal window

```
$ rm -rf ~/.slack  # Delete the download binary
$ rm /usr/local/bin/slack  # Delete the command alias (replacing `slack` with a command name)
```

</TabItem>
<TabItem value="win" label="Windows uninstallation">


The command binary is stored in `$HOME\AppData\Local\slack-cli` or `$HOME\.slack-cli`. This binary can be removed with the following command:

```
rd -r $HOME\AppData\Local\slack-cli
```

where `$HOME` is substituted for the full path, e.g. `C:\Users\<your_username>`.

:::warning

**As with installation, PowerShell is required for uninstallation of the Slack CLI from Windows machines.** An alternative shell will not work.

:::

Removing the command from the `$env:path` can be done with the following command:

```
$env:Path = ($env:Path -split ';' -ne  "$HOME\AppData\Local\slack-cli\bin") -join ';'
```

Removing the command from the system path can be done with the following command:

```
[System.Environment]::SetEnvironmentVariable('Path', (([System.Environment]::GetEnvironmentVariable('Path', [System.EnvironmentVariableTarget]::User) -split ';' -ne '$HOME\AppData\Local\slack-cli\bin') -join ';'), [System.EnvironmentVariableTarget]::User)
```

Finally, general configurations can be removed with the following command:

```
rd -r $HOME\.slack
```

To check that uninstallation was successful, run the following commands and verify that you receive an error &mdash; in this case, that's a good thing!

```
slack version
echo $env:path
```

</TabItem>
</Tabs>

Until next time!
