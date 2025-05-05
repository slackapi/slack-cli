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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/logger"
	"github.com/slackapi/slack-cli/internal/shared"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	ActivityIdleTimeoutDefault     = 5 // Minutes
	ActivityLimitDefault           = 100
	ActivityMinLevelDefault        = "info"
	ActivityPollingIntervalDefault = 3 // Seconds
	logTemplate                    = "%s [%s] [%s] (Trace=%s) %s"
)

// actor verb object

// Activity Reads activity logs in a ticker loop, sleeping in intervals as defined by the args. Returns if the context is canceled or an error occurs. Expected to be used in a goroutine.
func Activity(
	ctx context.Context,
	clients *shared.ClientFactory,
	log *logger.Logger,
	args types.ActivityArgs,
) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "cmd.activity")
	defer span.Finish()

	// Get Auth Token
	var token = config.GetContextToken(ctx)
	if strings.TrimSpace(token) == "" {
		return slackerror.New(slackerror.ErrAuthToken)
	}

	if args.Browser {
		clients.Browser().OpenURL(fmt.Sprintf("https://app.%s/app-settings/%s/%s/app-logs", strings.Split(clients.Config.APIHostResolved, "//")[1], args.TeamID, args.AppID))
		return nil
	}

	authSession, err := clients.API().ValidateSession(ctx, token)
	if err != nil {
		return err
	}
	if authSession.UserName != nil && authSession.TeamName != nil {
		clients.IO.PrintInfo(ctx, false, "%s%s", style.Emoji("sparkles"), style.Secondary(fmt.Sprintf("%s of %s", *authSession.UserName, *authSession.TeamName)))
	}

	ctx = config.SetContextToken(ctx, token)

	activityRequest := types.ActivityRequest{
		AppID:              args.AppID,
		Limit:              args.Limit,
		MinimumDateCreated: args.MinDateCreated,
		MaximumDateCreated: args.MaxDateCreated,
		MinimumLogLevel:    args.MinLevel,
		EventType:          args.EventType,
		ComponentType:      args.ComponentType,
		ComponentID:        args.ComponentID,
		Source:             args.Source,
		TraceID:            args.TraceID,
	}

	latestCreatedTimestamp, _, err := printLatestActivity(ctx, clients, token, activityRequest, token)
	if err != nil {
		return err
	}

	if !args.TailArg {
		return nil
	}

	var lastNonZeroResultsTimestamp = time.Now()
	duration := time.Millisecond * time.Duration(args.PollingIntervalMS)
	clients.IO.PrintDebug(ctx, "Setting up a ticker to poll activity on a %s interval", duration)
	ticker := time.NewTicker(duration)
	for {
		select {
		case <-ctx.Done():
			clients.IO.PrintDebug(ctx, "Activity watcher context canceled, returning.")
			return nil
		case <-ticker.C:
			// Try to grab new logs using the last logs timestamp
			activityRequest.MinimumDateCreated = latestCreatedTimestamp + 1
			newLatestCreatedTimestamp, count, err := printLatestActivity(ctx, clients, token, activityRequest, token)
			if err != nil {
				return slackerror.New(slackerror.ErrStreamingActivityLogs).WithRootCause(err)
			}

			if count > 0 {
				lastNonZeroResultsTimestamp = time.Now()
			}

			// If we received new logs that are more recent, we update our timestamp for the next request's filtering
			if newLatestCreatedTimestamp > latestCreatedTimestamp {
				latestCreatedTimestamp = newLatestCreatedTimestamp
			}

			// If we've haven't received anything for the length of our timeout duration, we bail out.
			if time.Since(lastNonZeroResultsTimestamp) > time.Duration(args.IdleTimeoutM)*time.Minute {
				clients.IO.PrintInfo(ctx, true, "%s%s", style.Emoji("wave"), style.Secondary("Closing due to inactivity. Au revoir!"))
				return nil
			}
		}
	}
}

func printLatestActivity(ctx context.Context, clients *shared.ClientFactory, token string, args types.ActivityRequest, xoxpToken string) (latestCreated int64, num int, e error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "getLatestActivity")
	defer span.Finish()

	var result, err = clients.API().Activity(ctx, xoxpToken, args)
	if err != nil {
		return 0, 0, err
	}

	var activities = result.Activities
	latestTimestamp := args.MinimumDateCreated

	// process in reverse
	for i := len(activities) - 1; i >= 0; i-- {
		var activity = activities[i]

		if activity.Created > latestTimestamp {
			latestTimestamp = activity.Created
		}

		clients.IO.PrintInfo(ctx, false, prettifyActivity(activity))
	}

	return latestTimestamp, len(result.Activities), nil
}

func prettifyActivity(activity api.Activity) (log string) {
	msg := ""

	switch activity.EventType {
	case types.DatastoreRequestResult:
		msg = datastoreRequestResultToString(activity)
	case types.ExternalAuthMissingFunction:
		msg = externalAuthMissingFunctionToString(activity)
	case types.ExternalAuthMissingSelectedAuth:
		msg = externalAuthMissingSelectedAuthToString(activity)
	case types.ExternalAuthResult:
		msg = externalAuthResultToString(activity)
	case types.ExternalAuthStarted:
		msg = externalAuthStartedToString(activity)
	case types.ExternalAuthTokenFetchResult:
		msg = externalAuthTokenFetchResult(activity)
	case types.FunctionDeployment:
		msg = functionDeploymentToString(activity)
	case types.FunctionExecutionOutput:
		msg = functionExecutionOutputToString(activity)
	case types.TriggerPayloadReceived:
		msg = triggerPayloadReceivedOutputToString(activity)
	case types.FunctionExecutionResult:
		msg = functionExecutionResultToString(activity)
	case types.FunctionExecutionStarted:
		msg = functionExecutionStartedToString(activity)
	case types.TriggerExecuted:
		msg = triggerExecutedToString(activity)
	case types.WorkflowBillingResult:
		msg = workflowBillingResultToString(activity)
	case types.WorkflowBotInvited:
		msg = workflowBotInvitedToString(activity)
	case types.WorkflowCreatedFromTemplate:
		msg = workflowCreatedFromTemplateToString(activity)
	case types.WorkflowExecutionResult:
		msg = workflowExecutionResultToString(activity)
	case types.WorkflowExecutionStarted:
		msg = workflowExecutionStartedToString(activity)
	case types.WorkflowPublished:
		msg = workflowPublishedToString(activity)
	case types.WorkflowStepExecutionResult:
		msg = workflowStepExecutionResultToString(activity)
	case types.WorkflowStepStarted:
		msg = workflowStepStartedToString(activity)
	case types.WorkflowUnpublished:
		msg = workflowUnpublishedToString(activity)
	default:
		payload := []byte("")
		if activity.Payload != nil {
			var err error
			if payload, err = json.Marshal(activity.Payload); err != nil {
				payload = []byte("")
			}
		}
		msg = fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, payload)
	}

	switch activity.Level {
	case types.WARN:
		return style.Styler().Yellow(msg).String()
	case types.ERROR, types.FATAL:
		return style.Styler().Red(msg).String()
	}

	return msg
}

func datastoreRequestResultToString(activity api.Activity) (result string) {
	var (
		reqType       string
		datastoreName string
		details       string
		msg           string
	)
	if t, ok := activity.Payload["request_type"]; ok {
		reqType = fmt.Sprintf("%v", t)
	}
	if n, ok := activity.Payload["datastore_name"]; ok {
		datastoreName = fmt.Sprintf("Datastore:%v", n)
	}
	if d, ok := activity.Payload["details"]; ok {
		details = fmt.Sprintf("%v", d)
	}
	if activity.Level == types.ERROR || activity.Level == types.FATAL {
		var apiErr string
		if e, ok := activity.Payload["error"]; ok {
			apiErr = fmt.Sprintf("%v", e)
		}
		msg = fmt.Sprintf("Datastore %s failed with '%s'\n\t%s", reqType, details, apiErr)
	} else {
		msg = fmt.Sprintf("Datastore %s succeeded with '%s'", reqType, details)
	}

	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, datastoreName, activity.TraceID, msg)
}

func externalAuthMissingFunctionToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Step function '%s' is missing", activity.Payload["function_id"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func externalAuthMissingSelectedAuthToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Missing mapped token for workflow '%s'", activity.Payload["code"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func externalAuthResultToString(activity api.Activity) (result string) {
	s := "completed"
	switch activity.Level {
	case types.ERROR:
		s = "failed"
	}

	// If there's an error, we append it with a tab to make it stand out and readable
	msg := fmt.Sprintf("Auth %s for user '%s' on team '%s' for app '%s' and provider '%s'", s, activity.Payload["user_id"], activity.Payload["team_id"], activity.Payload["app_id"], activity.Payload["provider_key"])
	switch activity.Level {
	case types.ERROR:
		msg = msg + "\n\t" + strings.ReplaceAll(activity.Payload["code"].(string), "\n", "\n\t")
		msg = msg + "\n\t\t" + strings.ReplaceAll(activity.Payload["extra_message"].(string), "\n", "\n\t\t")
	}

	return style.Styler().Gray(13, msg).String()
}

func externalAuthStartedToString(activity api.Activity) (result string) {
	s := "succeeded"
	switch activity.Level {
	case types.ERROR:
		s = "failed"
	}

	msg := fmt.Sprintf("Auth start %s by user '%s' on team '%s' for app '%s' and provider '%s'", s, activity.Payload["user_id"], activity.Payload["team_id"], activity.Payload["app_id"], activity.Payload["provider_key"])
	switch activity.Level {

	// If there's an error, we append it with a tab to make it stand out and readable
	case types.ERROR:
		msg = msg + "\n\t" + strings.ReplaceAll(activity.Payload["code"].(string), "\n", "\n\t")
	}

	return style.Styler().Gray(13, msg).String()
}

func externalAuthTokenFetchResult(activity api.Activity) (result string) {
	s := "succeeded"
	switch activity.Level {
	case types.ERROR:
		s = "failed"
	}

	msg := fmt.Sprintf("Token fetch %s for user '%s' on team '%s' for app '%s' and provider '%s'", s, activity.Payload["user_id"], activity.Payload["team_id"], activity.Payload["app_id"], activity.Payload["provider_key"])
	switch activity.Level {

	// If there's an error, we append it with a tab to make it stand out and readable
	case types.ERROR:
		msg = msg + "\n\t" + strings.ReplaceAll(activity.Payload["code"].(string), "\n", "\n\t")
	}

	return style.Styler().Gray(13, msg).String()
}

func functionDeploymentToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Application %sd by user '%s' on team '%s'", activity.Payload["action"], activity.Payload["user_id"], activity.Payload["team_id"])
	msg = fmt.Sprintf("%s %s [%s] %s", style.Emoji("cloud"), activity.CreatedPretty(), activity.Level, msg)
	return style.Styler().Gray(13, msg).String()
}

func functionExecutionOutputToString(activity api.Activity) (result string) {
	log := activity.Payload["log"]
	//strings.ReplaceAll needs string input only, it doesn't accept json encoded interfaces.
	//While Sprintf can take the interfaces like json encoded data and outputs as string.
	msg := strings.ReplaceAll(fmt.Sprintf("Function output:\n%s", log), "\n", "\n\t")
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func triggerPayloadReceivedOutputToString(activity api.Activity) (result string) {
	log := activity.Payload["log"]
	//strings.ReplaceAll needs string input only, it doesn't accept json encoded interfaces.
	//While Sprintf can take the interfaces like json encoded data and outputs as string.
	msg := strings.ReplaceAll(fmt.Sprintf("Trigger payload:\n%s", log), "\n", "\n\t")
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func functionExecutionResultToString(activity api.Activity) (result string) {
	s := "completed"
	switch activity.Level {
	case types.ERROR, types.FATAL:
		s = "failed"
	}

	msg := fmt.Sprintf("Function '%s' (%s function) %s", activity.Payload["function_name"], activity.Payload["function_type"], s)

	// If there's an error, we append it with a tab to make it stand out and readable
	if err, ok := activity.Payload["error"]; ok {
		msg = msg + "\n\t" + strings.ReplaceAll(err.(string), "\n", "\n\t")
	}

	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func functionExecutionStartedToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Function '%s' (%s function) started", activity.Payload["function_name"], activity.Payload["function_type"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func triggerExecutedToString(activity api.Activity) (result string) {
	msgs := []string{}

	// Additional consideration for errors
	if activity.Level == types.INFO {
		trigger, ok := activity.Payload["trigger"].(map[string]interface{})
		if ok {
			caser := cases.Title(language.English)
			msgs = append(msgs, fmt.Sprintf("%s trigger successfully started execution of function '%s'", caser.String(trigger["type"].(string)), activity.Payload["function_name"]))
		} else {
			msgs = append(msgs, fmt.Sprintf("Trigger successfully started execution of function '%s'", activity.Payload["function_name"]))
		}
	} else if activity.Level == types.ERROR {
		msgs = append(msgs, fmt.Sprintf("Trigger for workflow '%s' failed: %s", activity.Payload["function_name"], activity.Payload["reason"]))
		// Default format for the errors message is the raw, json-encoded string
		var payloadErrors = []string{fmt.Sprintf("• %s", activity.Payload["errors"])}

		// Try to format the errors as a bullet list by parsing the 'errors' as a JSON array of strings
		if payloadErrorsAsString, ok := activity.Payload["errors"].(string); ok {
			if err := json.Unmarshal([]byte(payloadErrorsAsString), &payloadErrors); err == nil {
				for _, payloadError := range payloadErrors {
					msgs = append(msgs, fmt.Sprintf("  - %s", payloadError))
				}
			}
			// TODO: and if err is not nil?
		}
	}

	// Format the messages in the log template
	for i, msg := range msgs {
		msgs[i] = fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
	}

	return strings.Join(msgs, "\n")
}

func workflowBillingResultToString(activity api.Activity) (result string) {
	msg := ""
	if workflowName, ok := activity.Payload["workflow_name"]; ok {
		msg = fmt.Sprintf("Workflow '%s'", workflowName)
	} else {
		msg = "Workflow"
	}

	if activity.Payload["is_billing_result"] == true {
		msg = fmt.Sprintf("%s billing reason '%s'", msg, activity.Payload["billing_reason"])
	} else {
		msg = fmt.Sprintf("%s is excluded from billing", msg)
	}
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowBotInvitedToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Channel %s detected in workflow configuration. Bot user %s automatically invited.", activity.Payload["channel_id"], activity.Payload["bot_user_id"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowCreatedFromTemplateToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Workflow '%s' created from template '%s'", activity.Payload["workflow_name"], activity.Payload["template_id"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowExecutionResultToString(activity api.Activity) (result string) {
	s := "completed"
	switch activity.Level {
	case types.ERROR, types.FATAL:
		s = "failed"
	}

	msg := fmt.Sprintf("Workflow '%s' %s", activity.Payload["workflow_name"], s)

	// If there's an error, we append it with a tab to make it stand out and readable
	if err, ok := activity.Payload["error"]; ok {
		msg = msg + "\n\t" + strings.ReplaceAll(err.(string), "\n", "\n\t")
	}
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowExecutionStartedToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Workflow '%s' started", activity.Payload["workflow_name"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowPublishedToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Workflow '%s' published", activity.Payload["workflow_name"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowStepExecutionResultToString(activity api.Activity) (result string) {
	s := "completed"
	switch activity.Level {
	case types.ERROR, types.FATAL:
		s = "failed"
	}

	msg := fmt.Sprintf("Workflow step '%s' %s", activity.Payload["function_name"], s)
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowStepStartedToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Workflow step %.0f of %.0f started", activity.Payload["current_step"], activity.Payload["total_steps"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}

func workflowUnpublishedToString(activity api.Activity) (result string) {
	msg := fmt.Sprintf("Workflow '%s' unpublished", activity.Payload["workflow_name"])
	return fmt.Sprintf(logTemplate, activity.CreatedPretty(), activity.Level, activity.ComponentID, activity.TraceID, msg)
}
