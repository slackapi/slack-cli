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

// Datastorer is a named collection of key-value pairs for an app
type Datastorer interface {
	Name() string   // Name returns the name of the datastore
	SetName(string) // SetName sets the name of the datastore
	AppID() string  // AppID returns the app ID for the datastore
}

type AppDatastoreQuery struct {
	Datastore            string                 `json:"datastore,omitempty"`
	App                  string                 `json:"app,omitempty"`
	Expression           string                 `json:"expression,omitempty"`
	ExpressionAttributes map[string]interface{} `json:"expression_attributes,omitempty"`
	ExpressionValues     map[string]interface{} `json:"expression_values,omitempty"`
	Limit                int                    `json:"limit,omitempty"`
	Cursor               string                 `json:"cursor,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreQuery) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreQuery) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreQuery) AppID() string {
	return datastore.App
}

type AppDatastoreQueryResult struct {
	Datastore  string                   `json:"datastore,omitempty"`
	Items      []map[string]interface{} `json:"items,omitempty"`
	NextCursor string                   `json:"next_cursor,omitempty"`
}

type AppDatastoreCount struct {
	Datastore            string                 `json:"datastore,omitempty"`
	App                  string                 `json:"app,omitempty"`
	Expression           string                 `json:"expression,omitempty"`
	ExpressionAttributes map[string]interface{} `json:"expression_attributes,omitempty"`
	ExpressionValues     map[string]interface{} `json:"expression_values,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreCount) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreCount) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreCount) AppID() string {
	return datastore.App
}

type AppDatastoreCountResult struct {
	Datastore string `json:"datastore,omitempty"`
	Count     int    `json:"count,omitempty"`
}

type AppDatastorePut struct {
	Datastore string                 `json:"datastore,omitempty"`
	App       string                 `json:"app,omitempty"`
	Item      map[string]interface{} `json:"item,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastorePut) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastorePut) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastorePut) AppID() string {
	return datastore.App
}

type AppDatastorePutResult struct {
	Datastore string                 `json:"datastore,omitempty"`
	Item      map[string]interface{} `json:"item,omitempty"`
}

type AppDatastoreBulkPut struct {
	Datastore string                   `json:"datastore,omitempty"`
	App       string                   `json:"app,omitempty"`
	Items     []map[string]interface{} `json:"items,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreBulkPut) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreBulkPut) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreBulkPut) AppID() string {
	return datastore.App
}

type AppDatastoreBulkPutResult struct {
	Datastore   string                   `json:"datastore,omitempty"`
	FailedItems []map[string]interface{} `json:"failed_items,omitempty"`
}

type AppDatastoreUpdate struct {
	Datastore string                 `json:"datastore,omitempty"`
	App       string                 `json:"app,omitempty"`
	Item      map[string]interface{} `json:"item,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreUpdate) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreUpdate) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreUpdate) AppID() string {
	return datastore.App
}

type AppDatastoreUpdateResult struct {
	Datastore string                 `json:"datastore,omitempty"`
	Item      map[string]interface{} `json:"item,omitempty"`
}

type AppDatastoreDelete struct {
	Datastore string `json:"datastore,omitempty"`
	App       string `json:"app,omitempty"`
	ID        string `json:"id,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreDelete) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreDelete) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreDelete) AppID() string {
	return datastore.App
}

type AppDatastoreDeleteResult struct {
	Datastore string `json:"datastore,omitempty"`
	ID        string `json:"id,omitempty"`
}

type AppDatastoreBulkDelete struct {
	Datastore string   `json:"datastore,omitempty"`
	App       string   `json:"app,omitempty"`
	IDs       []string `json:"ids,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreBulkDelete) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreBulkDelete) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreBulkDelete) AppID() string {
	return datastore.App
}

type AppDatastoreBulkDeleteResult struct {
	Datastore   string   `json:"datastore,omitempty"`
	FailedItems []string `json:"failed_items,omitempty"`
}

type AppDatastoreGet struct {
	Datastore string `json:"datastore,omitempty"`
	App       string `json:"app,omitempty"`
	ID        string `json:"id,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreGet) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreGet) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreGet) AppID() string {
	return datastore.App
}

type AppDatastoreGetResult struct {
	Datastore string                 `json:"datastore,omitempty"`
	Item      map[string]interface{} `json:"item,omitempty"`
}

type AppDatastoreBulkGet struct {
	Datastore string   `json:"datastore,omitempty"`
	App       string   `json:"app,omitempty"`
	IDs       []string `json:"ids,omitempty"`
}

// Name returns the name of the datastore
func (datastore *AppDatastoreBulkGet) Name() string {
	return datastore.Datastore
}

// SetName sets the name of the datastore
func (datastore *AppDatastoreBulkGet) SetName(name string) {
	datastore.Datastore = name
}

// AppID returns the app ID for the datastore
func (datastore *AppDatastoreBulkGet) AppID() string {
	return datastore.App
}

type AppDatastoreBulkGetResult struct {
	Datastore   string                   `json:"datastore,omitempty"`
	Items       []map[string]interface{} `json:"items,omitempty"`
	FailedItems []string                 `json:"failed_items,omitempty"`
}
