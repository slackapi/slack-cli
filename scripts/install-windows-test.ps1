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
$qualifiedCommand = "Microsoft.PowerShell.Archive\Expand-Archive"
$installerPaths = @(
  (Join-Path $PSScriptRoot "install-windows.ps1"),
  (Join-Path $PSScriptRoot "install-windows-dev.ps1")
)

foreach ($installerPath in $installerPaths) {
  $tokens = $null
  $parseErrors = $null
  $ast = [System.Management.Automation.Language.Parser]::ParseFile(
    $installerPath,
    [ref]$tokens,
    [ref]$parseErrors
  )

  if ($parseErrors.Count -gt 0) {
    throw "PowerShell parser errors in $installerPath"
  }

  $archiveCommands = $ast.FindAll({
      param($node)
      $node -is [System.Management.Automation.Language.CommandAst] -and
        $node.GetCommandName() -like "*Expand-Archive"
    }, $true)

  if ($archiveCommands.Count -ne 1 -or $archiveCommands[0].GetCommandName() -ne $qualifiedCommand) {
    throw "$installerPath must call $qualifiedCommand exactly once"
  }
}

$testRoot = Join-Path ([System.IO.Path]::GetTempPath()) "slack-cli-install-$([guid]::NewGuid())"
$sourcePath = Join-Path $testRoot "source"
$destinationPath = Join-Path $testRoot "destination"
$archivePath = Join-Path $testRoot "slack_cli.zip"

try {
  New-Item -ItemType Directory -Path (Join-Path $sourcePath "bin") -Force | Out-Null
  Set-Content -LiteralPath (Join-Path $sourcePath "bin\slack.exe") -Value "test"
  Microsoft.PowerShell.Archive\Compress-Archive -Path (Join-Path $sourcePath "*") -DestinationPath $archivePath

  function Expand-Archive {
    throw "The shadowing command should not be called"
  }

  Microsoft.PowerShell.Archive\Expand-Archive -LiteralPath $archivePath -DestinationPath $destinationPath -Force

  if (!(Test-Path -LiteralPath (Join-Path $destinationPath "bin\slack.exe"))) {
    throw "The qualified archive command did not extract bin\slack.exe"
  }
}
finally {
  if (Test-Path -LiteralPath $testRoot) {
    Remove-Item -LiteralPath $testRoot -Recurse -Force
  }
}
