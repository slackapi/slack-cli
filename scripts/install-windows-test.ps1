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

$installers = @(
  "install-windows.ps1",
  "install-windows-dev.ps1"
)

foreach ($name in $installers) {
  $installer = Join-Path $PSScriptRoot $name

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

  if ($archiveCommands.Count -lt 1) {
    throw "$installer calls no Expand-Archive; expected $qualifiedCommand (issue #651)"
  }
  $unqualified = $archiveCommands | Where-Object { $_.GetCommandName() -ne $qualifiedCommand }
  if ($unqualified) {
    throw "$installer must call $qualifiedCommand (found unqualified Expand-Archive; issue #651)"
  }

  Write-Host "$name calls $qualifiedCommand ($($archiveCommands.Count)x), all qualified (issue #651 guard passed)"
}
