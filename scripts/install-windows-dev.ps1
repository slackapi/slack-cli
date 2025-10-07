# Copyright 2022-2025 Salesforce, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

param(
  [Parameter(HelpMessage = "Alias of Slack CLI")]
  [string]$Alias,

  [Parameter(HelpMessage = "Version of Slack CLI")]
  [string]$Version = "dev",

  [Parameter(HelpMessage = "Skip Git installation")]
  [bool]$SkipGit = $false
)

# As this script is for internal usage only, we should set SLACK_DISABLE_TELEMETRY environment variable
[System.Environment]::SetEnvironmentVariable('SLACK_DISABLE_TELEMETRY', $true)

Function delay ([float]$seconds, [string]$message, [string]$newlineOption) {
  if ($newlineOption -eq "-n") {
    Write-Host -NoNewline $message
  }
  else {
    Write-Host $message
  }
  Start-Sleep -Seconds $seconds
}

function check_slack_binary_exist() {
  param(
    [Parameter(HelpMessage = "Alias of Slack CLI")]
    [string]$Alias,

    [Parameter(HelpMessage = "Version of Slack CLI")]
    [string]$Version,

    [Parameter(HelpMessage = "Display diagnostic information")]
    [boolean]$Diagnostics
  )
  $FINGERPRINT = "d41d8cd98f00b204e9800998ecf8427e"
  $SLACK_CLI_NAME = "slack"
  if ($alias) {
    $SLACK_CLI_NAME = $alias
  }
  if (Get-Command $SLACK_CLI_NAME -ErrorAction SilentlyContinue) {
    if ($Diagnostics) {
      delay 0.3 "Checking if ``$SLACK_CLI_NAME`` already exists on this system..."
      delay 0.2 "Heads up! A binary called ``$SLACK_CLI_NAME`` was found!"
      delay 0.3 "Now checking if it's the same Slack CLI..."
    }
    if ($Version) {
      $SLACK_CLI_VERSION = $Version
      $message = "It is the same Slack CLI! Switching over to v$Version..."
    }
    if ($Diagnostics) {
      delay 0.3 "$message`n"
    }
  }
  return $SLACK_CLI_NAME
}

function install_slack_cli {
  param(
    [Parameter(HelpMessage = "Alias of Slack CLI")]
    [string]$Alias,

    [Parameter(HelpMessage = "Version of Slack CLI")]
    [string]$Version
  )

  delay 0.6 "Hello and welcome! Now beginning to install the..."

  delay 0.1 "      ________ _     _    _____ _    __    _____ _    ________"
  delay 0.1 "     /  ______/ |   / \ /  ____/ | /  /  /  ____/ | /___   __/"
  delay 0.1 "    /______  |  |  / _ \  |   |      /   | |   |  |    |  |   "
  delay 0.1 "     ____ /  |  |___ __ \ |____  |\  \   | |____  |__ _|  |___"
  delay 0.1 "   /_______ /|______/  \_\ ____/_| \__\    _____/______/_____/"
  delay 0.2 ""

  $confirmed_alias = check_slack_binary_exist $Alias $Version $true
  $error.clear()
  try {
    if ($Version) {
      $SLACK_CLI_VERSION = $Version
    }
    else {
      Write-Host "Finding the latest Slack CLI release version"
      $cli_info = Invoke-RestMethod -Uri "https://api.slack.com/slackcli/metadata.json"
      $SLACK_CLI_VERSION = $cli_info.'slack-cli'.releases.version[0]
    }
  }
  catch {
    Write-Error "Installer cannot find latest Slack CLI release version"
    throw
  }

  $slack_cli_dir = "${Home}\AppData\Local\slack-cli"
  try {
    if (!(Test-Path $slack_cli_dir)) {
      try {
        New-Item $slack_cli_dir -ItemType Directory | Out-Null
      }
      catch {
        $alternative_slack_cli_dir = "${Home}\.slack-cli"
        if (!(Test-Path $alternative_slack_cli_dir)) {
          try {
            New-Item $alternative_slack_cli_dir -ItemType Directory | Out-Null
            $slack_cli_dir = $alternative_slack_cli_dir
          }
          catch {
            Write-Error "Installer cannot create folder in $($alternative_slack_cli_dir). `nPlease manually create $($slack_cli_dir) folder and re-run the installation script"
            throw
          }
        }
      }
    }
  }
  catch {
    Write-Error "Installer cannot create folder for Slack CLI, `nPlease manually create $($slack_cli_dir) folder and re-run the installation script"
    throw
  }

  if ($Version -eq "dev") {
    Write-Host "Downloading the latest development build..."
  }
  else {
    Write-Host "Downloading Slack CLI v$SLACK_CLI_VERSION..."
  }
  try {
    Invoke-WebRequest -Uri "https://downloads.slack-edge.com/slack-cli/slack_cli_$($SLACK_CLI_VERSION)_windows_64-bit.zip" -OutFile "$($slack_cli_dir)\slack_cli.zip"
  }
  catch {
    Write-Error "Installer cannot download Slack CLI"
    throw
  }

  $slack_cli_bin_dir = "$($slack_cli_dir)\bin"
  $slack_cli_binary_path = "$($slack_cli_dir)\bin\slack.exe"
  $slack_cli_new_binary_path = "$($slack_cli_dir)\bin\${confirmed_alias}.exe"

  delay 0.3 "Extracting the executable to:`n   $slack_cli_new_binary_path"
  Expand-Archive "$($slack_cli_dir)\slack_cli.zip" -DestinationPath "$($slack_cli_dir)" -Force
  Move-Item -Path $slack_cli_binary_path -Destination $slack_cli_new_binary_path -Force

  $User = [System.EnvironmentVariableTarget]::User
  $Path = [System.Environment]::GetEnvironmentVariable('Path', $User)
  if (!(";${Path};".ToLower() -like "*;${slack_cli_bin_dir};*".ToLower())) {
    Write-Host "Adding ``$confirmed_alias.exe`` to your Path environment variable"
    [System.Environment]::SetEnvironmentVariable('Path', $Path.TrimEnd(';') + ";${slack_cli_bin_dir}", $User)
    $Env:Path = $Env:Path.TrimEnd(';') + ";$slack_cli_bin_dir"
  }
  Remove-Item "$($slack_cli_dir)\slack_cli.zip"
}

function install_git {
  param(
    [Parameter(HelpMessage = "Skip Git installation")]
    [bool]$SkipGit = $false
  )
  if ($SkipGit) {
    Write-Host "Skipping the check for a Git installation!"
  }
  else {
    try {
      git | Out-Null
      Write-Host "Git is already installed. Nice!"
    }
    catch [System.Management.Automation.CommandNotFoundException] {
      Write-Host "Git is not installed. Installing now..."

      $MIN_GIT_VERSION = "2.40.0"
      $exePath = "$env:TEMP\git.exe"

      Invoke-WebRequest -Uri https://github.com/git-for-windows/git/releases/download/v$($MIN_GIT_VERSION).windows.1/Git-$($MIN_GIT_VERSION)-64-bit.exe -UseBasicParsing -OutFile $exePath

      Start-Process $exePath -ArgumentList '/VERYSILENT /NORESTART /NOCANCEL /SP- /CLOSEAPPLICATIONS /RESTARTAPPLICATIONS /COMPONENTS="icons,ext\reg\shellhere,assoc,assoc_sh"' -Wait

      [Environment]::SetEnvironmentVariable('Path', "$([Environment]::GetEnvironmentVariable('Path', 'Machine'));C:\Program Files\Git\bin", 'Machine')

      foreach ($level in "Machine", "User") {
        [Environment]::GetEnvironmentVariables($level).GetEnumerator() | % {
          if ($_.Name -match 'Path$') {
            $_.Value = ($((Get-Content "Env:$($_.Name)") + ";$($_.Value)") -split ';' | Select -unique) -join ';'
          }
          $_
        } | Set-Content -Path { "Env:$($_.Name)" }
      }
      Write-Host "Git is installed and ready!"
    }
  }
}

function terms_of_service {
  param(
    [Parameter(HelpMessage = "Alias of Slack CLI")]
    [string]$Alias
  )
  $confirmed_alias = check_slack_binary_exist $Alias $Version $false
  # if (Get-Command $confirmed_alias) {
  Write-Host "`nUse of the Slack CLI should comply with the Slack API Terms of Service:"
  Write-Host "   https://slack.com/terms-of-service/api"
  # }
}

function feedback_message {
  param(
    [Parameter(HelpMessage = "Alias of Slack CLI")]
    [string]$Alias
  )
  $confirmed_alias = check_slack_binary_exist $Alias $Version $false
  if (Get-Command $confirmed_alias) {
    Write-Host "`nWe would love to know how things are going. Really. All of it."
    Write-Host "   Survey your development experience with ``$confirmed_alias feedback``"
  }
}

function next_step_message {
  param(
    [Parameter(HelpMessage = "Alias of Slack CLI")]
    [string]$Alias
  )
  $confirmed_alias = check_slack_binary_exist $Alias $Version $false
  if (Get-Command $confirmed_alias -ErrorAction SilentlyContinue) {
    try {
      $confirmed_alias | Out-Null
      Write-Host "`nYou're all set! Relaunch your terminal to ensure changes take effect."
      Write-Host "   Then, authorize your CLI in your workspace with ``$confirmed_alias login``.`n"
    }
    catch {
      Write-Error "Slack CLI was not installed."
      Write-Host "`nFind help troubleshooting: https://docs.slack.dev/tools/slack-cli"
      throw
    }
  }
}

trap {
  Write-Host "`nWe would love to know how things are going. Really. All of it."
  Write-Host "Submit installation issues: https://github.com/slackapi/slack-cli/issues"
  exit 7
}

install_slack_cli $Alias $Version
Write-Host "`nAdding developer tooling for an enhanced experience..."
install_git $SkipGit
Write-Host "Sweet! You're all set to start developing!"
terms_of_service $Alias
# feedback_message $Alias
# next_step_message $Alias
