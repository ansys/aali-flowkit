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
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

type Response struct {
	Criteria []sharedtypes.MaterialCriterionWithGuid
	Tokens   int
}

type LlmCriteria struct {
	Criteria []sharedtypes.MaterialLlmCriterion
}

// SerializeResponse formats the criteria to a response suitable for the UI clients in string format
//
// Tags:
//   - @displayName: Serialize response for clients
//
// Parameters:
//   - criteriaSuggestions: the list of criteria with their identities
//   - tokens: tokens consumed by the request
//
// Returns:
//   - result: string representation of the response in JSON format
func SerializeResponse(criteriaSuggestions []sharedtypes.MaterialCriterionWithGuid, tokens int) (result string) {
	response := Response{Criteria: criteriaSuggestions, Tokens: tokens}

	responseJson, err := json.Marshal(response)
	if err != nil {
		logging.Log.Debugf(&logging.ContextMap{}, "Failed to serialize suggested criteria into json: %v", err)
		panic("Failed to serialize suggested criteria into json")
	}

	return string(responseJson)
}

// AddGuidsToAttributes adds GUIDs to the attributes in the criteria
//
// Tags:
//   - @displayName: Add GUIDs to criteria suggestions
//
// Parameters:
//   - criteriaSuggestions: the list of criteria without identities
//   - availableAttributes: the list of available attributes with their identities
//
// Returns:
//   - criteriaWithGuids: the list of criteria with their identities
func AddGuidsToAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion, availableAttributes []sharedtypes.MaterialAttribute) (criteriaWithGuids []sharedtypes.MaterialCriterionWithGuid) {
	attributeMap := make(map[string]string)
	for _, attr := range availableAttributes {
		attributeMap[strings.ToLower(attr.Name)] = attr.Guid
	}

	for _, criterion := range criteriaSuggestions {
		lowerAttrName := strings.ToLower(criterion.AttributeName)
		guid, exists := attributeMap[lowerAttrName]

		if !exists {
			logging.Log.Debugf(&logging.ContextMap{}, "Could not find attribute to match: %s", lowerAttrName)
			panic("Could not find attribute to match")
		}

		criteriaWithGuids = append(criteriaWithGuids, sharedtypes.MaterialCriterionWithGuid{
			AttributeName: criterion.AttributeName,
			AttributeGuid: guid,
			Explanation:   criterion.Explanation,
			Confidence:    criterion.Confidence,
		})
	}

	return criteriaWithGuids
}

// FilterOutNonExistingAttributes filters out criteria suggestions that do not match any of the available attributes based on their names
//
// Tags:
//   - @displayName: Filter out non-existing attributes
//
// Parameters:
//   - criteriaSuggestions: current list of criteria suggestions
//   - availableAttributes: the list of available attributes
//
// Returns:
//   - filtered: the list of criteria suggestions excluding those that do not match any of the available attributes
func FilterOutNonExistingAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion, availableAttributes []sharedtypes.MaterialAttribute) (filtered []sharedtypes.MaterialLlmCriterion) {
	attributeMap := make(map[string]bool)
	for _, attr := range availableAttributes {
		attributeMap[strings.ToLower(attr.Name)] = true
	}

	var filteredCriteria []sharedtypes.MaterialLlmCriterion
	for _, suggestion := range criteriaSuggestions {
		if attributeMap[strings.ToLower(suggestion.AttributeName)] {
			filteredCriteria = append(filteredCriteria, suggestion)
		} else {
			logging.Log.Debugf(&logging.ContextMap{}, "Filtered out non existing attribute: %s", suggestion.AttributeName)
		}
	}

	return filteredCriteria
}

// FilterOutDuplicateAttributes filters out duplicate attributes from the criteria suggestions based on their names
//
// Tags:
//   - @displayName: Filter out duplicate attributes
//
// Parameters:
//   - criteriaSuggestions: current list of criteria suggestions
//
// Returns:
//   - filtered: the list of criteria suggestions excluding duplicates based on attribute names
func FilterOutDuplicateAttributes(criteriaSuggestions []sharedtypes.MaterialLlmCriterion) (filtered []sharedtypes.MaterialLlmCriterion) {
	seen := make(map[string]bool)

	for _, suggestion := range criteriaSuggestions {
		lowerAttrName := strings.ToLower(suggestion.AttributeName)
		if !seen[lowerAttrName] {
			seen[lowerAttrName] = true
			filtered = append(filtered, suggestion)
		}
	}

	return filtered
}

// ExtractCriteriaSuggestions extracts criteria suggestions from the LLM response text
//
// Tags:
//   - @displayName: Extract criteria suggestions from LLM response
//
// Parameters:
//   - llmResponse: the text response from the LLM containing JSON with criteria suggestions
//
// Returns:
//   - criteriaSuggestions: the list of criteria suggestions extracted from the LLM response
func ExtractCriteriaSuggestions(llmResponse string) (criteriaSuggestions []sharedtypes.MaterialLlmCriterion) {
	criteriaText := ExtractJson(llmResponse)
	if criteriaText == "" {
		logging.Log.Debugf(&logging.ContextMap{}, "No valid JSON found in LLM response: %s", llmResponse)
		return nil
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Attempting to parse JSON:\n%s", criteriaText)

	var criteria LlmCriteria
	err := json.Unmarshal([]byte(criteriaText), &criteria)
	if err != nil {
		logging.Log.Debugf(&logging.ContextMap{}, "Failed to deserialize criteria JSON from LLM response: %v; Raw JSON: %s", err, criteriaText)
		return nil
	}

	if len(criteria.Criteria) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "Deserialized JSON successfully but found 0 criteria. Object: %+v", criteria)
	} else {
		logging.Log.Debugf(&logging.ContextMap{}, "Successfully extracted %d criteria.", len(criteria.Criteria))
	}
	return criteria.Criteria
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
//
// Returns:
//   - uniqueCriterion: a deduplicated list of extracted attributes (criteria) from all responses
//   - tokenCount: the total token count (input tokens × n + combined output tokens)
func PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput(input string, history []sharedtypes.HistoricMessage, systemPrompt string, modelIds []string, tokenCountModelName string, n int) (uniqueCriterion []sharedtypes.MaterialLlmCriterion, tokenCount int) {
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

	logging.Log.Debugf(&logging.ContextMap{}, "System prompt: %s", systemPrompt)

	// Collect all responses
	allResponses := runRequestsInParallel(n, sendRequest)

	var allCriteria []sharedtypes.MaterialLlmCriterion
	for _, response := range allResponses {
		criteria := ExtractCriteriaSuggestions(response)
		if criteria != nil {
			allCriteria = append(allCriteria, criteria...)
		}
	}

	// get input token count
	inputTokenCount := getTokenCount(tokenCountModelName, input)

	// get the output token count
	var combinedResponseText string
	for _, response := range allResponses {
		combinedResponseText += response
	}
	outputTokenCount := getTokenCount(tokenCountModelName, combinedResponseText)

	var totalTokenCount = inputTokenCount*n + outputTokenCount
	logging.Log.Debugf(&logging.ContextMap{}, "Total token count: %d", totalTokenCount)

	if len(allCriteria) == 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "No valid criteria found in any response")
		return []sharedtypes.MaterialLlmCriterion{}, outputTokenCount
	}

	// Only return unique duplicates
	uniqueCriterion = FilterOutDuplicateAttributes(allCriteria)

	return uniqueCriterion, totalTokenCount
}

func runRequestsInParallel(n int, sendRequest func() string) []string {
	responseChan := make(chan string, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					logging.Log.Errorf(&logging.ContextMap{}, "Recovered from panic in LLM request: %v", r)
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
		logging.Log.Debugf(&logging.ContextMap{}, "Raw LLM response: %s", response)
		allResponses = append(allResponses, response)
	}
	return allResponses
}

func getTokenCount(modelName, text string) int {
	count, err := openAiTokenCount(modelName, text)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting output token count: %v", err)
		panic(errorMessage)
	}
	return count
}


func ExtractJson(text string) (json string) {
	re := regexp.MustCompile("{[\\s\\S]*}")
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 1 {
		return strings.TrimSpace(matches[0])
	}

	logging.Log.Debugf(&logging.ContextMap{}, "No valid JSON found in response %s", text)
	return ""
}

// LogRequestSuccess writes a .Info log entry indicating that a request was completed successfully.
//
// Tags:
//   - @displayName: Log request success
//
// Parameters:
//   - none
//
// Returns:
//   - none
func LogRequestSuccess() {
	logging.Log.Infof(&logging.ContextMap{}, "Request successful")
	return
}

// LogRequestFailed writes a .Info log entry indicating that a request was not completed successfully.
//
// Tags:
//   - @displayName: Log request failed
//
// Parameters:
//   - none
//
// Returns:
//   - none
func LogRequestFailed() {
	logging.Log.Infof(&logging.ContextMap{}, "Request failed")
	return
}

// LogRequestFailedDebugWithMessage writes a .Debug log entry indicating that a request was not completed successfully with additional message.
//
// Tags:
//   - @displayName: Log request failed with message
//
// Parameters:
//   - msg1: the first part of the debug message
//   - msg2: the second part of the debug message
//
// Returns:
//   - none
func LogRequestFailedDebugWithMessage(msg1, msg2 string) {
	logging.Log.Debugf(&logging.ContextMap{}, "Request failed:%s %s", msg1, msg2)
	return
}

// CheckApiKeyAuthKvDb checks if the provided API key is authenticated against the KVDB.
//
// Tags:
//   - @displayName: Verify API Key
//
// Parameters:
//   - apiKey: The API key to check
//
// Returns:
//   - isAuthenticated: true if the API key is authenticated, false otherwise
func CheckApiKeyAuthKvDb(kvdbEndpoint string, apiKey string) (isAuthenticated bool) {

	// Check if the API key is empty
	if apiKey == "" {
		logging.Log.Warnf(&logging.ContextMap{}, "API key is empty")
		return false
	}

	// Check if the API key exists in the KVDB
	jsonString, exists, err := kvdbGetEntry(kvdbEndpoint, apiKey)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error in getting API key from KVDB: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Warnf(&logging.ContextMap{}, "API key does not exist in KVDB: %s", apiKey)
		return false
	}

	// Unmarshal the JSON string into materials customer object
	var customer materialsCustomerObject
	err = json.Unmarshal([]byte(jsonString), &customer)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error unmarshalling JSON string: %v", err)
		panic(err)
	}

	// Check if customer is denied access
	if customer.AccessDenied {
		logging.Log.Warnf(&logging.ContextMap{}, "Access denied for customer: %s", customer.CustomerName)
		return false
	}

	return true
}

// UpdateTotalTokenCountForCustomerKvDb updates the total token count for a customer in the KVDB
//
// Tags:
//   - @displayName: Update Customer Token Count
//
// Parameters:
//   - apiKey: The API key of the customer
//   - additionalTokenCount: The number of tokens to add to the customer's total token count
//
// Returns:
//   - tokenLimitReached: true if the new total token count exceeds the customer's token limit, false otherwise
func UpdateTotalTokenCountForCustomerKvDb(kvdbEndpoint string, apiKey string, additionalTokenCount int) (tokenLimitReached bool) {
	// Check if the API key is empty
	if apiKey == "" {
		logging.Log.Errorf(&logging.ContextMap{}, "API key is empty")
		panic("API key is empty")
	}

	// Get the current token count for the customer
	jsonString, exists, err := kvdbGetEntry(kvdbEndpoint, apiKey)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error getting customer object: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Errorf(&logging.ContextMap{}, "API key does not exist in KVDB: %s", apiKey)
		panic("API key does not exist in KVDB")
	}

	// Unmarshal the JSON string into materials customer object
	var customer materialsCustomerObject
	err = json.Unmarshal([]byte(jsonString), &customer)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error unmarshalling JSON string: %v", err)
		panic(err)
	}

	// Get new total token count
	customer.TotalTokenCount = customer.TotalTokenCount + additionalTokenCount

	// create json string from customer object
	newJsonString, err := json.Marshal(customer)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error marshalling updated customer object: %v", err)
		panic(err)
	}

	// Update the KVDB with the new JSON string
	err = kvdbSetEntry(kvdbEndpoint, apiKey, string(newJsonString))
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error updating customer token count in KVDB: %v", err)
		panic(err)
	}

	// Check if the new token count exceeds the limit
	if customer.TotalTokenCount > customer.TokenLimit {
		return true
	}

	return false
}

// DenyCustomerAccessAndSendWarningKvDb denies access to a customer and sends a warning if not already sent
//
// Tags:
//   - @displayName: Deny Customer Access and Send Warning
//
// Parameters:
//   - apiKey: The API key of the customer
//
// Returns:
//   - customerName: The name of the customer
//   - sendWarning: true if a warning was sent, false if it was already sent
func DenyCustomerAccessAndSendWarningKvDb(kvdbEndpoint string, apiKey string) (customerName string, sendWarning bool) {
	// Check if the API key is empty
	if apiKey == "" {
		logging.Log.Errorf(&logging.ContextMap{}, "API key is empty")
		panic("API key is empty")
	}

	// Get the current customer object from KVDB
	jsonString, exists, err := kvdbGetEntry(kvdbEndpoint, apiKey)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error getting customer object: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Errorf(&logging.ContextMap{}, "API key does not exist in KVDB: %s", apiKey)
		panic("API key does not exist in KVDB")
	}

	// Unmarshal the JSON string into materials customer object
	var customer materialsCustomerObject
	err = json.Unmarshal([]byte(jsonString), &customer)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error unmarshalling JSON string: %v", err)
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
		logging.Log.Errorf(&logging.ContextMap{}, "Error marshalling updated customer object: %v", err)
		panic(err)
	}

	// Update the KVDB with the new JSON string
	err = kvdbSetEntry(kvdbEndpoint, apiKey, string(newJsonString))
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error updating customer access in KVDB: %v", err)
		panic(err)
	}

	return customer.CustomerName, sendWarning
}

// ExtractDesignRequirementsAndSearchCriteria parses the user input JSON and returns the design requirements string
// and the list of available search criteria GUIDs.
//
// Tags:
//   - @displayName: Extract Design Requirements and Search Criteria
//
// Parameters:
//   - userInput: the user input JSON string
//
// Returns:
//   - designRequirements: the extracted design requirements string
//   - availableSearchCriteria: the extracted list of attribute GUIDs
func ExtractDesignRequirementsAndSearchCriteria(userInput string) (designRequirements string, availableSearchCriteria []string) {    
	type promptInput struct {
        UserDesignRequirements      string   `json:"userDesignRequirements"`
        AvailableSearchCriteria []string `json:"availableSearchCriteria"`
    }

    var input promptInput
    if err := json.Unmarshal([]byte(userInput), &input); err != nil {
        panic("failed to parse user input: " + err.Error())
    }

    return input.UserDesignRequirements, input.AvailableSearchCriteria
}

// AddAvailableAttributesToSystemPrompt adds available attributes to the system prompt template.
//
// Tags:
//   - @displayName: Add Available Attributes to System Prompt
//
// Parameters:
//   - userDesignRequirements: design requirements provided by the user
//   - availableSearchCriteria: the list of available search criteria (GUIDs)
//   - availableAttributes: the list of all available attributes
//   - systemPromptTemplate: the prompt template string to modify
//
// Returns:
//   - string: the full system prompt to send to the LLM, including available attributes
func AddAvailableAttributesToSystemPrompt(userDesignRequirements string, systemPromptTemplate string, allAvailableAttributes []sharedtypes.MaterialAttribute, availableSearchCriteria []string) string {
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

    // 2) Convert filtered attributes to serialized JSON
    attributesJson, err := json.MarshalIndent(filteredAttributes, "", "  ")
    if err != nil {
        panic("failed to serialize available attributes: " + err.Error())
    }

    // 3) Replace ***ATTRIBUTES*** with this serialized attributes JSON
    fullSystemPrompt := strings.Replace(systemPromptTemplate, "***ATTRIBUTES***", string(attributesJson), 1)

	logging.Log.Debugf(&logging.ContextMap{}, "Full system prompt with attributes: %s", fullSystemPrompt)

	return fullSystemPrompt
}
