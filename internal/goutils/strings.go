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

package goutils

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
)

// HashString will compute a short sha1 identity for a text blob
func HashString(s string) (string, error) {
	var hash = sha1.New()
	_, err := hash.Write([]byte(s))
	if err != nil {
		return "", err
	}
	hashSum := hash.Sum(nil)
	return fmt.Sprintf("%x", hashSum), nil
}

// ExtractFirstJSONFromString returns the first bracket-balanced json-looking substring in a string
// It returns the empty string if the string has no json-looking substring or does not have a
// json-looking substring with balanced brackets
//
//		Sometimes, when we run hooks, they return the expected json object but with some other
//		warning messages that we don't need.  This causes errors when unmarshalling.  This function
//	 helps us extract exactly what we need and ignore the rest.
func ExtractFirstJSONFromString(s string) string {
	s = strings.TrimSpace(s) // remove any white spaces
	var start = strings.Index(s, "{")
	if start == -1 {
		return "" // there is no json-looking substring so return immediately
	}

	var end = -1 // tracks the matching closing-bracket `}` for the opening-bracket `{`

	var stack = []rune{}
	for i := start; i < len(s); i++ {
		if s[i] == '{' {
			stack = append(stack, rune(s[i]))
		} else if s[i] == '}' {
			stack = stack[:len(stack)-1] // pop last item
			end = i                      // we found a potential ending bracket
		}

		if len(stack) == 0 {
			// we know we have found a matching closing bracket for the opening one
			break
		}
	}

	if end == -1 || len(stack) > 0 {
		// either we could not find a matching closing bracket `}` for the opening `{`
		// OR there the brackets found were not balanced
		return ""
	}

	return s[start : end+1]
}

// RedactPII will replace all occurrences in a string that match any of the expressions in a regex list
func RedactPII(s string) string {
	var regexListNoQuotes = []*regexp.Regexp{
		// Escape token values as "token=xoxp-123"
		regexp.MustCompile(`(?P<keys>(?:\w)*token(?:\=s*))(?P<values>(([\w\s.-]*)))`),
		// Escape oauth_authorize_url for apps apis
		regexp.MustCompile(`(?P<keys>(?:\w)*oauth_authorize_url(?:\=s*))(?P<values>(([\w\s.-]*)))`),
		// Escape provider_key for 3p auth
		regexp.MustCompile(`(?P<keys>(?:\w)*provider_key(?:\=s*))(?P<values>(([\w\s.-]*)))`),
		// Escape authorizations for 3p auth
		regexp.MustCompile(`(?P<keys>(?:\w)*authorizations(?:\=s*))(?P<values>(([\w\s.-]*)))`),
		// Escape authorization_url for 3p auth
		regexp.MustCompile(`(?P<keys>(?:\w)*authorization_url(?:\=s*))(?P<values>(([\w\s.-]*)))`),
		// Escape secret for 3p auth
		regexp.MustCompile(`(?P<keys>(?:\w)*secret(?:\=s*))(?P<values>(([\w\s.-]*)))`),
		// Escape variables for 3p auth client_id
		regexp.MustCompile(`(?P<keys>(?:\w)*client_id(?:\=s*))(?P<values>(([\w\s.-]*)))`),
		// Escape variables for 3p auth add client secrets
		regexp.MustCompile(`(?P<keys>(?:\w)*secret(?:\ s*))(?P<values>(([\w\s.-]*)))`),
		// Escape variables for env and it's alias for add/remove command
		regexp.MustCompile(`(?P<keys>(?:\w)*(env|var|vars|variable|variables|auth) (add|remove)(?:\ s*))(?P<values>(([\w\s.-]*)))`),
		// Add more regex here
	}
	// regexListWithQuotes will find sensitive data within quotes and later we escape with "..."
	var regexListWithQuotes = []*regexp.Regexp{
		// Escape token values based on hash keys with keyword "token" from JSON string
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*token(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Escape user name as `"user":"cheng chen"`
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*user(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Escape oauth_authorize_url
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*oauth_authorize_url(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Escape provider_key
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*provider_key(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Escape authorizations
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*authorizations(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Escape authorization_url
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*authorization_url(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Escape secret
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*secret(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Escape variables
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*variables(?:\"|\')(?:\:\s*))(?:\[\{)?(?P<values>(\[.*?\]))(?:\}])?`),
		// Escape client_id
		regexp.MustCompile(`(?P<keys>(?:\"|\')(?:\w)*client_id(?:\"|\')(?:\:\s*))(?:\"|\')?(?P<values>([\w\s.-]*))(?:\"|\')?`),
		// Add more regex here
	}
	var regexListOfWords = []*regexp.Regexp{
		// Escape App Token (xapp)
		regexp.MustCompile(`(?P<words>((xapp-[\w.-]*)))`),
		// Escape Bot Token (xoxb)
		regexp.MustCompile(`(?P<words>((xoxb-[\w.-]*)))`),
		// Escape User Token (xoxp)
		regexp.MustCompile(`(?P<words>((xoxp-[\w.-]*)))`),
		// Escape Refresh Token (xoxe)
		regexp.MustCompile(`(?P<words>((xoxe-[\w.-]*)))`),
	}

	for _, re := range regexListNoQuotes {
		// Keep token name and replace value with ...
		s = re.ReplaceAllString(s, "$1...")
	}
	for _, re := range regexListWithQuotes {
		// Keep token name and replace value with "...""
		s = re.ReplaceAllString(s, "$1\"...\"")
	}
	for _, re := range regexListOfWords {
		// Replace matched words with "..."
		s = re.ReplaceAllString(s, "...")
	}
	home, _ := os.UserHomeDir()
	s = strings.ReplaceAll(s, home, "...")
	return s
}

// AddLogWhenValExist returns a formatted string if value exists
func AddLogWhenValExist(title string, val string) string {
	if len(strings.TrimSpace(val)) > 0 {
		return fmt.Sprintf("%s: [%s]\n", title, val)
	}
	return ""
}

// UpperCaseTrimAll returns a formatted named_entites for trigger ACLs
func UpperCaseTrimAll(namedEntities string) string {
	return strings.ReplaceAll(strings.ToUpper(namedEntities), " ", "")
}

// ToHTTPS returns url with https protocol
func ToHTTPS(urlAddr string) string {
	u, err := url.Parse(urlAddr)
	if err != nil {
		return urlAddr
	}
	u.Scheme = "https"
	return u.String()
}
