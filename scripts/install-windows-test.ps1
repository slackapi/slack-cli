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

# Guard against the Pscx-shadowing regression from https://github.com/slackapi/slack-cli/issues/651:
# Pscx 3.3.2 ships its own Expand-Archive that shadows the built-in, so the extraction call must be
# qualified as Microsoft.PowerShell.Archive\Expand-Archive. We assert this statically (parse the AST,
# not run the installer) because the installers' post-install courtesy check hangs under pwsh 7 on
# current Windows runner images -- `& slack _fingerprint | Tee-Object -Variable | Out-Null` blocks
# with no console -- so a real end-to-end install is not yet runnable in CI. Tracked separately.
$qualifiedCommand = "Microsoft.PowerShell.Archive\Expand-Archive"

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

  # The invariant is "no unqualified extraction call," not a fixed count: every
  # *Expand-Archive invocation must be the qualified form, and there must be at
  # least one (so a removed/renamed extraction can't pass vacuously). This won't
  # break if an installer legitimately extracts more than once.
  if ($archiveCommands.Count -lt 1) {
    throw "$installer calls no Expand-Archive; expected $qualifiedCommand (issue #651)"
  }
  $unqualified = $archiveCommands | Where-Object { $_.GetCommandName() -ne $qualifiedCommand }
  if ($unqualified) {
    throw "$installer must call $qualifiedCommand (found unqualified Expand-Archive; issue #651)"
  }

  Write-Host "$name calls $qualifiedCommand ($($archiveCommands.Count)x), all qualified (issue #651 guard passed)"
}
