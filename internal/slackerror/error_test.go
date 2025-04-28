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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_AddAPIMethod(t *testing.T) {
	tests := map[string]struct {
		err         Error
		newEndpoint string
	}{
		"previously empty endpoint": {
			err:         Error{APIEndpoint: ""},
			newEndpoint: "http://example2.com",
		},
		"previously non-empty endpoint": {
			err:         Error{APIEndpoint: "http://example1.com"},
			newEndpoint: "http://example2.com",
		},
		"empty endpoint being set": {
			err:         Error{APIEndpoint: "http://example1.com"},
			newEndpoint: "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_ = tt.err.AddAPIMethod(tt.newEndpoint)
			require.Equal(t, tt.newEndpoint, tt.err.APIEndpoint)
		})
	}
}

func Test_AddDetail(t *testing.T) {
	tests := map[string]struct {
		err    *Error
		detail ErrorDetail
	}{
		"previously empty details": {
			err:    New("new"),
			detail: ErrorDetail{Message: "test"},
		},
		"previously non-empty endpoint": {
			err:    &Error{Code: "some_code", Details: ErrorDetails{{Message: "detail1"}}},
			detail: ErrorDetail{Message: "detail"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			e := tt.err.AddDetail(tt.detail)
			require.Contains(t, e.Details, tt.detail)
		})
	}
}

func Test_AppendRemediation(t *testing.T) {
	tests := map[string]struct {
		initialMsg  string
		msgToAppend string
		expected    string
	}{
		"previously empty error remediation message": {
			initialMsg:  "",
			msgToAppend: "hello world",
			expected:    "hello world",
		},
		"previously non-empty remediation messages": {
			initialMsg:  "hello",
			msgToAppend: "world",
			expected:    "hello\nworld",
		},
		"appending an empty string": {
			initialMsg:  "hello world",
			msgToAppend: "",
			expected:    "hello world",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := New("").WithRemediation("%s", tt.initialMsg).AppendRemediation(tt.msgToAppend)
			require.Equal(t, tt.expected, err.Remediation)
		})
	}
}

func Test_JSONUnmarshalErrorTest(t *testing.T) {
	tests := map[string]struct {
		data     []byte
		err      error
		expected Error
	}{
		"err is nil": {
			data: []byte("{bad: json}"),
			err:  nil,
		},
		"err not nil": {
			data: []byte("{bad: json}"),
			err:  New("because i am hungry"),
		},
		"appending an empty string": {
			data: []byte("{bad: json}"),
			err:  New(ErrAccessDenied),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := JSONUnmarshalError(tt.err, tt.data)
			if tt.err == nil {
				require.Nil(t, err)
			} else {
				require.Equal(t, ErrUnableToParseJSON, err.Code)
				require.Contains(t, err.Error(), string(tt.data[:]))

				transformedErr := ToSlackError(tt.err)
				require.Contains(t, err.Error(), transformedErr.Code)
				require.Contains(t, err.Error(), transformedErr.Message)
				require.Contains(t, err.Error(), transformedErr.Remediation)
			}
		})
	}
}

func Test_New(t *testing.T) {
	tests := map[string]struct {
		msgOrCode string
		expected  Error
	}{
		"simple case -  non existent error": {
			msgOrCode: "fake",
			expected:  Error{Message: "fake"},
		},
		"known error - add error using error code": {
			msgOrCode: ErrAccessDenied,
			expected:  ErrorCodeMap[ErrAccessDenied],
		},
		"known error - add error using the error message": {
			msgOrCode: ErrorCodeMap[ErrAccessDenied].Message, // get the message associated with access denied error
			expected:  ErrorCodeMap[ErrAccessDenied],
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := New(tt.msgOrCode)
			require.Equal(t, tt.expected.Error(), err.Error())
		})
	}
}

func Test_NewAPIError(t *testing.T) {
	testErr := ErrorCodeMap[ErrInternal]

	tests := map[string]struct {
		code        string
		details     ErrorDetails
		apiMethod   string
		description string
		expected    Error
	}{
		"empty error code": {
			code:        "",
			apiMethod:   "",
			description: "",
			expected:    Error{Code: "unknown_error", Description: ""},
		},
		"no-empty but unknown error code": {
			code:        "fake",
			description: "fake",
			expected:    Error{Code: "fake", Description: "fake"},
		},
		"no-empty but unknown error code + api endpoint": {
			code:        "fake",
			description: "fake",
			apiMethod:   "fakeMethod",
			expected:    Error{Code: "fake", Description: "fake", APIEndpoint: "fakeMethod"},
		},
		"error code  + description + details + api endpoint": {
			code:        "fake",
			description: "fake",
			details:     ErrorDetails{{Code: "c1", Message: "m1"}},
			apiMethod:   "fakeMethod",
			expected:    Error{Code: "fake", Description: "fake", APIEndpoint: "fakeMethod", Details: ErrorDetails{{Code: "c1", Message: "m1"}}},
		},
		"known error + add details": {
			code:      ErrInternal,
			details:   ErrorDetails{{Code: "c1", Message: "m1"}},
			apiMethod: "fakeMethod",
			expected:  *testErr.AddAPIMethod("fakeMethod").WithDetails(ErrorDetails{{Code: "c1", Message: "m1"}}),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := NewAPIError(tt.code, tt.description, tt.details, tt.apiMethod)
			require.Equal(t, tt.expected.Code, err.Code)
			require.Equal(t, tt.expected.Description, err.Description)
			require.Equal(t, tt.expected.Message, err.Message)
			require.Equal(t, tt.expected.Details, err.Details)
			require.Equal(t, tt.expected.APIEndpoint, err.APIEndpoint)
			require.Equal(t, tt.expected.Error(), err.Error())
		})
	}
}

func Test_Error(t *testing.T) {
	tests := map[string]struct {
		err *Error
	}{
		"nil case": {
			err: nil,
		},
		"code only": {
			err: &Error{Code: "code"},
		},
		"message only": {
			err: &Error{Message: "msg"},
		},
		"code + message only": {
			err: &Error{Code: "code", Message: "msg"},
		},
		"code + message + remediation only": {
			err: &Error{Code: "code", Message: "msg", Remediation: "remediation"},
		},
		"code + message + remediation + api endpoint only": {
			err: &Error{Code: "code", Message: "msg", Remediation: "remediation", APIEndpoint: "host"},
		},
		"code + message + remediation + api endpoint + details": {
			err: &Error{Code: "code", Message: "msg", Remediation: "remediation", APIEndpoint: "host", Details: ErrorDetails{{Message: "m1", Code: "c1", Remediation: "r1", Pointer: "p1"}}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res := tt.err.Error()

			if tt.err == nil {
				require.Empty(t, res, "When slack error is nil, the string should be empty")
			} else {
				require.Contains(t, res, tt.err.Code)
				require.Contains(t, res, tt.err.Message)
				require.Contains(t, res, tt.err.APIEndpoint)
				require.Contains(t, res, tt.err.Remediation)
				for _, d := range tt.err.Details {
					require.Contains(t, res, d.Message)
					require.Contains(t, res, d.Code)
					require.Contains(t, res, d.Remediation)
					require.Contains(t, res, d.Pointer)
				}
			}
		})
	}
}

func Test_recursiveUnwrap(t *testing.T) {
	err1 := Error{Code: "error1"}
	err2 := Error{Code: "error2", Cause: &err1}
	err3 := Error{Code: "error3", Cause: &err2}

	tests := map[string]struct {
		err      Error
		expected error
	}{
		"simple case": {
			err:      err1,
			expected: &err1,
		},
		"single nested": {
			err:      err2,
			expected: &err1,
		},
		"multinested err": {
			err:      err3,
			expected: &err1,
		},
		"unknown cause": {
			err:      Error{Code: "error1", Cause: errors.New("unknown")},
			expected: &Error{Message: "unknown"},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.recursiveUnwrap())
		})
	}
}

func Test_ToSlackError(t *testing.T) {
	testErrCode := ErrAccessDenied
	testErr := ErrorCodeMap[testErrCode]
	testErrMsg := testErr.Message

	tests := map[string]struct {
		err      error
		expected Error
	}{
		"an error code that exists": {
			err:      errors.New(testErrCode),
			expected: testErr,
		},
		"an error message that we know about": {
			err:      errors.New(testErrMsg),
			expected: testErr,
		},
		"an error code we don't know about": {
			err:      errors.New("error-message-from-out-of-nowhere"),
			expected: Error{Message: "error-message-from-out-of-nowhere"},
		},
		"empty string error": {
			err:      errors.New(""),
			expected: Error{Message: ""},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := ToSlackError(tt.err)
			require.Equal(t, tt.expected, *err)
		})
	}
}

func Test_Unwrap(t *testing.T) {
	err1 := Error{Code: "error1"}
	err2 := Error{Code: "error2", Cause: &err1}
	err3 := Error{Code: "error2", Cause: &err2}

	tests := map[string]struct {
		err      Error
		expected error
	}{
		"simple case": {
			err:      err2,
			expected: &err1,
		},
		"no nested err": {
			err:      err1,
			expected: nil,
		},
		"same nested err": {
			err:      err3,
			expected: &err2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.err.Unwrap())
		})
	}
}

func Test_WithCode(t *testing.T) {
	tests := map[string]struct {
		oldCode string
		newCode string
	}{
		"previously empty error code": {
			oldCode: "",
			newCode: "newCode1",
		},
		"previously non-empty codes": {
			oldCode: "oldCode1",
			newCode: "newCode1",
		},
		"empty error code being set": {
			oldCode: "oldCode1",
			newCode: "",
		},
		"non-changing error code": {
			oldCode: "code1",
			newCode: "code1",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := New(tt.oldCode).WithCode(tt.newCode)
			require.Equal(t, tt.newCode, err.Code)
		})
	}
}

func Test_WithMessage(t *testing.T) {
	tests := map[string]struct {
		oldMsg string
		newMsg string
	}{
		"previously empty error message": {
			oldMsg: "",
			newMsg: "newMsg1",
		},
		"previously non-empty messages": {
			oldMsg: "oldMsg1",
			newMsg: "newMsg1",
		},
		"empty error message being set": {
			oldMsg: "oldMsg1",
			newMsg: "",
		},
		"non-changing error message": {
			oldMsg: "msg",
			newMsg: "msg",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := New(tt.oldMsg).WithMessage("%s", tt.newMsg)
			require.Equal(t, tt.newMsg, err.Message)
		})
	}
}

func Test_WithRemediation(t *testing.T) {
	tests := map[string]struct {
		oldMsg string
		newMsg string
	}{
		"previously empty error remediation message": {
			oldMsg: "",
			newMsg: "newMsg1",
		},
		"previously non-empty remediation messages": {
			oldMsg: "oldMsg1",
			newMsg: "newMsg1",
		},
		"empty error remediation message being set": {
			oldMsg: "oldMsg1",
			newMsg: "",
		},
		"non-changing errorremediation message": {
			oldMsg: "msg",
			newMsg: "msg",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := New("").WithRemediation("%s", tt.newMsg)
			require.Equal(t, tt.newMsg, err.Remediation)
		})
	}
}

func Test_WithRootCause(t *testing.T) {
	tests := map[string]struct {
		err       Error
		rootCause error
	}{
		"simple case - error with no initial error details": {
			err:       *New("test"),
			rootCause: New("i am the cause"),
		},
		"simple case 2 - error with no initial error details": {
			err:       *New("test"),
			rootCause: &Error{Code: "code", Message: "msg"},
		},
		"add rootcause error (which does not have error details) to an error with previous details": {
			err:       *New("test"),
			rootCause: &Error{Code: "code", Message: "msg", Details: ErrorDetails{{Message: "d1"}}},
		},
		"add rootcause error (which has its own error details) to an error with previous details": {
			err:       Error{Code: "code", Message: "msg", Details: ErrorDetails{{Message: "d1"}, {Message: "d2"}}},
			rootCause: &Error{Code: "code", Message: "msg", Details: ErrorDetails{{Message: "d3"}, {Message: "d4"}}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			errDetailsCountBefore := len(tt.err.Details) // keep track of the details count before adding the rootcause
			err := tt.err.WithRootCause(tt.rootCause)    // call the function

			rootCauseErr := ToSlackError(tt.rootCause) // cast the root cause error to a slack error for easy comparison

			// During rootCause addition, the rootCause error's message is added as a detail
			// then in addition, all its details are also carried over
			// so the total count of details we have should be the sum of the len(err.Details) + len(rootCause.Details) + 1
			require.Equal(t, errDetailsCountBefore+len(rootCauseErr.Details)+1, len(err.Details))

			require.Contains(t, err.Error(), rootCauseErr.Code)
			require.Contains(t, err.Error(), rootCauseErr.Message)
			for _, d := range rootCauseErr.Details {
				require.Contains(t, tt.err.Details, d) // see to it that all the details in rootCause were carried over
			}
		})
	}
}

func Test_nil_error(t *testing.T) {
	var testErr *Error = nil
	require.Nil(t, testErr.WithCode("test"))
	require.Nil(t, testErr.WithRootCause(New("test")))
	require.Nil(t, testErr.WithMessage("test"))
	require.Nil(t, testErr.WithRemediation("test"))
	require.Nil(t, testErr.WithDetails(ErrorDetails{}))
	require.Nil(t, testErr.AddDetail(ErrorDetail{}))
}
