# Copyright 2022-2026 Salesforce, Inc.
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

$ErrorActionPreference = "Stop"

$qualifiedCommand = "Microsoft.PowerShell.Archive\Expand-Archive" # https://github.com/slackapi/slack-cli/issues/651
$installDir = Join-Path $env:LOCALAPPDATA "slack-cli"

$cases = @(
  @{ Installer = "install-windows.ps1"; Alias = "slack-test"; Version = "4.6.0"; ExpectVersion = $true },
  @{ Installer = "install-windows-dev.ps1"; Alias = "slack-dev-test"; Version = "dev"; ExpectVersion = $false }
)

function Remove-Install {
  if (Test-Path -LiteralPath $installDir) {
    Remove-Item -LiteralPath $installDir -Recurse -Force
  }
}

foreach ($case in $cases) {
  $installer = Join-Path $PSScriptRoot $case.Installer
  $aliasBinary = Join-Path $installDir "bin\$($case.Alias).exe"

  $tokens = $null
  $parseErrors = $null
  $ast = [System.Management.Automation.Language.Parser]::ParseFile(
    $installer,
    [ref]$tokens,
    [ref]$parseErrors
  )

  if ($parseErrors.Count -gt 0) {
    throw "PowerShell parser errors in $installer"
  }

  $archiveCommands = $ast.FindAll({
      param($node)
      $node -is [System.Management.Automation.Language.CommandAst] -and
        $node.GetCommandName() -like "*Expand-Archive"
    }, $true)

  if ($archiveCommands.Count -ne 1 -or $archiveCommands[0].GetCommandName() -ne $qualifiedCommand) {
    throw "$installer must call $qualifiedCommand exactly once"
  }

  try {
    Remove-Install

    & $installer -Alias $case.Alias -Version $case.Version -SkipGit $true

    if (!(Test-Path -LiteralPath $aliasBinary)) {
      throw "$($case.Installer) did not place $($case.Alias).exe at $aliasBinary (extraction failed?)"
    }

    $versionOutput = & $aliasBinary --version
    if ($case.ExpectVersion -and $versionOutput -notmatch [regex]::Escape($case.Version)) {
      throw "Version mismatch: expected '$($case.Version)' in output, got '$versionOutput'"
    }

    Write-Host "$($case.Installer) E2E passed: $($case.Alias).exe installed, version '$versionOutput'"
  }
  finally {
    Remove-Install
  }
}
