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

package iostreams

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
)

// Constants
const dateFormat = "20060102"

// Create new log file with suffix as current date
var currentTime = time.Now().UTC()
var filename = "slack-debug-" + currentTime.Format(dateFormat) + ".log"
var permissions = os.O_APPEND | os.O_CREATE | os.O_WRONLY

// InitLogFile will start the debug info to .slack/logs/slack-debug-[date].log with debug-ability data
func (io *IOStreams) InitLogFile(ctx context.Context) error {
	logFolder, _ := io.config.SystemConfig.LogsDir(ctx)

	// Before write to the log, we will delete any logs older than a week
	defer func() {
		// fail silently
		_ = deleteLogFilesOlderThanOneWeek(ctx, logFolder)
	}()

	// Ensure the error log file exists
	var errorLogFilePath string = filepath.Join(logFolder, filename)

	// TODO: this os reference should probably be something that is dependency injected for easier testing
	errorLogFile, err := os.OpenFile(errorLogFilePath, permissions, 0600)
	if err != nil {
		return err
	}

	logger := log.New(errorLogFile, "INFO ", log.Ldate|log.Ltime|log.Lmicroseconds)

	// Added line break for each session
	logger.Println("------------------------------------")

	// Log the Slack-CLI version, User's OS, SessionID, TraceID
	// But format data before writing them to the log file
	formatAndWriteDataToLogFile(logger, map[string]string{
		"Command":               goutils.RedactPII(strings.Join(os.Args[0:], " ")),
		"SessionID":             config.GetContextSessionID(ctx),
		"Slack-CLI-TraceID":     config.GetContextTraceID(ctx),
		"Slack-CLI Version":     io.config.Version,
		"Operating System (OS)": runtime.GOOS,
		"System ID":             io.config.SystemID,
		"Project ID":            io.config.ProjectID,
	})

	defer errorLogFile.Close()
	return nil
}

// FlushToLogFile will append the error to .slack/logs/slack-debug-[date].log
func (io *IOStreams) FlushToLogFile(ctx context.Context, prefix, errStr string) error {

	// TODO: we should switch this to the project directory later
	logFolder, err := io.config.SystemConfig.LogsDir(ctx)
	if err != nil {
		return err
	}

	// Ensure the error log file exists
	var errorLogFilePath string = filepath.Join(logFolder, filename)

	errorLogFile, err := io.fs.OpenFile(errorLogFilePath, permissions, 0600)
	if err != nil {
		log.Println(err)
		return err
	}
	defer errorLogFile.Close()

	logger := log.New(errorLogFile, fmt.Sprintf("%s ", strings.ToUpper(prefix)), log.Ldate|log.Ltime|log.Lmicroseconds)

	logger.Println(style.RemoveANSI(goutils.RedactPII(errStr)))

	if prefix != "debug" {
		// if its not a debug log, we should inform the user that
		// we logged an error in the errorLogFilePath
		io.Stderr.Println(style.Secondary(fmt.Sprintf("Check %s for error logs", errorLogFilePath)))
	}

	return nil
}

// FinishLogFile will end the debug info to .slack/logs/slack-debug-[date].log with debug-ability data
func (io *IOStreams) FinishLogFile(ctx context.Context) {
	logFolder, _ := io.config.SystemConfig.LogsDir(ctx)

	// Ensure the error log file exists
	var errorLogFilePath string = filepath.Join(logFolder, filename)

	// TODO: this os reference should probably be something that is dependency injected for easier testing
	errorLogFile, err := os.OpenFile(errorLogFilePath, permissions, 0600)
	if err != nil {
		log.Println(err)
	}

	logger := log.New(errorLogFile, "INFO ", log.Ldate|log.Ltime|log.Lmicroseconds)

	// format the TeamID, UserID and write to them to the log file
	formatAndWriteDataToLogFile(logger, map[string]string{
		"TeamID": config.GetContextTeamID(ctx),
		"UserID": config.GetContextUserID(ctx),
	})

	defer errorLogFile.Close()
}

// isOlderThanOneWeek returns bool if passed timestamp is older than 1 week
func isOlderThanOneWeek(t time.Time) bool {
	return time.Since(t) > 7*24*time.Hour
}

// deleteLogFilesOlderThanOneWeek deletes files with prefix "slack-debug-" that are older than 1 week in a directory
func deleteLogFilesOlderThanOneWeek(ctx context.Context, dir string) error {
	// TODO: this os reference should probably be something that is dependency injected for easier testing
	tmpfiles, err := os.ReadDir(dir)
	if err != nil {
		return slackerror.Wrap(err, "failed to return a list of log files in slack log directory")
	}
	// TODO - List files that cannot be accessed and deleted @cchensh
	for _, file := range tmpfiles {
		path := filepath.Join(dir, file.Name())
		// TODO: this os reference should probably be something that is dependency injected for easier testing
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.Mode().IsRegular() {
			// Only delete log files older than 1 week and with "slack-debug-" prefix
			if isOlderThanOneWeek(info.ModTime()) && strings.HasPrefix(file.Name(), "slack-debug-") {
				path := filepath.Join(dir, file.Name())
				// TODO: this os reference should probably be something that is dependency injected for easier testing
				if err := os.Remove(path); err != nil {
					continue
				}
			}
		}
	}
	return nil
}

// formatAndWriteDataToLogFile will format data in map and flush them to the log file
func formatAndWriteDataToLogFile(logger *log.Logger, data map[string]string) {
	if logger == nil {
		return
	}

	for id, idValue := range data {
		valueToLog := goutils.AddLogWhenValExist(id, idValue)
		if len(valueToLog) > 0 {
			logger.Printf("%s", valueToLog)
		}
	}
}
