// File: aali-flowkit/pkg/externalfunctions/ace.go
//
// This package provides external functions for the AALI flowkit system.
// Functions in this package follow Go best practices including:
// - Proper error handling without panics
// - Efficient string building using strings.Builder
// - Descriptive function and variable names
// - Early returns to reduce nesting
// - Structured parameters for functions with many arguments
package externalfunctions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

// SearchExamples performs a search in the Example collection name.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Examples
//
// Parameters:
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//   - userQuery: the user query to be used for the query.
//
// Returns:
//   - generatedCode: the generated code as a string
func SearchExamples(ansysProduct string, collectionName string, maxRetrievalCount int, denseWeight float64, sparseWeight float64, userQuery string) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamples STARTED - ansysProduct: %s, collectionName: %s, maxRetrievalCount: %d, userQuery: %s", ansysProduct, collectionName, maxRetrievalCount, userQuery)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamples COMPLETED - duration: %v", duration)
	}()

	userMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		As a first step, you need to search the Examples Vector DB to find any relevant examples. Check if the examples contain enough information to generate the code.
		If you are sure that the examples are enough, return "true". If you need more examples, return "false".

		The format in the following text, do not add anything else (no extra keys, no extra texts, or formatting (including no code fences)):
		true/false`, ansysProduct)

	outputFields := []string{"text", "document_name", "previous_chunk", "next_chunk", "guid"}
	historyMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: userMessage,
		},
	}

	// Time the database query
	dbStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamples - Database query STARTED for query: %s", userQuery)
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, userQuery, denseWeight, sparseWeight, "")
	dbDuration := time.Since(dbStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamples - Database query COMPLETED - duration: %v, results count: %d", dbDuration, len(scoredPoints))

	if len(scoredPoints) == 0 {
		logging.Log.Warnf(&logging.ContextMap{}, "No results found for query: %s", userQuery)
		return ""
	}

	var exampleBuilder strings.Builder
	for _, scoredPoint := range scoredPoints {
		entry := scoredPoint.Payload
		exampleName := entry["document_name"].GetStringValue()
		exampleText := entry["text"].GetStringValue()
		exampleRefs, _ := getExampleReferences(exampleName, "aali") //example_refs_info

		exampleBuilder.WriteString(fmt.Sprintf("Example: {%s}\n{%s}\n\n", exampleName, exampleText))
		exampleBuilder.WriteString(fmt.Sprintf("Example {%s} References: {%s}\n\n", exampleName, exampleRefs))
	}
	exampleString := exampleBuilder.String()

	// Time the LLM request
	llmStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamples - LLM request STARTED")
	result, _ := PerformGeneralRequest(exampleString, historyMessage, false, "")
	llmDuration := time.Since(llmStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamples - LLM request COMPLETED - duration: %v", llmDuration)

	// Convert string result to boolean using strconv.ParseBool
	response, err := strconv.ParseBool(result)
	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting result to boolean: %v", err)
		return ""
	}
	if !response {
		logging.Log.Error(&logging.ContextMap{}, "Example response is false")
		return ""
	}
	return exampleString
}

// SearchMethods performs a search in the Elements Database.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Methods
//
// Parameters:
//   - tableOfContents: the table of contents to be used in the system message
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//   - userQuery: the user query to be used for the query.
//
// Returns:
//   - examplesString: the formatted examples string containing the method examples and references
func SearchMethods(tableOfContents string, ansysProduct string, collectionName string, maxRetrievalCount int, denseWeight float64, sparseWeight float64, userQuery string) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods STARTED - ansysProduct: %s, collectionName: %s, maxRetrievalCount: %d, userQuery: %s", ansysProduct, collectionName, maxRetrievalCount, userQuery)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods COMPLETED - duration: %v", duration)
	}()

	userMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided. As a first step, you need to search the Ansys API Reference Vector DB to find the relevant Method. Return the optimal search query to search the %s API Reference vector database. Make sure that you do not remove any relevant information from the original query. The format in the following JSON format, do not add anything else (no extra keys, no extra texts, or formatting (including no code fences)): {{ 'response': 'optimal vector db search query' }}`, ansysProduct, ansysProduct)

	multipleCandidatesForUserMessage := fmt.Sprintf(`In Ansys Fluent-Pyfluent you must create a script to efficiently execute the instructions provided. Propose multiple candidate queries (up to 5 highly relevant variations) that will be useful to completing the user's instruction. If available, scour the User Guide table of contents that can help you generate domain-relevant queries: %s IMPORTANT: - Do not remove any critical domain terms from the user's query. - NO FILLER WORDS OR PHRASES. - Localize to the user's intent if possible (e.g., structural or thermal context). - Keep your answer under 5 meaningful variations max. Return them in valid JSON with the following structure exactly (no extra keys, no extra texts, or formatting (including no code fences)): {{ "candidate_queries": [ "query_variant_1", "query_variant_2", "... up to 5" ] }} `, tableOfContents)

	historyMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: multipleCandidatesForUserMessage,
		},
	}

	// Time the first LLM request for candidate queries
	llmStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - LLM request (candidate queries) STARTED")
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, "")
	llmDuration := time.Since(llmStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - LLM request (candidate queries) COMPLETED - duration: %v", llmDuration)

	messageJSON, err := jsonStringToObject(result)
	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		return ""
	}

	candidateQueries, ok := messageJSON["candidate_queries"].([]interface{})
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "candidate_queries is not a slice")
		return ""
	}

	// If no candidate queries were found, use the original query
	bestQuery := ""
	if len(candidateQueries) == 0 {
		logging.Log.Warnf(&logging.ContextMap{}, "No candidate queries found, using original query: %s", userQuery)
		candidateQueries = []interface{}{userQuery}
	} else {
		rankingUserMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided. You have proposed multiple potential queries. Now, please rank these queries in terms of likely effectiveness for searching the %s API to fulfill the user's intent. Then return only the best overall query in JSON (no extra keys, no extra texts, or formatting (including no code fences)). Format: {{ 'response': 'the single best query to scour the API reference to generate code'}} Consider which query would retrieve the most relevant methods or functionalities.`, ansysProduct, ansysProduct)

		var candidateBuilder strings.Builder
		for i, query := range candidateQueries {
			if i > 0 {
				candidateBuilder.WriteString("\n")
			}
			candidateBuilder.WriteString(fmt.Sprintf(`"- %s"`, query))
		}
		candidateQueriesString := candidateBuilder.String()
		historyMessage := []sharedtypes.HistoricMessage{
			sharedtypes.HistoricMessage{
				Role:    "user",
				Content: rankingUserMessage,
			},
		}

		// Time the ranking LLM request
		llmRankingStartTime := time.Now()
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - LLM request (ranking queries) STARTED")
		result, _ := PerformGeneralRequest("Candidate queries:\n"+candidateQueriesString, historyMessage, false, "")
		llmRankingDuration := time.Since(llmRankingStartTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - LLM request (ranking queries) COMPLETED - duration: %v", llmRankingDuration)

		messageJSON, err = jsonStringToObject(result)

		if err != nil {
			logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
			return ""
		}

		bestQuery, ok = messageJSON["response"].(string)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "response is not a string")
			return ""
		}
	}
	if bestQuery == "" {
		historyMessage := []sharedtypes.HistoricMessage{
			sharedtypes.HistoricMessage{
				Role:    "user",
				Content: userMessage,
			},
		}
		logging.Log.Error(&logging.ContextMap{}, "Best query is empty")
		// Time the fallback LLM request
		llmFallbackStartTime := time.Now()
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - LLM request (fallback) STARTED")
		result, _ := PerformGeneralRequest(userQuery, historyMessage, false, "")
		llmFallbackDuration := time.Since(llmFallbackStartTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - LLM request (fallback) COMPLETED - duration: %v", llmFallbackDuration)
		messageJSON, err = jsonStringToObject(result)

		if err != nil {
			logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
			return ""
		}
		bestQuery, ok = messageJSON["response"].(string)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "response is not a string")
			return ""
		}
	}
	logging.Log.Warnf(&logging.ContextMap{}, "Best query found: %s", bestQuery)

	outputFields := []string{"text", "document_name", "previous_chunk", "next_chunk", "guid"}

	// Time the database query
	dbStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - Database query STARTED for best query: %s", bestQuery)
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, bestQuery, denseWeight, sparseWeight, "")
	dbDuration := time.Since(dbStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - Database query COMPLETED - duration: %v, results count: %d", dbDuration, len(scoredPoints))

	// Format results as requested
	var exampleBuilder strings.Builder
	for _, scoredPoint := range scoredPoints {
		entry := scoredPoint.Payload
		name := entry["document_name"].GetStringValue()
		exampleRefs, _ := getExampleReferences(name, "aali") //example_refs_info
		if exampleRefs != "" || entry["text"] != nil {
			// Format the examples as a string
			exampleBuilder.WriteString(fmt.Sprintf("Example: {%s}\n{%s}\n\n", entry["document_name"], entry["text"]))
			exampleBuilder.WriteString(fmt.Sprintf("Example {%s} References: {%s}\n\n", entry["document_name"], exampleRefs))
		}
	}
	if exampleBuilder.Len() == 0 {
		return ""
	}
	// return checkWhetherOneOfTheMethodsFits(collectionName, historyMessage, ansysProduct, denseWeight, sparseWeight, maxRetrievalCount, exampleBuilder.String())
	return exampleBuilder.String()
}

// SearchDocumentation performs a general query in the User Guide.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Documentation
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - queryString: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - historyMessage: the history of messages to be used in the query
//
// Returns:
//   - userResponse: the formatted user response string
func SearchDocumentation(collectionName string, maxRetrievalCount int, queryString string, denseWeight float64, sparseWeight float64, ansysProduct string, historyMessage []sharedtypes.HistoricMessage, tableOfContentsString string) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation STARTED - ansysProduct: %s, collectionName: %s, maxRetrievalCount: %d, queryString: %s", ansysProduct, collectionName, maxRetrievalCount, queryString)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation COMPLETED - duration: %v", duration)
	}()

	userMessage := fmt.Sprintf(`In %s: You need to write a script that finds the most relevant chapter or subchapter in the Ansys User Guide to help answer the User Query.
	- Focus only on technical content; ignore Interface and Introduction sections.
	- The section name doesn’t have to match perfectly—just find the best one to explore.
	- Indicate whether the section needs more references by returning a boolean (true or false).
	- Don’t repeat subchapters already used—pick new ones.
	- List chapter details in order of relevance.
	- Return only the JSON object in this format (no extra text, quotes, or formatting):
		[
			{
			"index": "<Index of the Chapter>",
			"sub_chapter_name": "<Chapter/Subchapter Name>",
			"section_name": "<Full path in Table of Contents>",
			"get_references": <true or false>
			}
		]
	- Example output:
		[
				{
					"index": "18.5.1",
					"sub_chapter_name": "Structural Results",
					"section_name": "ds_using_select_results_structural_types.xml::Deformation",
					"get_references": true
				},
		]
		`, ansysProduct, queryString)

	historyMessage = append(historyMessage, sharedtypes.HistoricMessage{
		Role:    "user",
		Content: userMessage,
	})

	historyMessage = append(historyMessage, sharedtypes.HistoricMessage{
		Role:    "user",
		Content: "Ansys User Guide:" + tableOfContentsString,
	})

	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - User message for LLM request: %s", userMessage)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Table of Contents: %s", tableOfContentsString)

	// Time the LLM request for chapter selection
	llmStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - LLM request (chapter selection) STARTED")
	message, _ := PerformGeneralRequest("User Query:\n"+queryString, historyMessage, false, "")
	llmDuration := time.Since(llmStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - LLM request (chapter selection) COMPLETED - duration: %v", llmDuration)

	// Log first 500 chars of response for debugging without overwhelming logs
	if len(message) > 500 {
		logging.Log.Infof(&logging.ContextMap{}, "LLM response preview (first 500 chars): %s...", message[:500])
	} else {
		logging.Log.Infof(&logging.ContextMap{}, "LLM response: %s", message)
	}

	// messageJSON is expected to be a slice of map[string]interface{} (JSON array)
	var chapters []map[string]interface{}

	// Clean and validate JSON before parsing
	cleanedMessage := strings.TrimSpace(message)
	if cleanedMessage == "" {
		logging.Log.Error(&logging.ContextMap{}, "Empty LLM response received")
		return ""
	}

	// Extract JSON array if wrapped in other text
	startIdx := strings.Index(cleanedMessage, "[")
	endIdx := strings.LastIndex(cleanedMessage, "]")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		logging.Log.Error(&logging.ContextMap{}, "No valid JSON array found in response")
		return ""
	}

	jsonContent := cleanedMessage[startIdx : endIdx+1]
	err := json.Unmarshal([]byte(jsonContent), &chapters)
	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON array: %v", err)
		logging.Log.Error(&logging.ContextMap{}, "Failed JSON content: %s", jsonContent)
		return ""
	}

	if len(chapters) == 0 {
		logging.Log.Warn(&logging.ContextMap{}, "No chapters found in LLM response")
		return ""
	}
	// Build unique sections map more efficiently
	uniqueSection := make(map[string]map[string]interface{}, len(chapters))
	for _, item := range chapters {
		name, ok := item["sub_chapter_name"].(string)
		if !ok {
			logging.Log.Warn(&logging.ContextMap{}, "Skipping chapter with invalid sub_chapter_name")
			continue
		}
		if _, exists := uniqueSection[name]; !exists {
			uniqueSection[name] = item
		}
	}

	if len(uniqueSection) == 0 {
		logging.Log.Warn(&logging.ContextMap{}, "No valid unique sections found")
		return ""
	}

	// Pre-allocate string builder with estimated capacity
	var guideSectionsBuilder strings.Builder
	guideSectionsBuilder.Grow(len(uniqueSection) * 1000) // Estimate 1KB per section

	// Process sections in parallel-friendly way (though Go's map iteration is random)
	for _, item := range uniqueSection {
		// Validate all required fields upfront
		sectionName, sectionOk := item["section_name"].(string)
		subChapterName, subChapterOk := item["sub_chapter_name"].(string)
		index, indexOk := item["index"].(string)
		getReferences, refOk := item["get_references"].(bool)

		if !sectionOk {
			logging.Log.Error(&logging.ContextMap{}, "section_name is not a string, skipping section")
			continue
		}
		if !subChapterOk {
			logging.Log.Error(&logging.ContextMap{}, "sub_chapter_name is not a string, skipping section")
			continue
		}
		if !indexOk {
			logging.Log.Error(&logging.ContextMap{}, "index is not a string, skipping section")
			continue
		}
		if !refOk {
			logging.Log.Error(&logging.ContextMap{}, "get_references is not a boolean, skipping section")
			continue
		}

		// Write section header
		guideSectionsBuilder.WriteString(fmt.Sprintf("Index: %s, Title: %s, Section Name: %s\n", index, subChapterName, sectionName))

		var userResponse strings.Builder
		userResponse.Grow(500) // Pre-allocate for efficiency

		if getReferences {
			// Time the user guide query
			guideQueryStart := time.Now()
			scoredPoints := queryUserGuideName(sectionName, uint64(3), collectionName)
			logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - User guide query for '%s' took %v", sectionName, time.Since(guideQueryStart))

			realSectionName := sectionName
			if len(scoredPoints) > 0 {
				payload := scoredPoints[0].GetPayload()
				userResponse.WriteString("With section texts: ")
				userResponse.WriteString(payload["text"].GetStringValue())
				userResponse.WriteString("\n")
				realSectionName = payload["section_name"].GetStringValue()
			} else {
				logging.Log.Warnf(&logging.ContextMap{}, "No results found for section: %s", sectionName)
				continue // Skip sections with no results to save time
			}

			// Time the graph database query
			graphQueryStart := time.Now()
			escapedSectionName := strings.ReplaceAll(realSectionName, `\`, `\\`)
			escapedSectionName = strings.ReplaceAll(escapedSectionName, `"`, `\"`)
			query := fmt.Sprintf("MATCH (n:UserGuide {name: \"%s\"})-[:References]->(reference) RETURN reference.name AS section_name", escapedSectionName)
			parameters := aali_graphdb.ParameterMap{}
			result := GeneralGraphDbQuery(query, parameters)
			logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Graph query for '%s' took %v", realSectionName, time.Since(graphQueryStart))

			if len(result) == 0 {
				logging.Log.Warnf(&logging.ContextMap{}, "No references found for section: %s", sectionName)
				// Continue with just the section text instead of skipping entirely
			} else {
				// Limit references to improve performance (max 3 as already limited in loop)
				for index, record := range result {
					if index > 2 {
						break
					}
					referenceName := record["section_name"].(string)
					userResponse.WriteString("With references: ")
					userResponse.WriteString(referenceName)
					userResponse.WriteString("\n")

					// Time individual section queries
					refQueryStart := time.Now()
					sections := queryUserGuideName(referenceName, uint64(3), collectionName)
					logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Reference query for '%s' took %v", referenceName, time.Since(refQueryStart))

					for _, section := range sections {
						if text := section.Payload["text"].GetStringValue(); text != "" {
							userResponse.WriteString("With reference section texts: ")
							userResponse.WriteString(text)
							userResponse.WriteString("\n")
						}
					}
				}
			}
		} else {
			logging.Log.Infof(&logging.ContextMap{}, "Skipping references for section: %s", sectionName)
			// Time the simplified query
			simpleQueryStart := time.Now()
			scoredPoints := queryUserGuideName(sectionName, uint64(5), collectionName)
			logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Simple query for '%s' took %v", sectionName, time.Since(simpleQueryStart))

			if len(scoredPoints) > 0 {
				payload := scoredPoints[0].Payload
				userResponse.WriteString(payload["text"].GetStringValue())
				userResponse.WriteString("\n")
			} else {
				logging.Log.Warnf(&logging.ContextMap{}, "No results found for section: %s", sectionName)
			}
		}

		guideSectionsBuilder.WriteString(userResponse.String())
		guideSectionsBuilder.WriteString("\n\n\n-------------------\n\n\n")
	}

	userGuideInformation := "Retrieved information from user guide:\n\n\n" + guideSectionsBuilder.String()
	// return checkWhetherUserInformationFits(ansysProduct, userGuideInformation, historyMessage)
	return userGuideInformation
}

// SearchExamplesForMethod performs a search in the Example.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Examples for Method
//
// Parameters:
//   - exampleCollectionName: the name of the collection to which the data objects will be added.
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - historyMessage: the history of messages to be used in the query
//   - methodName: the name of the method to be used in the query
//   - maxExamples: the maximum number of examples to be retrieved.
//
// Returns:
//   - returns method examples as a formatted string or empty string
func SearchExamplesForMethod(collectionName string, ansysProduct string, historyMessage []sharedtypes.HistoricMessage, methodName string, maxExamples int) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamplesForMethod STARTED - methodName: %s, collectionName: %s, maxExamples: %d", methodName, collectionName, maxExamples)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamplesForMethod COMPLETED - duration: %v", duration)
	}()

	// Time the method lookup
	methodDbStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamplesForMethod - Database query (method lookup) STARTED for: %s", methodName)
	nresult := getElementByName(methodName, "Method")
	methodDbDuration := time.Since(methodDbStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamplesForMethod - Database query (method lookup) COMPLETED - duration: %v, results count: %d", methodDbDuration, len(nresult))

	if len(nresult) == 0 {
		logging.Log.Warnf(&logging.ContextMap{}, "No method found with name: %s", methodName)
		return ""
	}

	result, ok := nresult[0]["n"].(map[string]interface{})
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "Failed to parse method result for: %s", methodName)
		return ""
	}

	apiExample := result["example"]
	if apiExample == nil {
		logging.Log.Warnf(&logging.ContextMap{}, "No API example found for method: %s", methodName)
		return ""
	}

	// Time the examples lookup
	examplesDbStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamplesForMethod - Database query (examples lookup) STARTED for method: %s", methodName)
	examples := getExampleNodesFromElement("Method", methodName, collectionName)
	examplesDbDuration := time.Since(examplesDbStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamplesForMethod - Database query (examples lookup) COMPLETED - duration: %v, results count: %d", examplesDbDuration, len(examples))

	if len(examples) == 0 {
		return ""
	}

	var outputBuilder strings.Builder
	outputBuilder.WriteString(fmt.Sprintf("For the api method: %s the following examples were found:\n\n", methodName))

	if apiExample != nil {
		outputBuilder.WriteString(fmt.Sprintf("API Example: %s\n\n", apiExample))
	}

	for i, example := range examples {
		if i >= maxExamples {
			break // Limit the number of examples to maxExamples
		}
		outputBuilder.WriteString(fmt.Sprintf("Example: %s\n%s\n\n", example["name"], example["text"]))

		exampleRefs, _ := getExampleReferences(example["name"].(string), "aali") //example_refs_info
		outputBuilder.WriteString(fmt.Sprintf("%s-------------------\n\n", exampleRefs))
	}
	return outputBuilder.String()
}

// GenerateCode performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Generate Code
//
// Parameters:
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - methodName: the name of the method to be used in the query
//   - examples : the examples to be used in the query
//   - historyMessages: the history of messages to be used in the query
//   - userQuery: the user query to be used for the query
//
// Returns:
//   - Code as a string
func GenerateCode(ansysProduct string, methods string, examples string, methods_from_user_guide string, historyMessages []sharedtypes.HistoricMessage, userQuery string) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GenerateCode STARTED - ansysProduct: %s, userQuery: %s", ansysProduct, userQuery)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GenerateCode COMPLETED - duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "Generating code for ansysProduct: %s, methods: %s, examples: %s, methods_from_user_guide: %s", ansysProduct, methods, examples, methods_from_user_guide)
	userMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		Use the API definition and the related APIs found. Do you best to generate the code based on the information available.

		Methods: %s
		Examples: %s
		Methods from User Guide: %s

		- Generate the code that solves the user query.
		- Use the examples and methods to generate the code.
		- If you are sure about the code, return the code in markdown format.
		- If you are not sure about the code, return "Please provide more information about the user query and the methods to be used."
		- Strictly follow the user's query and the methods provided. Don't add any extra information or comments.

		Respond with the following format, do not add anything else:
		The generated Python code only`, ansysProduct, methods, examples, methods_from_user_guide)

	historyMessages = append(historyMessages, sharedtypes.HistoricMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Time the LLM request for code generation
	llmStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GenerateCode - LLM request (code generation) STARTED")
	result, _ := PerformGeneralRequest(userQuery, historyMessages, false, "")
	llmDuration := time.Since(llmStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GenerateCode - LLM request (code generation) COMPLETED - duration: %v", llmDuration)

	if result == "" {
		return result
	}

	// Format the result as markdown code block
	logging.Log.Infof(nil, "ending of the flow time %v", time.Now().Format(time.RFC3339))
	return fmt.Sprintf("%s", result)

}

// StringReplacementArgs holds the arguments for string replacement
type StringReplacementArgs struct {
	Input        string
	Placeholder1 string
	Placeholder2 string
	Placeholder3 string
	Placeholder4 string
	Placeholder5 string
}

// StringReplaceWithArray replaces placeholders in a string with provided values.
// It counts the number of %s placeholders and uses the corresponding number of replacement values.
//
// Tags:
//   - @displayName: Replace a string with an array of strings
//
// Parameters:
//   - args: StringReplacementArgs containing input string and replacement values
//
// Returns:
//   - the input string with the replacements applied
func StringReplaceWithArray(args StringReplacementArgs) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: StringReplaceWithArray STARTED")
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: StringReplaceWithArray COMPLETED - duration: %v", duration)
	}()

	input := args.Input

	// Count the number of placeholders in the input string
	placeholderCount := countPlaceholders(input)
	if placeholderCount == 0 {
		return input
	}

	// Build replacements slice based on placeholder count
	replacements := buildReplacements(placeholderCount, args)

	return fmt.Sprintf(input, replacements...)
}

// ConvertJSONToCustomize converts JSON to customize format
//
// Tags:
//   - @displayName: Convert JSON to customize format
//
// Parameters:
//   - object: the object
//
// Returns:
//   - the value of the field as a string
//
// Example output:
// 01. Getting started (section Name -> getting_started\\getting_started_contents.md)
// 02. User guide (section Name -> user_guide\\user_guide_contents.md)
// 03. API reference (section Name -> api\\api_contents.md)
// 04. Contributing to PyFluent (section Name -> contributing\\contributing_contents.md)
// 05. Release notes (section Name -> changelog.md)
func ConvertJSONToCustomize() string {
	object := GeneralGraphDbQuery("MATCH (chapter:UserGuide {level:1}) WHERE chapter.parent = 'index.md' OPTIONAL MATCH (section:UserGuide {level:2}) WHERE section.parent = chapter.document_name OPTIONAL MATCH (subsection:UserGuide {level:3}) WHERE subsection.parent = section.document_name RETURN chapter.title AS chapter_title, chapter.document_name AS chapter_doc, section.title AS section_title, section.document_name AS section_doc, subsection.title AS subsection_title, subsection.document_name AS subsection_doc ORDER BY chapter.title, section.title, subsection.title", aali_graphdb.ParameterMap{})
	startTime := time.Now()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: ConvertJSONToCustomize STARTED")
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: ConvertJSONToCustomize COMPLETED - duration: %v", duration)
	}()

	return convertJSONToCustomizeHelper(object, 0, "")
}
