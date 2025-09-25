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
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/russross/blackfriday/v2"

	"github.com/ansys/aali-flowkit/pkg/meshpilot/ampgraphdb"

	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/qdrant/go-client/qdrant"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

/*************************************************************************/
/* 						Private Functions 								 */
/*************************************************************************/
func cleanJSONBlock(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") && strings.HasSuffix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```JSON")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	// Sometimes models wrap whole JSON in quotes
	s = strings.Trim(s, "\"")
	return s
}

// updateActionsMatching applies key/value updates to every action where matchKey == matchValue.
func updateActionsMatching(actions []map[string]string, matchKey, matchValue string, updates map[string]string) {
	for i := range actions {
		if val, ok := actions[i][matchKey]; ok && val == matchValue {
			for k, v := range updates {
				actions[i][k] = v
			}
		}
	}
}

/*************************************************************************/
/* 						Extract Info from LLM Output 					 */
/*************************************************************************/

// ExtractMapFieldValueFromJSON finds the relevant description by prompting
//
// Tags:
//   - @displayName: ExtractMapFieldValueFromJSON
//
// Parameters:
//   - message: the message from llm
//   - extractField: the field to extract from the llm output
//
// Returns:
//   - fieldValue: the extracted field value
func ExtractMapFieldValueFromJSON(message, extractField string) (fieldValue interface{}) {
	ctx := &logging.ContextMap{}
	parsed := ParseMapFromJSON(message)

	fieldValue, ok := parsed[extractField]
	if !ok {
		errMsg := fmt.Sprintf("ExtractMapFieldValueFromJSON: %s not found in LLM output", extractField)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}
	return
}

// ExtractMapFieldStringValueFromJSON finds the relevant description by prompting
//
// Tags:
//   - @displayName: ExtractMapFieldStringValueFromJSON
//
// Parameters:
//   - message: the message from llm
//   - extractField: the field to extract from the llm output
//
// Returns:
//   - fieldValue: the extracted field value
func ExtractMapFieldStringValueFromJSON(message, extractField string) (fieldValue string) {
	ctx := &logging.ContextMap{}
	parsed := ParseMapFromJSON(message)

	fieldValue, ok := parsed[extractField].(string)
	if !ok {
		errMsg := fmt.Sprintf("ExtractMapFieldStringValueFromJSON: %s not found in LLM output", extractField)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	return
}

// ExtractPropertiesFieldsFromJSON finds the relevant description by prompting
//
// Tags:
//   - @displayName: ExtractPropertiesFromJSON
//
// Parameters:
//   - message: the message from llm
//   - extractField: the field to extract from the llm output
//
// Returns:
//   - propertyDetails: the extracted field value
func ExtractPropertiesFromJSON(message string) (propertyDetails []map[string]string) {
	ctx := &logging.ContextMap{}
	parsed := ParseMapFromJSON(message)

	propertyDetails = []map[string]string{}
	for key, value := range parsed {
		propertyData := map[string]string{}
		propertyData["PropertyName"] = key
		propertyData["PropertyUnits"] = ""
		propertyData["PropertyValue"] = ""
		switch v := value.(type) {
		case string:
		case int, int64, int32:
			propertyData["PropertyValue"] = fmt.Sprintf("%d", v)
		case float32, float64:
			propertyData["PropertyValue"] = fmt.Sprintf("%f", v)
		case map[string]interface{}:
			if propertyValue, ok := v["value"].(string); !ok {
				if intVal, ok := v["value"].(int); ok {
					propertyData["PropertyValue"] = fmt.Sprintf("%d", intVal)
				} else if floatVal, ok := v["value"].(float64); ok {
					propertyData["PropertyValue"] = fmt.Sprintf("%f", floatVal)
				} else {
					propertyData["PropertyValue"] = ""
					logging.Log.Infof(ctx, "Key: %s, Value 'value' is of a different type or missing: %T", key, v["value"])
				}
			} else {
				propertyData["PropertyValue"] = propertyValue
			}

			if propertyUnits, ok := v["units"].(string); !ok {
				propertyData["PropertyUnits"] = ""
				logging.Log.Infof(ctx, "Key: %s, Value 'units' is of a different type or missing: %T", key, v["units"])
			} else {
				propertyData["PropertyUnits"] = propertyUnits
			}
			propertyDetails = append(propertyDetails, propertyData)
		default:
			logging.Log.Infof(ctx, "Key: %s, Value is of a different type: %T", key, v)
			continue
		}
	}
	return
}

// FindRelevantDescription finds the relevant description by prompting
//
// Tags:
//   - @displayName: FindRelevantDescription
//
// Parameters:
//   - descriptions: the list of descriptions
//   - message: the message from llm
//
// Returns:
//   - relevantDescription: the relevant desctiption
func FindRelevantDescription(descriptions []string, message string) (relevantDescription string) {

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
	cleanedContent := cleanJSONBlock(message)

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

/*************************************************************************/
/* 						Update Action Fields 							 */
/*************************************************************************/
func UpdateActionField(actions []map[string]string, matchKey, matchValue, updateKey, updateValue string) []map[string]string {
	ctx := &logging.ContextMap{}

	if len(actions) == 0 {
		logging.Log.Warn(ctx, "UpdateAction: no actions to update")
		return actions
	}

	updateActionsMatching(actions, matchKey, matchValue, map[string]string{
		updateKey: updateValue,
	})

	return actions
}

func UpdateActionFields(actions []map[string]string, matchKey, matchValue string, updates map[string]string) []map[string]string {
	ctx := &logging.ContextMap{}

	if len(actions) == 0 {
		logging.Log.Warn(ctx, "UpdateAction: no actions to update")
		return actions
	}

	updateActionsMatching(actions, matchKey, matchValue, updates)

	return actions
}

func UpdateActionsWithProperties(actions []map[string]string, properties []map[string]string, identifierKey string, valueKey string, unitsKey string) []map[string]string {
	ctx := &logging.ContextMap{}

	if len(actions) == 0 {
		logging.Log.Warn(ctx, "UpdateActionsWithProperties: no actions to update")
		return actions
	}

	if len(properties) == 0 {
		logging.Log.Warn(ctx, "UpdateActionsWithProperties: no properties to apply")
		return actions
	}

	for _, prop := range properties {
		propName := prop["PropertyName"]
		propValue := prop["PropertyValue"]
		propUnits := prop["PropertyUnits"]

		for i := range actions {
			if actions[i][identifierKey] == propName {
				if propUnits != "" {
					actions[i][unitsKey] = propUnits
				}
				actions[i][valueKey] = propValue
			}
		}
	}

	return actions
}

/*************************************************************************/
/* 						Generate Custom Action 						     */
/*************************************************************************/

// GenerateAction generates special action based on user inputs
//
// Tags:
//   - @displayName: GenerateAction
//
// Parameters:
//   - keys: keys to create customize action
//   - values: values to create customize action
//   - message: message to send to the client
//
// Returns:
//   - result: the actions in json format
func GenerateAction(keys []string, values []string, message string) (actions []map[string]string) {
	ctx := &logging.ContextMap{}

	logging.Log.Info(ctx, "Generate Special Action...")

	if len(keys) != len(values) {
		errMsg := fmt.Sprintf("keys and values length mismatch: %d vs %d", len(keys), len(values))
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	// Validate keys to contain only letters
	validKey := regexp.MustCompile(`^[a-zA-Z]+$`)
	for _, key := range keys {
		if !validKey.MatchString(key) {
			errMsg := fmt.Sprintf("invalid key format: %s", key)
			logging.Log.Error(ctx, errMsg)
			panic(errMsg)
		}
	}

	action := make(map[string]string)
	for i, key := range keys {
		action[key] = values[i]
	}

	actions = []map[string]string{action}
	return
}

/*************************************************************************/
/* 				 	Vector Database Related Functions 					 */
/*************************************************************************/

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

// PerformSimilaritySearchForSubqueries performs similarity search for each sub-query and returns Q&A pairs
//
// Tags:
//   - @displayName: PerformSimilaritySearchForSubqueries
//
// Parameters:
//   - subQueries: the list of expanded sub-queries
//   - collection: the vector database collection name
//   - similaritySearchResults: the number of similarity search results
//   - similaritySearchMinScore: the minimum similarity score threshold
//
// Returns:
//   - uniqueQAPairs: the unique Q&A pairs from similarity search results
func PerformSimilaritySearchForSubqueries(subQueries []string, collection string, similaritySearchResults int, similaritySearchMinScore float64) (uniqueQAPairs []map[string]interface{}) {
	ctx := &logging.ContextMap{}
	uniqueQAPairs = []map[string]interface{}{}
	uniqueQuestions := make(map[string]bool)

	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logging.Log.Error(ctx, fmt.Sprintf("unable to create qdrant client: %v", err))
		return
	}

	for _, subQuery := range subQueries {
		logging.Log.Debugf(ctx, "Processing sub-query: %s", subQuery)
		embeddedVector, _ := PerformVectorEmbeddingRequest(subQuery, false)
		if len(embeddedVector) == 0 {
			logging.Log.Warnf(ctx, "Failed to get embedding for sub-query: %s", subQuery)
			continue
		}

		limit := uint64(similaritySearchResults)
		scoreThreshold := float32(similaritySearchMinScore)
		query := qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryDense(embeddedVector),
			Limit:          &limit,
			ScoreThreshold: &scoreThreshold,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
		}

		scoredPoints, err := client.Query(context.TODO(), &query)
		if err != nil {
			logging.Log.Warnf(ctx, "Qdrant query failed: %v", err)
			continue
		}

		for _, scoredPoint := range scoredPoints {
			payload, err := qdrant_utils.QdrantPayloadToType[map[string]interface{}](scoredPoint.GetPayload())
			if err != nil {
				logging.Log.Warnf(ctx, "Failed to parse payload: %v", err)
				continue
			}
			question, _ := payload["question"].(string)
			answer, _ := payload["answer"].(string)
			if question == "" {
				continue
			}
			if !uniqueQuestions[question] {
				qaPair := map[string]interface{}{
					"question": question,
					"answer":   answer,
				}
				uniqueQAPairs = append(uniqueQAPairs, qaPair)
				uniqueQuestions[question] = true
			}
		}
	}
	for i, qa := range uniqueQAPairs {
		logging.Log.Debugf(ctx, "Unique QA Pair #%d: Question: %s, Answer: %s", i+1, qa["question"], qa["answer"])
	}
	logging.Log.Infof(ctx, "Simple similarity search complete. Found %d unique Q&A pairs from %d sub-queries", len(uniqueQAPairs), len(subQueries))
	return uniqueQAPairs
}

/*************************************************************************/
/* 				 	Graph Database Related Functions 					 */
/*************************************************************************/

// FetchPropertiesFromPathDescription get properties from path description
//
// Tags:
//   - @displayName: FetchPropertiesFromPathDescription
//
// Parameters:
//   - db_name: the graph database name
//   - description: the desctiption of path
//   - query: the cypher query to get properties from description
//
// Returns:
//   - properties: the list of descriptions
func FetchPropertiesFromPathDescription(db_name, description, query string) (properties []string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Properties From Path Descriptions...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

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
//   - db_name: the graph database name
//   - description: the desctiption of path
//   - query: the cypher query to get node descriptions from description
//
// Returns:
//   - actionDescriptions: action descriptions
func FetchNodeDescriptionsFromPathDescription(db_name, description, query string) (actionDescriptions string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Node Descriptions From Path Descriptions...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

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
//   - db_name: the graph database name
//   - description: the desctiption of path
//   - query: the cypher query to get actions from description
//
// Returns:
//   - actions: the list of actions to execute
func FetchActionsPathFromPathDescription(db_name, description, query string) (actions []map[string]string) {
	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Fetching Actions From Path Descriptions...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	actions, err = ampgraphdb.GraphDbDriver.GetActions(description, query)
	if err != nil {
		errorMessage := fmt.Sprintf("Error fetching actions from path description: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

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
func GetSolutionsToFixProblem(db_name, fmFailureCode, primeMeshFailureCode, query string) (solutions string) {

	ctx := &logging.ContextMap{}

	logging.Log.Infof(ctx, "Get Solutions To Fix Problem...")

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, db_name)

	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
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

// GenerateMKSummariesforTags retrieves unique MK summaries for the provided tags from the graph database.
//
// Tags:
//   - @displayName: GenerateMKSummariesforTags
//
// Parameters:
//   - query: the user query
//   - dbName: the name of the database
//   - tags: the list of tags
//
// Returns:
//   - allTagsSummaries: the list of unique MK summaries
func GenerateMKSummariesforTags(dbName string, tags []string, GetTagIdByNameQuery string, GetMKSummaryFromDBQuery string) (allTagsSummaries []string) {
	ctx := &logging.ContextMap{}

	err := ampgraphdb.EstablishConnection(config.GlobalConfig.GRAPHDB_ADDRESS, dbName)
	if err != nil {
		errMsg := fmt.Sprintf("error initializing graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	uniqueSummaries := make(map[string]bool)
	for _, tag := range tags {
		// Inline GetTagIdByName
		id, err := ampgraphdb.GraphDbDriver.GetTagIdByName(tag, GetTagIdByNameQuery)
		if err != nil {
			logging.Log.Warnf(ctx, "No tag_id found for tag %s (error: %v)", tag, err)
			continue
		}
		if id != "" {
			logging.Log.Infof(ctx, "Found tag_id %s for tag %s", id, tag)
			// Inline GetMKSummaryFromDB
			sum, err := ampgraphdb.GraphDbDriver.GetMKSummaryFromDB(id, GetMKSummaryFromDBQuery)
			if err != nil {
				logging.Log.Warnf(ctx, "Error getting MK summary for tag_id %s: %v", id, err)
				continue
			}
			if sum != "" {
				uniqueSummaries[sum] = true
			}
		} else {
			logging.Log.Warnf(ctx, "No tag_id found for tag %s", tag)
		}
	}

	allTagsSummaries = make([]string, 0, len(uniqueSummaries))
	for summary := range uniqueSummaries {
		allTagsSummaries = append(allTagsSummaries, summary)
	}

	logging.Log.Infof(ctx, "Metatag extraction complete. Tags: %v, Summaries found: %d", tags, len(allTagsSummaries))
	return allTagsSummaries
}

/*************************************************************************/
/* 				 	Parse JSON to Object 					 	 		 */
/*************************************************************************/

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

// ParseMapFromJSON update action as per user instruction
//
// Tags:
//   - @displayName: ParseMapFromJSON
//
// Parameters:
//   - message: the message from the llm
//
// Returns:
//   - structuredOutput: the list of synthesized actions
func ParseMapFromJSON(message string) (structuredOutput map[string]interface{}) {
	ctx := &logging.ContextMap{}

	if strings.TrimSpace(message) == "" {
		errMsg := "ParseMapFromJSON: empty message"
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	cleaned := cleanJSONBlock(message)

	if err := json.Unmarshal([]byte(cleaned), &structuredOutput); err != nil {
		errMsg := fmt.Sprintf("ParseMapFromJSON: unmarshal failed: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	return
}

// ProcessJSONListOutput parses the response and returns the tags slice.
//
// Tags:
//   - @displayName: ProcessJSONListOutput
//
// Parameters:
//   - response: the JSON response string
//
// Returns:
//   - tags: the list of items extracted from the response
func ProcessJSONListOutput(response string) (generatedList []string) {
	ctx := &logging.ContextMap{}

	err := json.Unmarshal([]byte(response), &generatedList)
	if err != nil {
		logging.Log.Errorf(ctx, "Error decoding JSON response: %v", err)
		return []string{}
	}
	logging.Log.Debugf(ctx, "Generated List: %s", strings.Join(generatedList, ", "))
	if len(generatedList) == 0 {
		logging.Log.Error(ctx, "No items generated.")
		return nil
	}
	return generatedList
}

/*************************************************************************/
/* 				 	Slash Command Functions 					 		 */
/*************************************************************************/

// ParseSlashCommand retrieves the Slash Input from the input string.
//
// Tags:
//   - @displayName: ParseSlashCommand
//
// Parameters:
//   - userInput: the input string containing the Slash Input message in JSON format
//
// Returns:
//   - slashCmd: the slash command if found, otherwise an empty string
//   - targetCmd: the target command if found, otherwise an empty string
//   - hasCmd: boolean indicating if a slash command or target command was found
func ParseSlashCommand(userInput string) (slashCmd, targetCmd string, hasCmd bool, hasContext bool) {

	targetRe := regexp.MustCompile(`@[A-Za-z][\w]*`)
	slashRe := regexp.MustCompile(`/[A-Za-z][\w]*`)

	target := targetRe.FindString(userInput)
	slash := slashRe.FindString(userInput)

	if target != "" {
		target = target[1:]
	}
	if slash != "" {
		slash = slash[1:]
	}

	switch {
	case slash != "" && target != "":
		hasCmd = true
	case slash != "" && target == "":
		hasCmd = true
	default:
		hasCmd = false
	}

	remaining := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(userInput, "/"+slash, ""), "@"+target, ""))
	hasContext = remaining != ""

	logging.Log.Debugf(&logging.ContextMap{}, "User Command: Slash: %s, Target: %s, Has Command: %t, Has Context: %t, Remaining: %s", slash, target, hasCmd, hasContext, remaining)

	return slash, target, hasCmd, hasContext
}

// SynthesizeSlashCommand synthesize actions based on user instruction
//
// Tags:
//   - @displayName: SynthesizeSlashCommand
//
// Parameters:
//   - slashCmd: the slash command
//   - targetCmd: the target command
//
// Returns:
//   - result: the synthesized string
func SynthesizeSlashCommand(slashCmd, targetCmd, finalizeResult, message, key1, key2, value string) (result string) {
	ctx := &logging.ContextMap{}

	var actions []map[string]string

	if finalizeResult != "" {
		var parsedResult map[string]interface{}
		err := json.Unmarshal([]byte(finalizeResult), &parsedResult)
		if err != nil {
			errorMessage := fmt.Sprintf("failed to unmarshal finalizeResult: %v", err)
			logging.Log.Error(ctx, errorMessage)
			panic(errorMessage)
		}

		if parsedActions, ok := parsedResult["Actions"].([]interface{}); ok {
			for _, action := range parsedActions {
				if actionMap, ok := action.(map[string]interface{}); ok {
					updatedAction := map[string]string{}
					for k, v := range actionMap {
						if strVal, ok := v.(string); ok {
							updatedAction[k] = strVal
						}
					}
					updatedAction[key1] = targetCmd
					updatedAction[key2] = value
					updatedAction["Argument"] = slashCmd
					actions = append(actions, updatedAction)
				}
			}
		}
	} else {
		actions = []map[string]string{
			{
				key1:       targetCmd,
				key2:       value,
				"Argument": slashCmd,
			},
		}
	}

	finalMessage := map[string]interface{}{
		"Message": message,
		"Actions": actions,
	}

	resultStream, err := json.Marshal(finalMessage)
	if err != nil {
		errorMessage := fmt.Sprintf("failed to marshal final message: %v", err)
		logging.Log.Error(ctx, errorMessage)
		panic(errorMessage)
	}

	result = string(resultStream)
	logging.Log.Infof(ctx, "SynthesizeSlashCommand result: %s", result)
	logging.Log.Infof(ctx, "successfully synthesized slash command")

	return result
}

/****************************************************************************/
/*		 				Constructing Prompts 								*/
/****************************************************************************/

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
func GenerateSubWorkflowPrompt(userInstruction, systemPromptTemplate, userPromptTemplate string, subworkflows []map[string]string) (systemPrompt string, userPrompt string) {
	ctx := &logging.ContextMap{}

	// Retrieve subworkflows (name and description)
	var subworkflowListStr strings.Builder
	for i, sw := range subworkflows {
		swName, nameOk := sw["Name"]
		swDesc, descOk := sw["Description"]
		if nameOk && descOk {
			subworkflowListStr.WriteString(fmt.Sprintf("%d. %s - %s\n", i+1, swName, swDesc))
		}
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

// GenerateSynthesizeAnswerfromMetaKnowlwdgeUserPrompt generates a user prompt for synthesizing an answer from meta knowledge.
//
// Tags:
//   - @displayName: GenerateSynthesizeAnswerfromMetaKnowlwdgeUserPrompt
//
// Parameters:
//   - SynthesizeAnswerUserPromptTemplate: the template string with placeholders for original query, expanded sub-queries, and retrieved Q&A pairs
//   - originalQuery: the user's original query
//   - expandedQueries: the expanded sub-queries
//   - retrievedQAPairs: the retrieved Q&A pairs
//
// Returns:
//   - userPrompt: the formatted user prompt
func GenerateSynthesizeAnswerfromMetaKnowlwdgeUserPrompt(SynthesizeAnswerUserPromptTemplate string, originalQuery string, expandedQueries []string, retrievedQAPairs []map[string]interface{}) (userPrompt string) {
	ctx := &logging.ContextMap{}

	expandedQueriesStr := fmt.Sprintf("[%s]", strings.Join(expandedQueries, ", "))
	qaPairsBytes, _ := json.MarshalIndent(retrievedQAPairs, "", "  ")
	qaPairsStr := string(qaPairsBytes)

	userPrompt = fmt.Sprintf(SynthesizeAnswerUserPromptTemplate, originalQuery, expandedQueriesStr, qaPairsStr)
	logging.Log.Debugf(ctx, "Generated Synthesize Answer User Prompt: %s", userPrompt)
	return
}

/*************************************************************************/
/* 						Miscellaneous 									 */
/*************************************************************************/

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

// FinalizeResult converts actions to json string to send back data
//
// Tags:
//   - @displayName: FinalizeResult
//
// Parameters:
//   - actions: the executable actions
//   - message: message to send to the client
//
// Returns:
//   - result: the actions in json format
func FinalizeResult(actions []map[string]string, message string) (result string) {
	ctx := &logging.ContextMap{}

	if actions == nil {
		actions = []map[string]string{}
	}

	finalMessage := map[string]interface{}{
		"Message": message,
		"Actions": actions,
	}

	bytesStream, err := json.Marshal(finalMessage)
	if err != nil {
		panic(fmt.Sprintf("failed to convert actions to json: %v", err))
	}

	result = string(bytesStream)
	logging.Log.Info(ctx, "successfully converted actions to json")
	return
}

// SynthesizeActionsTool3 update action as per user instruction
// // Tags:
//   - @displayName: SynthesizeActionsTool3
//
// Parameters:
//   - message_1: the first message from the llm
//   - message_2: the second message from the llm
//   - actions: the list of actions
//
// Returns:
//   - updatedActions: the list of synthesized actions
func SynthesizeActionsTool3(message_1, message_2, target_object, key1, key2, target1, target2 string, actions []map[string]string) (updatedActions []map[string]string) {
	ctx := &logging.ContextMap{}

	// Clean up the input messages
	message_1 = strings.TrimSpace(strings.Trim(message_1, "\""))
	message_2 = strings.TrimSpace(strings.Trim(message_2, "\""))
	target_object = strings.TrimSpace(strings.Trim(target_object, "\""))

	logging.Log.Debugf(ctx, "Tool 3 Synthesize Message 1: %s\n", message_1)
	logging.Log.Debugf(ctx, "Tool 3 Synthesize Message 2: %s\n", message_2)
	logging.Log.Debugf(ctx, "Tool 3 Target Object: %s\n", target_object)

	// Initialize updatedActions with the input actions
	updatedActions = actions

	if target_object == target1 {
		// Check the first dictionary in actions
		if len(updatedActions) > 0 {
			firstAction := updatedActions[0]
			if _, ok := firstAction[key1]; ok {
				// Replace the value with the input message
				firstAction[key1] = message_1
			}

			if _, ok := firstAction[key2]; ok && len(message_2) != 0 {
				firstAction[key2] = message_2
			}
		}
	} else if target_object == target2 {
		// Check if there is a third dictionary in actions
		if len(updatedActions) > 2 {
			thirdAction := updatedActions[2]
			if _, ok := thirdAction[key1]; ok {
				// Replace the value with the input message
				thirdAction[key1] = message_1
			}

			updatedActions = []map[string]string{thirdAction}
		} else {
			logging.Log.Warnf(ctx, "No third action found in updatedActions for target_object: %s", target_object)
		}
	} else {
		// Skip if target_object is neither APP_TOOL_ACTIONS_TARGET_5 nor APP_TOOL_ACTIONS_TARGET_6
		logging.Log.Infof(ctx, "Skipping action synthesis for target_object: %s", target_object)
	}

	logging.Log.Debugf(ctx, "The Updated Actions: %q\n", updatedActions)

	return
}
