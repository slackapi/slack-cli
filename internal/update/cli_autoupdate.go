// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

type cliUpgrade struct {
	// The path to the compressed file we've downloaded that contains the new CLI to install.
	// The backup and new version directories are created in the same directory as this file.
	// This file will be removed after extraction.
	UpgradeArchivePath string
	// The resolved path to the existing binary we're going to upgrade. This file will be replaced
	// with a hard link to the new binary, after its been copied to the backup directory.
	ExistingBinaryPath string
	// The name of the current executable. After upgrade, we'll run the `version` command
	// against this executable and verify the reported version matches `NewVersion`.
	ExecutablePath string
	// The version of the current binary being upgraded.
	CurrentVersion string
	// The new version of the binary.
	NewVersion string

	// A debug logging function, to cut down on dependencies being passed around
	PrintDebug func(format string, a ...interface{})
}

// InstallUpdate grabs the latest version of the CLI binary and installs it onto the system
func (c *CLIDependency) InstallUpdate(ctx context.Context) error {
	fmt.Print(style.SectionSecondaryf("Starting the auto-update..."))

	fmt.Print(style.SectionSecondaryf("Downloading version %s...", c.releaseInfo.Version))

	architecture := runtime.GOARCH
	operatingSys := runtime.GOOS
	c.clients.IO.PrintDebug(ctx, "Architecture: %s\nOS: %s", architecture, operatingSys)
	fileName, err := getUpdateFileName(c.releaseInfo.Version, operatingSys, architecture)
	if err != nil {
		return err
	}

	c.clients.IO.PrintDebug(ctx, "Downloading %s", fileName)

	link := "https://downloads.slack-edge.com/slack-cli/" + fileName

	configDir, err := c.clients.Config.SystemConfig.SlackConfigDir(ctx)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to get config directory")
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	c.clients.IO.PrintDebug(ctx, "Downloading cli from %s", link)
	dstFilePath := path.Join(configDir, fileName)
	err = download(link, dstFilePath)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to download the binary")
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}
	c.clients.IO.PrintDebug(ctx, "Downloaded to: %s", dstFilePath)

	executablePath, err := os.Executable()
	if err != nil {
		err = slackerror.Wrapf(err, "failed to get current process name")
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	currentBinaryPath, err := getCurrentResolvedBinaryPath(ctx, c.clients, executablePath)
	if err != nil {
		return err
	}

	cliUpgradeData := cliUpgrade{
		UpgradeArchivePath: dstFilePath,
		CurrentVersion:     c.version,
		NewVersion:         c.releaseInfo.Version,
		ExecutablePath:     executablePath,
		ExistingBinaryPath: currentBinaryPath,
		PrintDebug: func(format string, a ...interface{}) {
			c.clients.IO.PrintDebug(ctx, format, a...)
		},
	}

	return UpgradeFromLocalFile(cliUpgradeData)
}

// UpgradeFromLocalFile performs an upgrade given an upgrade configuration.
func UpgradeFromLocalFile(cliUpgradeData cliUpgrade) error {

	baseDir := filepath.Dir(cliUpgradeData.UpgradeArchivePath)

	// Folder in the Slack config folder that contains the different versions of the CLI
	binaryFolderPath := filepath.Join(baseDir, "bin")
	updatedBinaryFolderPath := filepath.Join(binaryFolderPath, cliUpgradeData.NewVersion)

	fmt.Print(style.SectionSecondaryf("Extracting the download: %s", updatedBinaryFolderPath))
	updatedBinaryPath, err := ExtractArchive(cliUpgradeData.UpgradeArchivePath, updatedBinaryFolderPath)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to extract the upgrade archive")
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	// Deletes the archive file after extraction
	err = os.Remove(cliUpgradeData.UpgradeArchivePath)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to remove the downloaded archive")
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	backupPathToOldBinary, err := backupBinary(cliUpgradeData, binaryFolderPath)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to backup binary")
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	fmt.Print(style.SectionSecondaryf("Updating to version %s: %s", cliUpgradeData.NewVersion, cliUpgradeData.ExistingBinaryPath))

	cliUpgradeData.PrintDebug("Creating hardlink for updated %s binary to %s and placing it in %s",
		cliUpgradeData.NewVersion, updatedBinaryPath, cliUpgradeData.ExistingBinaryPath)
	err = os.Link(updatedBinaryPath, cliUpgradeData.ExistingBinaryPath)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to create hardlink")
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	binaryName := filepath.Base(cliUpgradeData.ExistingBinaryPath)
	cliUpgradeData.PrintDebug("Successfully updated %s from %s to %s", binaryName, cliUpgradeData.CurrentVersion, cliUpgradeData.NewVersion)
	fmt.Print(style.SectionSecondaryf("Successfully updated to version %s", cliUpgradeData.NewVersion))

	err = verifyUpgrade(cliUpgradeData)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to verify new install")
		restoreErr := restoreBinary(updatedBinaryFolderPath, backupPathToOldBinary, cliUpgradeData.ExistingBinaryPath)
		if restoreErr != nil {
			restoreErr = slackerror.Wrapf(err, "failed to restore backup")
			return slackerror.Wrapf(restoreErr, slackerror.ErrCLIAutoUpdate)
		}
		return err
	}

	return nil
}

// download does an HTTP GET to the given url and places its response payload in dstFilePath
func download(url string, dstFilePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = slackerror.Wrapf(err, "http status code %d, not 200", resp.StatusCode)
		return slackerror.Wrap(err, slackerror.ErrCLIAutoUpdate)
	}

	// Create the file
	out, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return err
}

// backupBinary copies the existing binary into a backups folder in binaryFolderPath, returning a path to the backed up binary
func backupBinary(cliUpgradeData cliUpgrade, binaryFolderPath string) (string, error) {
	// Create folder for the current binary to be moved to
	backupFolderForOldBinaries := filepath.Join(binaryFolderPath, "backups")
	backupFolderForCurrentBinary := filepath.Join(backupFolderForOldBinaries, cliUpgradeData.CurrentVersion)
	fmt.Print(style.SectionSecondaryf("Backing up current install: %s", backupFolderForCurrentBinary))

	err := os.MkdirAll(backupFolderForCurrentBinary, os.ModePerm)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to make backup directory")
		return "", slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	binaryName := filepath.Base(cliUpgradeData.ExistingBinaryPath)
	cliUpgradeData.PrintDebug("Binary Name: %s", binaryName)

	// move existing binary to a backup location
	newPathToOldBinary := filepath.Join(backupFolderForOldBinaries, cliUpgradeData.CurrentVersion, binaryName)
	err = os.Rename(cliUpgradeData.ExistingBinaryPath, newPathToOldBinary)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to move current binary to backup directory")
		return "", slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}
	cliUpgradeData.PrintDebug("Moved old %s binary from %s to: %s", cliUpgradeData.CurrentVersion, cliUpgradeData.ExistingBinaryPath, newPathToOldBinary)
	return newPathToOldBinary, nil
}

// restoreBinary undoes a partial upgrade and replaces the old binary in its original location
func restoreBinary(updatedBinaryFolderPath string, newPathToOldBinary string, originalOldBinaryLocation string) error {
	// Delete the folder create for the new binary
	err := os.RemoveAll(updatedBinaryFolderPath)
	if err != nil {
		return err
	}

	// Delete the current hardlink that points to the updated binary
	err = os.Remove(originalOldBinaryLocation)
	if err != nil {
		return err
	}

	// Create a hardlink from the old binary's location to the original old binary location
	err = os.Link(newPathToOldBinary, originalOldBinaryLocation)
	if err != nil {
		return err
	}

	return nil
}

// getUpdateFilename returns name of the archive that contains the upgrade for
// the given version and OS.
//
// All possible OS/architecture combinations can be listed with the command:
//
//	go tool dist list | column -c 75 | column -t
func getUpdateFileName(version, operatingSys, architecture string) (filename string, err error) {
	switch operatingSys {
	case "darwin":
		switch architecture {
		case "amd64":
			filename = fmt.Sprintf("slack_cli_%s_macOS_amd64.zip", version)
		case "arm64":
			filename = fmt.Sprintf("slack_cli_%s_macOS_arm64.zip", version)
		default:
			filename = fmt.Sprintf("slack_cli_%s_macOS_64-bit.zip", version)
		}
	case "linux":
		filename = fmt.Sprintf("slack_cli_%s_linux_64-bit.tar.gz", version)
	case "windows":
		filename = fmt.Sprintf("slack_cli_%s_windows_64-bit.zip", version)
	default:
		err = slackerror.New(fmt.Sprintf("auto-updating for the operating system (%s) is unsupported", operatingSys))
		err = slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}
	return
}

// getCurrentResolvedBinaryPath returns the resolved path to the current executable which should be overwritten by the upgrade,
// following symlinks
func getCurrentResolvedBinaryPath(ctx context.Context, clients *shared.ClientFactory, executablePath string) (string, error) {
	// the executablePath may either be the symlink path or the actual path to the binary
	// Gets the path to the current binary
	clients.IO.PrintDebug(ctx, "Path to current binary: %s", executablePath)

	// Get the name of the current binary
	binaryName := filepath.Base(executablePath)
	clients.IO.PrintDebug(ctx, "Binary Name: %s", binaryName)

	// Gets the path to the actual binary if the executablePath is a symlink, otherwise it just returns the executablePath
	// We always want the actual binary path, not the symlink. By replacing the actual binary with the updated binary, the symlink will be updated as well
	pathToOldBinary, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		err = slackerror.Wrapf(err, "failed to get path to actual binary")
		return "", slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}
	clients.IO.PrintDebug(ctx, "isSymLink: %s", pathToOldBinary)
	if pathToOldBinary == executablePath {
		clients.IO.PrintDebug(ctx, "Binary is not a symlink")
	} else {
		clients.IO.PrintDebug(ctx, "Binary is a symlink")
	}

	fmt.Print(style.SectionSecondaryf("Found current install path: %s", pathToOldBinary))
	return pathToOldBinary, nil
}

// verifyUpgrade runs the `version` command on the newly upgraded binary and ensures that it returns the expected version number
func verifyUpgrade(cliUpgradeData cliUpgrade) error {
	command := exec.Command(cliUpgradeData.ExecutablePath, "--version")

	envars := os.Environ()

	// We remove the SLACK_TEST_VERSION env var from the environment, so that the version check will not fail
	for i, envar := range envars {
		// replace the env var we want to remove with the last element, then remove the last element
		if strings.Contains(envar, "SLACK_TEST_VERSION") {
			envars[i] = envars[len(envars)-1]
			envars = envars[:len(envars)-1]
		}
	}

	// we clear the environment variables here to avoid having them passed to the new binary when it's tested
	// this is particularly relevant when the SLACK_TEST_VERSION var is set to mock a version
	command.Env = envars

	commandOutput, err := command.CombinedOutput()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			err = slackerror.New(fmt.Sprintf("running '%s version' after upgrade failed with error: %d.",
				cliUpgradeData.ExecutablePath, exitError.ExitCode()))
			return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
		}
	}

	cliUpgradeData.PrintDebug("Output of %s %s: \n \n %s", command.Args[0], command.Args[1], string(commandOutput))

	fmt.Print(style.SectionSecondaryf("Verifying the update..."))
	versionMatches := strings.Contains(string(commandOutput), cliUpgradeData.NewVersion)
	if !versionMatches {
		err = slackerror.New(fmt.Sprintf("versions did not match. Expected version %s was not in the version output: %s.",
			cliUpgradeData.NewVersion, string(commandOutput)))
		return slackerror.Wrapf(err, slackerror.ErrCLIAutoUpdate)
	}

	return nil
}
