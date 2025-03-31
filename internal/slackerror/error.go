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

package slackerror

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/slackapi/slack-cli/internal/style"
)

const MaxRecursiveErrDepth = 20

// errorMessageMap generates a map of all our errors listed in the errors.go file
func errorMessageMap() map[string]Error {
	var errorMessageMap = map[string]Error{}
	for _, e := range ErrorCodeMap {
		if e.Code != "" {
			errorMessageMap[strings.ToLower(e.Message)] = e
		}
	}
	return errorMessageMap
}

var ErrorMessageMap = errorMessageMap()

// A custom Error for the Slack CLI and API errors
type Error struct {
	Code        string       `json:"code,omitempty"`
	Message     string       `json:"error,omitempty"`
	Description string       `json:"slack_cli_error_description,omitempty"`
	Remediation string       `json:"remediation,omitempty"`
	Details     ErrorDetails `json:"errors,omitempty"`
	Cause       error        // TODO - Refactor 'Cause' to be 'Error' so that native Go error unwrapping works with a slackerror

	// useful for api errors
	ApiEndpoint string
}

type constraint struct {
	Type     string      `json:"type,omitempty"`
	Expected interface{} `json:"expected,omitempty"`
	Got      interface{} `json:"got,omitempty"`
}

type ErrorDetail struct {
	Code             string                 `json:"code,omitempty"`
	Constraint       constraint             `json:"constraint,omitempty"`
	Message          string                 `json:"message,omitempty"`
	Pointer          string                 `json:"pointer,omitempty"`
	Remediation      string                 `json:"remediation,omitempty"`
	RelatedComponent string                 `json:"related_component,omitempty"`
	Item             map[string]interface{} `json:"item,omitempty"`
}

type ErrorDetails []ErrorDetail

// Error stringifies the custom Error struct
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	var err = e.recursiveUnwrap()
	if err == nil {
		return ""
	}

	if err.Message == "" {
		var errSearch = getErrorIfKnown(err.Code)
		if errSearch != nil {
			err.Message = errSearch.Message
			err.Remediation = errSearch.Remediation
		}
	}

	var code, remediation string

	if err.Code != "" {
		code = fmt.Sprintf(" (%s)", err.Code)
	}

	if err.Remediation != "" {
		remediation = fmt.Sprintf("\n%s", style.Sectionf(style.TextSection{
			Emoji:     "bulb",
			Text:      "Suggestion",
			Secondary: []string{err.Remediation},
		}))
	}

	if err.Description != "" {
		err.Message = err.Description
	}

	var errDetails []string
	for ii, d := range err.Details {
		var code, pointer, item, remediation, padding string
		if d.Code != "" {
			code = fmt.Sprintf(" (%s)", d.Code)
		}
		if d.Pointer != "" {
			pointer = fmt.Sprintf("\nSource: %s", d.Pointer)
		}
		if d.Item != nil {
			var buff bytes.Buffer
			encoder := json.NewEncoder(&buff)
			encoder.SetEscapeHTML(false)
			_ = encoder.Encode(d.Item)
			item = fmt.Sprintf("\nItem: %s", buff.String())
		}
		if d.Remediation != "" {
			remediation = fmt.Sprintf("\nSuggestion: %s", d.Remediation)
		}
		if len(err.Details) > 1 && ii != len(err.Details)-1 {
			padding = "\n"
		}
		errDetails = append(errDetails, strings.TrimSpace(style.Sectionf(style.TextSection{
			Text: fmt.Sprintf(
				"%s%s%s%s%s",
				d.Message,
				code,
				pointer,
				item,
				remediation,
			),
		}))+padding)
	}

	var errStr, apiErr string

	if err.ApiEndpoint != "" {
		apiMethodText := style.Secondary("The following error was returned by the " + err.ApiEndpoint + " Slack API method")
		apiErr = fmt.Sprintf("%s\n\n%s", apiMethodText, style.Emoji("prohibited"))
	}

	errStr = strings.TrimRight(style.Sectionf(style.TextSection{
		Text:      fmt.Sprintf("%s%s", err.Message, style.Warning(code)),
		Secondary: errDetails,
	}), "\n")

	errStr = fmt.Sprintf("%s%s\n%s", apiErr, errStr, remediation)

	return errStr
}

// AddApiMethod will set the api endpoint for a slack error
// It is useful for rendering API errors
func (e *Error) AddApiMethod(endpoint string) *Error {
	if e == nil {
		return e
	}
	e.ApiEndpoint = endpoint
	return e
}

// AddDetail will add an error detail to a slack error
func (e *Error) AddDetail(detail ErrorDetail) *Error {
	if e == nil {
		return e
	}
	e.Details = append(e.Details, detail)
	return e
}

// AppendMessage will append to the error Message if one exists already
// If none exists, it will set message as the error Message
func (e *Error) AppendMessage(message string) *Error {
	if e == nil {
		return e
	} else if e.Message == "" {
		e.Message = message
	} else if message != "" {
		e.Message = fmt.Sprintf("%s\n%s", e.Message, message)
	}
	return e
}

// AppendRemediation will append to the Remediation if one exists already
// If none exists, it will set remediationMsg as the error's Remidiation
func (e *Error) AppendRemediation(remediationMsg string) *Error {
	if e == nil {
		return e
	} else if e.Remediation == "" {
		e.Remediation = remediationMsg

	} else if remediationMsg != "" {
		e.Remediation = fmt.Sprintf("%s\n%s", e.Remediation, remediationMsg)
	}
	return e
}

// recursiveUnwrap will find the inner most error in a wrapped error.
func (e *Error) recursiveUnwrap() *Error {
	if e == nil {
		return e
	}

	// we want to put a max recursive depth of 20 to avoid a stack overflow.
	return e.recursiveUnwrapWithLimit(MaxRecursiveErrDepth)
}

// RecursiveUnwrapWithLimit is a helper function to recursively unwrap an error
// and returns the innermost error or the error at the maxDepth specified.
func (e *Error) recursiveUnwrapWithLimit(maxDepth int) *Error {
	if e == nil {
		return e
	}

	// base case
	if e.Cause == nil || maxDepth == 0 {
		// if we get to an error with no cause then we have the innermost error so we should stop here.
		// OR if we have exausted the depth allowed then we should also stop here and return this error
		return e
	}

	err := ToSlackError(e.Cause)
	if err == nil {
		return nil
	}

	return err.recursiveUnwrapWithLimit(maxDepth - 1)
}

// Unwrap will return the imminent root cause of an error
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// WithCode will set the error code for a slack error
func (e *Error) WithCode(code string) *Error {
	if e == nil {
		return e
	}
	e.Code = code
	return e
}

// WithDetails will set the error details for a slack error
func (e *Error) WithDetails(details ErrorDetails) *Error {
	if e == nil {
		return e
	}
	e.Details = details
	return e
}

// WithMessage will set the error message for a slack error
func (e *Error) WithMessage(msg string, args ...interface{}) *Error {
	if e == nil {
		return e
	}
	e.Message = fmt.Sprintf(msg, args...)
	return e
}

// WithRemediation will set the remediation text for a slack error
func (e *Error) WithRemediation(remediationMsg string, args ...interface{}) *Error {
	if e == nil {
		return e
	}
	e.Remediation = fmt.Sprintf(remediationMsg, args...)
	return e
}

// WithRootCause will add err as part of the error details of this error.
func (e *Error) WithRootCause(err error) *Error {
	if e == nil || err == nil {
		return e
	}

	transformedErr := ToSlackError(err)
	e.Details = append(e.Details, ErrorDetail{
		Code:        transformedErr.Code,
		Message:     transformedErr.Message,
		Remediation: transformedErr.Remediation,
	})

	e.Details = append(e.Details, transformedErr.Details...)
	return e
}

// New returns an Error struct.  It first searches our known errors based on the error message to find it.
// If the error exists it returns the known error struct.  Otherwise it creates a new Error and returns that.
func New(msgOrCode string) *Error {
	err := getErrorIfKnown(msgOrCode)
	if err != nil {
		return err
	}

	return &Error{Message: msgOrCode}
}

// ToSlackError will cast an error to our custom Error type
func ToSlackError(err error) *Error {
	if err == nil {
		return nil
	}
	slackErr, ok := err.(*Error)
	if ok {
		return slackErr
	}

	slackErr = getErrorIfKnown(err.Error())
	if slackErr == nil {
		slackErr = &Error{
			Message: err.Error(),
		}
	}

	return slackErr
}

// Wrap returns an error and the supplied message.
// If err is nil, Wrap returns a new error.
// Wrap returns a new error with message and the cause as err.
func Wrap(cause error, message string) *Error {
	var wrappedErr *Error

	// search for the error message
	errSearch := getErrorIfKnown(message)
	if errSearch == nil {
		// we don't know about this error
		// so let's create a new one
		wrappedErr = &Error{
			Message: message,
		}
	} else {
		wrappedErr = errSearch
	}

	wrappedErr.Cause = cause
	return wrappedErr
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &Error{
		Cause:   err,
		Message: fmt.Sprintf(format, args...),
	}
}

// NewApiError returns an API error with details and error code
func NewApiError(errorCode string, description string, details ErrorDetails, apiEndpoint string) *Error {
	err := getErrorIfKnown(errorCode)
	if err == nil {
		if errorCode == "" {
			err = &Error{Code: "unknown_error"}
		} else {
			err = &Error{Code: errorCode}
		}
	}
	err.Description = description
	err.Details = append(err.Details, details...)
	err.ApiEndpoint = apiEndpoint
	return err
}

// getErrorIfKnown looks through the map of error codes and messages
// and returns an Error that matches the search term.
// if not found it will return nil
func getErrorIfKnown(searchTerm string) *Error {
	// turn the searchTerm to lowercase
	searchTerm = strings.ToLower(searchTerm)

	// let's first see if we have this error in our Error Map as an error message
	err, ok := ErrorMessageMap[searchTerm]
	if ok {
		return &err
	}

	// next let's see if we have this error in our Error Map as an error code
	err, ok = ErrorCodeMap[searchTerm]
	if ok {
		return &err
	}

	return nil
}

// IsErrorType returns true if the given error is of the type represented by the given error code
func IsErrorType(err error, code string) bool {
	errFromCode := ErrorCodeMap[code]
	return reflect.TypeOf(err) == reflect.TypeOf(&errFromCode)
}

// Is checks if errorCode is the code for an error, err
func Is(err error, errorCode string) bool {
	return strings.Contains(err.Error(), errorCode)
}

// JsonUnmarshalError returns a human readable json unmarshal error for CLI users
// Simply displaying the json.Unmarshal errors have proven to be very confusing already.
// This attempts to ensure the user understands that the CLI is trying to parse a JSON
// but is running into some issues.
func JsonUnmarshalError(err error, data []byte) *Error {
	if err == nil {
		return nil
	}

	contentToParse := style.Secondary(string(data[:]))
	jsonErr := New(ErrUnableToParseJson)
	jsonErr.Message = strings.Replace(jsonErr.Message, "<json>", contentToParse, 1)

	transformedErr := ToSlackError(err)
	return jsonErr.AddDetail(
		ErrorDetail{
			Code:        transformedErr.Code,
			Message:     transformedErr.Message,
			Remediation: transformedErr.Remediation,
		},
	)
}
