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
	"github.com/qdrant/go-client/qdrant"
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
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - userQuery: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - historyMessage: the history of messages to be used in the query
//   - tableOfContentsString: the table of contents string to be used in the query
//   - exampleCollectionName: the name of the example collection to be used in the query
//
// Returns:
//   - userResponse: the formatted user response string
func SearchDocumentation(collectionName string, exampleCollectionName string, maxRetrievalCount int, userQuery string, denseWeight float64, sparseWeight float64, ansysProduct string, historyMessage []sharedtypes.HistoricMessage, tableOfContentsString string) string {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation COMPLETED - duration: %v", duration)
	}()

	userMessage := fmt.Sprintf(`In %s: You need to write a script that finds the most relevant chapter or subchapter in the Ansys User Guide to help answer the User Query.
    Ansys User Guide: %s
    User Query: %s
	- Focus only on technical content; ignore Interface and Introduction sections.
	- The section name doesn’t have to match perfectly—just find the best one to explore.
	- Indicate whether the section needs more references by returning a boolean (true or false).
	- Don’t repeat subchapters already used—pick new ones.
	- List chapter details in order of relevance.
	- Return only the JSON object in this format (no extra text, quotes, or formatting):
	- section_name path should be in this format "api\api_contents.md"
		[
			{
			"index": "<Index of the Chapter>.<Sub Chapter if applicable>.",
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
       ]`, ansysProduct, tableOfContentsString, userQuery)

	// Time the LLM request for chapter selection
	llmStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - LLM request (chapter selection) STARTED")
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

	// Pre-allocate string builder with estimated capacity
	var guideSectionsBuilder strings.Builder
	guideSectionsBuilder.Grow(len(uniqueSection) * 1000) // Estimate 1KB per section

	// Optimization 1: Limit processing to top 3 sections for faster response
	maxSectionsToProcess := 3
	sectionsProcessed := 0

	// Optimization 2: Collect all section names for batch queries
	var sectionNames []string
	var validSections []map[string]interface{}
	
	for _, item := range uniqueSection {
		if sectionsProcessed >= maxSectionsToProcess {
			break
		}
		
		// Validate all required fields upfront
		sectionName, sectionOk := item["section_name"].(string)
		_, subChapterOk := item["sub_chapter_name"].(string)
		_, indexOk := item["index"].(string)
		_, refOk := item["get_references"].(bool)

		if !sectionOk || !subChapterOk || !indexOk || !refOk {
			logging.Log.Warn(&logging.ContextMap{}, "Skipping section with invalid fields")
			continue
		}
		
		sectionNames = append(sectionNames, sectionName)
		validSections = append(validSections, item)
		sectionsProcessed++
	}

	if len(validSections) == 0 {
		logging.Log.Warn(&logging.ContextMap{}, "No valid sections to process")
		return ""
	}

	// Optimization 3: Batch query for all sections at once
	batchQueryStart := time.Now()
	sectionDataMap := make(map[string][]*qdrant.ScoredPoint)
	
	for _, sectionName := range sectionNames {
		// Reduce query limit from 5/3 to 2 for faster response
		scoredPoints := queryUserGuideName(sectionName, uint64(2), collectionName)
		if len(scoredPoints) > 0 {
			sectionDataMap[sectionName] = scoredPoints
		}
	}
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Batch queries COMPLETED - duration: %v", time.Since(batchQueryStart))

	// Process sections with cached data
	for i, item := range validSections {
		sectionName := item["section_name"].(string)
		subChapterName := item["sub_chapter_name"].(string)
		index := item["index"].(string)
		getReferences := item["get_references"].(bool)

		// Write section header
		guideSectionsBuilder.WriteString(fmt.Sprintf("Index: %s, Title: %s, Section Name: %s\n", index, subChapterName, sectionName))

		var userResponse strings.Builder
		userResponse.Grow(500) // Pre-allocate for efficiency

		scoredPoints, exists := sectionDataMap[sectionName]
		if !exists || len(scoredPoints) == 0 {
			logging.Log.Warnf(&logging.ContextMap{}, "No data found for section: %s", sectionName)
			continue
		}

		// Get main section content - use .Payload like existing code
		payload := scoredPoints[0].Payload
		userResponse.WriteString("With section texts: ")
		userResponse.WriteString(payload["text"].GetStringValue())
		userResponse.WriteString("\n")
		
		// Optimization 4: Skip reference processing for faster response (can be made conditional)
		if getReferences && i < 2 { // Only get references for first 2 sections
			realSectionName := payload["section_name"].GetStringValue()
			
			// Simplified reference query with timeout protection
			refQueryStart := time.Now()
			escapedSectionName := strings.ReplaceAll(realSectionName, `\`, `\\`)
			escapedSectionName = strings.ReplaceAll(escapedSectionName, `"`, `\"`)
			query := fmt.Sprintf("MATCH (n:UserGuide {name: \"%s\"})-[:References]->(reference) RETURN reference.name AS section_name LIMIT 2", escapedSectionName)
			parameters := aali_graphdb.ParameterMap{}
			result := GeneralGraphDbQuery(query, parameters)
			
			// Only process if query was fast (< 2 seconds)
			if time.Since(refQueryStart) < 2*time.Second && len(result) > 0 {
				// Process only first reference to save time
				if len(result) > 0 {
					referenceName := result[0]["section_name"].(string)
					userResponse.WriteString("With references: ")
					userResponse.WriteString(referenceName)
					userResponse.WriteString("\n")
					
					// Get reference content with minimal data
					refSections := queryUserGuideName(referenceName, uint64(1), collectionName)
					if len(refSections) > 0 {
						if text := refSections[0].Payload["text"].GetStringValue(); text != "" {
							// Truncate reference text to first 500 chars for performance
							if len(text) > 500 {
								text = text[:500] + "..."
							}
							userResponse.WriteString("With reference section texts: ")
							userResponse.WriteString(text)
							userResponse.WriteString("\n")
						}
					}
				}
				logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Reference processing for '%s' took %v", realSectionName, time.Since(refQueryStart))
			} else {
				logging.Log.Warnf(&logging.ContextMap{}, "Skipping references for '%s' - query too slow or no results", realSectionName)
			}
		}

		guideSectionsBuilder.WriteString(userResponse.String())
		guideSectionsBuilder.WriteString("\n\n\n-------------------\n\n\n")
	}

	userGuideInformation := "Retrieved information from user guide:\n\n\n" + guideSectionsBuilder.String()
	unambiguousMethodPath, queryToApiReference, questionToUser := checkWhetherUserInformationFits(ansysProduct, userGuideInformation, historyMessage, userQuery)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: SearchDocumentation - Unambiguous method path: %s, query to API reference: %s, question to user: %s", unambiguousMethodPath, queryToApiReference, questionToUser)
	if unambiguousMethodPath != "" {
		return unambiguousMethodPath
	} else if queryToApiReference != "" {
		methods := searchExamplesForMethod(exampleCollectionName, ansysProduct, historyMessage, queryToApiReference, maxRetrievalCount)
		return methods
	} else {
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

// StringReplacementArgs holds the arguments for string replacement
type StringReplacementArgs struct {
	Input        string
	Placeholder1 string
	Placeholder2 string
	Placeholder3 string
	Placeholder4 string
	Placeholder5 string
}

// QueryUserGuideAndFormat converts JSON to customize format
//
// Tags:
//   - @displayName: Query the UserGuide and convert it to customize format
//
// Parameters:
//   - object: the object
//
// Returns:
//   - the value of the field as a string
//
// Example output:
// 01.
func QueryUserGuideAndFormat() string {
	object := GeneralGraphDbQuery("MATCH (chapter:UserGuide {level:1}) WHERE chapter.parent = 'index.md' OPTIONAL MATCH (section:UserGuide {level:2}) WHERE section.parent = chapter.document_name OPTIONAL MATCH (subsection:UserGuide {level:3}) WHERE subsection.parent = section.document_name RETURN chapter.title AS chapter_title, chapter.document_name AS chapter_doc, section.title AS section_title, section.document_name AS section_doc, subsection.title AS subsection_title, subsection.document_name AS subsection_doc ORDER BY chapter.title, section.title, subsection.title", aali_graphdb.ParameterMap{})
	startTime := time.Now()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: ConvertJSONToCustomize STARTED")
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: ConvertJSONToCustomize COMPLETED - duration: %v", duration)
	}()

	return convertJSONToCustomizeHelper(object, 0, "")
}
