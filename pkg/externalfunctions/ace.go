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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/config"
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
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_REWRITE_QUERY_HISTORY - Input: historyMessage=%v, userQuery=%s", historyMessage, userQuery)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_REWRITE_QUERY_HISTORY - Duration: %v", duration)
	}()

	systemMessage := `You are heful assistant who will look at the latest 5 history chat and assitant reponse and userquery as new input and create a redefined user query and query itself shoudld be sufficient to understand the user query and provide the answer.
	Response: Just query, do not add anything else, do not add any extra keys, no extra texts, or formatting (including no code fences).`
	result := PerformGeneralRequestNoStreaming(userQuery, historyMessage, systemMessage)

	if result != "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_REWRITE_QUERY_HISTORY - Output: %s", result)
		return result
	} else {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_REWRITE_QUERY_HISTORY - Output: %s", userQuery)
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
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_EXAMPLES - Input: libraryName=%s, maxRetrievalCount=%d, denseWeight=%f, sparseWeight=%f, userQuery=%s", libraryName, maxRetrievalCount, denseWeight, sparseWeight, userQuery)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_SEARCH_EXAMPLES - Duration: %v", duration)
	}()

	outputFields := []string{"text", "document_name", "previous_chunk", "next_chunk", "guid"}
	collectionName := fmt.Sprintf("%s_examples", libraryName)
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, userQuery, denseWeight, sparseWeight, "")

	if len(scoredPoints) == 0 {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_EXAMPLES - Output: (empty)")
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

	ansysProduct := pyansysProduct["name"][libraryName]
	// User message to verify the results got from the DB is relevant or not to solve the problem
	userMessage := fmt.Sprintf(`In %s: You need to verify the examples returned from the database is relevant or not to solve the problem.

		If you are sure that the examples are relevant, return "true". If you need more examples, return "false".

		The format in the following text, do not add anything else (no extra keys, no extra texts, or formatting (including no code fences)):
		true/false

	`, ansysProduct)
	historyMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: userMessage,
		},
	}
	result, _ := PerformGeneralRequest(exampleString, historyMessage, false, "")

	// Convert string result to boolean using strconv.ParseBool
	response, err := strconv.ParseBool(result)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_EXAMPLES - Output: (error)")
		return ""
	}
	if !response {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_EXAMPLES - Output: (false response)")
		return ""
	}

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_EXAMPLES - Output: %s", exampleString)
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
func SearchMethods(libraryName string, maxRetrievalCount int, denseWeight float64, sparseWeight float64, userQuery string) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_METHODS - Input: libraryName=%s, maxRetrievalCount=%d, denseWeight=%f, sparseWeight=%f, userQuery=%s", libraryName, maxRetrievalCount, denseWeight, sparseWeight, userQuery)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_SEARCH_METHODS - Duration: %v", duration)
	}()

	bestQuery := userQuery
	historyMessage := []sharedtypes.HistoricMessage{}
	outputFields := []string{"text", "document_name", "previous_chunk", "next_chunk", "guid"}

	collectionName := fmt.Sprintf("%s_elements", libraryName)
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, bestQuery, denseWeight, sparseWeight, "")

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

	if exampleBuilder.Len() == 0 {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_METHODS - Output: (empty)")
		return ""
	}

	ansysProduct := pyansysProduct["name"][libraryName]
	result := checkWhetherOneOfTheMethodsFits(collectionName, historyMessage, ansysProduct, denseWeight, sparseWeight, maxRetrievalCount, exampleBuilder.String())

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_METHODS - Output: %s", result)
	return result
}

// Function to get the raw data from congnitive services without any processing
//
// Tags:
//   - @displayName: Get Raw Data from Cognitive Services for user guide
//
// Parameters:
//   - libraryName: the name of the library to be used in the system message
//   - userQuery: the user query to be used for the query.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//
// Returns:
//   - response: the response from the cognitive services as a string
func GetRawDataFromCognitiveServicesForDocumentation(libraryName string, userQuery string, maxRetrievalCount int) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Input: libraryName=%s, userQuery=%s, maxRetrievalCount=%d", libraryName, userQuery, maxRetrievalCount)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Duration: %v", duration)
	}()

	ansysProduct := pyansysProduct["name"][libraryName]

	// 0. Rewrite user query
	userMessage := fmt.Sprintf(`In %s: The following user query may be brief, ambiguous, or lacking technical detail.
		Please rewrite it as a clear, detailed, and specific question suitable for retrieving relevant and precise information from a technical knowledge base about {product}.
		If necessary, add clarifying context, standard terminology, or related technical concepts commonly used in {product} documentation, without changing the original intent of the user's question.

		User Query: "%s"

		Return your response as a JSON object with a single key "unified_query".
		For example:
		"unified_query": "<your generated query here>"`, ansysProduct, userQuery)

	historyMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: userMessage,
		},
	}

	// Make llm call to rewrite the query
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, "")
	messageJSON, err := jsonStringToObject(result)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (JSON parse error)")
		return ""
	}
	rewrittenQuery, ok := messageJSON["unified_query"].(string)
	if !ok {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (unified_query not string)")
		return ""
	}
	if rewrittenQuery == "" {
		rewrittenQuery = userQuery
	}

	// 1. Get embedding
	embReq, _ := json.Marshal(map[string]string{
		"model": "text-embedding-3-large",
		"input": rewrittenQuery,
	})

	req, _ := http.NewRequest("POST",
		config.GlobalConfig.AZURE_EMBEDDING_URL,
		bytes.NewBuffer(embReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.GlobalConfig.AZURE_EMBEDDING_TOKEN)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var embResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&embResp)
	embedding := embResp["data"].([]interface{})[0].(map[string]interface{})["embedding"]

	// 2. Vector search
	searchReq, _ := json.Marshal(map[string]interface{}{
		"vectorQueries": []map[string]interface{}{{
			"kind": "vector", "k": maxRetrievalCount, "vector": embedding, "fields": "content_vctr",
		}},
		"filter": fmt.Sprintf("product eq '%s' and version eq '%s' and typeOFasset eq 'documentation'", libraryName, pyansysProduct["version"][libraryName]),
		"top":    5,
		"select": "content,product,physics,sourceURL_lvl1,sourceTitle_lvl1,typeOFasset",
	})

	req, err = http.NewRequest("POST",
		config.GlobalConfig.AZURE_COGNITIVE_SERVICE_API,
		bytes.NewBuffer(searchReq))
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (search request error)")
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.GlobalConfig.AZURE_COGNITIVE_SERVICE_TOKEN)

	resp, err = client.Do(req)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (search error)")
		return ""
	}
	defer resp.Body.Close()

	var searchResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&searchResp)

	// 3. Format and print
	results := searchResp["value"].([]interface{})
	chunks := make(map[string]interface{})

	for i, result := range results {
		r := result.(map[string]interface{})
		chunks[fmt.Sprintf("chunk_%d", i+1)] = map[string]interface{}{
			"context":          r["content"],
			"product":          r["product"],
			"physics":          r["physics"],
			"sourceURL_lvl1":   r["sourceURL_lvl1"],
			"sourceTitle_lvl1": r["sourceTitle_lvl1"],
			"typeOfAsset":      r["typeOFasset"],
		}
	}

	output, err := json.MarshalIndent(chunks, "", "  ")
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (JSON marshal error)")
		return ""
	}

	// // 4. Process the output
	// processingMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
	// 	Use the API definition and the related APIs found. Do your best to generate the code based on the information available.
	// 	API Search Results: %s
	// 	- STRICT: Only use the context provided in this system message. Do NOT think outside this context, do NOT add anything else, do NOT invent or hallucinate anything beyond the provided information.
	// 	- Generate the code that solves the user query using only the API Search Results.
	// 	- If you are not able to generate the code using the context provided, Send "I am not able to generate the code with the information provided."
	// 	- If you are sure about the code, return the code in markdown format.
	// 	- If you are not sure about the code, return "Please provide more information about the user query and the methods to be used."
	// 	Respond with the following format, do not add anything else:
	// 	The generated Python code only`, ansysProduct, string(output))
	// processingHistoryMessage := []sharedtypes.HistoricMessage{
	// 	sharedtypes.HistoricMessage{
	// 		Role:    "user",
	// 		Content: processingMessage,
	// 	},
	// }
	// result, _ = PerformGeneralRequest(userQuery, processingHistoryMessage, false, "")
	// logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: %s", result)
	// return result

	return string(output)
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
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Input: libraryName=%s, maxRetrievalCount=%d, userQuery=%s, denseWeight=%f, sparseWeight=%f, historyMessage=%v, tableOfContentsString=%s", libraryName, maxRetrievalCount, userQuery, denseWeight, sparseWeight, historyMessage, tableOfContentsString)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_SEARCH_DOCUMENTATION - Duration: %v", duration)
	}()

	ansysProduct := pyansysProduct["name"][libraryName]
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
	message, _ := PerformGeneralRequest(userMessage, historyMessage, false, "")

	// messageJSON is expected to be a slice of map[string]interface{} (JSON array)
	var chapters []map[string]interface{}

	// Clean and validate JSON before parsing
	cleanedMessage := strings.TrimSpace(message)
	if cleanedMessage == "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: (empty response)")
		return ""
	}

	// Extract JSON array if wrapped in other text
	startIdx := strings.Index(cleanedMessage, "[")
	endIdx := strings.LastIndex(cleanedMessage, "]")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: (invalid JSON)")
		return ""
	}

	jsonContent := cleanedMessage[startIdx : endIdx+1]
	err := json.Unmarshal([]byte(jsonContent), &chapters)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: (JSON parse error)")
		return ""
	}

	if len(chapters) == 0 {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: (no chapters)")
		return ""
	}
	// Build unique sections map more efficiently
	uniqueSection := make(map[string]map[string]interface{}, len(chapters))
	for _, item := range chapters {
		name, ok := item["sub_chapter_name"].(string)
		if !ok {
			continue
		}
		if _, exists := uniqueSection[name]; !exists {
			uniqueSection[name] = item
		}
	}

	if len(uniqueSection) == 0 {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: (no unique sections)")
		return ""
	}

	var guideSectionsBuilder strings.Builder

	for _, item := range uniqueSection {
		sectionName, sectionOk := item["section_name"].(string)
		subChapterName, subChapterOk := item["sub_chapter_name"].(string)
		index, indexOk := item["index"].(string)
		getReferences, refOk := item["get_references"].(bool)

		if !sectionOk || !subChapterOk || !indexOk || !refOk {
			continue
		}

		guideSectionsBuilder.WriteString(fmt.Sprintf("Index: %s, Title: %s, Section Name: %s\n", index, subChapterName, sectionName))

		var userResponse strings.Builder
		collectionName := fmt.Sprintf("%s_user_guide", libraryName)
		scoredPoints := queryUserGuideName(sectionName, uint64(3), collectionName) // changed this to 3 from 5
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
			realSectionName := scoredPoints[0].Payload["section_name"].GetStringValue()
			escapedSectionName := strings.ReplaceAll(realSectionName, `\`, `\\`)
			escapedSectionName = strings.ReplaceAll(escapedSectionName, `"`, `\"`)
			query := fmt.Sprintf("MATCH (n:UserGuide {name: \"%s\"})-[:References]->(reference) RETURN reference.name AS section_name LIMIT 5", escapedSectionName)
			parameters := aali_graphdb.ParameterMap{}
			result := GeneralGraphDbQuery(query, parameters, libraryName)

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
		}

		guideSectionsBuilder.WriteString(userResponse.String())
		guideSectionsBuilder.WriteString("\n\n\n-------------------\n\n\n")
	}

	userGuideInformation := "Retrieved information from user guide:\n\n\n" + guideSectionsBuilder.String()
	unambiguousMethodPath, queryToApiReference, questionToUser := checkWhetherUserInformationFits(ansysProduct, userGuideInformation, historyMessage, userQuery)

	if unambiguousMethodPath != "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: %s", unambiguousMethodPath)
		return unambiguousMethodPath
	} else if queryToApiReference != "" {
		exampleCollectionName := fmt.Sprintf("%s_examples", libraryName)
		methods := searchExamplesForMethod(exampleCollectionName, ansysProduct, historyMessage, queryToApiReference, maxRetrievalCount, libraryName)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: %s", methods)
		return methods
	} else {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_DOCUMENTATION - Output: %s", questionToUser)
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
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GENERATE_CODE - Input: methods=%s, examples=%s, methods_from_user_guide=%s, historyMessages=%v, userQuery=%s, libraryName=%s", methods, examples, methods_from_user_guide, historyMessages, userQuery, libraryName)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_GENERATE_CODE - Duration: %v", duration)
	}()

	ansysProduct := pyansysProduct["name"][libraryName]
	userMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		Use the API definition and the related APIs found. Do your best to generate the code based on the information available.

		Methods: %s
		Examples: %s
		Methods from User Guide: %s

		- STRICT: You are a code generation chatbot only create python code with respect to pyansys packages no documentation or reference purely python code
		- Generate the code that solves the user query using only the Methods, Examples and Methods from User Guide.
		- If you are not able to generate the code using the context provided, and Methods from User Guide has question instead of required context, Send the question as response.
		- If you are sure about the code, return the code in markdown format.
		- If you are not sure about the code and  Methods from User Guide does not have any question, return "Please provide more information about the user query and the methods to be used."
		- If you think the context provided is okay to create a script, then do so. (Do logical thinking and provide the answer if required but always stay within the context and provide the answer only if you are sure about it.)
		- DO ONLY what user asks dont add additional parameter or anything else.

		Respond with the following format, do not add anything else:
		The generated Python code only`, ansysProduct, methods, examples, methods_from_user_guide)

	// - STRICT: Only use the context provided in this system message. Do NOT think outside this context, do NOT add anything else, do NOT invent or hallucinate anything beyond the provided information.

	historyMessages = append(historyMessages, sharedtypes.HistoricMessage{
		Role:    "user",
		Content: userMessage,
	})

	result, _ := PerformGeneralRequest(userQuery, historyMessages, false, "")

	if result == "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GENERATE_CODE - Output: (empty)")
		return result
	}

	output := fmt.Sprintf("%s", result)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GENERATE_CODE - Output: %s", output)
	return output
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
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_QUERY_USER_GUIDE_FORMAT - Input: libraryName=%s", libraryName)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_QUERY_USER_GUIDE_FORMAT - Duration: %v", duration)
	}()

	object := GeneralGraphDbQuery("MATCH (chapter:UserGuide {level:1}) WHERE chapter.parent = 'index.md' OPTIONAL MATCH (section:UserGuide {level:2}) WHERE section.parent = chapter.document_name OPTIONAL MATCH (subsection:UserGuide {level:3}) WHERE subsection.parent = section.document_name RETURN chapter.title AS chapter_title, chapter.document_name AS chapter_doc, section.title AS section_title, section.document_name AS section_doc, subsection.title AS subsection_title, subsection.document_name AS subsection_doc ORDER BY chapter.title, section.title, subsection.title", aali_graphdb.ParameterMap{}, libraryName)

	result := convertJSONToCustomizeHelper(object, 0, "")
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_QUERY_USER_GUIDE_FORMAT - Output: %s", result)
	return result
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
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_MAKE_API_REQUEST - Input: requestType=%s, endpoint=%s, header=%v, query=%s, libraryName=%s", requestType, endpoint, header, query, libraryName)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_MAKE_API_REQUEST - Duration: %v", duration)
	}()

	queryParams := map[string]string{
		"Content-Type": "application/json",
	}
	if libraryName == "" {
		libraryName = "pyfluent"
	}
	// Sample json body
	// `{"key": "value", "number": 123}`
	jsonBody := fmt.Sprintf(`{"query": "%s", "product": "%s" }`, query, libraryName)
	if requestType == "" {
		requestType = "POST"
	}
	if endpoint == "" {
		endpoint = "https://dev-codegen.azurewebsites.net/code_gen"
	}
	success, returnJsonBody := SendRestAPICall(requestType, endpoint, header, queryParams, jsonBody)
	if !success {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_MAKE_API_REQUEST - Output: (API request failed)")
		return ""
	}
	// CHeck the returnJsonBody is valid json
	var result map[string]interface{}
	err := json.Unmarshal([]byte(returnJsonBody), &result)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_MAKE_API_REQUEST - Output: (JSON parse error)")
		return ""
	}
	if code, ok := result["code"].(string); ok {
		code = PerformGeneralRequestNoStreaming("The code generated is: "+code, []sharedtypes.HistoricMessage{}, "You are a helpful assistant that helps to generate python code in markdown format. Do not add anything else, do not add any extra keys, no extra texts, or formatting (including no code fences). Remove the docs in the code and only provide the code.")
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_MAKE_API_REQUEST - Output: %s", code)
		return code
	}
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_MAKE_API_REQUEST - Output: (no code found)")
	return ""
}

// API to get the data from congnitive services
//
// Tags:
//   - @displayName: Get Data from Cognitive Services
//
// Parameters:
//   - libraryName: the name of the library to be used in the system message
//   - userQuery: the user query to be used for the query.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//
// Returns:
//   - response: the response from the cognitive services as a string
func GetDataFromCognitiveServices(libraryName string, userQuery string, maxRetrievalCount int) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_DATA_COGNITIVE_SERVICES - Input: libraryName=%s, userQuery=%s, maxRetrievalCount=%d", libraryName, userQuery, maxRetrievalCount)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_GET_DATA_COGNITIVE_SERVICES - Duration: %v", duration)
	}()

	ansysProduct := pyansysProduct["name"][libraryName]

	userMessage := fmt.Sprintf(`In %s: The following user query may be brief, ambiguous, or lacking technical detail.
		Please rewrite it as a clear, detailed, and specific question suitable for retrieving relevant and precise information from a technical knowledge base about {product}.
		If necessary, add clarifying context, standard terminology, or related technical concepts commonly used in {product} documentation, without changing the original intent of the user's question.

		User Query: "%s"

		Return your response as a JSON object with a single key "unified_query".
		For example:
		"unified_query": "<your generated query here>"`, ansysProduct, userQuery)

	historyMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: userMessage,
		},
	}

	// Make llm call to rewrite the query
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, "")
	messageJSON, err := jsonStringToObject(result)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_DATA_COGNITIVE_SERVICES - Output: (JSON parse error)")
		return ""
	}
	rewrittenQuery, ok := messageJSON["unified_query"].(string)
	if !ok {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_DATA_COGNITIVE_SERVICES - Output: (unified_query not string)")
		return ""
	}
	if rewrittenQuery == "" {
		rewrittenQuery = userQuery
	}
	// Make rest call to cognitive services
	jsonBody := fmt.Sprintf(`{"query": "%s", "product": "%s", "top_k": %d}`, rewrittenQuery, libraryName, maxRetrievalCount)
	endpoint := "https://codegen-rm.azurewebsites.net/run_search"
	header := map[string]string{
		"Content-Type": "application/json",
	}
	success, returnJsonBody := SendRestAPICall("POST", endpoint, header, map[string]string{}, jsonBody)
	if !success {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_DATA_COGNITIVE_SERVICES - Output: (API request failed)")
		return ""
	}
	// Make llm call to process the response from cognitive services
	processingMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		Use the API definition and the related APIs found. Do your best to generate the code based on the information available.
		API Search Results: %s
		- STRICT: Only use the context provided in this system message. Do NOT think outside this context, do NOT add anything else, do NOT invent or hallucinate anything beyond the provided information.
		- Generate the code that solves the user query using only the API Search Results.
		- If you are not able to generate the code using the context provided, Send "I am not able to generate the code with the information provided."
		- If you are sure about the code, return the code in markdown format.
		- If you are not sure about the code, return "Please provide more information about the user query and the methods to be used."
		Respond with the following format, do not add anything else:
		The generated Python code only`, ansysProduct, returnJsonBody)
	processingHistoryMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: processingMessage,
		},
	}
	result, _ = PerformGeneralRequest(userQuery, processingHistoryMessage, false, "")
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_DATA_COGNITIVE_SERVICES - Output: %s", result)
	return result
}

// Function to get the raw data from congnitive services without any processing
//
// Tags:
//   - @displayName: Get Raw Data from Cognitive Services
//
// Parameters:
//   - libraryName: the name of the library to be used in the system message
//   - userQuery: the user query to be used for the query.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//
// Returns:
//   - response: the response from the cognitive services as a string
func GetRawDataFromCognitiveServices(libraryName string, userQuery string, maxRetrievalCount int) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Input: libraryName=%s, userQuery=%s, maxRetrievalCount=%d", libraryName, userQuery, maxRetrievalCount)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Duration: %v", duration)
	}()

	ansysProduct := pyansysProduct["name"][libraryName]

	// 0. Rewrite user query
	userMessage := fmt.Sprintf(`In %s: The following user query may be brief, ambiguous, or lacking technical detail.
		Please rewrite it as a clear, detailed, and specific question suitable for retrieving relevant and precise information from a technical knowledge base about {product}.
		If necessary, add clarifying context, standard terminology, or related technical concepts commonly used in {product} documentation, without changing the original intent of the user's question.

		User Query: "%s"

		Return your response as a JSON object with a single key "unified_query".
		For example:
		"unified_query": "<your generated query here>"`, ansysProduct, userQuery)

	historyMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: userMessage,
		},
	}

	// Make llm call to rewrite the query
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, "")
	messageJSON, err := jsonStringToObject(result)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (JSON parse error)")
		return ""
	}
	rewrittenQuery, ok := messageJSON["unified_query"].(string)
	if !ok {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (unified_query not string)")
		return ""
	}
	if rewrittenQuery == "" {
		rewrittenQuery = userQuery
	}

	// 1. Get embedding
	embReq, _ := json.Marshal(map[string]string{
		"model": "text-embedding-3-large",
		"input": rewrittenQuery,
	})

	req, _ := http.NewRequest("POST",
		config.GlobalConfig.AZURE_EMBEDDING_URL,
		bytes.NewBuffer(embReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.GlobalConfig.AZURE_EMBEDDING_TOKEN)

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var embResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&embResp)
	embedding := embResp["data"].([]interface{})[0].(map[string]interface{})["embedding"]

	// 2. Vector search
	searchReq, _ := json.Marshal(map[string]interface{}{
		"vectorQueries": []map[string]interface{}{{
			"kind": "vector", "k": maxRetrievalCount, "vector": embedding, "fields": "content_vctr",
		}},
		"filter": fmt.Sprintf("product eq '%s' and version eq '%s' and typeOFasset ne 'documentation'", libraryName, pyansysProduct["version"][libraryName]),
		"top":    5,
		"select": "content,product,physics,sourceURL_lvl1,sourceTitle_lvl1,typeOFasset",
	})

	req, err = http.NewRequest("POST",
		config.GlobalConfig.AZURE_COGNITIVE_SERVICE_API,
		bytes.NewBuffer(searchReq))
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (search request error)")
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.GlobalConfig.AZURE_COGNITIVE_SERVICE_TOKEN)

	resp, err = client.Do(req)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (search error)")
		return ""
	}
	defer resp.Body.Close()

	var searchResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&searchResp)

	// 3. Format and print
	results := searchResp["value"].([]interface{})
	chunks := make(map[string]interface{})

	for i, result := range results {
		r := result.(map[string]interface{})
		chunks[fmt.Sprintf("chunk_%d", i+1)] = map[string]interface{}{
			"context":          r["content"],
			"product":          r["product"],
			"physics":          r["physics"],
			"sourceURL_lvl1":   r["sourceURL_lvl1"],
			"sourceTitle_lvl1": r["sourceTitle_lvl1"],
			"typeOfAsset":      r["typeOFasset"],
		}
	}

	output, err := json.MarshalIndent(chunks, "", "  ")
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: (JSON marshal error)")
		return ""
	}

	// // 4. Process the output
	// processingMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
	// 	Use the API definition and the related APIs found. Do your best to generate the code based on the information available.
	// 	API Search Results: %s
	// 	- STRICT: Only use the context provided in this system message. Do NOT think outside this context, do NOT add anything else, do NOT invent or hallucinate anything beyond the provided information.
	// 	- Generate the code that solves the user query using only the API Search Results.
	// 	- If you are not able to generate the code using the context provided, Send "I am not able to generate the code with the information provided."
	// 	- If you are sure about the code, return the code in markdown format.
	// 	- If you are not sure about the code, return "Please provide more information about the user query and the methods to be used."
	// 	Respond with the following format, do not add anything else:
	// 	The generated Python code only`, ansysProduct, string(output))
	// processingHistoryMessage := []sharedtypes.HistoricMessage{
	// 	sharedtypes.HistoricMessage{
	// 		Role:    "user",
	// 		Content: processingMessage,
	// 	},
	// }
	// result, _ = PerformGeneralRequest(userQuery, processingHistoryMessage, false, "")
	// logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GET_RAW_DATA_COGNITIVE_SERVICES - Output: %s", result)
	// return result

	return string(output)
}

// Preprocess the user input for better LLM understanding
//
// Tags:
//   - @displayName: User Query Rewriting
//
// Parameters:
// - userQuery: The original user query to be rewritten.
// - libraryName: The name of the library being queried.
// - historyMessages: the history of messages to be used in the query
//
// Returns:
// - The rewritten user query.
func PreprocessTheInput(userQuery string, libraryName string, historyMessages []sharedtypes.HistoricMessage) string {
	startTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_PREPROCESS_INPUT - Input: userQuery=%s, libraryName=%s", userQuery, libraryName)
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_PREPROCESS_INPUT - Duration: %v", duration)
	}()

	userQuery = strings.TrimSpace(userQuery)
	ansysProduct := pyansysProduct["name"][libraryName]

	// userMessage := fmt.Sprintf(`In %s: The following user query may be brief, ambiguous, or lacking technical detail.
	// 	Please rewrite it as a clear, detailed, and specific question suitable for retrieving relevant and precise information from a technical knowledge base about %s.
	// 	If necessary, add clarifying context, standard terminology, or related technical concepts commonly used in %s documentation, without changing the original intent of the user's question.

	// 	User Query: "%s"

	// 	Return your response as a JSON object with a single key "unified_query".
	// 	For example:
	// 	"unified_query": "<your generated query here>"`, ansysProduct, ansysProduct, ansysProduct, userQuery)

	userMessage := fmt.Sprintf(`In %s: Your task is to deeply analyze the provided history (last 5 chat messages and assistant responses) to understand the user's context, intent, and requirements. Use this context to rewrite the user query so it is fully optimized for searching the Vector Database for relevant Methods or Examples.
History: %v
User Query: %s
Do not remove any relevant information from the original query. If the history clarifies or expands the user's intent, incorporate that context into the unified query.
Return your response as a JSON object with a single key "unified_query".
Start the unififed query with 'Using %s,' if not mentioned
For example:
"unified_query": "<your generated query here>"
`, ansysProduct, historyMessages, userQuery, libraryName)
	historyMessage := append(historyMessages,
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: userMessage,
		},
	)

	// Make llm call to rewrite the query
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, "")
	messageJSON, err := jsonStringToObject(result)
	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_PREPROCESS_INPUT - Output: %s (JSON parse error)", userQuery)
		return userQuery
	}
	rewrittenQuery, ok := messageJSON["unified_query"].(string)
	if !ok {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_PREPROCESS_INPUT - Output: %s (unified_query not string)", userQuery)
		return userQuery
	}
	if rewrittenQuery == "" {
		rewrittenQuery = userQuery
	}
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_PREPROCESS_INPUT - Output: %s", rewrittenQuery)
	return rewrittenQuery
}

// Write welcome message based on library name
//
// Tags:
//   - @displayName: Welcome Message Generation
//
// Parameters:
// - libraryName: The name of the library for which the welcome message is generated.
//
// Returns:
// - The generated welcome message.
func GenerateWelcomeMessage(libraryName string) string {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_GENERATE_WELCOME_MESSAGE - Duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GENERATE_WELCOME_MESSAGE - Input: libraryName=%s", libraryName)

	result := fmt.Sprintf("Welcome to the %s chatbot! How can I assist you today?", libraryName)

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_GENERATE_WELCOME_MESSAGE - Output: %s", result)
	return result
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
	if libraryName == "" {
		libraryName = "pyfluent"
	}
	// Sample json body
	// `{"key": "value", "number": 123}`
	jsonBody := fmt.Sprintf(`{"query": "%s", "product": "%s"}`, query, libraryName)
	if requestType == "" {
		requestType = "POST"
	}
	if endpoint == "" {
		endpoint = "https://dev-codegen.azurewebsites.net/code_gen"
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
		logging.Log.Infof(&logging.ContextMap{}, "API request successful, code: %s", code)
		code = PerformGeneralRequestNoStreaming("The code generated is: "+code, []sharedtypes.HistoricMessage{}, "You are a helpful assistant that helps to generate python code in markdown format. Do not add anything else, do not add any extra keys, no extra texts, or formatting (including no code fences). Remove the docs in the code and only provide the code.")
		return code
	}
	return ""
}

// API to get the data from congnitive services
//
// Tags:
//   - @displayName: Get Data from Cognitive Services
//
// Parameters:
//   - libraryName: the name of the library to be used in the system message
//   - userQuery: the user query to be used for the query.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//
// Returns:
//   - response: the response from the cognitive services as a string
func GetDataFromCognitiveServices(libraryName string, userQuery string, maxRetrievalCount int) string {
	startTime := time.Now()
	ansysProduct := pyansysProduct[libraryName]
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND GetDataFromCognitiveServices - started %v", time.Now())
	defer func() {
		duration := time.Since(startTime)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GetDataFromCognitiveServices COMPLETED - duration: %v", duration)
	}()

	userMessage := fmt.Sprintf(`In %s: The following user query may be brief, ambiguous, or lacking technical detail.
		Please rewrite it as a clear, detailed, and specific question suitable for retrieving relevant and precise information from a technical knowledge base about {product}.
		If necessary, add clarifying context, standard terminology, or related technical concepts commonly used in {product} documentation, without changing the original intent of the user's question.

		User Query: "%s"

		Return your response as a JSON object with a single key "unified_query".
		For example:
		"unified_query": "<your generated query here>"`, ansysProduct, userQuery)

	historyMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: userMessage,
		},
	}

	// Make llm call to rewrite the query
	llmStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GetDataFromCognitiveServices - LLM request (rewrite query) STARTED")
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, "")
	llmDuration := time.Since(llmStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GetDataFromCognitiveServices - LLM request (rewrite query) COMPLETED - duration: %v", llmDuration)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND GetDataFromCognitiveServices - llm ended %v", time.Now())
	messageJSON, err := jsonStringToObject(result)
	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		return ""
	}
	rewrittenQuery, ok := messageJSON["unified_query"].(string)
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "unified_query is not a string")
		return ""
	}
	if rewrittenQuery == "" {
		logging.Log.Warn(&logging.ContextMap{}, "Rewritten query is empty, using original query: %s", userQuery)
		rewrittenQuery = userQuery
	}
	logging.Log.Infof(&logging.ContextMap{}, "Rewritten query: %s", rewrittenQuery)
	// Make rest call to cognitive services
	jsonBody := fmt.Sprintf(`{"query": "%s", "product": "%s", "top_k": %d}`, rewrittenQuery, libraryName, maxRetrievalCount)
	endpoint := "https://codegen-rm.azurewebsites.net/run_search"
	header := map[string]string{
		"Content-Type": "application/json",
	}
	success, returnJsonBody := SendRestAPICall("POST", endpoint, header, map[string]string{}, jsonBody)
	logging.Log.Infof(&logging.ContextMap{}, "success: %v, returnJsonBody: %s", success, returnJsonBody)
	if !success {
		logging.Log.Errorf(&logging.ContextMap{}, "API request to cognitive services failed")
		return ""
	}
	// Make llm call to process the response from cognitive services
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND GetDataFromCognitiveServices - rest call ended %v", time.Now())
	llmProcessingStartTime := time.Now()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GetDataFromCognitiveServices - LLM request (process response) STARTED")
	processingMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		Use the API definition and the related APIs found. Do your best to generate the code based on the information available.
		API Search Results: %s
		- STRICT: Only use the context provided in this system message. Do NOT think outside this context, do NOT add anything else, do NOT invent or hallucinate anything beyond the provided information.
		- Generate the code that solves the user query using only the API Search Results.
		- If you are not able to generate the code using the context provided, Send "I am not able to generate the code with the information provided."
		- If you are sure about the code, return the code in markdown format.
		- If you are not sure about the code, return "Please provide more information about the user query and the methods to be used."
		Respond with the following format, do not add anything else:
		The generated Python code only`, ansysProduct, returnJsonBody)
	processingHistoryMessage := []sharedtypes.HistoricMessage{
		sharedtypes.HistoricMessage{
			Role:    "user",
			Content: processingMessage,
		},
	}
	result, _ = PerformGeneralRequest(userQuery, processingHistoryMessage, false, "")
	llmProcessingDuration := time.Since(llmProcessingStartTime)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING: GetDataFromCognitiveServices - LLM request (process response) COMPLETED - duration: %v", llmProcessingDuration)
	logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING_IND GetDataFromCognitiveServices - llm processing ended %v", time.Now())
	logging.Log.Infof(&logging.ContextMap{}, "result: %s", result)
	return result
}
