// Copyright 2022-2026 Salesforce, Inc.
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

package config

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/cache"
	"github.com/slackapi/slack-cli/internal/experiment"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

const (
	// ProjectHooksJSONFilename is the project-level hooks.json filename
	ProjectHooksJSONFilename = "hooks.json"

	// ProjectConfigDirName is the name of the project-level configuration directory
	ProjectConfigDirName = ".slack"

	// ProjectConfigJSONFilename is the project-level config.json filename
	ProjectConfigJSONFilename = "config.json"
)

//go:embed dotgitignore
var dotGitignoreFileData []byte

// ProjectConfigManager is the interface for interacting with the project config
type ProjectConfigManager interface {
	InitProjectID(ctx context.Context, overwriteExistingProjectID bool) (string, error)
	GetProjectID(ctx context.Context) (string, error)
	SetProjectID(ctx context.Context, projectID string) (string, error)
	GetManifestSource(ctx context.Context) (ManifestSource, error)
	GetSurveyConfig(ctx context.Context, name string) (SurveyConfig, error)
	SetSurveyConfig(ctx context.Context, name string, surveyConfig SurveyConfig) error

	Cache() cache.Cacher
}

// ProjectConfig is the project-level config file
type ProjectConfig struct {
	Experiments []experiment.Experiment `json:"experiments,omitempty"`
	Manifest    *ManifestConfig         `json:"manifest,omitempty"`
	ProjectID   string                  `json:"project_id,omitempty"`
	Surveys     map[string]SurveyConfig `json:"surveys,omitempty"`

	// fs is the file system module that's shared by all packages and enables testing & mock of the file system
	fs afero.Fs

	// os is the `os` package that's shared by all packages and enables testing & mocking
	os types.Os
}

// NewProjectConfig read and writes to the project-level configuration file
func NewProjectConfig(fs afero.Fs, os types.Os) *ProjectConfig {
	projectConfig := &ProjectConfig{
		fs: fs,
		os: os,
	}

	return projectConfig
}

// InitProjectID will set the project_id in the project-level config when it's unset
// and then returns the project_id.
func (c *ProjectConfig) InitProjectID(ctx context.Context, overwriteExistingProjectID bool) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "InitProjectID")
	defer span.Finish()

	// Check that current directory is a project
	if _, err := GetProjectDirPath(c.fs, c.os); err != nil {
		return "", err
	}

	projectID, err := c.GetProjectID(ctx)
	if err != nil {
		return "", err
	}

	// Return if the Project ID exists and we don't need to overwrite it
	if projectID != "" && !overwriteExistingProjectID {
		return projectID, nil
	}

	projectID, err = c.SetProjectID(ctx, uuid.New().String())

	return projectID, err
}

// GetProjectID reads the project_id from the project-level config file
func (c *ProjectConfig) GetProjectID(ctx context.Context) (string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetProjectID")
	defer span.Finish()

	var projectConfig, err = ReadProjectConfigFile(ctx, c.fs, c.os)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(projectConfig.ProjectID), nil
}

// SetProjectID sets the project_id to a random UUID string in the project-level config file
func (c *ProjectConfig) SetProjectID(ctx context.Context, projectID string) (string, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "SetProjectID")
	defer span.Finish()

	var projectConfig, err = ReadProjectConfigFile(ctx, c.fs, c.os)
	if err != nil {
		return "", err
	}

	projectConfig.ProjectID = projectID

	_, err = WriteProjectConfigFile(ctx, c.fs, c.os, projectConfig)
	if err != nil {
		return "", err
	}

	return projectConfig.ProjectID, nil
}

// GetManifestSource finds the manifest source preference for the project
func (c *ProjectConfig) GetManifestSource(ctx context.Context) (ManifestSource, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetManifestSource")
	defer span.Finish()

	var projectConfig, err = ReadProjectConfigFile(ctx, c.fs, c.os)
	if err != nil {
		return "", err
	}

	if projectConfig.Manifest != nil {
		source := ManifestSource(strings.TrimSpace(projectConfig.Manifest.Source))
		switch {
		case source.Equals(ManifestSourceLocal), source.Equals(ManifestSourceRemote):
			return source, nil
		case !source.Exists():
			return ManifestSourceLocal, nil
		default:
			return "", slackerror.New(slackerror.ErrProjectConfigManifestSource)
		}
	}

	return ManifestSourceLocal, nil
}

// SetManifestSource saves the manifest source preference for the project
func SetManifestSource(ctx context.Context, fs afero.Fs, os types.Os, source ManifestSource) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "SetManifestSource")
	defer span.Finish()
	projectConfig, err := ReadProjectConfigFile(ctx, fs, os)
	if err != nil {
		return err
	}
	if projectConfig.Manifest == nil {
		projectConfig.Manifest = &ManifestConfig{}
	}
	projectConfig.Manifest.Source = source.String()
	_, err = WriteProjectConfigFile(ctx, fs, os, projectConfig)
	if err != nil {
		return err
	}
	return nil
}

// GetSurveyConfig reads the survey for the given survey ID from the project-level config file
func (c *ProjectConfig) GetSurveyConfig(ctx context.Context, name string) (SurveyConfig, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "GetSurveyConfig")
	defer span.Finish()

	var projectConfig, err = ReadProjectConfigFile(ctx, c.fs, c.os)
	if err != nil {
		return SurveyConfig{}, err
	}

	survey, ok := projectConfig.Surveys[name]
	if !ok {
		return SurveyConfig{}, slackerror.New(slackerror.ErrSurveyConfigNotFound)
	}

	return survey, nil
}

// SetSurveyConfig writes the survey for the given survey ID from the project-level config file
func (c *ProjectConfig) SetSurveyConfig(ctx context.Context, name string, surveyConfig SurveyConfig) error {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "SetSurveyConfig")
	defer span.Finish()

	var projectConfig, err = ReadProjectConfigFile(ctx, c.fs, c.os)
	if err != nil {
		return err
	}

	projectConfig.Surveys[name] = SurveyConfig{
		AskedAt:     surveyConfig.AskedAt,
		CompletedAt: surveyConfig.CompletedAt,
	}

	_, err = WriteProjectConfigFile(ctx, c.fs, c.os, projectConfig)
	if err != nil {
		return err
	}

	return nil
}

// ReadProjectConfigFile reads the project-level config.json file
func ReadProjectConfigFile(ctx context.Context, fs afero.Fs, os types.Os) (ProjectConfig, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "ReadProjectConfigFile")
	defer span.Finish()

	var projectConfig ProjectConfig

	projectDirPath, err := GetProjectDirPath(fs, os)
	if err != nil {
		return projectConfig, err
	}

	if !ProjectConfigJSONFileExists(fs, os, projectDirPath) {
		return projectConfig, nil
	}

	var projectConfigFilePath = GetProjectConfigJSONFilePath(projectDirPath)
	projectConfigFileBytes, err := afero.ReadFile(fs, projectConfigFilePath)
	if err != nil {
		return projectConfig, err
	}

	err = json.Unmarshal(projectConfigFileBytes, &projectConfig)
	if err != nil {
		return projectConfig, slackerror.New(slackerror.ErrUnableToParseJSON).
			WithMessage("Failed to parse contents of project-level config file").
			WithRootCause(err).
			WithRemediation("Check that %s is valid JSON", style.HomePath(projectConfigFilePath))
	}
	if projectConfig.Surveys == nil {
		projectConfig.Surveys = map[string]SurveyConfig{}
	}

	return projectConfig, nil
}

// WriteProjectConfigFile writes the project-level config.json file
func WriteProjectConfigFile(ctx context.Context, fs afero.Fs, os types.Os, projectConfig ProjectConfig) (string, error) {
	var span opentracing.Span
	span, _ = opentracing.StartSpanFromContext(ctx, "WriteProjectConfigFile")
	defer span.Finish()

	projectConfigBytes, err := json.MarshalIndent(projectConfig, "", "  ")
	if err != nil {
		return "", err
	}

	projectDirPath, err := GetProjectDirPath(fs, os)
	if err != nil {
		return "", err
	}

	projectConfigFilePath := GetProjectConfigJSONFilePath(projectDirPath)
	err = afero.WriteFile(fs, projectConfigFilePath, projectConfigBytes, 0644)
	if err != nil {
		return "", err
	}

	return projectConfigFilePath, nil
}

// ProjectConfigJSONFileExists returns true if the .slack/config.json file exists
func ProjectConfigJSONFileExists(fs afero.Fs, os types.Os, projectDirPath string) bool {
	var projectConfigFilePath = GetProjectConfigJSONFilePath(projectDirPath)
	_, err := fs.Stat(projectConfigFilePath)
	return !os.IsNotExist(err)
}

// GetProjectDirPath returns the path to the project directory or an error if not a Slack project
// TODO(@mbrooks) Standardize the definition of a validate project directory and merge with `cmdutil.ValidProjectDirectoryOrExit`
func GetProjectDirPath(fs afero.Fs, os types.Os) (string, error) {
	currentDir, _ := os.Getwd()

	// Verify project-level hooks.json exists
	projectHooksJSONPath := GetProjectHooksJSONFilePath(currentDir)
	if _, err := fs.Stat(projectHooksJSONPath); os.IsNotExist(err) {

		// Fallback check for slack.json and .slack/slack.json file
		// DEPRECATED(semver:major): remove both fallbacks next major release
		projectSlackJSONPath := filepath.Join(currentDir, "slack.json")
		if _, err := fs.Stat(projectSlackJSONPath); err == nil {
			return currentDir, nil
		}
		projectDotSlackSlackJSONPath := filepath.Join(currentDir, ".slack", "slack.json")
		if _, err := fs.Stat(projectDotSlackSlackJSONPath); err == nil {
			return currentDir, nil
		}
		return "", slackerror.New(slackerror.ErrInvalidAppDirectory)
	}

	return currentDir, nil
}

// Cache loads the cached project values
func (c *ProjectConfig) Cache() cache.Cacher {
	path, err := GetProjectDirPath(c.fs, c.os)
	if err != nil {
		return &cache.Cache{}
	}
	return cache.NewCache(c.fs, c.os, path)
}

// GetProjectConfigDirPath returns the path to the project's config directory
func GetProjectConfigDirPath(projectDirPath string) string {
	return filepath.Join(projectDirPath, ProjectConfigDirName)
}

// CreateProjectConfigDir creates a .slack/ directory in projectDirPath and returns the path
func CreateProjectConfigDir(ctx context.Context, fs afero.Fs, projectDirPath string) (configDirPath string, err error) {
	configDirPath = GetProjectConfigDirPath(projectDirPath)

	if info, err := fs.Stat(configDirPath); err == nil && info.IsDir() {
		return configDirPath, os.ErrExist
	}

	// Create a .slack directory
	if err := fs.MkdirAll(configDirPath, 0755); err != nil {
		return configDirPath, err
	}

	return configDirPath, nil
}

// GetProjectConfigJSONFilePath returns the path to the project's config file
func GetProjectConfigJSONFilePath(projectDirPath string) string {
	return filepath.Join(projectDirPath, ProjectConfigDirName, ProjectConfigJSONFilename)
}

// CreateProjectConfigJSONFile creates a project-level config.json file or returns error if exists
func CreateProjectConfigJSONFile(fs afero.Fs, projectDirPath string) (path string, err error) {
	configJSONFilePath := GetProjectConfigJSONFilePath(projectDirPath)

	var projectConfig = ProjectConfig{}
	projectConfigBytes, err := json.MarshalIndent(projectConfig, "", "  ")
	if err != nil {
		return configJSONFilePath, err
	}

	// Check if config.json file exists
	if _, err = fs.Stat(configJSONFilePath); err == nil {
		return configJSONFilePath, os.ErrExist
	}

	// Write new config.json file
	err = afero.WriteFile(fs, configJSONFilePath, projectConfigBytes, 0o0644)
	if err != nil {
		return configJSONFilePath, err
	}

	return configJSONFilePath, nil
}

// GetProjectHooksJSONFilePath returns the path to the project's hooks file
func GetProjectHooksJSONFilePath(projectDirPath ...string) string {
	return filepath.Join(filepath.Join(projectDirPath...), ProjectConfigDirName, ProjectHooksJSONFilename)
}

// CreateProjectHooksJSONFile writes data to a new project hooks.json file
func CreateProjectHooksJSONFile(fs afero.Fs, projectDirPath string, data []byte) (path string, err error) {
	var hooksJSONFilePath = GetProjectHooksJSONFilePath(projectDirPath)

	// Check if hooks.json already exists
	if _, err = fs.Stat(hooksJSONFilePath); err == nil {
		return hooksJSONFilePath, os.ErrExist
	}

	// Check if slack.json already exists (deprecated)
	slackJSONFilePath := filepath.Join(projectDirPath, "slack.json")
	if _, err = fs.Stat(slackJSONFilePath); err == nil {
		return slackJSONFilePath, os.ErrExist
	}

	// Create the hooks.json file
	if err = afero.WriteFile(fs, hooksJSONFilePath, data, 0644); err != nil {
		return hooksJSONFilePath, err
	}

	return hooksJSONFilePath, nil
}

// GetProjectConfigDirDotGitIgnoreFilePath returns the file path to the .gitignore file
// located in the project's config directory (e.g. project-name/.slack/.gitignore)
func GetProjectConfigDirDotGitIgnoreFilePath(projectDirPath string) string {
	return filepath.Join(GetProjectConfigDirPath(projectDirPath), ".gitignore")
}

// CreateProjectConfigDirDotGitIgnoreFile creates a new .gitignore file
// located in the project's config directory (e.g. project-name/.slack/.gitignore).
// Returns a os.ErrExists when the file already exists
func CreateProjectConfigDirDotGitIgnoreFile(fs afero.Fs, projectDirPath string) (path string, err error) {
	var gotGitignoreFilePath = GetProjectConfigDirDotGitIgnoreFilePath(projectDirPath)

	// Check if file already exists
	if _, err = fs.Stat(gotGitignoreFilePath); err == nil {
		return gotGitignoreFilePath, os.ErrExist
	}

	// Create the file
	if err = afero.WriteFile(fs, gotGitignoreFilePath, dotGitignoreFileData, 0644); err != nil {
		return gotGitignoreFilePath, err
	}

	return gotGitignoreFilePath, nil
}
