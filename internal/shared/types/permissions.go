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

package types

import (
	"fmt"

	"github.com/spf13/pflag"
)

type AppCollaboratorPermission string

// App collaborator permission types
const (
	OWNER  AppCollaboratorPermission = "owner"
	READER AppCollaboratorPermission = "reader"
)

// StringToAppCollaboratorPermission returns the human readable word
// of the app collaborator permission
func StringToAppCollaboratorPermission(input string) (AppCollaboratorPermission, error) {
	switch input {
	case "owner":
		return OWNER, nil
	case "reader":
		return READER, nil
	default:
		return "", fmt.Errorf("invalid")
	}
}

// AppCollaboratorPermissionF formats the permission for use in a sentence
func (acp AppCollaboratorPermission) AppCollaboratorPermissionF() string {
	switch acp {
	case OWNER:
		return fmt.Sprintf("an %s collaborator", string(acp))
	case READER:
		return fmt.Sprintf("a %s collaborator", string(acp))
	default:
		return "a collaborator"
	}
}

type Permission string

// Consumed in function distribution ACLs and trigger run ACLs
// distribution type or access/permission type: 'everyone' | 'app_collaborators' | 'named_entities';
const (
	PermissionNamedEntities    Permission = "named_entities"
	PermissionAppCollaborators Permission = "app_collaborators"
	PermissionEveryone         Permission = "everyone"
)

// FunctionPermissions holds information for setting multiple function distributions
type FunctionPermissions struct {
	FunctionMap map[string]struct {
		Type    Permission `yaml:"type" json:"type"`
		UserIDs []string   `yaml:"user_ids,omitempty" json:"user_ids,omitempty"`
	} `yaml:"function_distribution" json:"function_distribution"`
}

func (d Permission) IsValid() bool {
	switch d {
	case PermissionAppCollaborators, PermissionNamedEntities, PermissionEveryone:
		return true
	}
	return false
}

func (d Permission) ToString() (userFriendlyString string) {
	switch d {
	case PermissionNamedEntities:
		userFriendlyString = "specific entities"
	case PermissionAppCollaborators:
		userFriendlyString = "app collaborators"
	case PermissionEveryone:
		userFriendlyString = "everyone"
	}
	return
}

// isNamedEntityFlag returns true if named entity flags are set
func IsNamedEntityFlag(flags *pflag.FlagSet) bool {
	return !flags.Lookup("everyone").Changed &&
		(flags.Lookup("channels").Changed || flags.Lookup("organizations").Changed ||
			flags.Lookup("users").Changed || flags.Lookup("workspaces").Changed)
}

// GetAccessTypeDescriptionForEveryone returns the user-friendly output
// for the "everyone" access type. It will clarify who everyone is given
// how the app has been installed to the team (workspace or org).
func GetAccessTypeDescriptionForEveryone(app App) string {
	if app.IsEnterpriseApp() {
		return "everyone in all workspaces in this org granted to this app"
	} else {
		return "everyone in the workspace"
	}
}
