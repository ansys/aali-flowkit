// Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package externalfunctions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/russross/blackfriday/v2"

	"github.com/ansys/aali-flowkit/pkg/meshpilot/ampgraphdb"
	"github.com/ansys/aali-flowkit/pkg/meshpilot/azure"

	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/qdrant/go-client/qdrant"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

// SimilartitySearchOnPathDescriptions do similarity search on path description
//
// Tags:
//   - @displayName: SimilartitySearchOnPathDescriptions
//
// Parameters:
//   - instruction: the user query
//   - toolName: the tool name
//
// Returns:
//   - descriptions: the list of descriptions
func SimilartitySearchOnPathDescriptions(instruction string, toolName string) (descriptions []string) {
	descriptions = []string{}
	ctx := &logging.ContextMap{}

	db_endpoint := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["MESHPILOT_DB_ENDPOINT"]
	logging.Log.Debugf(ctx, "DB Endpoint: %q", db_endpoint)

	toolName1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_1_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolName2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_1_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 2 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolName3, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_3_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 3 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolName4, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_2_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 4 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolName5, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_3_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 5 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolName6, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_4_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 6 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolName7, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_5_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 7 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolName8, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_6_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool name 8 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	collection1Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["COLLECTION_1_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load collection name 1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	collection2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["COLLECTION_2_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load collection name 2 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	collection3Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["COLLECTION_3_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load collection name 3 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	collection4Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["COLLECTION_4_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load collection name 4 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	collection5Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["COLLECTION_5_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load collection name 5 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	collection6Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["COLLECTION_6_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load collection name 6 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	collection_name := ""
	if toolName == toolName5 {
		collection_name = collection2Name
	} else if toolName == toolName6 {
		collection_name = collection3Name
	} else if toolName == toolName4 {
		collection_name = collection4Name
	} else if toolName == toolName7 {
		collection_name = collection5Name
	} else if toolName == toolName8 {
		collection_name = collection6Name
	} else if toolName == toolName1 ||
		toolName == toolName2 ||
		toolName == toolName3 {
		collection_name = collection1Name
	} else {
		errorMessage := fmt.Sprintf("Invalid Tool Name: %q", toolName)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	db_url := fmt.Sprintf("%s%s%s", db_endpoint, "/qdrant/similar_descriptions/from/", collection_name)
	logging.Log.Debugf(ctx, "Constructed URL: %s", db_url)

	body := map[string]string{
		"query": instruction,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to marshal request body: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	logging.Log.Debugf(ctx, "Request Body: %s", string(bodyBytes))

	req, err := http.NewRequest("POST", db_url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to create request: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to send request: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to read response body: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	logging.Log.Debugf(ctx, "Response: %s", string(responseBody))

	var response struct {
		Descriptions []string `json:"descriptions"`
	}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to unmarshal response: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	descriptions = response.Descriptions
	logging.Log.Debugf(ctx, "Descriptions: %q", descriptions)
	return
}

// FindRelevantPathDescription finds the relevant description by prompting
//
// Tags:
//   - @displayName: FindRelevantPathDescription
//
// Parameters:
//   - descriptions: the list of descriptions
//   - message: the message from llm
//
// Returns:
//   - relevantDescription: the relevant desctiption
func FindRelevantPathDescription(descriptions []string, message string) (relevantDescription string) {

	relevantDescription = ""
	ctx := &logging.ContextMap{}

	if len(descriptions) == 0 {
		logging.Log.Error(ctx, "no descriptions provided to this function")
		return
	}

	if len(descriptions) == 1 {
		relevantDescription = descriptions[0]
		return
	}

	if len(message) == 0 {
		errorMessage := fmt.Sprintf("no message found from the choice")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Log the response content for debugging
	logging.Log.Debugf(ctx, "Response Content: %s", message)

	// Strip backticks and "json" label from the response content
	cleanedContent := strings.TrimSpace(message)
	if strings.HasPrefix(cleanedContent, "```json") && strings.HasSuffix(cleanedContent, "```") {
		cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
		cleanedContent = strings.TrimSuffix(cleanedContent, "```")
		cleanedContent = strings.TrimSpace(cleanedContent)
	}

	var output *struct {
		Index int `json:"index"`
	}

	err := json.Unmarshal([]byte(cleanedContent), &output)
	if err != nil {
		logging.Log.Errorf(ctx, "Failed to unmarshal response content: %s, error: %v", cleanedContent, err)
		logging.Log.Warn(ctx, "Falling back to the first description as relevant.")
		relevantDescription = descriptions[0]
		return
	}

	logging.Log.Debugf(ctx, "The Index: %d", output.Index)

	if output.Index < len(descriptions) && output.Index >= 0 {
		relevantDescription = descriptions[output.Index]
	} else {
		errorMessage := fmt.Sprintf("Output Index: %d, out of range( 0, %d )", output.Index, len(descriptions))
		logging.Log.Error(ctx, errorMessage)
		logging.Log.Warn(ctx, "Falling back to the first description as relevant.")
		relevantDescription = descriptions[0]
	}

	logging.Log.Infof(ctx, "The relevant description: %s", relevantDescription)

	return
}

// FetchPropertiesFromPathDescription get properties from path description
//
// Tags:
//   - @displayName: FetchPropertiesFromPathDescription
//
// Parameters:
//   - description: the desctiption of path
//
// Returns:
//   - properties: the list of descriptions
func FetchPropertiesFromPathDescription(db_name, description string) (properties []string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Properties From Path Descriptions...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	query := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_PROPERTIES_QUERY"]

	properties, err = ampgraphdb.GraphDbDriver.GetProperties(description, query)

	if err != nil {
		errorMessage := fmt.Sprintf("Error fetching properties from path description: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	logging.Log.Debugf(ctx, "Propetries: %q\n", properties)
	return
}

// FetchNodeDescriptionsFromPathDescription get node descriptions from path description
//
// Tags:
//   - @displayName: FetchNodeDescriptionsFromPathDescription
//
// Parameters:
//   - description: the desctiption of path
//
// Returns:
//   - actionDescriptions: action descriptions
func FetchNodeDescriptionsFromPathDescription(db_name, description string) (actionDescriptions string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Node Descriptions From Path Descriptions...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	// Get environment variables
	query := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_STATE_NODE_QUERY"]

	summaries, err := ampgraphdb.GraphDbDriver.GetSummaries(description, query)

	if err != nil {
		errorMessage := fmt.Sprintf("Error fetching summaries from path description: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actionDescriptions = summaries
	logging.Log.Debugf(ctx, "Summaries: %q\n", actionDescriptions)

	return
}

// FetchActionsPathFromPathDescription fetch actions from path description
//
// Tags:
//   - @displayName: FetchActionsPathFromPathDescription
//
// Parameters:
//   - description: the desctiption of path
//   - nodeLabel: the label of the node
//
// Returns:
//   - actions: the list of actions to execute
func FetchActionsPathFromPathDescription(db_name, description, nodeLabel string) (actions []map[string]string) {
	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Actions From Path Descriptions...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	// Get the node label 1 from the configuration
	nodeLabel1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_LABEL_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load node label 1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get the node label 2 from the configuration
	nodeLabel2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_LABEL_2"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load node label 2 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	var query string
	if nodeLabel == nodeLabel1 {
		query, exists = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_ACTIONS_QUERY_LABEL_1"]
	} else if nodeLabel == nodeLabel2 {
		query, exists = config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_ACTIONS_QUERY_LABEL_2"]
	} else {
		errorMessage := fmt.Sprintf("Invalid Node Label: %q", nodeLabel)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actions, err = ampgraphdb.GraphDbDriver.GetActions(description, query)
	if err != nil {
		errorMessage := fmt.Sprintf("Error fetching actions from path description: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	return
}

// SynthesizeActionsTool2 update action as per user instruction
//
// Tags:
//   - @displayName: SynthesizeActionsTool2
//
// Parameters:
//   - message: the message from the llm
//   - actions: the list of actions
//
// Returns:
//   - updatedActions: the list of synthesized actions
func SynthesizeActionsTool2(message string, actions []map[string]string) (updatedActions []map[string]string) {

	ctx := &logging.ContextMap{}

	updatedActions = actions

	if len(message) == 0 {
		errorMessage := fmt.Sprintf("the message is empty, cannot synthesize actions")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	logging.Log.Debugf(ctx, "The Message: %s\n", message)

	// Clean the response content
	cleanedContent := strings.TrimSpace(message)
	if strings.HasPrefix(cleanedContent, "```json") && strings.HasSuffix(cleanedContent, "```") {
		cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
		cleanedContent = strings.TrimSuffix(cleanedContent, "```")
		cleanedContent = strings.TrimSpace(cleanedContent)
	}

	var output struct {
		ScopePattern string `json:"ScopePattern"`
	}

	err := json.Unmarshal([]byte(cleanedContent), &output)
	if err != nil {
		errorMessage := fmt.Sprintf("SynthesizeActionsTool2: Failed to unmarshal response: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	scopePattern := output.ScopePattern

	logging.Log.Debugf(ctx, "scopePattern: %q\n", scopePattern)

	// Get synthesize actions find key from configuration
	synthesizeActionsFindKey, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_FIND_KEY"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize actions find key from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	synthesizeActionsValue, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_TOOL2_VALUE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize actions find key from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	synthesizeActionsReplaceKey, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_REPLACE_KEY_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize actions find key from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Updated actions from output
	for i := 0; i < len(updatedActions); i++ {
		updateAction := updatedActions[i]
		for key, value := range updateAction {
			if key == synthesizeActionsFindKey && value == synthesizeActionsValue {
				updateAction[synthesizeActionsReplaceKey] = scopePattern
			}
		}
	}

	logging.Log.Debugf(ctx, "The Updated Actions: %q\n", updatedActions)

	return
}

// SynthesizeActionsTool11 synthesize actions based on user instruction
//
// Tags:
//   - @displayName: SynthesizeActionsTool11
//
// Parameters:
//   - content: the llm content
//
// Returns:
//   - result: the synthesized string
func SynthesizeActionsTool11(content string) (result string) {
	ctx := &logging.ContextMap{}

	var out struct {
		UnitSystem string `json:"UnitSystem"`
	}
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		panic(fmt.Sprintf("unmarshal UnitSystem failed: %v", err))
	}
	unitSystem := out.UnitSystem
	logging.Log.Infof(ctx, "Synthesized UnitSystem: %s", unitSystem)

	message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_11_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_ACTION_TOOL_11_SUCCESS_MESSAGE from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actionKey1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTIONS_KEY_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTIONS_KEY_1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	actionKey2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTIONS_KEY_2"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTIONS_KEY_2 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	actionValue1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_11_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTION_11_NAME from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	actionValue2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTIONS_TARGET_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTIONS_TARGET_1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actions := []map[string]string{
		{
			actionKey1:      actionValue1,
			actionKey2:      actionValue2,
			"ArgumentUnits": unitSystem,
		},
	}

	finalMessage := map[string]interface{}{
		"Message": message,
		"Actions": actions,
	}

	resultStream, err := json.Marshal(finalMessage)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to marshal final message for tool 13: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	result = string(resultStream)
	logging.Log.Infof(ctx, "SynthesizeActionsTool13 result: %s", result)
	logging.Log.Infof(ctx, "successfully synthesized actions for tool 13")

	return
}

// SynthesizeActionsTool12 synthesize actions based on user instruction
//
// Tags:
//   - @displayName: SynthesizeActionsTool12
//
// Parameters:
//   - content: the llm content
//
// Returns:
//   - result: the synthesized string
func SynthesizeActionsTool12(content string) (result string) {
	ctx := &logging.ContextMap{}

	var out struct {
		Argument string `json:"Argument"`
	}

	if err := json.Unmarshal([]byte(content), &out); err != nil {
		panic(fmt.Sprintf("unmarshal Argument failed: %v", err))
	}

	Argument := out.Argument
	logging.Log.Infof(ctx, "Synthesized Argument: %s", Argument)

	message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_12_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_ACTION_TOOL_12_SUCCESS_MESSAGE from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actionKey1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTIONS_KEY_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTIONS_KEY_1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actionKey2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTIONS_KEY_2"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTIONS_KEY_2 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actionValue1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_12_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTION_12_NAME from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actionValue2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTIONS_TARGET_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load APP_TOOL_ACTIONS_TARGET_1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	actions := []map[string]string{
		{
			actionKey1: actionValue1,
			actionKey2: actionValue2,
			"Argument": Argument,
		},
	}

	finalMessage := map[string]interface{}{
		"Message": message,
		"Actions": actions,
	}

	resultStream, err := json.Marshal(finalMessage)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to marshal final message for tool 14: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	result = string(resultStream)
	logging.Log.Infof(ctx, "SynthesizeActionsTool14 result: %s", result)
	logging.Log.Infof(ctx, "successfully synthesized actions for tool 14")

	return result
}

// SynthesizeActions update action as per user instruction
//
// Tags:
//   - @displayName: SynthesizeActions
//
// Parameters:
//   - message: the message from the llm
//   - properties: the list of properties
//   - actions: the list of actions
//
// Returns:
//   - updatedActions: the list of synthesized actions
func SynthesizeActions(message string, properties []string, actions []map[string]string) (updatedActions []map[string]string) {

	ctx := &logging.ContextMap{}

	updatedActions = actions

	if len(properties) == 0 {
		logging.Log.Infof(ctx, "No properties to synthesize actions")
		return
	}

	if len(message) == 0 {
		errorMessage := fmt.Sprintf("the message is empty, cannot synthesize actions")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	logging.Log.Debugf(ctx, "The Message: %s\n", message)

	var output map[string]interface{}

	// Attempt to unmarshal the response
	err := json.Unmarshal([]byte(message), &output)
	if err != nil {
		// Log the error and fallback to an empty output
		logging.Log.Errorf(ctx, "Failed to unmarshal response content: %s, error: %v", message, err)

		// Attempt to clean the response and retry unmarshaling
		cleanedContent := strings.TrimSpace(message)
		if strings.HasPrefix(cleanedContent, "```json") && strings.HasSuffix(cleanedContent, "```") {
			cleanedContent = strings.TrimPrefix(cleanedContent, "```json")
			cleanedContent = strings.TrimSuffix(cleanedContent, "```")
			cleanedContent = strings.TrimSpace(cleanedContent)
		}

		err = json.Unmarshal([]byte(cleanedContent), &output)
		if err != nil {
			logging.Log.Errorf(ctx, "Failed to unmarshal cleaned response content: %s, error: %v", cleanedContent, err)
			logging.Log.Warn(ctx, "Returning an empty output as fallback.")
			output = make(map[string]interface{})
		}
	}

	logging.Log.Debugf(ctx, "The LLM Output of properties processing: %q\n", output)

	// Get synthesize actions find key from configuration
	synthesizeActionsFindKey, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_FIND_KEY"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize actions find key from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get synthesize actions replace key 1 from configuration
	synthesizeActionsReplaceKey1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_REPLACE_KEY_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize actions replace key 1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get synthesize actions replace key 2 from configuration
	synthesizeActionsReplaceKey2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_REPLACE_KEY_2"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize actions replace key 2 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get synthesize output key 1 from configuration
	synthesizeOutputKey1, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_OUTPUT_KEY_1"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize output key 1 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get synthesize output key 2 from configuration
	synthesizeOutputKey2, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_PROMPT_TEMPLATE_SYNTHESIZE_OUTPUT_KEY_2"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load synthesize output key 2 from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Update the actions
	for key, value := range output {
		switch v := value.(type) {
		case string:
			updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey1, v)
		case int:
		case int64:
		case int32:
			convValue := fmt.Sprintf("%d", v)
			updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey1, convValue)
		case float32:
		case float64:
			convValue := fmt.Sprintf("%f", v)
			updateMeshPilotActionProperty(updatedActions, synthesizeActionsFindKey, key, synthesizeActionsReplaceKey1, convValue)
		case map[string]interface{}:
			outerKey := key

			if val, ok := v[synthesizeOutputKey1].(string); ok {
				updateMeshPilotActionProperty(
					updatedActions,
					synthesizeActionsFindKey,
					outerKey,
					synthesizeActionsReplaceKey1,
					val,
				)
			}
			if units, ok := v[synthesizeOutputKey2].(string); ok {
				updateMeshPilotActionProperty(
					updatedActions,
					synthesizeActionsFindKey,
					outerKey,
					synthesizeActionsReplaceKey2,
					units,
				)
			}
		default:
			logging.Log.Infof(ctx, "Key: %s, Value is of a different type: %T", key, v)
		}
	}

	logging.Log.Debugf(ctx, "The SynthesizeActions Updated Actions: %q\n", updatedActions)

	return
}

// FinalizeResult converts actions to json string to send back data
//
// Tags:
//   - @displayName: FinalizeResult
//
// Parameters:
//   - actions: the executable actions
//   - toolName: tool name to create customize messages
//
// Returns:
//   - result: the actions in json format
func FinalizeResult(actions []map[string]string, toolName string) (result string) {

	ctx := &logging.ContextMap{}

	var hasActions bool
	var message string

	if len(actions) > 0 {
		hasActions = true
	} else {
		hasActions = false
		actions = make([]map[string]string, 0)
	}

	// Get tool 2 name from the configuration
	toolAction1Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_1_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 1 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 4 name from the configuration
	toolAction2Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_2_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 2 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 5 name from the configuration
	toolAction3Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_3_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 3 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 6 name from the configuration
	toolAction4Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_4_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 4 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 7 name from the configuration
	toolAction5Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_5_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 5 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 8 name from the configuration
	toolAction6Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_6_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 6 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction14Name, exexists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_14_NAME"]
	if !exexists {
		errorMessage := fmt.Sprintf("failed to load action tool 14 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool action success message from configuration
	toolAction15Name, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_TOOL_ACTION_15_NAME"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 15 name from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 2 action message from configuration
	toolAction1Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_1_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 1 message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 2 no action message from configuration
	toolAction1NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_1_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load action tool 1 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction2Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_2_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 4 action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 4 no action message from configuration
	toolAction2NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_2_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 4 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction3Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_3_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 5 action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 5 no action message from configuration
	toolAction3NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_3_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 5 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction4Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_4_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 6 action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 6 no action message from configuration
	toolAction4NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_4_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 6 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction5Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_5_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 7 action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 7 no action message from configuration
	toolAction5NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_5_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 7 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction6Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_6_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 8 action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Get tool 8 no action message from configuration
	toolAction6NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_6_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 8 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction14Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_14_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 16 action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction14NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_14_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 16 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction15Message, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_15_SUCCESS_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 17 action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	toolAction15NoActionMessage, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_ACTION_TOOL_15_NO_ACTION_MESSAGE"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load tool 17 no action message from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	if toolName == toolAction1Name {
		if hasActions {
			message = toolAction1Message
		} else {
			message = toolAction1NoActionMessage
		}
	} else if toolName == toolAction2Name {
		if hasActions {
			message = toolAction2Message
		} else {
			message = toolAction2NoActionMessage
		}
	} else if toolName == toolAction3Name {
		if hasActions {
			message = toolAction3Message
		} else {
			message = toolAction3NoActionMessage
		}
	} else if toolName == toolAction4Name {
		if hasActions {
			message = toolAction4Message
		} else {
			message = toolAction4NoActionMessage
		}
	} else if toolName == toolAction5Name {
		if hasActions {
			message = toolAction5Message
		} else {
			message = toolAction5NoActionMessage
		}
	} else if toolName == toolAction6Name {
		if hasActions {
			message = toolAction6Message
		} else {
			message = toolAction6NoActionMessage
		}
	} else if toolName == toolAction14Name {
		if hasActions {
			message = toolAction14Message
		} else {
			message = toolAction14NoActionMessage
		}
	} else if toolName == toolAction15Name {
		if hasActions {
			message = toolAction15Message
		} else {
			message = toolAction15NoActionMessage
		}
	} else {
		errorMessage := fmt.Sprintf("Invalid toolName %s", toolName)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	finalMessage := make(map[string]interface{})
	finalMessage["Message"] = message
	finalMessage["Actions"] = actions

	// Marshal the actions
	bytesStream, err := json.Marshal(finalMessage)

	if err != nil {
		errorMessage := fmt.Sprintf("failed to convert actions to json: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	result = string(bytesStream)
	logging.Log.Info(ctx, "successfully converted actions to json")

	return
}

// GetSolutionsToFixProblem do similarity search on path description
//
// Tags:
//   - @displayName: GetSolutionsToFixProblem
//
// Parameters:
//   - fmFailureCode: FM failure Code
//   - primeMeshFailureCode: Prime Mesh Failure Code
//
// Returns:
//   - solutions: the list of solutions in json
func GetSolutionsToFixProblem(db_name, fmFailureCode, primeMeshFailureCode string) (solutions string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Get Solutions To Fix Problem...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	query, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_DATABASE_GET_SOLUTIONS_QUERY"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load query from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	solutionsVec, err := ampgraphdb.GraphDbDriver.GetSolutions(fmFailureCode, primeMeshFailureCode, query)
	if err != nil {
		errorMessage := fmt.Sprintf("Error fetching solutions from path description: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	byteStream, err := json.Marshal(solutionsVec)
	if err != nil {
		errorMessage := fmt.Sprintf("Error marshalling solutions: %v\n", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	solutions = string(byteStream)
	logging.Log.Info(ctx, "found solutions to fix problem...")
	return
}

// GetSelectedSolution get user selected solutions from the options provided
//
// Tags:
//   - @displayName: GetSelectedSolution
//
// Parameters:
//   - arguments: these are the arguments ReAct found based on user choice
//
// Returns:
//   - solution: the selected solution
func GetSelectedSolution(arguments string) (solution string) {

	ctx := &logging.ContextMap{}

	var output struct {
		Solution string `json:"solution_description"`
	}

	err := json.Unmarshal([]byte(arguments), &output)

	if err != nil {
		errorMessage := fmt.Sprintf("failed to un marshal index output")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	solution = output.Solution
	logging.Log.Infof(ctx, "Selected Solution: %s", solution)
	return
}

// AppendToolHistory this function append tool history
//
// Tags:
//   - @displayName: AppendToolHistory
//
// Parameters:
//   - toolHistory: the tool history
//   - toolId: the tool id
//   - toolName: the tool name
//   - toolArguments: the tool arguments
//   - toolResponse: the tool response
//
// Returns:
//   - updatedToolHistory: the updated tool history
func AppendToolHistory(toolHistory []map[string]string, toolId, toolName, toolArguments, toolResponse string) (updatedToolHistory []map[string]string) {
	ctx := &logging.ContextMap{}

	// populate tool history
	for _, tool := range toolHistory {
		toolId := tool["toolId"]
		content := tool["content"]
		tool := map[string]string{
			"toolId":  toolId,
			"content": content,
		}
		updatedToolHistory = append(updatedToolHistory, tool)
	}

	// populate current tool
	if len(toolId) > 0 && len(toolResponse) > 0 {
		// append assistant tool call
		tool := map[string]string{
			"role":          "assistant",
			"content":       toolResponse,
			"toolId":        toolId,
			"toolName":      toolName,
			"toolArguments": toolArguments,
		}
		updatedToolHistory = append(updatedToolHistory, tool)

		// append tool
		tool = map[string]string{
			"role":    "tool",
			"content": toolResponse,
			"toolId":  toolId,
		}
		updatedToolHistory = append(updatedToolHistory, tool)
	}

	logging.Log.Info(ctx, fmt.Sprintf("Updated Tool History: %q", updatedToolHistory))
	return
}

// AppendMeshPilotHistory this function append mesh pilot history
//
// Tags:
//   - @displayName: AppendMeshPilotHistory
//
// Parameters:
//   - history: the tool history
//   - role: the tool id
//   - content: the tool name
//
// Returns:
//   - updatedHistory: the updated mesh pilot history
func AppendMeshPilotHistory(history []map[string]string, role, content string) (updatedHistory []map[string]string) {
	ctx := &logging.ContextMap{}

	updatedHistory = []map[string]string{}

	for _, item := range history {
		updatedHistory = append(updatedHistory, item)
	}

	updatedHistory = append(updatedHistory, map[string]string{
		"role":    role,
		"content": content,
	})

	logging.Log.Debugf(ctx, "Updated history: %q", updatedHistory)
	return
}

// ParseHistory this function parses history from json to map
//
// Tags:
//   - @displayName: ParseHistory
//
// Parameters:
//   - historyJson: history in json format
//
// Returns:
//   - history: the parsed history
func ParseHistory(historyJson string) (history []map[string]string) {
	ctx := &logging.ContextMap{}

	history = []map[string]string{}

	// convert json to map
	var historyMap []map[string]string
	err := json.Unmarshal([]byte(historyJson), &historyMap)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to unmarshal history json: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// populate history
	for _, item := range historyMap {
		history = append(history, item)
	}
	logging.Log.Debugf(ctx, "Parsed history: %q", history)
	return
}

// FinalizeResult converts actions to json string to send back data
//
// Tags:
//   - @displayName: GetActionsFromConfig
//
// Parameters:
//   - toolName: tool name to create customize messages
//
// Returns:
//   - result: the actions in json format
func GetActionsFromConfig(toolName string) (result string) {
	ctx := &logging.ContextMap{}

	logging.Log.Info(ctx, "Get Actions From Config...")
	logging.Log.Infof(ctx, "Tool Name: %q", toolName)

	// Configuration keys for different tools, for now only tool 9 and tool 11
	configKeys := map[string]map[string]string{
		"tool9": {
			"resultName":    "APP_ACTION_TOOL7_RESULT_NAME",
			"resultMessage": "APP_ACTION_TOOL7_RESULT_MESSAGE",
			"actionValue1":  "APP_ACTIONS_VALUE_1_ACTION_TOOL7",
			"actionValue2":  "APP_TOOL_ACTIONS_TARGET_2",
		},
		"tool11": {
			"resultName":    "APP_ACTION_TOOL9_RESULT_NAME",
			"resultMessage": "APP_ACTION_TOOL9_RESULT_MESSAGE",
			"actionValue1":  "APP_ACTIONS_VALUE_1_ACTION_TOOL9",
			"actionValue2":  "APP_TOOL_ACTIONS_TARGET_1",
		},
		"tool12": {
			"resultName":    "APP_ACTION_TOOL10_RESULT_NAME",
			"resultMessage": "APP_ACTION_TOOL10_RESULT_MESSAGE",
			"actionValue1":  "APP_ACTIONS_VALUE_1_ACTION_TOOL10",
			"actionValue2":  "APP_TOOL_ACTIONS_TARGET_2",
		},
		"tool15": {
			"resultName":    "APP_ACTION_TOOL13_RESULT_NAME",
			"resultMessage": "APP_ACTION_TOOL13_RESULT_MESSAGE",
			"actionValue1":  "APP_ACTIONS_VALUE_1_ACTION_TOOL13",
			"actionValue2":  "APP_TOOL_ACTIONS_TARGET_3",
		},
		"tool18": {
			"resultName":    "APP_ACTION_TOOL16_RESULT_NAME",
			"resultMessage": "APP_ACTION_TOOL16_RESULT_MESSAGE",
			"actionValue1":  "APP_ACTIONS_VALUE_1_ACTION_TOOL16",
			"actionValue2":  "APP_TOOL_ACTIONS_TARGET_1",
		},
	}

	// Help function to get the config value
	getConfigValue := func(key string, errorMsg string) string {
		value, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES[key]
		if !exists {
			errorMessage := fmt.Sprintf("%s: %s", errorMsg, key)
			logging.Log.Error(ctx, errorMessage)
			panic(errorMessage)
		}
		return value
	}

	// Get tool result name from the configuration
	tool9ResultName := getConfigValue(configKeys["tool9"]["resultName"], "failed to load tool 9 result name from the configuration")
	tool11ResultName := getConfigValue(configKeys["tool11"]["resultName"], "failed to load tool 11 result name from the configuration")
	tool12ResultName := getConfigValue(configKeys["tool12"]["resultName"], "failed to load tool 12 result name from the configuration")
	tool15ResultName := getConfigValue(configKeys["tool15"]["resultName"], "failed to load tool 15 result name from the configuration")
	tool18ResultName := getConfigValue(configKeys["tool18"]["resultName"], "failed to load tool 18 result name from the configuration")

	// Get tool result message from the configuration
	tool9ResultMessage := getConfigValue(configKeys["tool9"]["resultMessage"], "failed to load tool 9 result message from the configuration")
	tool11ResultMessage := getConfigValue(configKeys["tool11"]["resultMessage"], "failed to load tool 11 result message from the configuration")
	tool12ResultMessage := getConfigValue(configKeys["tool12"]["resultMessage"], "failed to load tool 12 result message from the configuration")
	tool15ResultMessage := getConfigValue(configKeys["tool15"]["resultMessage"], "failed to load tool 15 result message from the configuration")
	tool18ResultMessage := getConfigValue(configKeys["tool18"]["resultMessage"], "failed to load tool 18 result message from the configuration")

	// Get tool action success message from configuration
	toolActionSuccessMessage := getConfigValue("APP_TOOL_ACTION_SUCCESS_MESSAGE", "failed to load tool action success message from the configuration")
	actionKey1 := getConfigValue("APP_TOOL_ACTIONS_KEY_1", "failed to load tool action key 1 from the configuration")
	actionKey2 := getConfigValue("APP_TOOL_ACTIONS_KEY_2", "failed to load tool action key 2 from the configuration")

	// Initialize action values and message
	var actionValue1, actionValue2, selectedMessage string

	// Based on the tool name, set the action values and message
	if toolName == tool9ResultName {
		actionValue1 = getConfigValue(configKeys["tool9"]["actionValue1"], "failed to load tool 9 action value 1 from the configuration")
		actionValue2 = getConfigValue(configKeys["tool9"]["actionValue2"], "failed to load tool 9 action value 2 from the configuration")
		selectedMessage = tool9ResultMessage
	} else if toolName == tool11ResultName {
		actionValue1 = getConfigValue(configKeys["tool11"]["actionValue1"], "failed to load tool 11 action value 1 from the configuration")
		actionValue2 = getConfigValue(configKeys["tool11"]["actionValue2"], "failed to load tool 11 action value 2 from the configuration")
		selectedMessage = tool11ResultMessage
	} else if toolName == tool12ResultName {
		actionValue1 = getConfigValue(configKeys["tool12"]["actionValue1"], "failed to load tool 12 action value 1 from the configuration")
		actionValue2 = getConfigValue(configKeys["tool12"]["actionValue2"], "failed to load tool 12 action value 2 from the configuration")
		selectedMessage = tool12ResultMessage
	} else if toolName == tool15ResultName {
		actionValue1 = getConfigValue(configKeys["tool15"]["actionValue1"], "failed to load tool 15 action value 1 from the configuration")
		actionValue2 = getConfigValue(configKeys["tool15"]["actionValue2"], "failed to load tool 15 action value 2 from the configuration")
		selectedMessage = tool15ResultMessage
	} else if toolName == tool18ResultName {
		actionValue1 = getConfigValue(configKeys["tool18"]["actionValue1"], "failed to load tool 18 action value 1 from the configuration")
		actionValue2 = getConfigValue(configKeys["tool18"]["actionValue2"], "failed to load tool 18 action value 2 from the configuration")
		selectedMessage = tool18ResultMessage
	} else {
		errorMessage := fmt.Sprintf("Invalid toolName %s", toolName)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	message := toolActionSuccessMessage
	if toolName == tool9ResultName || toolName == tool11ResultName || toolName == tool12ResultName || toolName == tool15ResultName || toolName == tool18ResultName {
		message = selectedMessage
		actions := []map[string]string{
			{
				actionKey1: actionValue1,
				actionKey2: actionValue2,
			},
		}
		finalMessage := map[string]interface{}{
			"Message": message,
			"Actions": actions,
		}
		bytesStream, err := json.Marshal(finalMessage)
		if err != nil {
			errorMessage := fmt.Sprintf("failed to convert actions to json: %v", err)
			logging.Log.Error(ctx, errorMessage)
			panic(errorMessage)
		}
		result = string(bytesStream)
		logging.Log.Infof(ctx, "successfully converted actions to json: %q", result)
	} else {
		errorMessage := fmt.Sprintf("Invalid toolName %s", toolName)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	return result
}

// SimilartitySearchOnPathDescriptions (Qdrant) do similarity search on path description
//
// Tags:
//   - @displayName: SimilartitySearchOnPathDescriptions (Qdrant)
//
// Parameters:
//   - instruction: the user query
//   - toolName: the tool name
//
// Returns:
//   - descriptions: the list of descriptions
func SimilartitySearchOnPathDescriptionsQdrant(vector []float32, collection string, similaritySearchResults int, similaritySearchMinScore float64) (descriptions []string) {
	descriptions = []string{}

	logCtx := &logging.ContextMap{}

	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	limit := uint64(similaritySearchResults)
	scoreThreshold := float32(similaritySearchMinScore)
	query := qdrant.QueryPoints{
		CollectionName: collection,
		Query:          qdrant.NewQueryDense(vector),
		Limit:          &limit,
		ScoreThreshold: &scoreThreshold,
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadInclude("Description"),
	}

	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}
	logging.Log.Debugf(logCtx, "Got %d points from qdrant query", len(scoredPoints))

	for i, scoredPoint := range scoredPoints {
		logging.Log.Debugf(&logging.ContextMap{}, "Result #%d:", i)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity score: %v", scoredPoint.Score)
		dbResponse, err := qdrant_utils.QdrantPayloadToType[map[string]interface{}](scoredPoint.GetPayload())

		if err != nil {
			errMsg := fmt.Sprintf("error converting qdrant payload to dbResponse: %q", err)
			logging.Log.Errorf(logCtx, "%s", errMsg)
			panic(errMsg)
		}

		description, ok := dbResponse["Description"].(string)
		if !ok {
			logging.Log.Errorf(&logging.ContextMap{}, "Description not found or not a string for scored point #%d", i)
			continue
		}
		logging.Log.Debugf(&logging.ContextMap{}, "Description: %s", description)

		descriptions = append(descriptions, description)
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Descriptions: %q", descriptions)
	return
}

// ParseHistoryToHistoricMessages this function to convert chat history to historic messages
//
// Tags:
//   - @displayName: ParseHistoryToHistoricMessages
//
// Parameters:
//   - historyJson: chat history in json format
//
// Returns:
//   - history: the history in sharedtypes.HistoricMessage format
func ParseHistoryToHistoricMessages(historyJson string) (history []sharedtypes.HistoricMessage) {
	ctx := &logging.ContextMap{}

	var historyMaps []map[string]string
	err := json.Unmarshal([]byte(historyJson), &historyMaps)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to unmarshal history json: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	for _, msg := range historyMaps {
		role, _ := msg["role"]
		content, _ := msg["content"]
		history = append(history, sharedtypes.HistoricMessage{
			Role:    role,
			Content: content,
		})
	}
	return history
}

// GenerateActionsSubWorkflowPrompt generates system and user prompts for subworkflow identification.
//
// Tags:
//   - @displayName: GenerateActionsSubWorkflowPrompt
//
// Parameters:
//   - userInstruction: user instruction
//
// Returns:
//   - systemPrompt: the system prompt
//   - userPrompt: the user prompt
func GenerateActionsSubWorkflowPrompt(userInstruction string) (systemPrompt string, userPrompt string) {
	ctx := &logging.ContextMap{}

	// Retrieve subworkflows (name and description)
	subworkflows := azure.GetActionsSubworkflows()
	var subworkflowListStr strings.Builder
	for i, sw := range subworkflows {
		subworkflowListStr.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, sw.Name, sw.Description))
	}

	// Retrieve prompt templates from configuration
	systemPromptTemplate, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_SUBWORKFLOW_IDENTIFICATION_SYSTEM_PROMPT"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load system prompt template from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	userPromptTemplate, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_SUBWORKFLOW_IDENTIFICATION_USER_PROMPT"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load user prompt template from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Format the prompts
	systemPrompt = fmt.Sprintf(systemPromptTemplate, subworkflowListStr.String())
	userPrompt = fmt.Sprintf(userPromptTemplate, userInstruction)

	logging.Log.Debugf(ctx, "Generated System Prompt: %s", systemPrompt)

	logging.Log.Debugf(ctx, "Generated User Prompt: %s", userPrompt)
	return systemPrompt, userPrompt
}

// GenerateHelperSubWorkflowPrompt generates system and user prompts for helper subworkflow identification.
//
// Tags:
//   - @displayName: GenerateHelperSubWorkflowPrompt
//
// Parameters:
//   - userInstruction: user instruction
//
// Returns:
//   - systemPrompt: the system prompt
//   - userPrompt: the user prompt
func GenerateHelperSubWorkflowPrompt(userInstruction string) (systemPrompt string, userPrompt string) {
	ctx := &logging.ContextMap{}

	// Retrieve subworkflows (name and description)
	subworkflows := azure.GetHelperSubworkflows()
	var subworkflowListStr strings.Builder
	for i, sw := range subworkflows {
		subworkflowListStr.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, sw.Name, sw.Description))
	}

	// Retrieve prompt templates from configuration
	systemPromptTemplate, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_SUBWORKFLOW_IDENTIFICATION_SYSTEM_PROMPT"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load system prompt template from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}
	userPromptTemplate, exists := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES["APP_SUBWORKFLOW_IDENTIFICATION_USER_PROMPT"]
	if !exists {
		errorMessage := fmt.Sprintf("failed to load user prompt template from the configuration")
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	// Format the prompts
	systemPrompt = fmt.Sprintf(systemPromptTemplate, subworkflowListStr.String())
	userPrompt = fmt.Sprintf(userPromptTemplate, userInstruction)

	logging.Log.Debugf(ctx, "Generated System Prompt: %s", systemPrompt)

	logging.Log.Debugf(ctx, "Generated User Prompt: %s", userPrompt)
	return systemPrompt, userPrompt
}

// GenerateUserPrompt generates user instruction prompt based on the provided template.
//
// Tags:
//   - @displayName: GenerateUserPrompt
//
// Parameters:
//   - userInstruction: user instruction
//   - userPromptTemplate: user prompt template
//
// Returns:
//   - userPrompt: the user prompt
func GenerateUserPrompt(userInstruction string, userPromptTemplate string) (userPrompt string) {
	ctx := &logging.ContextMap{}

	userPrompt = fmt.Sprintf(userPromptTemplate, userInstruction)

	logging.Log.Debugf(ctx, "Generated User Prompt: %s", userPrompt)

	return
}

// GenerateUserPromptWithContext generates user instruction prompt based on the provided template with instruction and context.
//
// Tags:
//   - @displayName: GenerateUserPromptWithContext
//
// Parameters:
//   - userInstruction: user instruction
//   - context: user context
//   - userPromptTemplate: user prompt template
//
// Returns:
//   - userPrompt: the user prompt
func GenerateUserPromptWithContext(userInstruction string, context string, userPromptTemplate string) (userPrompt string) {
	ctx := &logging.ContextMap{}

	userPrompt = fmt.Sprintf(userPromptTemplate, userInstruction, context)

	logging.Log.Debugf(ctx, "Generated User Prompt With Context: %s", userPrompt)

	return
}

// GenerateUserPromptWithList generates user instruction prompt based on the provided template, instruction, list.
//
// Tags:
//   - @displayName: GenerateUserPromptWithList
//
// Parameters:
//   - userInstruction: user instruction
//   - userList: list of items to include in the prompt
//   - userPromptTemplate: user prompt template
//
// Returns:
//   - userPrompt: the user prompt
func GenerateUserPromptWithList(userInstruction string, userList []string, userPromptTemplate string) (userPrompt string) {
	ctx := &logging.ContextMap{}

	userPrompt = fmt.Sprintf(userPromptTemplate, userList, userInstruction)

	logging.Log.Debugf(ctx, "Generated User Prompt: %s", userPrompt)

	return
}

// ProcessSubworkflowIdentificationOutput this function process output of llm
//
// Tags:
//   - @displayName: ProcessSubworkflowIdentificationOutput
//
// Parameters:
//   - llmOutput: the llm output for subworkflow identification
//
// Returns:
//   - status: status of processing
//   - workflowName: the identified subworkflow name
func ProcessSubworkflowIdentificationOutput(llmOutput string) (status string, workflowName string) {
	ctx := &logging.ContextMap{}

	// Clean up the output in case it is wrapped in code block
	cleaned := strings.TrimSpace(llmOutput)
	if strings.HasPrefix(cleaned, "```json") && strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}

	// Parse JSON
	var result struct {
		Subworkflow string `json:"subworkflow"`
	}
	err := json.Unmarshal([]byte(cleaned), &result)
	if err != nil {
		logging.Log.Errorf(ctx, "Failed to parse LLM output as JSON: %v, content: %s", err, cleaned)
		return "failure", ""
	}

	// Check if subworkflow is valid
	if result.Subworkflow == "" || strings.ToLower(result.Subworkflow) == "none" {
		logging.Log.Warnf(ctx, "No valid subworkflow found in LLM output: %s", cleaned)
		return "failure", ""
	}

	logging.Log.Debugf(ctx, "Identified Subworkflow: %s", result.Subworkflow)

	return "success", result.Subworkflow
}

// ProcessMainAgentOutput this function process output of llm
//
// Tags:
//   - @displayName: ProcessMainAgentOutput
//
// Parameters:
//   - llmOutput: the llm output for main agent
//
// Returns:
//   - messageTo: send the message to this recipient
//   - message: message to send to the recipient
func ProcessMainAgentOutput(llmOutput string) (messageTo string, message string) {
	ctx := &logging.ContextMap{}

	// Clean up the output in case it is wrapped in code block
	cleaned := strings.TrimSpace(llmOutput)
	if strings.HasPrefix(cleaned, "```json") && strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}

	// Parse JSON
	var result struct {
		MessageTo string `json:"messageTo"`
		Message   string `json:"message"`
	}

	err := json.Unmarshal([]byte(cleaned), &result)
	if err != nil {
		errorMessage := fmt.Sprintf("Failed to parse LLM output as JSON: %v, content: %s", err, cleaned)
		logging.Log.Error(ctx, errorMessage)
		return "user", "Main Agent failed to parse the request. Please try again."
	}

	// Check if subworkflow is valid
	if result.MessageTo == "" || result.Message == "" {
		errorMessage := fmt.Sprintf("No valid messageTo or Message property found in LLM output: %s", cleaned)
		logging.Log.Error(ctx, errorMessage)
		return "user", "Main Agent failed to understand the request. Can you please try to rephrase your question?"
	}

	logging.Log.Debugf(ctx, "Message To: %s", result.MessageTo)
	logging.Log.Debugf(ctx, "Message: %s", result.Message)

	return result.MessageTo, result.Message
}

// MarkdownToHTML this function converts markdown to html
//
// Tags:
//   - @displayName: MarkdownToHTML
//
// Parameters:
//   - markdown: content in markdown format
//
// Returns:
//   - html: content in html format
func MarkdownToHTML(markdown string) (html string) {
	logging.Log.Info(&logging.ContextMap{}, "Converting Markdown to HTML...")
	// Use blackfriday to convert markdown to HTML
	logging.Log.Debugf(&logging.ContextMap{}, "Markdown content: %s", markdown)
	html = string(blackfriday.Run([]byte(markdown)))
	return html
}

// FinalizeMessage this function takes message and generate response schema
//
// Tags:
//   - @displayName: FinalizeMessage
//
// Parameters:
//   - message: final message
//
// Returns:
//   - result: response schema sent to chat interface
func FinalizeMessage(message string) (result string) {
	ctx := &logging.ContextMap{}
	logging.Log.Info(ctx, "Finalizing message...")

	actions := make([]map[string]string, 0)

	finalMessage := make(map[string]interface{})
	finalMessage["Message"] = message
	finalMessage["Actions"] = actions

	// Marshal the actions
	bytesStream, err := json.Marshal(finalMessage)

	if err != nil {
		errorMessage := fmt.Sprintf("failed to convert actions to json: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	result = string(bytesStream)
	logging.Log.Debugf(ctx, "Final message: %s", result)
	logging.Log.Info(ctx, "successfully converted actions to json")

	return result
}

// SelectedSolution this function parse the selected solution
//
// Tags:
//   - @displayName: SelectedSolution
//
// Parameters:
//   - selectedSolution: selected solution found by the LLM
//
// Returns:
//   - solution: parsed solution
func SelectedSolution(selectedSolution string) (solution string) {
	logging.Log.Info(&logging.ContextMap{}, "Parsing Selected solution...")
	// Use blackfriday to convert markdown to HTML
	logging.Log.Debugf(&logging.ContextMap{}, "Selected Solution: %s", selectedSolution)
	solution = selectedSolution
	return solution
}
