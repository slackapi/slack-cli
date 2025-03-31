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

package triggers

import (
	"github.com/slackapi/slack-cli/internal/shared/types"
)

var fakeAppID = "A1234"
var fakeAppTeamID = "T1234"
var fakeAppEnterpriseID = "E1234"
var fakeAppUserID = "U1234"
var fakeTriggerID = "Ft123"
var fakeTriggerName = "My Trigger"

var fakeApp = types.App{
	TeamDomain: "test",
	AppID:      fakeAppID,
	TeamID:     fakeAppTeamID,
	UserID:     fakeAppUserID,
}

func createFakeTrigger(fakeTriggerID string, fakeTriggerName string, fakeAppID string, triggerType string) types.DeployedTrigger {
	var fakeTrigger types.DeployedTrigger
	switch triggerType {
	case "event":
		fakeTrigger = types.DeployedTrigger{
			ID:   fakeTriggerID,
			Type: "event",
		}
	case "shortcut":
		fakeTrigger = types.DeployedTrigger{
			ID:          fakeTriggerID,
			Type:        "shortcut",
			Name:        fakeTriggerName,
			ShortcutUrl: "https://app.slack.com/app/" + fakeAppID + "/shortcut/" + fakeTriggerID,
		}
	case "scheduled":
		fakeTrigger = types.DeployedTrigger{
			ID:   fakeTriggerID,
			Type: "scheduled",
		}
	}
	fakeTrigger.Workflow.AppID = fakeAppID
	return fakeTrigger
}
