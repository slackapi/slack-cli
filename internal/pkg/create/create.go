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

package create

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/cmd/doctor"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/archiveutil"
	"github.com/slackapi/slack-cli/internal/cmdutil"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/deputil"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/slacktrace"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

// CreateArgs are the arguments passed into the Create function
type CreateArgs struct {
	AppName   string
	Template  Template
	GitBranch string
}

// Create will create a new Slack app on the file system and app manifest on the Slack API.
func Create(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, createArgs CreateArgs) (appDirPath string, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.create")
	defer span.Finish()

	// Get the current directory to use as the base for the project
	workingDirPath, err := os.Getwd()
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
	}

	// Get the app selection and accompanying app directory name (this may change when we find the unique directory name)
	appDirName, err := getAppDirName(createArgs.AppName)
	if err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
	}

	// Get the project's full directory path
	projectDirPath := ""
	if filepath.IsLocal(appDirName) {
		projectDirPath = filepath.Join(workingDirPath, appDirName)
		projectDirPath, err = getAvailableDir(ctx, projectDirPath)
		if err != nil {
			return "", slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
		}
		appDirPath, err = filepath.Rel(workingDirPath, projectDirPath)
		if err != nil {
			return "", slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
		}
	} else {
		projectDirPath = filepath.Join(appDirName)
		projectDirPath, err = getAvailableDir(ctx, projectDirPath)
		if err != nil {
			return "", slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
		}
		appDirPath, err = filepath.Abs(projectDirPath)
		if err != nil {
			return "", slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
		}
	}

	// Update the app's directory name now that the unique directory is created
	appDirName = filepath.Base(projectDirPath)

	// Print a bunch of information about the progress of the command to traces
	// and debugs and the standard output here
	clients.IO.PrintTrace(ctx, slacktrace.CreateStart)
	clients.IO.PrintTrace(ctx, slacktrace.CreateProjectPath, projectDirPath)
	clients.IO.PrintDebug(ctx, fmt.Sprintf("creating a new project called '%s'", appDirName))
	clients.IO.PrintDebug(ctx, fmt.Sprintf("cloning project from template '%s'", createArgs.Template.path))
	if createArgs.GitBranch != "" {
		clients.IO.PrintDebug(ctx, fmt.Sprintf("cloning project from branch '%s'", createArgs.GitBranch))
	}
	clients.IO.PrintDebug(ctx, fmt.Sprintf("writing project to path '%s'", projectDirPath))
	projectDetails := []string{
		fmt.Sprintf("Cloning template %s", style.Highlight(createArgs.Template.GetTemplatePath())),
	}
	projectPathF := style.HomePath(projectDirPath)
	projectDetails = append(
		projectDetails,
		fmt.Sprintf("To path %s", style.Highlight(projectPathF)),
	)
	clients.IO.PrintInfo(ctx, false, "\n%s", style.Sectionf(style.TextSection{
		Emoji:     "open_file_folder",
		Text:      "Created a new Slack project",
		Secondary: projectDetails,
	}))

	// Create the project from a templateURL
	if err := createApp(ctx, projectDirPath, createArgs.Template, createArgs.GitBranch, log, clients.Fs); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrAppCreate)
	}

	// Change into the project directory to configure defaults and dependencies
	// then return to the starting directory
	if err = os.Chdir(projectDirPath); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrAppDirectoryAccess)
	}
	defer func() {
		_ = os.Chdir(workingDirPath)
	}()

	// Update default project files' app name, bot name, etc
	if err := app.UpdateDefaultProjectFiles(clients.Fs, projectDirPath, appDirName); err != nil {
		return "", slackerror.Wrap(err, slackerror.ErrProjectFileUpdate)
	}

	// Install project dependencies to add CLI support and cache dev dependencies.
	// CLI created projects always default to config.ManifestSourceLocal.
	InstallProjectDependencies(ctx, clients, projectDirPath, config.ManifestSourceLocal)
	clients.IO.PrintTrace(ctx, slacktrace.CreateDependenciesSuccess)

	// Notify listeners that app directory is created
	log.Log("info", "on_app_create_completion")

	return appDirPath, nil
}

// generateRandomAppName will create a random app name based on two words and a number
func generateRandomAppName() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	var firstRandomNum = rand.Intn(len(adjectives))
	var secondRandomNum = rand.Intn(len(animals))
	var randomName = fmt.Sprintf("%s-%s-%d", adjectives[firstRandomNum], animals[secondRandomNum], rand.Intn(1000))
	return randomName
}

// getAppDirName will validate and return the app's directory name
func getAppDirName(appName string) (string, error) {
	if len(appName) <= 0 {
		return generateRandomAppName(), nil
	}

	// trim whitespace
	appName = strings.ReplaceAll(appName, " ", "")

	// name cannot be a reserved word
	if goutils.Contains(reserved, appName, false) {
		return "", fmt.Errorf("the app name you entered is reserved")
	}
	return appName, nil
}

// getAvailableDir will return a unique directory path.
// If dirPath already exists, then a unique numbered path will be appended to the path.
func getAvailableDir(ctx context.Context, dirPath string) (string, error) {
	var span, _ = opentracing.StartSpanFromContext(ctx, "getAvailableDirectory")
	defer span.Finish()

	// Verify that the parent directory path exists (we will not create missing directories)
	if exists, err := parentDirExists(dirPath); err != nil {
		return "", err
	} else if !exists {
		return "", slackerror.New(slackerror.ErrAppDirectoryAccess)
	}

	// If the directory does not exist, then it is available
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return dirPath, nil
	}

	// Attempt to find a unique directory name by adding a number to the end starting at 1
	for i := 1; i < 100; i++ {
		var uniqueDirPath = fmt.Sprintf("%s-%d", dirPath, i)
		if _, err := os.Stat(uniqueDirPath); os.IsNotExist(err) {
			return uniqueDirPath, nil
		}
	}

	return "", slackerror.New("directory already exists and failed to create a unique directory")
}

// parentDirExists will check if the parent directory of dirPath exists
func parentDirExists(dirPath string) (bool, error) {
	parentDirPath := filepath.Dir(dirPath)
	if _, err := os.Stat(parentDirPath); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

// createApp will create the app directory using the default app template or a specified template URL.
func createApp(ctx context.Context, dirPath string, template Template, gitBranch string, log *logger.Logger, fs afero.Fs) error {
	log.Data["templatePath"] = template.path
	log.Data["isGit"] = template.isGit
	log.Data["gitBranch"] = gitBranch
	log.Data["isSample"] = template.IsSample()

	// Notify listeners
	log.Log("info", "on_app_create_template")

	if template.isGit {
		doctorSection, err := doctor.CheckGit(ctx)
		if doctorSection.HasError() || err != nil {
			zipFileURL := generateGitZipFileURL(template.path, gitBranch)
			if zipFileURL == "" {
				return slackerror.New(slackerror.ErrGitZipDownload)
			}
			resp, err := http.Get(zipFileURL)
			if err != nil {
				return slackerror.New(slackerror.ErrGitZipDownload)
			}

			defer resp.Body.Close()

			zipFile := dirPath + ".zip"
			out, err := os.Create(zipFile)
			if err != nil {
				return slackerror.Wrap(err, "error copying remote template")
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return slackerror.Wrap(err, "error copying remote template")
			}

			_, err = archiveutil.Unzip(zipFile, dirPath)
			if err != nil {
				err = slackerror.Wrapf(err, "failed to extract the remote template archive")
				return slackerror.Wrapf(err, slackerror.ErrGitZipDownload)
			}

			entries, _ := os.ReadDir(dirPath)
			tmpFolder := filepath.Join(dirPath, entries[0].Name())

			copyDirectoryOpts := goutils.CopyDirectoryOpts{
				Src: tmpFolder,
				Dst: dirPath,
			}
			if err := goutils.CopyDirectory(copyDirectoryOpts); err != nil {
				return slackerror.Wrap(err, "error copying remote template")
			}
			_ = fs.RemoveAll(tmpFolder)
			err = os.Remove(zipFile)
			if err != nil {
				err = slackerror.Wrapf(err, "failed to remove the remote template archive")
				return slackerror.Wrapf(err, slackerror.ErrGitZipDownload)
			}

		} else {
			// Use go-git to clone repo
			cloneOptions := git.CloneOptions{
				URL:   template.path,
				Depth: 1,
			}
			// Set ReferenceName to be the branch
			if gitBranch != "" {
				cloneOptions = git.CloneOptions{
					URL:           template.path,
					ReferenceName: plumbing.NewBranchReferenceName(gitBranch),
					Depth:         1,
				}
			}
			_, err := git.PlainClone(dirPath, false, &cloneOptions)
			if err != nil {
				if err.Error() == "authentication required" {
					gitArgs := createGitArgs(template.path, dirPath, gitBranch)
					gitCloneCommand := exec.Command("git", gitArgs...)
					if output, err := gitCloneCommand.CombinedOutput(); err != nil {
						errMsg := fmt.Sprintf(
							"%s\n%s\n\n%s",
							"The following output was printed during the clone command",
							fmt.Sprintf("%s %s", "git", strings.Join(gitArgs, " ")),
							strings.TrimSpace(string(output)),
						)
						return slackerror.New(slackerror.ErrGitClone).
							WithMessage("An error occurred while cloning the repository").
							WithDetails(slackerror.ErrorDetails{
								slackerror.ErrorDetail{
									Message: errMsg,
								},
							})
					}
				} else {
					return slackerror.New(slackerror.ErrGitClone)
				}
			}
		}
		// Remove .github folder if it's a sample app
		if template.IsSample() {
			_ = fs.RemoveAll(filepath.Join(dirPath, ".github"))
		}

	} else if template.isLocal {
		// Copy the local directory template
		//
		// Existing history is ignored when starting on a new app from a template.
		// Vendored dependencies are skipped since these can error from symlinks.
		copyDirectoryOpts := goutils.CopyDirectoryOpts{
			Src:               template.path,
			Dst:               dirPath,
			IgnoreDirectories: []string{".git", ".venv", "node_modules"},
			IgnoreFiles:       []string{".DS_Store"},
		}
		if err := goutils.CopyDirectory(copyDirectoryOpts); err != nil {
			return slackerror.Wrap(err, "error copying local template")
		}
	}

	// Clean the new project by removing unnecessary files, such as a .git directory
	// TODO - what other files should we remove? Keeping `.gitignore` is important for the project integrity
	err := fs.RemoveAll(filepath.Join(dirPath, ".git"))
	if err != nil {
		return slackerror.Wrap(err, "Error removing .git directory from project")
	}

	return nil
}

// InstallProjectDependencies installs the project runtime dependencies or
// continues with next steps if that fails. You can specify the manifestSource
// for the project configuration file (default: ManifestSourceLocal)
func InstallProjectDependencies(
	ctx context.Context,
	clients *shared.ClientFactory,
	projectDirPath string,
	manifestSource config.ManifestSource,
) []string {
	var outputs []string

	// Start the spinner
	spinnerText := fmt.Sprintf(
		"Installing project dependencies %s",
		style.Secondary("(this may take a few seconds)"),
	)
	spinner := style.NewSpinner(clients.IO.WriteErr())
	spinner.Update(spinnerText, "").Start()

	// Stop the spinner when the function returns
	defer func() {
		spinnerText = style.Sectionf(style.TextSection{
			Text:      "Installed project dependencies",
			Secondary: outputs,
		})
		spinner.Update(spinnerText, "package").Stop()
	}()

	// Initialize the project runtime
	if err := clients.InitRuntime(ctx, projectDirPath); err != nil {
		clients.IO.PrintDebug(ctx, "Error detecting the runtime of the project: %s", err)
	} else {
		clients.IO.PrintDebug(ctx, fmt.Sprintf("Detected a project using %s", style.Highlight(clients.Runtime.Name())))
	}

	// Create a .slack directory
	dotSlackDirPath, err := config.CreateProjectConfigDir(ctx, clients.Fs, projectDirPath)
	dotSlackDirPathRel, _ := filepath.Rel(filepath.Dir(projectDirPath), dotSlackDirPath)
	dotSlackDirPathRelStyled := style.Highlight(dotSlackDirPathRel)

	switch {
	case os.IsExist(err):
		outputs = append(outputs, fmt.Sprintf("Found %s", dotSlackDirPathRelStyled))
	case err != nil:
		outputs = append(outputs, fmt.Sprintf("Error adding the directory %s: %s", dotSlackDirPathRel, err))
	default:
		outputs = append(outputs, fmt.Sprintf("Added %s", dotSlackDirPathRelStyled))
	}

	// Create .slack/.gitignore file
	if clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		gitignoreFilePath, err := config.CreateProjectConfigDirDotGitIgnoreFile(clients.Fs, projectDirPath)
		dotSlackDirPathRel, _ := filepath.Rel(filepath.Dir(projectDirPath), gitignoreFilePath)
		dotSlackDirPathRelStyled := style.Highlight(dotSlackDirPathRel)

		switch {
		case os.IsExist(err):
			outputs = append(outputs, fmt.Sprintf("Found %s", dotSlackDirPathRelStyled))
		case err != nil:
			outputs = append(outputs, fmt.Sprintf("Error adding the file %s: %s", dotSlackDirPathRel, err))
		default:
			outputs = append(outputs, fmt.Sprintf("Added %s", dotSlackDirPathRelStyled))
		}
	}

	// Create .slack/config.json file
	if clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		configJSONFilePath, err := config.CreateProjectConfigJSONFile(clients.Fs, projectDirPath)
		configJSONFilePathRel, _ := filepath.Rel(filepath.Dir(projectDirPath), configJSONFilePath)
		configJSONFilePathRelStyled := style.Highlight(configJSONFilePathRel)

		switch {
		case os.IsExist(err):
			outputs = append(outputs, fmt.Sprintf("Found %s", configJSONFilePathRelStyled))
		case err != nil:
			outputs = append(outputs, fmt.Sprintf("Error adding the file %s: %s", configJSONFilePathRel, err))
		default:
			outputs = append(outputs, fmt.Sprintf("Added %s", configJSONFilePathRelStyled))
		}
	}

	// Create .slack/hooks.json file
	if clients.Config.WithExperimentOn(experiment.BoltFrameworks) {
		var hooksJSONTemplate = []byte("{}")
		if clients.Runtime != nil {
			hooksJSONTemplate = clients.Runtime.HooksJSONTemplate()
		}

		hooksJSONFilePath, err := config.CreateProjectHooksJSONFile(clients.Fs, projectDirPath, hooksJSONTemplate)
		hooksJSONFilePathRel, _ := filepath.Rel(filepath.Dir(projectDirPath), hooksJSONFilePath)
		hooksJSONFilePathRelStyled := style.Highlight(hooksJSONFilePathRel)

		switch {
		case os.IsExist(err):
			outputs = append(outputs, fmt.Sprintf("Found %s", hooksJSONFilePathRelStyled))
		case err != nil:
			outputs = append(outputs, fmt.Sprintf("Error adding the file %s: %s", hooksJSONFilePathRel, err))
		default:
			outputs = append(outputs, fmt.Sprintf("Added %s", hooksJSONFilePathRelStyled))
		}
	}

	// Set "project_id" in .slack/config.json
	if projectID, err := clients.Config.ProjectConfig.InitProjectID(ctx, true); err != nil {
		clients.IO.PrintDebug(ctx, "Error initializing a project_id: %s", err)
	} else {
		// Used by Logstash
		clients.Config.ProjectID = projectID

		// Used by OpenTracing
		if span := opentracing.SpanFromContext(ctx); span != nil {
			span.SetTag("project_id", clients.Config.ProjectID)
			ctx = opentracing.ContextWithSpan(ctx, span)
		}
	}

	// Set "manifest.source" in .slack/config.json
	if !manifestSource.Exists() {
		manifestSource = config.ManifestSourceLocal
	}

	// Non-ROSI projects default to ManifestSourceRemote
	if clients.Config.WithExperimentOn(experiment.BoltInstall) {
		isSlackHostedProject := cmdutil.IsSlackHostedProject(ctx, clients) == nil
		if !isSlackHostedProject {
			manifestSource = config.ManifestSourceRemote
		}
	}

	if err := clients.Config.ProjectConfig.SetManifestSource(ctx, manifestSource); err != nil {
		clients.IO.PrintDebug(ctx, "Error setting manifest source in project-level config: %s", err)
	} else {
		configJSONFilename := config.ProjectConfigJSONFilename
		manifestSourceStyled := style.Highlight(manifestSource.Human())

		outputs = append(outputs, fmt.Sprintf(
			"Updated %s manifest source to %s",
			configJSONFilename,
			manifestSourceStyled,
		))
	}

	// Install the runtime's dependencies
	if clients.Runtime == nil {
		return outputs
	}

	output, err := clients.Runtime.InstallProjectDependencies(ctx, projectDirPath, clients.HookExecutor, clients.IO, clients.Fs, clients.Os)
	if err != nil {
		clients.IO.PrintDebug(ctx, "error installing project dependencies: %s", err)

		// Output raw installation message
		if len(strings.TrimSpace(output)) > 0 {
			outputs = append(outputs, output)
			outputs = append(outputs, " ")
		}

		// Output error returned
		if len(err.Error()) > 0 {
			outputs = append(outputs,
				"Error: "+style.Highlight(err.Error()),
				" ",
			)
		}

		// Output suggested next steps
		outputs = append(outputs, style.Darken("Manually install project dependencies before proceeding with development"))
	} else {
		// Output raw installation message
		if len(strings.TrimSpace(output)) > 0 {
			outputs = append(outputs, strings.TrimSpace(output))
		}
	}

	return outputs
}

// generateGitZipFileURL will return template's GitHub zip file download link
// In the future, this function can be extended to support other Git hosts, such as GitLab.
// TODO, @cchensh, we should get prepared for other non-Git hosts and refactor the create pkg
func generateGitZipFileURL(templateURL string, gitBranch string) string {
	zipURL := strings.Replace(templateURL, ".git", "", -1) + "/archive/refs/heads/"

	if gitBranch == "" {
		mainURL := zipURL + "main.zip"
		masterURL := zipURL + "master.zip"
		zipURL = deputil.URLChecker(mainURL)
		if zipURL == "" {
			zipURL = deputil.URLChecker(masterURL)
		}
	} else {
		zipURL = zipURL + gitBranch + ".zip"
	}
	return zipURL
}

func createGitArgs(templatePath string, dirPath string, gitBranch string) []string {
	gitArgs := []string{"clone", "--depth=1", templatePath, dirPath}
	gitBranch = strings.Trim(gitBranch, " ")

	// GitBranchFlag
	if gitBranch != "" {
		gitArgs = append(gitArgs, "--branch", gitBranch)
	}

	return gitArgs
}
