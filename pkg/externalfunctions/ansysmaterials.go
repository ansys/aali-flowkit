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
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
)

type Response struct {
	Criteria []sharedtypes.MaterialCriterionWithGuid
	Tokens   int
}

type LlmCriteria struct {
	Criteria []sharedtypes.MaterialLlmCriterion
}

// StartTrace generates a new trace ID and span ID for tracing
//
// Tags:
//   - @displayName: Start new trace
//
// Parameters:
//   - str: a string
//
// Returns:
//   - traceID: a 128-bit trace ID in decimal format
//   - spanID: a 64-bit span ID in decimal format
func StartTrace() (traceID string, spanID string) {
	traceID = generateTraceID()
	spanID = generateSpanID()
	ctx := &logging.ContextMap{}
	ctx.Set(logging.ContextKey("dd.trace_id"), traceID)
	ctx.Set(logging.ContextKey("dd.span_id"), spanID)
	ctx.Set(logging.ContextKey("dd.trace_idVisible"), traceID)
	ctx.Set(logging.ContextKey("dd.span_idVisible"), spanID)
	logging.Log.Infof(ctx, "Starting new trace with trace ID: %s and span ID: %s", traceID, spanID)

	return traceID, spanID
}

// generateTraceID generates a 128-bit trace ID in decimal format
func generateTraceID() string {
	id := uuid.New()
	traceID := new(big.Int).SetBytes(id[:])
	return traceID.String()
}

// generateSpanID generates a 64-bit span ID in decimal format
func generateSpanID() string {
	id := uuid.New()

	// Take first 64 bits
	spanID := binary.BigEndian.Uint64(id[:8])
	return strconv.FormatUint(spanID, 10)
}

func CreateChildSpan(ctx *logging.ContextMap, traceID string, parentSpanID string) (childSpanID string) {
	// Generate a new span ID for the child
	childSpanID = generateSpanID()

	// Update the context with trace and span information
	ctx.Set(logging.ContextKey("dd.trace_id"), traceID)
	ctx.Set(logging.ContextKey("dd.span_id"), childSpanID)
	ctx.Set(logging.ContextKey("dd.parent_id"), parentSpanID)
	ctx.Set(logging.ContextKey("dd.trace_idVisible"), traceID)
	ctx.Set(logging.ContextKey("dd.span_idVisible"), childSpanID)
	ctx.Set(logging.ContextKey("dd.parent_idVisible"), parentSpanID)

	// logging.Log.Infof(ctx, "Starting child span with trace ID: %s, span ID: %s, and parent span ID: %s", traceID, childSpanID, parentSpanID)

	return childSpanID
}

// SerializeResponse formats the criteria to a response suitable for the UI clients in string format
//
// Tags:
//   - @displayName: Serialize response for clients
//
// Parameters:
//   - tokens: tokens consumed by the request
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - result: string representation of the response in JSON format
//   - childSpanID: the child span ID created for this operation
func SerializeResponse(criteriaSuggestions []sharedtypes.MaterialCriterionWithGuid, tokens int, traceID string, spanID string) (result string, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	response := Response{Criteria: criteriaSuggestions, Tokens: tokens}

	responseJson, err := json.Marshal(response)
	if err != nil {
		logging.Log.Debugf(ctx, "Failed to serialize suggested criteria into json: %v", err)
		panic("Failed to serialize suggested criteria into json")
	}

	return string(responseJson), childSpanID
}

// AddGuidsToAttributes adds GUIDs to the attributes in the criteria
//
// Tags:
//   - @displayName: Add GUIDs to criteria suggestions
//
// Parameters:
//   - criteriaSuggestions: the list of criteria without identities
//   - availableAttributes: the list of available attributes with their identities
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - criteriaWithGuids: the list of criteria with their identities
//   - childSpanID: the child span ID created for this operation
func AddGuidsToAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion, availableAttributes []sharedtypes.MaterialAttribute, traceID string, spanID string) (criteriaWithGuids []sharedtypes.MaterialCriterionWithGuid, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	attributeMap := make(map[string]string)
	for _, attr := range availableAttributes {
		attributeMap[strings.ToLower(attr.Name)] = attr.Guid
	}

	for _, criterion := range criteriaSuggestions {
		lowerAttrName := strings.ToLower(criterion.AttributeName)
		guid, exists := attributeMap[lowerAttrName]

		if !exists {
			logging.Log.Debugf(ctx, "Could not find attribute to match: %s", lowerAttrName)
			continue // This might have been an hallucinated attribute, skip it
		}

		criteriaWithGuids = append(criteriaWithGuids, sharedtypes.MaterialCriterionWithGuid{
			AttributeName: criterion.AttributeName,
			AttributeGuid: guid,
			Explanation:   criterion.Explanation,
			Confidence:    criterion.Confidence,
		})
	}

	return criteriaWithGuids, childSpanID
}

// FilterOutNonExistingAttributes filters out criteria suggestions that do not match any of the available attributes based on their GUIDs
//
// Tags:
//   - @displayName: Filter out non-existing attributes
//
// Parameters:
//   - criteriaSuggestions: current list of criteria suggestions
//   - availableSearchCriteria: the list of available search criteria (GUIDs)
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - filtered: the list of criteria suggestions excluding those that do not match any of the available search criteria
//   - childSpanID: the child span ID created for this operation
func FilterOutNonExistingAttributes(criteriaSuggestions []sharedtypes.MaterialCriterionWithGuid, availableSearchCriteria []string, traceID string, spanID string) (filtered []sharedtypes.MaterialCriterionWithGuid, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	attributeGuidMap := make(map[string]bool)
	for _, attr := range availableSearchCriteria {
		attributeGuidMap[strings.ToLower(attr)] = true
	}

	var filteredCriteria []sharedtypes.MaterialCriterionWithGuid
	for _, suggestion := range criteriaSuggestions {
		if attributeGuidMap[strings.ToLower(suggestion.AttributeGuid)] {
			filteredCriteria = append(filteredCriteria, suggestion)
		} else {
			logging.Log.Debugf(ctx, "Filtered out non existing attribute: %s", suggestion.AttributeName)
		}
	}

	return filteredCriteria, childSpanID
}

// FilterOutDuplicateAttributes filters out duplicate attributes from the criteria suggestions based on their names
//
// Tags:
//   - @displayName: Filter out duplicate attributes
//
// Parameters:
//   - criteriaSuggestions: current list of criteria suggestions
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - filtered: the list of criteria suggestions excluding duplicates based on attribute names
//   - childSpanID: the child span ID created for this operation
func FilterOutDuplicateAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion, traceID string, spanID string) (filtered []sharedtypes.MaterialLlmCriterion, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	seen := make(map[string]bool)

	for _, suggestion := range criteriaSuggestions {
		lowerAttrName := strings.ToLower(suggestion.AttributeName)
		if !seen[lowerAttrName] {
			seen[lowerAttrName] = true
			filtered = append(filtered, suggestion)
		}
	}

	return filtered, childSpanID
}

// ExtractCriteriaSuggestions extracts criteria suggestions from the LLM response text
//
// Tags:
//   - @displayName: Extract criteria suggestions from LLM response
//
// Parameters:
//   - llmResponse: the text response from the LLM containing JSON with criteria suggestions
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - criteriaSuggestions: the list of criteria suggestions extracted from the LLM response
//   - childSpanID: the child span ID created for this operation
func ExtractCriteriaSuggestions(llmResponse string, traceID string, spanID string) (criteriaSuggestions []sharedtypes.MaterialLlmCriterion, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	criteriaText, _ := ExtractJson(llmResponse, traceID, spanID)
	if criteriaText == "" {
		logging.Log.Debugf(ctx, "No valid JSON found in LLM response: %s", llmResponse)
		return nil, childSpanID
	}

	logging.Log.Debugf(ctx, "Attempting to parse JSON:\n%s", criteriaText)

	var criteria LlmCriteria
	err := json.Unmarshal([]byte(criteriaText), &criteria)
	if err != nil {
		logging.Log.Debugf(ctx, "Failed to deserialize criteria JSON from LLM response: %v; Raw JSON: %s", err, criteriaText)
		return nil, childSpanID
	}

	if len(criteria.Criteria) == 0 {
		logging.Log.Debugf(ctx, "Deserialized JSON successfully but found 0 criteria. Object: %+v", criteria)
	} else {
		logging.Log.Debugf(ctx, "Successfully extracted %d criteria.", len(criteria.Criteria))
	}
	return criteria.Criteria, childSpanID
}

// PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput performs multiple general LLM requests
// using specific models, extracts structured attributes (criteria) from the responses, and returns the total token count
// using the specified OpenAI token counting model. This version does not stream responses.
//
// Tags:
//   - @displayName: Multiple General LLM Requests (Specific Models, No Stream, Attribute Extraction, OpenAI Token Output)
//
// Parameters:
//   - input: the user input string
//   - history: the conversation history for context
//   - systemPrompt: the system prompt to guide the LLM
//   - modelIds: the model IDs of the LLMs to query
//   - tokenCountModelName: the model name used for token count calculation
//   - n: number of parallel requests to perform
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - uniqueCriterion: a deduplicated list of extracted attributes (criteria) from all responses
//   - tokenCount: the total token count (input tokens × n + combined output tokens)
//   - childSpanID: the child span ID created for this operation
func PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput(input string, history []sharedtypes.HistoricMessage, systemPrompt string, modelIds []string, tokenCountModelName string, n int, traceID string, spanID string) (uniqueCriterion []sharedtypes.MaterialLlmCriterion, tokenCount int, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Helper function to send a request and get the response as string
	sendRequest := func() string {
		responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil)
		defer close(responseChannel)

		var responseStr string
		for response := range responseChannel {
			if response.Type == "error" {
				panic(response.Error)
			}
			responseStr += *(response.ChatData)
			if *(response.IsLast) {
				break
			}
		}
		return responseStr
	}

	logging.Log.Debugf(ctx, "System prompt: %s", systemPrompt)

	// Collect all responses with child span for parallel execution
	allResponses := runRequestsInParallel(n, sendRequest, traceID, childSpanID)

	// Extract criteria from all responses with child span
	var allCriteria []sharedtypes.MaterialLlmCriterion
	for _, response := range allResponses {
		criteria, _ := ExtractCriteriaSuggestions(response, traceID, childSpanID)
		if criteria != nil {
			allCriteria = append(allCriteria, criteria...)
		}
	}

	// get input token count
	inputTokenCount, _ := getTokenCount(tokenCountModelName, input, traceID, childSpanID)
	promptTokenCount, _ := getTokenCount(tokenCountModelName, systemPrompt, traceID, childSpanID)

	// get the output token count
	combinedResponseText := strings.Join(allResponses, "\n")
	outputTokenCount, _ := getTokenCount(tokenCountModelName, combinedResponseText, traceID, childSpanID)

	var totalTokenCount = (promptTokenCount+inputTokenCount)*n + outputTokenCount
	logging.Log.Debugf(ctx, "Output token count: %d", outputTokenCount)
	logging.Log.Debugf(ctx, "Total token count: %d", totalTokenCount)

	if len(allCriteria) == 0 {
		logging.Log.Debugf(ctx, "No valid criteria found in any response")
		return []sharedtypes.MaterialLlmCriterion{}, outputTokenCount, childSpanID
	}

	// Only return unique duplicates
	uniqueCriterion, _ = FilterOutDuplicateAttributes(allCriteria, traceID, childSpanID)

	return uniqueCriterion, totalTokenCount, childSpanID
}

func runRequestsInParallel(n int, sendRequest func() string, traceID string, spanID string) []string {
	ctx := &logging.ContextMap{}
	_ = CreateChildSpan(ctx, traceID, spanID)
	responseChan := make(chan string, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logging.Log.Errorf(ctx, "Recovered from panic in LLM request: %v", r)
				}
			}()
			response := sendRequest()
			responseChan <- response
		}()
	}

	go func() {
		wg.Wait()
		close(responseChan)
	}()

	var allResponses []string
	for response := range responseChan {
		logging.Log.Debugf(ctx, "Raw LLM response: %s", response)
		allResponses = append(allResponses, response)
	}
	return allResponses
}

// getTokenCount gets the token count for the given text using the specified model
//
// Parameters:
//   - modelName: the model name used for token count calculation
//   - text: the text to count tokens for
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - count: the token count
//   - childSpanID: the child span ID created for this operation
func getTokenCount(modelName, text string, traceID string, spanID string) (count int, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	logging.Log.Debugf(ctx, "Getting token count for model: %s", modelName)

	tokenCount, err := openAiTokenCount(modelName, text)
	if err != nil {
		logging.Log.Errorf(ctx, "Error getting token count: %v", err)
		errorMessage := fmt.Sprintf("Error getting output token count: %v", err)
		panic(errorMessage)
	}

	logging.Log.Debugf(ctx, "Token count: %d", tokenCount)
	return tokenCount, childSpanID
}

func ExtractJson(text string, traceID string, spanID string) (json string, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	re := regexp.MustCompile("{[\\s\\S]*}")
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 1 {
		return strings.TrimSpace(matches[0]), childSpanID
	}

	logging.Log.Debugf(ctx, "No valid JSON found in response %s", text)
	return "", childSpanID
}

// LogRequestSuccess writes a .Info log entry indicating that a request was completed successfully.
//
// Tags:
//   - @displayName: Log request success
//
// Parameters:
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - childSpanID: the child span ID created for this operation
func LogRequestSuccess(traceID string, spanID string) (childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	logging.Log.Infof(ctx, "Request successful")
	return childSpanID
}

// LogRequestFailed writes a .Info log entry indicating that a request was not completed successfully.
//
// Tags:
//   - @displayName: Log request failed
//
// Parameters:
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - childSpanID: the child span ID created for this operation
func LogRequestFailed(traceID string, spanID string) (childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	logging.Log.Infof(ctx, "Request failed")
	return childSpanID
}

// LogRequestFailedDebugWithMessage writes a .Debug log entry indicating that a request was not completed successfully with additional message.
//
// Tags:
//   - @displayName: Log request failed with message
//
// Parameters:
//   - msg1: the first part of the debug message
//   - msg2: the second part of the debug message
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - childSpanID: the child span ID created for this operation
func LogRequestFailedDebugWithMessage(msg1, msg2 string, traceID string, spanID string) (childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	logging.Log.Debugf(ctx, "Request failed:%s %s", msg1, msg2)
	return childSpanID
}

// CheckApiKeyAuthKvDb checks if the provided API key is authenticated against the KVDB.
//
// Tags:
//   - @displayName: Verify API Key
//
// Parameters:
//   - kvdbEndpoint: the KVDB endpoint
//   - apiKey: The API key to check
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - isAuthenticated: true if the API key is authenticated, false otherwise
//   - childSpanID: the child span ID created for this operation
func CheckApiKeyAuthKvDb(kvdbEndpoint string, apiKey string, traceID string, spanID string) (isAuthenticated bool, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	// Check if the API key is empty
	if apiKey == "" {
		logging.Log.Warnf(ctx, "API key is empty")
		return false, childSpanID
	}

	// Check if the API key exists in the KVDB
	jsonString, exists, err := kvdbGetEntry(kvdbEndpoint, apiKey)
	if err != nil {
		logging.Log.Errorf(ctx, "Error in getting API key from KVDB: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Warnf(ctx, "API key does not exist in KVDB: %s", apiKey)
		return false, childSpanID
	}

	// Unmarshal the JSON string into materials customer object
	var customer materialsCustomerObject
	err = json.Unmarshal([]byte(jsonString), &customer)
	if err != nil {
		logging.Log.Errorf(ctx, "Error unmarshalling JSON string: %v", err)
		panic(err)
	}

	// Check if customer is denied access
	if customer.AccessDenied {
		logging.Log.Warnf(ctx, "Access denied for customer: %s", customer.CustomerName)
		return false, childSpanID
	}

	return true, childSpanID
}

// UpdateTotalTokenCountForCustomerKvDb updates the total token count for a customer in the KVDB
//
// Tags:
//   - @displayName: Update Customer Token Count
//
// Parameters:
//   - kvdbEndpoint: the KVDB endpoint
//   - apiKey: The API key of the customer
//   - additionalTokenCount: The number of tokens to add to the customer's total token count
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - tokenLimitReached: true if the new total token count exceeds the customer's token limit, false otherwise
//   - childSpanID: the child span ID created for this operation
func UpdateTotalTokenCountForCustomerKvDb(kvdbEndpoint string, apiKey string, additionalTokenCount int, traceID string, spanID string) (tokenLimitReached bool, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	// Check if the API key is empty
	if apiKey == "" {
		logging.Log.Errorf(ctx, "API key is empty")
		panic("API key is empty")
	}

	// Get the current token count for the customer
	jsonString, exists, err := kvdbGetEntry(kvdbEndpoint, apiKey)
	if err != nil {
		logging.Log.Errorf(ctx, "Error getting customer object: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Errorf(ctx, "API key does not exist in KVDB: %s", apiKey)
		panic("API key does not exist in KVDB")
	}

	// Unmarshal the JSON string into materials customer object
	var customer materialsCustomerObject
	err = json.Unmarshal([]byte(jsonString), &customer)
	if err != nil {
		logging.Log.Errorf(ctx, "Error unmarshalling JSON string: %v", err)
		panic(err)
	}

	// Get new total token count
	customer.TotalTokenCount = customer.TotalTokenCount + additionalTokenCount

	// create json string from customer object
	newJsonString, err := json.Marshal(customer)
	if err != nil {
		logging.Log.Errorf(ctx, "Error marshalling updated customer object: %v", err)
		panic(err)
	}

	// Update the KVDB with the new JSON string
	err = kvdbSetEntry(kvdbEndpoint, apiKey, string(newJsonString))
	if err != nil {
		logging.Log.Errorf(ctx, "Error updating customer token count in KVDB: %v", err)
		panic(err)
	}

	// Check if the new token count exceeds the limit
	if customer.TotalTokenCount > customer.TokenLimit {
		return true, childSpanID
	}

	return false, childSpanID
}

// DenyCustomerAccessAndSendWarningKvDb denies access to a customer and sends a warning if not already sent
//
// Tags:
//   - @displayName: Deny Customer Access and Send Warning
//
// Parameters:
//   - kvdbEndpoint: the KVDB endpoint
//   - apiKey: The API key of the customer
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - customerName: The name of the customer
//   - sendWarning: true if a warning was sent, false if it was already sent
//   - childSpanID: the child span ID created for this operation
func DenyCustomerAccessAndSendWarningKvDb(kvdbEndpoint string, apiKey string, traceID string, spanID string) (customerName string, sendWarning bool, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	// Check if the API key is empty
	if apiKey == "" {
		logging.Log.Errorf(ctx, "API key is empty")
		panic("API key is empty")
	}

	// Get the current customer object from KVDB
	jsonString, exists, err := kvdbGetEntry(kvdbEndpoint, apiKey)
	if err != nil {
		logging.Log.Errorf(ctx, "Error getting customer object: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Errorf(ctx, "API key does not exist in KVDB: %s", apiKey)
		panic("API key does not exist in KVDB")
	}

	// Unmarshal the JSON string into materials customer object
	var customer materialsCustomerObject
	err = json.Unmarshal([]byte(jsonString), &customer)
	if err != nil {
		logging.Log.Errorf(ctx, "Error unmarshalling JSON string: %v", err)
		panic(err)
	}

	// Check if warning has already been sent
	if !customer.WarningSent {
		sendWarning = true
		customer.WarningSent = true
	}

	// Deny access by setting the accessDenied flag to true
	customer.AccessDenied = true

	// create json string from customer object
	newJsonString, err := json.Marshal(customer)
	if err != nil {
		logging.Log.Errorf(ctx, "Error marshalling updated customer object: %v", err)
		panic(err)
	}

	// Update the KVDB with the new JSON string
	err = kvdbSetEntry(kvdbEndpoint, apiKey, string(newJsonString))
	if err != nil {
		logging.Log.Errorf(ctx, "Error updating customer access in KVDB: %v", err)
		panic(err)
	}

	return customer.CustomerName, sendWarning, childSpanID
}

// ExtractDesignRequirementsAndSearchCriteria parses the user input JSON and returns the design requirements string
// and the list of available search criteria GUIDs.
//
// Tags:
//   - @displayName: Extract Design Requirements and Search Criteria
//
// Parameters:
//   - userInput: the user input JSON string
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - designRequirements: the extracted design requirements string
//   - availableSearchCriteria: the extracted list of attribute GUIDs
//   - childSpanID: the child span ID created for this operation
func ExtractDesignRequirementsAndSearchCriteria(userInput string, traceID string, spanID string) (designRequirements string, availableSearchCriteria []string, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	type promptInput struct {
		UserDesignRequirements  string   `json:"userDesignRequirements"`
		AvailableSearchCriteria []string `json:"availableSearchCriteria"`
	}

	var input promptInput
	if err := json.Unmarshal([]byte(userInput), &input); err != nil {
		logging.Log.Debugf(ctx, "Failed to parse user input: %v", err)
		panic("failed to parse user input: " + err.Error())
	}

	logging.Log.Debugf(ctx, "Successfully extracted design requirements and %d search criteria", len(input.AvailableSearchCriteria))
	return input.UserDesignRequirements, input.AvailableSearchCriteria, childSpanID
}

// AddAvailableAttributesToSystemPrompt adds available attributes to the system prompt template.
//
// Tags:
//   - @displayName: Add Available Attributes to System Prompt
//
// Parameters:
//   - userDesignRequirements: design requirements provided by the user
//   - systemPromptTemplate: the prompt template string to modify
//   - allAvailableAttributes: the list of all available attributes
//   - availableSearchCriteria: the list of available search criteria (GUIDs)
//   - traceID: the trace ID in decimal format
//   - spanID: the span ID in decimal format
//
// Returns:
//   - fullSystemPrompt: the full system prompt to send to the LLM, including available attributes
//   - childSpanID: the child span ID created for this operation
func AddAvailableAttributesToSystemPrompt(userDesignRequirements string, systemPromptTemplate string, allAvailableAttributes []sharedtypes.MaterialAttribute, availableSearchCriteria []string, traceID string, spanID string) (fullSystemPrompt string, childSpanID string) {
	ctx := &logging.ContextMap{}
	childSpanID = CreateChildSpan(ctx, traceID, spanID)

	// 1) Filter allAvailableAttributes using availableSearchCriteria (GUIDs)
	guidSet := make(map[string]struct{}, len(availableSearchCriteria))
	for _, guid := range availableSearchCriteria {
		guidSet[guid] = struct{}{}
	}
	var filteredAttributes []sharedtypes.MaterialAttribute
	for _, attr := range allAvailableAttributes {
		if _, ok := guidSet[attr.Guid]; ok {
			filteredAttributes = append(filteredAttributes, attr)
		}
	}

	logging.Log.Debugf(ctx, "Filtered %d attributes from %d total attributes using %d search criteria",
		len(filteredAttributes), len(allAvailableAttributes), len(availableSearchCriteria))

	// 2) Extract names and create newline-separated list
	var attributeNames []string
	for _, attr := range filteredAttributes {
		attributeNames = append(attributeNames, attr.Name)
	}
	attributesList := strings.Join(attributeNames, "\n")

	// 3) Replace ***ATTRIBUTES*** with this serialized attributes JSON
	fullSystemPrompt = strings.Replace(systemPromptTemplate, "***ATTRIBUTES***", attributesList, 1)

	logging.Log.Debugf(ctx, "Successfully created system prompt with %d attributes", len(filteredAttributes))
	return fullSystemPrompt, childSpanID
}
