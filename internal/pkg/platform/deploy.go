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

package platform

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/cmd/triggers"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"

	"github.com/opentracing/opentracing-go"
)

// Deploy will package and upload an app to the Slack Platform
func Deploy(ctx context.Context, clients *shared.ClientFactory, showTriggers bool, log *logger.Logger, app types.App) (*logger.LogEvent, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.deploy")
	defer span.Finish()

	// Get auth token
	token := config.GetContextToken(ctx)
	if strings.TrimSpace(token) == "" {
		return nil, slackerror.New(slackerror.ErrAuthToken)
	}

	if app.IsNew() {
		err := slackerror.New("no app found to deploy")
		return nil, slackerror.Wrap(err, slackerror.ErrAppDeploy)
	}

	// Validate auth session
	authSession, err := clients.ApiInterface().ValidateSession(ctx, token)
	if err != nil {
		return nil, slackerror.Wrap(err, slackerror.ErrSlackAuth)
	}

	// Add enterprise_id returned from auth session to app if exists
	if authSession.EnterpriseID != nil {
		ctx = config.SetContextEnterpriseID(ctx, *authSession.EnterpriseID)
		app.EnterpriseID = *authSession.EnterpriseID
		clients.EventTracker.SetAuthEnterpriseID(*authSession.EnterpriseID)
		// TODO: should we also SetAppEnterpriseID for metrics tracking?
	}

	authString, _ := json.Marshal(authSession)
	log.Data["authSession"] = string(authString)
	if authSession.UserID != nil {
		ctx = config.SetContextUserID(ctx, *authSession.UserID)
		clients.EventTracker.SetAuthUserID(*authSession.UserID)
	}

	// Collect the installed app name for outputs
	//
	// This value is not used to update the manifest on app settings! Instead,
	// updating an app manifest should happen before an app is deployed.
	manifest, err := clients.AppClient().Manifest.GetManifestRemote(ctx, token, app.AppID)
	if err != nil {
		return nil, slackerror.Wrap(err, slackerror.ErrAppManifestAccess)
	}
	log.Data["appName"] = manifest.AppManifest.DisplayInformation.Name

	if showTriggers {
		// Generate an optional trigger when none exist
		_, err = triggers.TriggerGenerate(ctx, clients, app)
		if err != nil {
			return nil, slackerror.Wrap(err, slackerror.ErrAppDeploy)
		}
	}

	// deploy the app
	if _, err := deployApp(ctx, clients, log, app); err != nil {
		return nil, slackerror.Wrap(err, slackerror.ErrAppDeploy)
	}

	// Save app to apps.json
	if !clients.Config.SkipLocalFs() {
		if err := clients.AppClient().SaveDeployed(ctx, app); err != nil {
			return nil, slackerror.Wrap(err, slackerror.ErrAppDeploy)
		}
	}

	return log.SuccessEvent(), nil
}

func deployApp(ctx context.Context, clients *shared.ClientFactory, log *logger.Logger, app types.App) (types.App, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "deployApp")
	defer span.Finish()

	var token = config.GetContextToken(ctx)

	if app.AppID == "" {
		return app, fmt.Errorf("failed to deploy: %s", slackerror.New("app was not created"))
	}

	// TODO: use clients.os, ensure mock exists
	projDir, err := os.Getwd()
	if err != nil {
		return app, fmt.Errorf("error getting working directory: %s", err)
	}

	log.Data["appId"] = app.AppID
	log.Log("info", "on_app_deploy")

	if clients.Runtime == nil {
		return app, slackerror.New(slackerror.ErrRuntimeNotSupported).
			WithMessage("The project runtime is not supported by this CLI")
	}

	// create the zip archive
	var startPackage = time.Now()
	var result packageResult

	// TODO: Packaging the app can happen in parallel with getting the presigned S3 post params

	log.Log("info", "on_app_package")
	result, err = packageArchive(ctx, clients, projDir, app.AppID)
	if err != nil {
		return app, fmt.Errorf("error packaging project: %s", err)
	}
	var elapsedPackage = time.Since(startPackage)
	log.Data["packagedSize"] = fmt.Sprintf("%.3fMB", float64(result.Size)/1000000)
	log.Data["packagedTime"] = fmt.Sprintf("%.1fs", elapsedPackage.Seconds())

	defer os.Remove(result.Filename)
	log.Log("info", "on_app_package_completion")

	log.Log("info", "on_app_deploy_hosting")

	//upload zip to s3
	var startDeploy = time.Now()
	s3Params, err := clients.ApiInterface().GetPresignedS3PostParams(ctx, token, app.AppID)
	if err != nil {
		return app, slackerror.Wrapf(err, "failed generating s3 upload params %s", app.AppID)
	}

	fileName, err := clients.ApiInterface().UploadPackageToS3(ctx, clients.Fs, app.AppID, s3Params, result.Filename)
	if err != nil {
		return app, slackerror.Wrapf(err, "failed uploading the zip file to s3 %s", app.AppID)
	}

	// upload
	runtime := strings.ToLower(clients.Runtime.Name())
	err = clients.ApiInterface().UploadApp(ctx, token, runtime, app.AppID, fileName)
	if err != nil {
		return app, fmt.Errorf("error uploading app: %s", err)
	}
	var elapsedDeploy = time.Since(startDeploy)
	var deployTime = fmt.Sprintf("%.1fs", elapsedDeploy.Seconds())

	log.Data["deployTime"] = deployTime

	// Set the SLACK_API_URL environment variable for development workspaces
	//
	// Note: This errors silently to continue deployment without any problem
	var apiHost = clients.Config.ApiHostResolved
	if clients.AuthInterface().IsApiHostSlackDev(apiHost) {
		apiHostURL := fmt.Sprintf("%s/api/", apiHost)
		_ = clients.ApiInterface().AddVariable(ctx, token, app.AppID, "SLACK_API_URL", apiHostURL)
	}

	return app, nil
}

type packageResult struct {
	Filename string
	Size     int64
}

func packageArchive(ctx context.Context, clients *shared.ClientFactory, projectRootDir, appID string) (packageResult, error) {
	var (
		tmpDir string
		err    error
		span   opentracing.Span
	)

	span, ctx = opentracing.StartSpanFromContext(ctx, "packageArchive")
	defer span.Finish()

	clients.IO.PrintDebug(ctx, "packaging archive")

	// Make a copy of the project dir because deployment packaging may ignore or shrink some files
	tmpDir, err = os.MkdirTemp("", "slack-cli-package-*")
	if err != nil {
		return packageResult{}, err
	}
	defer func() {
		// TODO: use clients.os, ensure mock exists
		err = os.RemoveAll(tmpDir)
		if err != nil {
			clients.IO.PrintInfo(ctx, false, "Failed to remove temporary directory %s", err)
		}
	}()
	clients.IO.PrintDebug(ctx, "using %s as temp directory", tmpDir)

	// Prepare the app package based on the runtime
	var authTokens = clients.Config.DomainAuthTokens
	var preparePackageOpts = types.PreparePackageOpts{
		SrcDirPath: projectRootDir,
		DstDirPath: tmpDir,
		AuthTokens: authTokens,
	}
	if err := clients.Runtime.PreparePackage(clients.SDKConfig, clients.HookExecutor, preparePackageOpts); err != nil {
		return packageResult{}, slackerror.Wrap(err, "preparing the app package for deployment")
	}

	// Install the project's production dependencies with a timeout in case there are issues
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if _, err := clients.Runtime.InstallProjectDependencies(ctx, tmpDir, clients.HookExecutor, clients.IO, clients.Fs, clients.Os); err != nil {
		return packageResult{}, err
	}

	// Zip up the archive
	return bundleArchive(ctx, clients, tmpDir, appID)
}

// bundleArchive zips up the directory and provides the path to a file in a temp directory
func bundleArchive(ctx context.Context, clients *shared.ClientFactory, dir, appID string) (packageResult, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "bundleArchive")
	defer span.Finish()

	clients.IO.PrintDebug(ctx, "bundleArchive")
	var err error
	var packageFile *os.File

	zipPrefix := fmt.Sprintf("slack-cli-package.%s.*.zip", appID)
	// TODO: use clients.os, ensure mock exists
	packageFile, err = os.CreateTemp("", zipPrefix)
	if err != nil {
		return packageResult{}, err
	}
	clients.IO.PrintDebug(ctx, "writing contents to %s", packageFile.Name())
	var fileName = packageFile.Name()
	var zipWriter = zip.NewWriter(packageFile)
	defer packageFile.Close()
	defer zipWriter.Close()

	// ensure directory path is clean
	dir = filepath.Clean(dir)

	var walker = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		var file *os.File
		file, err = os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var strippedPath string
		strippedPath, err = filepath.Rel(dir, path) // strip absolute path
		if err != nil {
			return err
		}

		var f io.Writer
		f, err = zipWriter.Create(strippedPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	// walk the tree and build the zip
	// TODO: WalkDir introduced in go 1.16 which is apparently more efficient?
	err = filepath.Walk(dir, walker)
	if err != nil {
		os.Remove(fileName)
		return packageResult{}, err
	}

	// close up shop
	zipWriter.Close()
	packageFile.Close()

	// get the file size
	packageFileInfo, err := os.Stat(fileName)
	if err != nil {
		os.Remove(fileName)
		return packageResult{}, err
	}

	var result = packageResult{
		Filename: fileName,
		Size:     packageFileInfo.Size(),
	}
	clients.IO.PrintDebug(ctx, "packaging complete")
	return result, nil
}
