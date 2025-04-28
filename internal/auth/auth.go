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

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/slackapi/slack-cli/internal/api"
	"github.com/slackapi/slack-cli/internal/app"
	"github.com/slackapi/slack-cli/internal/config"
	"github.com/slackapi/slack-cli/internal/goutils"
	"github.com/slackapi/slack-cli/internal/iostreams"
	"github.com/slackapi/slack-cli/internal/shared/types"
	"github.com/slackapi/slack-cli/internal/slackerror"
	"github.com/slackapi/slack-cli/internal/style"
	"github.com/spf13/afero"
)

// Constants
const credentialsFileName = "credentials.json"
const defaultProdApiClientHost = "https://slack.com"
const defaultDevApiClientHost = "https://dev.slack.com"
const defaultHost = "slack.com"

var localBuildGitSHAInVersionRegex = regexp.MustCompile(`(?mi)-g[a-f0-9]{1,40}$`)

// Client can manage the state of the system's user/workspace authentications.
type Client struct {
	api       api.ApiInterface
	appClient *app.Client
	config    *config.Config
	io        iostreams.IOStreamer
	fs        afero.Fs
}

type AuthInterface interface {
	// AuthWithToken collects authentication information for the provided token
	AuthWithToken(ctx context.Context, token string) (types.SlackAuth, error)
	// AuthWithTeamDomain finds an auth with a given team domain
	//
	// FIXME: This is an unsafe method, since team domain does not guarantee unique auth if org and workspace are named the same
	AuthWithTeamDomain(ctx context.Context, teamDomain string) (types.SlackAuth, error)
	// AuthWithTeamID finds an auth with a given team ID
	AuthWithTeamID(ctx context.Context, teamID string) (types.SlackAuth, error)
	// Auths returns all the user's authorizatons as a slice
	Auths(ctx context.Context) ([]types.SlackAuth, error)

	// SetAuth will write new or overwrite existing auth record in credentials.json with provided auth
	SetAuth(ctx context.Context, auth types.SlackAuth) (types.SlackAuth, string, error)
	// SetSelectedAuth updates API configurations with relevant authentication details
	SetSelectedAuth(ctx context.Context, auth types.SlackAuth, config *config.Config, os types.Os)

	// DeleteAuth removes an auth from the list of auths in credentials.json
	DeleteAuth(context.Context, types.SlackAuth) (types.SlackAuth, error)
	// RevokeToken removes access for a given token, filtering known and safe errors
	RevokeToken(ctx context.Context, token string) error

	// ResolveApiHost returns the API Host based on the API Host Flag, Dev Flag, Project Config, and Stored Auth API Host.
	ResolveApiHost(ctx context.Context, apiHostFlag string, customAuth *types.SlackAuth) string
	// ResolveLogstashHost returns the error log stash host based on API Host and CLI version
	ResolveLogstashHost(ctx context.Context, apiHost string, cliVersion string) string

	// MapAuthTokensToDomains groups tokens by API host then delineates the host
	MapAuthTokensToDomains(ctx context.Context) string

	// IsApiHostSlackProd returns true if host is a development endpoint target
	IsApiHostSlackDev(host string) bool
	// IsApiHostSlackProd returns true if host is the production endpoint target
	IsApiHostSlackProd(host string) bool

	// FilterKnownAuthErrors catches known error codes that can be ignored to allow
	// the process to proceed safely without exiting.
	FilterKnownAuthErrors(ctx context.Context, err error) (bool, error)
}

// NewClient returns a new, empty instance of the Client
func NewClient(apiClient api.ApiInterface, appClient *app.Client, config *config.Config, io iostreams.IOStreamer, fs afero.Fs) *Client {
	var client = Client{
		api:       apiClient,
		appClient: appClient,
		config:    config,
		io:        io,
		fs:        fs,
	}

	return &client
}

// Getters
// AuthWithTeamDomain finds an auth with a given team domain
// FIXME: This is an unsafe method, since team domain does not guarantee unique auth if org and workspace are named the same
//
// TODO: Deprecate
func (c *Client) AuthWithTeamDomain(ctx context.Context, teamDomain string) (types.SlackAuth, error) {
	auths, err := c.Auths(ctx)
	if err != nil {
		return types.SlackAuth{}, err
	}

	for _, auth := range auths {
		if auth.TeamDomain == teamDomain {
			return auth, nil
		}
	}

	return types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound).
		WithMessage("No credentials found with the team domain \"%s\"", teamDomain)
}

// AuthWithTeamID finds an auth with a given team ID
func (c *Client) AuthWithTeamID(ctx context.Context, teamID string) (types.SlackAuth, error) {
	auths, err := c.Auths(ctx)
	if err != nil {
		return types.SlackAuth{}, err
	}

	for _, auth := range auths {
		if auth.TeamID == teamID {
			return auth, nil
		}
	}

	return types.SlackAuth{}, slackerror.New(slackerror.ErrCredentialsNotFound).
		WithMessage("No credentials found with the team ID \"%s\"", teamID)
}

// AuthWithToken collects authentication information for the provided token
func (c *Client) AuthWithToken(ctx context.Context, token string) (types.SlackAuth, error) {
	auth := types.SlackAuth{Token: token}

	session, err := c.api.ValidateSession(ctx, auth.Token)
	if err != nil {
		return types.SlackAuth{}, err
	}
	if session.URL != nil {
		if teamURL, err := url.Parse(*session.URL); err == nil {
			auth.TeamDomain = strings.Split(teamURL.Hostname(), ".")[0]
		}
	}
	if session.TeamID != nil {
		auth.TeamID = *session.TeamID
	}
	if session.UserID != nil {
		auth.UserID = *session.UserID
	}
	if session.EnterpriseID != nil {
		auth.EnterpriseID = *session.EnterpriseID
	}
	if session.IsEnterpriseInstall != nil {
		auth.IsEnterpriseInstall = *session.IsEnterpriseInstall
	}
	auth.LastUpdated = time.Now()

	return auth, nil
}

// Auths returns all the user's authorizatons as a slice
// Use this method outside this package to fetch a users cli authorizations
func (c *Client) Auths(ctx context.Context) ([]types.SlackAuth, error) {
	userAuthsMap, err := c.auths(ctx)
	if err != nil {
		return []types.SlackAuth{}, slackerror.New("Error reading credentials").WithRootCause(err)
	}
	userAuths := []types.SlackAuth{}
	for _, auth := range userAuthsMap {
		userAuths = append(userAuths, auth)
	}
	return userAuths, nil
}

// auths is an internal getter for a user's authorizations as a map of team_id to auth
func (c *Client) auths(ctx context.Context) (map[string]types.SlackAuth, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "auths")
	defer span.Finish()

	var auths types.AuthByTeamDomain

	c.io.PrintDebug(ctx, "reading credentials file")
	dir, err := c.config.SystemConfig.SlackConfigDir(ctx)
	if err != nil {
		return auths, err
	}
	var path string = filepath.Join(dir, credentialsFileName)

	if _, err := c.fs.Stat(path); os.IsNotExist(err) {
		return auths, err
	}

	c.io.PrintDebug(ctx, "found authorizations at %s reading", path)
	raw, err := afero.ReadFile(c.fs, path)
	if err != nil {
		return auths, err
	}

	err = json.Unmarshal(raw, &auths)
	if err != nil {
		return auths, slackerror.New(slackerror.ErrUnableToParseJSON).
			WithMessage("Failed to parse contents of credentials file").
			WithRootCause(err).
			WithRemediation("Check that %s is valid JSON", style.HomePath(path))
	}

	var updatedAuthsByName types.AuthByTeamDomain
	var updatedAuthsByTeamID types.AuthByTeamID
	var hasUpdateRotation, hasUpdateNaming bool

	updatedAuthsByName, hasUpdateRotation = c.rotateTokenAll(ctx, auths)

	// As of v2.4.0 we migrate users credentials json to storing auths by team_id
	updatedAuthsByTeamID, hasUpdateNaming = c.migrateToAuthByTeamID(ctx, updatedAuthsByName)

	if hasUpdateRotation || hasUpdateNaming {
		_, err = c.setAuths(ctx, types.AuthByTeamDomain(updatedAuthsByTeamID))
		if err != nil {
			return types.AuthByTeamID{}, err
		}
	}

	return updatedAuthsByTeamID, nil
}

// migrateToAuthByTeamID takes a map of auths keyed by team_domain or team_id and returns a map of auths
// guaranteed to be keyed by team id. Historically we have used non-unique team domain to store auths against.
// It was not ideal.
//
// This piece of migration logic ensures at read-time that credentials.json is updated to key auths against team_id.
// This change is effective as of slack-cli v2.4.0
//
// We can to remove this piece of logic in time after we are confident all users have their credentials.json updated
func (c *Client) migrateToAuthByTeamID(ctx context.Context, auths types.AuthByTeamDomain) (types.AuthByTeamID, bool) {
	var updatedAuths = types.AuthByTeamID{}
	var updated = false

	for authKey, auth := range auths {
		if !c.isTeamID(authKey) {
			// set the key in auths to use team_id
			updatedAuths[auth.TeamID] = auth
			updated = true
		} else {
			updatedAuths[authKey] = auth
		}
	}
	return updatedAuths, updated
}

func (c *Client) isTeamID(stringToCheck string) bool {
	teamIDSchemaV1 := "^[TE][A-Z0-9]{8,}$" // based on common_defs_schema.json
	teamIDSchemas := []string{teamIDSchemaV1}

	// stringToCheck must match at least one schema provided in teamIDSchemas
	for _, schema := range teamIDSchemas {
		matchTeamID := regexp.MustCompile(schema)
		if matchTeamID.Match([]byte(stringToCheck)) {
			return true
		}
	}
	return false
}

// rotateTokenAll attempts to rotate the token via token rotation of each
// user auth stored in credentials.json and returns the updated auths.
// If no changes are required it will return the same auths as provided unchanged
func (c *Client) rotateTokenAll(ctx context.Context, auths types.AuthByTeamDomain) (types.AuthByTeamDomain, bool) {
	//  track of all the auths including any updates done to their tokens
	var updatedAuths = types.AuthByTeamDomain{}
	var updated = false

	for authKey, auth := range auths {
		updatedAuths[authKey] = auth                                 // start with the un-updated auth.  We will override below if necessary
		updatedAuth, tokenIsUpdated, err := c.rotateToken(ctx, auth) // update the auth if possible
		if err != nil {
			// When the rotation does not succeed, we do not want to delete the entry.
			// We also do not want to stop the entire process: so we will not return here.
			// The user should go ahead with the bad token and the api will handle the
			// return of the appropriate error to the user.
			// We only want to warn the user about what we tried to do.
			c.io.PrintDebug(ctx, "Your auth token for '%s' is outdated. Tried refreshing the credentials but encountered the following error:\n%s", auth.TeamDomain, err.Error())
		} else if tokenIsUpdated {
			updated = true
			updatedAuths[authKey] = updatedAuth
		}
	}
	return updatedAuths, updated
}

// rotateToken takes an auth and returns an updated auth (via token rotation) where possible
// it returns the same auth if token rotation was not possible or the auth's token is still valid.
func (c *Client) rotateToken(ctx context.Context, auth types.SlackAuth) (types.SlackAuth, bool, error) {

	if !auth.ShouldRotateToken() {
		return auth, false /* tokenIsUpdated */, nil
	}

	// Store the current apiHost before rotation
	// We need this because we need to restore
	// the apiHost to what it was before rotating each of the user's auths
	activeApiHostBeforeRotation := c.api.Host()

	if auth.ApiHost != nil {
		c.api.SetHost(*auth.ApiHost)
	} else if auth.ApiHost == nil || c.api.Host() == "" {
		// always default to prod when we don't know what the api host is
		c.api.SetHost(defaultProdApiClientHost)
	}

	var result, err = c.api.RotateToken(ctx, auth)
	if err != nil {
		// handle token rotation failure by sending meaningful messages to the users and remove already expired auth
		return auth, false /* tokenIsUpdated */, err
	}

	auth.Token = result.Token
	auth.ExpiresAt = result.ExpiresAt
	auth.RefreshToken = result.RefreshToken
	auth.LastUpdated = time.Now()

	// now restore the previous default apiHost
	c.api.SetHost(activeApiHostBeforeRotation)

	return auth, true /* tokenIsUpdated */, nil
}

// setAuths sets the user's authorizations to the credentials file and returns the filepath of the credentials file in success cases
func (c *Client) setAuths(ctx context.Context, auths types.AuthByTeamDomain) (path string, err error) {
	var b []byte
	b, err = json.MarshalIndent(auths, "", "  ")

	if err != nil {
		return "", err
	}

	dir, err := c.config.SystemConfig.SlackConfigDir(ctx)
	if err != nil {
		return "", err
	}
	path = filepath.Join(dir, credentialsFileName)

	err = afero.WriteFile(c.fs, path, b, 0600)
	if err != nil {
		return path, err
	}

	return path, nil
}

// SetAuth will write new or overwrite existing auth record in credentials.json with provided auth
func (c *Client) SetAuth(ctx context.Context, auth types.SlackAuth) (types.SlackAuth, string, error) {
	var span opentracing.Span
	span, ctx = opentracing.StartSpanFromContext(ctx, "SetAuth")
	defer span.Finish()

	allAuths, err := c.auths(ctx)
	if err != nil {
		return types.SlackAuth{}, "", fmt.Errorf("Failed to fetch user authorizations %s", err)
	}
	allAuths[auth.TeamID] = auth

	// Now save the auth
	fileLocation, err := c.setAuths(ctx, allAuths)
	if err != nil {
		return types.SlackAuth{}, "", err
	}

	return auth, fileLocation, nil
}

// SetSelectedAuth updates API configurations with relevant authentication
// details
func (c *Client) SetSelectedAuth(ctx context.Context, auth types.SlackAuth, config *config.Config, os types.Os) {
	// Set the team flag to have confident checks if needed.
	config.TeamFlag = auth.TeamID

	// Update the API host to match the selected app or team authentication so API
	// calls are sent to the right host.
	//
	// Often set after standard selections but custom authentication must set this
	// unless the command is exiting right after, like 'login'.
	config.ApiHostResolved = c.ResolveApiHost(ctx, config.ApiHostFlag, &auth)
	config.LogstashHostResolved = c.ResolveLogstashHost(ctx, config.ApiHostResolved, config.Version)

	// Set environment variables for app development configurations and processes.
	if _, ok := os.LookupEnv("SLACK_API_URL"); !ok {
		_ = os.Setenv(
			"SLACK_API_URL",
			fmt.Sprintf("%s/api/", config.ApiHostResolved),
		)
	}
}

// DeleteAuth removes an auth from the list of auths in credentials.json
func (c *Client) DeleteAuth(ctx context.Context, auth types.SlackAuth) (types.SlackAuth, error) {
	// Ensure the auth exists
	toDelete, err := c.AuthWithTeamID(ctx, auth.TeamID)
	if err != nil {
		return toDelete, err
	}
	allAuths, err := c.auths(ctx)
	if err != nil {
		return toDelete, err
	}
	delete(allAuths, auth.TeamID)

	// Save to credentials.json
	_, err = c.setAuths(ctx, allAuths)
	if err != nil {
		return toDelete, err
	}
	return toDelete, nil
}

// IsApiHostSlackDev returns true if host is the Slack Dev endpoint (dev.slack.com, https://dev1234.api.slack.com, etc)
func (c *Client) IsApiHostSlackDev(host string) bool {
	if host == "" {
		return false
	}

	u, err := url.Parse(host)
	if err != nil {
		return false
	}

	subdomain := strings.Split(u.Hostname(), ".")[0]
	return strings.HasPrefix(subdomain, "dev")
}

// TODO (@kian) Make our Auths storage and checking convention consistent.Should this function also account for the nil case (ie. no host specified) being a prod auth?
// Either we should always save the api hostname even when it's a prod auth,
// or we should update this function to also return true if the api hostname
// is empty/undefined
// IsApiHostSlackProd returns true if host is the production endpoint target.
func (c *Client) IsApiHostSlackProd(host string) bool {
	return host == defaultProdApiClientHost
}

// ResolveApiHost returns the API Host based on the API Host Flag, Dev Flag, Project Config, and Stored Auth API Host.
func (c *Client) ResolveApiHost(ctx context.Context, apiHostFlag string, customAuth *types.SlackAuth) string {
	// TODO - Update this comment
	// Here is the order of relevance / importance:
	// 1. If the command is login
	//    apihost flag > dev flag > custom auth > default auth in credentials.json > prod
	// 2. Otherwise
	//    apihost flag > dev flag > custom auth >  prod

	// using Contains for isLoginCommand not a fool-proof function and cannot distinguish between
	// commands, flags, and args.
	var isLoginCommand = goutils.Contains(os.Args, "login", true)
	var apiHost string
	if apiHostFlag != "" {
		apiHost = goutils.ToHTTPS(apiHostFlag)
		c.config.SlackDevFlag = c.IsApiHostSlackDev(apiHostFlag)
	} else if c.config.SlackDevFlag {
		apiHost = defaultDevApiClientHost
	} else if customAuth != nil {
		// When a custom auth, we want to respect the APIHost
		// When not set, we default to prod
		if customAuth.ApiHost != nil {
			apiHost = goutils.ToHTTPS(*customAuth.ApiHost)
		} else {
			apiHost = defaultProdApiClientHost
		}
	} else if isLoginCommand {
		apiHost = defaultProdApiClientHost
	} else {
		apiHost = defaultProdApiClientHost
	}

	// when not on prod, warn the user to update SLACK_API_URL where possible
	if apiHost != defaultProdApiClientHost {
		// warn the user if the apihost is not the main slack production apihost to update SLACK_API_URL
		c.io.PrintDebug(
			ctx,
			"You're using a custom apihost. Run %s to add it to your app's Run on Slack environment",
			style.Commandf(fmt.Sprintf("var add SLACK_API_URL %s", apiHost), false),
		)
	}

	// warn the user if the apihost is not pointing to a dev instance
	// but they have the dev flag on
	if !c.IsApiHostSlackDev(apiHost) && c.config.SlackDevFlag {
		c.io.PrintWarning(ctx, "Warning: you are using the dev flag but you are signed into a production workspace or are using a custom apihost endpoint")
	}

	return apiHost
}

// TODO: how does this play together with ResolveApiHost above?
// ResolveLogstashHost returns the error log stash host based on API Host and CLI version
func (c *Client) ResolveLogstashHost(ctx context.Context, apiHost string, cliVersion string) string {
	c.io.PrintDebug(ctx, "Resolving logstash host, %s, %s", apiHost, cliVersion)
	if localBuildGitSHAInVersionRegex.Match([]byte(cliVersion)) {
		return "https://dev.slackb.com/events/cli"
	}
	if c.IsApiHostSlackProd(apiHost) {
		return "https://slackb.com/events/cli"
	}

	return "https://dev.slackb.com/events/cli"
}

// MapAuthTokensToDomains creates a delimited string of tokens for each api domain
// included in the available login contexts. If there are multiple login contexts
// for the the same api domain, the last one wins. It doesn't matter we just
// need one to be valid.
func (c *Client) MapAuthTokensToDomains(ctx context.Context) string {
	authTokenMap := map[string]string{}

	auths, err := c.Auths(ctx)
	if err != nil {
		return ""
	}

	for _, auth := range auths {
		if auth.ApiHost != nil {
			u, err := url.Parse(*auth.ApiHost)
			if err != nil {
				continue
			}
			authTokenMap[u.Host] = auth.Token
		} else {
			authTokenMap[defaultHost] = auth.Token
		}
	}
	authTokens := []string{}
	for domain, token := range authTokenMap {
		authTokens = append(authTokens, fmt.Sprintf("%s@%s", token, domain))
	}

	// Account for Deno bug that choose a root domain over more specific subdomain if encountered first
	// Sort so the longest domains are first
	sort.Slice(authTokens, func(i, j int) bool {
		return len(authTokens[i][strings.Index(authTokens[i], "@"):]) > len(authTokens[j][strings.Index(authTokens[j], "@"):])
	})

	return strings.Join(authTokens, ";")
}

// FilterKnownAuthErrors catches certain error codes that indicate an API error
// related to authentication happened, but the error might allow the process to
// proceed without exiting.
//
// If an error is caught with this filter, a "true" boolean is returned without
// the caught error. If no error is caught, this boolean will be "false" and the
// provided error is returned.
//
// Errors that are not caught should prevent certain commands from completing,
// such as deleting local authentication records with the "logout" command.
func (c *Client) FilterKnownAuthErrors(ctx context.Context, err error) (bool, error) {
	slackErr := slackerror.ToSlackError(err)
	if slackErr == nil {
		return false, err
	}
	switch slackErr.Code {
	case slackerror.ErrAlreadyLoggedOut:
		c.io.PrintDebug(ctx, "%s.", slackErr.Message)
		return true, nil
	case slackerror.ErrInvalidAuth:
		c.io.PrintDebug(ctx, "%s.", slackErr.Message)
		return true, nil
	case slackerror.ErrTokenExpired:
		c.io.PrintDebug(ctx, "%s.", slackErr.Message)
		return true, nil
	case slackerror.ErrTokenRevoked:
		c.io.PrintDebug(ctx, "%s.", slackErr.Message)
		return true, nil
	default:
		return false, err
	}
}
