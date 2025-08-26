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

// RewriteQueryWithHistory Rewrite the query based on the Histroy
//
// # The function returns the userquery
//
// Tags:
//   - @displayName: Rewrite Query With Respect to History
//
// Parameters:
//   - historyMessage: the history of messages to be used in the query
//   - UserQuery: the user query to be used for the query.
//
// Returns:
//   - UserQuery: formatted UserQuery
func RewriteQueryWithHistory(historyMessage []sharedtypes.HistoricMessage, userQuery string) string {
	systemMessage := `You are heful assistant who will look at the latest 5 history chat and assitant reponse and userquery as new input and create a redefined user query and query itself shoudld be sufficient to understand the user query and provide the answer.
	Response: Just query, do not add anything else, do not add any extra keys, no extra texts, or formatting (including no code fences).`
	result := PerformGeneralRequestNoStreaming(userQuery, historyMessage, systemMessage)
	if result != "" {
		logging.Log.Infof(&logging.ContextMap{}, "Rewritten query: %s", result)
		return result
	} else {
		return userQuery
	}

}

// SearchExamples performs a search in the Example collection name.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Examples
//
// Parameters:
//   - libraryName: the name of the library to be used in the system message
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//   - userQuery: the user query to be used for the query.
//
// Returns:
//   - generatedCode: the generated code as a string
func SearchExamples(libraryName string, maxRetrievalCount int, denseWeight float64, sparseWeight float64, userQuery string) string {
	startTime := time.Now()
	ansysProduct := pyansysProduct[libraryName]
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchExamples - started %v", time.Now())
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
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchExamples - preprocess ended %v", time.Now())
	collectionName := fmt.Sprintf("%s_examples", libraryName)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchExamples - Database query STARTED for user query: %s", collectionName)
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, userQuery, denseWeight, sparseWeight, "")
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchExamples - hybrid ended %v", time.Now())
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
		exampleRefs, _ := getExampleReferences(exampleName, libraryName) //example_refs_info

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
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchExamples - ended %v", time.Now())
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
//   - libraryName: the name of the library to be used in the system message
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//   - userQuery: the user query to be used for the query.
//
// Returns:
//   - examplesString: the formatted examples string containing the method examples and references
func SearchMethods(tableOfContents string, libraryName string, maxRetrievalCount int, denseWeight float64, sparseWeight float64, userQuery string) string {
	startTime := time.Now()
	ansysProduct := pyansysProduct[libraryName]
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchMethods - started %v", startTime)
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
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchMethods - preprocess ended %v", time.Now())
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - Database query STARTED for best query: %s", bestQuery)
	collectionName := fmt.Sprintf("%s_elements", libraryName)
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, bestQuery, denseWeight, sparseWeight, "")
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchMethods - hybridsearch ended %v", time.Now())
	dbDuration := time.Since(dbStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchMethods - Database query COMPLETED - duration: %v, results count: %d", dbDuration, len(scoredPoints))

	// Format results as requested
	var exampleBuilder strings.Builder
	for _, scoredPoint := range scoredPoints {
		entry := scoredPoint.Payload
		name := entry["document_name"].GetStringValue()
		exampleRefs, _ := getExampleReferences(name, libraryName) //example_refs_info
		if exampleRefs != "" || entry["text"] != nil {
			// Format the examples as a string
			exampleBuilder.WriteString(fmt.Sprintf("Example: {%s}\n{%s}\n\n", entry["document_name"], entry["text"]))
			exampleBuilder.WriteString(fmt.Sprintf("Example {%s} References: {%s}\n\n", entry["document_name"], exampleRefs))
		}
	}
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchMethods - postprocess ended %v", time.Now())
	if exampleBuilder.Len() == 0 {
		return ""
	}
	return checkWhetherOneOfTheMethodsFits(collectionName, historyMessage, ansysProduct, denseWeight, sparseWeight, maxRetrievalCount, exampleBuilder.String())
	// return exampleBuilder.String()
}

// SearchDocumentation performs a general query in the User Guide.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Documentation
//
// Parameters:
//   - libraryName: the name of the library to be used in the system message
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - userQuery: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//   - historyMessage: the history of messages to be used in the query
//   - tableOfContentsString: the table of contents string to be used in the query
//
// Returns:
//   - userResponse: the formatted user response string
func SearchDocumentation(libraryName string, maxRetrievalCount int, userQuery string, denseWeight float64, sparseWeight float64, historyMessage []sharedtypes.HistoricMessage, tableOfContentsString string) string {
	startTime := time.Now()
	ansysProduct := pyansysProduct[libraryName]
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - started %v", time.Now())
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation COMPLETED - duration: %v", duration)
	}()

	userMessage := fmt.Sprintf(`In %s: """You need to write a script that finds the most relevant chapter or subchapter in the Ansys User Guide to help answer the User Query.

		### Table of Contents:
		%s

		### User Query:
		%s

		### Instructions:
		- Focus only on technical content; ignore Interface/Introduction.  
		- The section name doesn’t have to match exactly; pick the closest relevant one.  
		- Avoid repeating previously used chapters/subchapters.  
		- Indicate if more references are needed: 'get_references: true/false'.  
		- Return only the JSON array in this format:

		json
		[
		{
			"index": "<Index of Chapter.Subchapter>",
			"sub_chapter_name": "<Name>",
			"section_name": "<Path like api\\api_contents.md>",
			"get_references": true/false
		}
		]
		`, ansysProduct, tableOfContentsString, userQuery)

	// Time the LLM request for chapter selection
	llmStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - LLM request (chapter selection) STARTED %s, %s", userMessage, historyMessage)
	message, _ := PerformGeneralRequest(userMessage, historyMessage, false, "")
	llmDuration := time.Since(llmStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - LLM request (chapter selection) COMPLETED - duration: %v", llmDuration)

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
	logging.Log.Infof(&logging.ContextMap{}, "Found %d chapters in LLM response and chapters %s", len(chapters), chapters)
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

	var guideSectionsBuilder strings.Builder

	for _, item := range uniqueSection {
		sectionName, sectionOk := item["section_name"].(string)
		subChapterName, subChapterOk := item["sub_chapter_name"].(string)
		index, indexOk := item["index"].(string)
		getReferences, refOk := item["get_references"].(bool)

		if !sectionOk || !subChapterOk || !indexOk || !refOk {
			logging.Log.Warn(&logging.ContextMap{}, "Skipping section with invalid fields")
			continue
		}

		guideSectionsBuilder.WriteString(fmt.Sprintf("Index: %s, Title: %s, Section Name: %s\n", index, subChapterName, sectionName))

		var userResponse strings.Builder

		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - preprocess ended %v", time.Now())

		collectionName := fmt.Sprintf("%s_user_guide", libraryName)

		scoredPoints := queryUserGuideName(sectionName, uint64(3), collectionName) // changed this to 3 from 5
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - db query ended %v", time.Now())
		for j, scoredPoint := range scoredPoints {
			if j >= 3 {
				break
			}
			payload := scoredPoint.Payload
			userResponse.WriteString(fmt.Sprintf("With section texts %d: ", j+1))
			userResponse.WriteString(payload["text"].GetStringValue())
			userResponse.WriteString("\n")
		}

		if getReferences && len(scoredPoints) > 0 {
			logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - references query started %v", time.Now())
			realSectionName := scoredPoints[0].Payload["section_name"].GetStringValue()
			escapedSectionName := strings.ReplaceAll(realSectionName, `\`, `\\`)
			escapedSectionName = strings.ReplaceAll(escapedSectionName, `"`, `\"`)
			query := fmt.Sprintf("MATCH (n:UserGuide {name: \"%s\"})-[:References]->(reference) RETURN reference.name AS section_name LIMIT 5", escapedSectionName)
			parameters := aali_graphdb.ParameterMap{}
			result := GeneralGraphDbQuery(query, parameters, libraryName)
			logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - references query ended %v", time.Now())

			for refIdx, reference := range result {
				if refIdx >= 3 {
					break
				}
				referenceName := reference["section_name"].(string)
				userResponse.WriteString(fmt.Sprintf("With references %d: ", refIdx+1))
				userResponse.WriteString(referenceName)
				userResponse.WriteString("\n")

				refSections := queryUserGuideName(referenceName, uint64(3), collectionName)
				if len(refSections) > 0 {
					if text := refSections[0].Payload["text"].GetStringValue(); text != "" {
						userResponse.WriteString(fmt.Sprintf("With reference section texts %d: ", refIdx+1))
						userResponse.WriteString(text)
						userResponse.WriteString("\n")
					}
				}
			}
			logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - references query ended %v", time.Now())
		}

		guideSectionsBuilder.WriteString(userResponse.String())
		guideSectionsBuilder.WriteString("\n\n\n-------------------\n\n\n")
	}

	userGuideInformation := "Retrieved information from user guide:\n\n\n" + guideSectionsBuilder.String()
	unambiguousMethodPath, queryToApiReference, questionToUser := checkWhetherUserInformationFits(ansysProduct, userGuideInformation, historyMessage, userQuery)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Unambiguous method path: %s, query to API reference: %s, question to user: %s", unambiguousMethodPath, queryToApiReference, questionToUser)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - postprocess ended %v", time.Now())
	if unambiguousMethodPath != "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - ended %v", time.Now())
		return unambiguousMethodPath
	} else if queryToApiReference != "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - example search started %v", time.Now())
		exampleCollectionName := fmt.Sprintf("%s_examples", libraryName)
		methods := searchExamplesForMethod(exampleCollectionName, ansysProduct, historyMessage, queryToApiReference, maxRetrievalCount, libraryName)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - example search ended %v", time.Now())
		return methods
	} else {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND SearchDocumentation - ended %v", time.Now())
		return questionToUser
	}

}

// GenerateCode performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Generate Code
//
// Parameters:
//   - methodName: the name of the method to be used in the query
//   - examples : the examples to be used in the query
//   - historyMessages: the history of messages to be used in the query
//   - userQuery: the user query to be used for the query
//   - libraryName: the name of the library to be used in the query
//
// Returns:
//   - Code as a string
func GenerateCode(methods string, examples string, methods_from_user_guide string, historyMessages []sharedtypes.HistoricMessage, userQuery string, libraryName string) string {
	startTime := time.Now()
	ansysProduct := pyansysProduct[libraryName]
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GenerateCode COMPLETED - duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "Generating code for ansysProduct: %s, methods: %s, examples: %s, methods_from_user_guide: %s, userQuery %s", ansysProduct, methods, examples, methods_from_user_guide, userQuery)
	userMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		Use the API definition and the related APIs found. Do your best to generate the code based on the information available.

		Methods: %s
		Examples: %s
		Methods from User Guide: %s

		- STRICT: Only use the context provided in this system message. Do NOT think outside this context, do NOT add anything else, do NOT invent or hallucinate anything beyond the provided information.
		- Generate the code that solves the user query using only the Methods, Examples and Methods from User Guide.
		- If you are not able to generate the code uisng the context provided, and Methods from User Guide has question instead of required context, Send the question as reponse.
		- If you are sure about the code, return the code in markdown format.
		- If you are not sure about the code and  Methods from User Guide does not have any question, return "Please provide more information about the user query and the methods to be used."

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
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: userQuery: %s, historyMessages %s", userQuery, historyMessages)
	llmDuration := time.Since(llmStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GenerateCode - LLM request (code generation) COMPLETED - duration: %v", llmDuration)

	if result == "" {
		return result
	}
	return fmt.Sprintf("%s", result)

}

// QueryUserGuideAndFormat converts JSON to customize format
//
// Tags:
//   - @displayName: Query the UserGuide and convert it to customize format
//
// Parameters:
//   - libraryName: the name of the library to be used in the system message
//
// Returns:
//   - the value of the field as a string
//
// Example output:
// 01.
func QueryUserGuideAndFormat(libraryName string) string {
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: QueryUserGuideAndFormat called")
	startTime := time.Now()
	object := GeneralGraphDbQuery("MATCH (chapter:UserGuide {level:1}) WHERE chapter.parent = 'index.md' OPTIONAL MATCH (section:UserGuide {level:2}) WHERE section.parent = chapter.document_name OPTIONAL MATCH (subsection:UserGuide {level:3}) WHERE subsection.parent = section.document_name RETURN chapter.title AS chapter_title, chapter.document_name AS chapter_doc, section.title AS section_title, section.document_name AS section_doc, subsection.title AS subsection_title, subsection.document_name AS subsection_doc ORDER BY chapter.title, section.title, subsection.title", aali_graphdb.ParameterMap{}, libraryName)

	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: ConvertJSONToCustomize COMPLETED - duration: %v", duration)
	}()

	return convertJSONToCustomizeHelper(object, 0, "")
}

// Make API Request to the URL with given method, headers, and body
//
// Tags:
//   - @displayName: Make API Request
//
// Parameters:
//   - requestType: the type of the request (GET, POST, etc.)
//   - endpoint: the URL to send the request to
//   - header: the headers to include in the request
//   - query: the user query to be used for the query.
//   - libraryName: the name of the library to be used in the query
//
// Returns:
//   - success: a boolean indicating whether the request was successful
//   - returnJsonBody: the JSON body of the response as a string
func MakeAPIRequest(requestType string, endpoint string, header map[string]string, query string, libraryName string) (code string) {

	queryParams := map[string]string{
		"Content-Type": "application/json",
	}
	// Sample json body
	// `{"key": "value", "number": 123}`
	jsonBody := fmt.Sprintf(`{"query": "%s", "product": "%s"}`, query)
	if requestType == "" {
		requestType = "POST"
	}
	if endpoint == "" {
		endpoint = "http://localhost:8000/code_gen"
	}
	success, returnJsonBody := SendRestAPICall(requestType, endpoint, header, queryParams, jsonBody)
	if !success {
		logging.Log.Errorf(&logging.ContextMap{}, "API request failed")
		return ""
	}
	// CHeck the returnJsonBody is valid json
	var result map[string]interface{}
	err := json.Unmarshal([]byte(returnJsonBody), &result)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error converting response to JSON object: %v", err)
		return ""
	}
	if code, ok := result["code"].(string); ok {
		return code
	}
	return ""
}
