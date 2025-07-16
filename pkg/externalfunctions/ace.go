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
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/qdrant/go-client/qdrant"
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
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		As a first step, you need to search the Examples Vector DB to find any relevant examples. Check if the examples contain enough information to generate the code.
		If you are sure that the examples are enough, return "true". If you need more examples, return "false".

		The format in the following JSON format, do not add anything else (no extra keys, no extra texts, or formatting (including no code fences)):
		boolean`, ansysProduct)

	outputFields := []string{"text", "document_name", "previous_chunk", "next_chunk", "guid"}
	historyMessage := []sharedtypes.HistoricMessage{}
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, userQuery, denseWeight, sparseWeight, "")
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
	result, _ := PerformGeneralRequest(exampleString, historyMessage, false, systemMessage)

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
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided. As a first step, you need to search the Ansys API Reference Vector DB to find the relevant Method. Return the optimal search query to search the %s API Reference vector database. Make sure that you do not remove any relevant information from the original query. The format in the following JSON format, do not add anything else (no extra keys, no extra texts, or formatting (including no code fences)): {{ 'response': 'optimal vector db search query' }}`, ansysProduct, ansysProduct)

	multipleCandidatesSystemMessage := fmt.Sprintf(`In Ansys Fluent-Pyfluent you must create a script to efficiently execute the instructions provided. Propose multiple candidate queries (up to 5 highly relevant variations) that will be useful to completing the user's instruction. If available, scour the User Guide table of contents that can help you generate domain-relevant queries: %s IMPORTANT: - Do not remove any critical domain terms from the user's query. - NO FILLER WORDS OR PHRASES. - Localize to the user's intent if possible (e.g., structural or thermal context). - Keep your answer under 5 meaningful variations max. Return them in valid JSON with the following structure exactly (no extra keys, no extra texts, or formatting (including no code fences)): {{ "candidate_queries": [ "query_variant_1", "query_variant_2", "... up to 5" ] }} `, tableOfContents)

	historyMessage := []sharedtypes.HistoricMessage{}

	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, multipleCandidatesSystemMessage)

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
		rankingSystemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided. You have proposed multiple potential queries. Now, please rank these queries in terms of likely effectiveness for searching the %s API to fulfill the user's intent. Then return only the best overall query in JSON (no extra keys, no extra texts, or formatting (including no code fences)). Format: {{ 'response': 'the single best query to scour the API reference to generate code'}} Consider which query would retrieve the most relevant methods or functionalities.`, ansysProduct, ansysProduct)

		var candidateBuilder strings.Builder
		for i, query := range candidateQueries {
			if i > 0 {
				candidateBuilder.WriteString("\n")
			}
			candidateBuilder.WriteString(fmt.Sprintf(`"- %s"`, query))
		}
		candidateQueriesString := candidateBuilder.String()

		result, _ := PerformGeneralRequest("Candidate queries:\n"+candidateQueriesString, historyMessage, false, rankingSystemMessage)

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
		logging.Log.Error(&logging.ContextMap{}, "Best query is empty")
		result, _ := PerformGeneralRequest(userQuery, historyMessage, false, systemMessage)
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
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, bestQuery, denseWeight, sparseWeight, "")

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
}

func checkWhetherOneOfTheMethodsFits(collectionName string, historyMessage []sharedtypes.HistoricMessage, ansysProduct string, denseWeight float64, sparseWeight float64, maxRetrievalCount int, methods string) string {
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
        In this step you must decide whether one of the options provided is unambiguously the right one. If so, return the full path of the MethodW. Otherwise return the explanation for the ambiguity.

        The format is as follows: "<full path of the Method, is mandatory to include the signature with parameters if present>"

        Important: If "unambiguous_method_found" is true, "unambiguous_method_path" must be provided.`, ansysProduct)
	message, _ := PerformGeneralRequest(methods, historyMessage, false, systemMessage)

	return message
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
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
		In this step you must find the right chapter / subchapter within the Ansys User Guide to retrieve more information in order to identify the correct Method. Return Chapter Details of the chapters you want to look into. Order the section names by relevance.
		Do focus on the task, Interface and Introduction sections should be ignored.
		If the section match isn't perfect, that's okay. The goal is to find the right chapter / subchapter to look into, so return a boolean to indicate if the section requires references.
		If you already selected a subchapter in a previous step, please don't select it again, explore other options from the table of contents.
		Respond in the following JSON format. Do only return the json object without any quotes around it (no extra keys, no extra texts, or formatting (including no code fences)):
		[
			"index":  "<Index of the Chapter>",
			"sub_chapter_name": "<Name of the Chapter / Subchapter>",
			"section_name": "<complete path as listed in table of contents>",
			"get_references": "<boolean true or false>"
		]
		Output Format is as follows:
		[
			{
				"index": "18.5.1",
				"sub_chapter_name": "Structural Results",
				"section_name": "ds_using_select_results_structural_types.xml::Deformation",
				"get_references": true
			},
			{
				"index": "17.10.1",
				"sub_chapter_name": "Using Result Trackers",
				"section_name": "ds_using_solve.xml::Structural Result Trackers",
				"get_references": false
			},
			{
				"index": "17.8",
				"sub_chapter_name": "Solving",
				"section_name": "ds_using_solve.xml::Specifying Solution Information",
				"get_references": true
			}
		]`, ansysProduct)
	message, _ := PerformGeneralRequest("User Guide Table of Contents:\n"+tableOfContentsString, historyMessage, false, systemMessage)
	// messageJSON is expected to be a slice of map[string]interface{} (JSON array)
	var chapters []map[string]interface{}
	err := json.Unmarshal([]byte(message), &chapters)
	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON array: %v", err)
		return ""
	}
	uniqueSection := map[string]map[string]interface{}{}
	for _, item := range chapters {
		name, ok := item["sub_chapter_name"].(string)
		if !ok {
			continue
		}
		if _, exists := uniqueSection[name]; !exists {
			uniqueSection[name] = item
		}
	}

	// Retrieve content for each unique section
	var guideSectionsBuilder strings.Builder
	for _, item := range uniqueSection {
		sectionName, ok := item["section_name"].(string)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "section_name is not a string")
			continue
		}
		subChapterName, ok := item["sub_chapter_name"].(string)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "sub_chapter_name is not a string")
			continue
		}
		index, ok := item["index"].(string)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "index is not a string")
			continue
		}
		guideSectionsBuilder.WriteString(fmt.Sprintf("Index: %s, Title: %s, Section Name: %s\n", sectionName, subChapterName, index))
		getReferences, ok := item["get_references"].(bool)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "get_references is not a boolean")
			continue
		}
		userResponse := ""
		// print collectionName
		if getReferences {
			// Query the user guide name
			scoredPoints := queryUserGuideName(sectionName, uint64(3), collectionName)
			realSectionName := sectionName
			if len(scoredPoints) > 0 {
				// Append the scored point to sections
				payload := scoredPoints[0].GetPayload()
				userResponse = "With section texts: " + payload["text"].GetStringValue() + "\n"
				realSectionName = payload["section_name"].GetStringValue()
			} else {
				logging.Log.Warnf(&logging.ContextMap{}, "No results found for section: %s", sectionName)
			}

			// Escape the string parameter properly for Cypher
			escapedSectionName := strings.ReplaceAll(realSectionName, `\`, `\\`)
			escapedSectionName = strings.ReplaceAll(escapedSectionName, `"`, `\"`)
			query := fmt.Sprintf("MATCH (n:UserGuide {name: \"%s\"})-[:References]->(reference) RETURN reference.name AS section_name", escapedSectionName)
			parameters := aali_graphdb.ParameterMap{}
			result := GeneralGraphDbQuery(query, parameters)
			if len(result) == 0 {
				logging.Log.Warnf(&logging.ContextMap{}, "No references found for section: %s", sectionName)
				continue
			}
			// Append section name from result
			logging.Log.Warnf(&logging.ContextMap{}, "Found references for collectionName: %s", collectionName)
			for index, record := range result {
				if index > 2 {
					break
				}
				sectionName := record["section_name"]
				userResponse += "With references: " + sectionName.(string) + "\n"
				sections := queryUserGuideName(sectionName.(string), uint64(3), collectionName)
				// Initialize an empty string to store the content
				content := ""

				// Iterate through the retrieved sections and append their text content
				for _, section := range sections {
					if section.Payload["text"].GetStringValue() != "" {
						content += section.Payload["text"].GetStringValue() + "\n"
					}
				}
				userResponse += "With reference section texts: " + content + "\n"
			}

		} else {
			logging.Log.Warnf(&logging.ContextMap{}, "Skipping references for section: %s", sectionName)
			scoredPoints := queryUserGuideName(sectionName, uint64(5), collectionName)
			content := ""
			if len(scoredPoints) > 0 {
				// Append the scored point to sections
				payload := scoredPoints[0].Payload
				content += payload["text"].GetStringValue() + "\n"
			} else {
				logging.Log.Warnf(&logging.ContextMap{}, "No results found for section: %s", sectionName)
			}
			userResponse = content
		}
		guideSectionsBuilder.WriteString(userResponse)
		guideSectionsBuilder.WriteString("\n\n\n-------------------\n\n\n")
	}

	userGuideInformation := "Retrieved information from user guide:\n\n\n" + guideSectionsBuilder.String()
	return checkWhetherUserInformationFits(ansysProduct, userGuideInformation, historyMessage)
}

func checkWhetherUserInformationFits(ansysProduct string, userQuery string, historyMessage []sharedtypes.HistoricMessage) string {
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
			In this step, you must evaluate the information retrieved from the User Guide and decide whether you have enough information to unambiguously identify the correct Method or whether you need some user input.
			If you need user input, return the query you would like to ask the user. If you have enough information, return the full path of the Method.
			Don't ask the user for information that is already provided in the query %s.
			
			Respond with the following JSON format: "<full path of the Method, is mandatory to include the signature with parameters if present>
			`, ansysProduct, userQuery)
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, systemMessage)

	return result
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
	nresult := getElementByName(methodName, "Method")
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
	examples := getExampleNodesFromElement("Method", methodName, collectionName)
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
	logging.Log.Infof(&logging.ContextMap{}, "Generating code for ansysProduct: %s, methods: %s, examples: %s, methods_from_user_guide: %s", ansysProduct, methods, examples, methods_from_user_guide)
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
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

	result, _ := PerformGeneralRequest(userQuery, historyMessages, false, systemMessage)

	if result == "" {
		return result
	}

	// Format the result as markdown code block
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

// countPlaceholders counts the number of %s placeholders in the input string
func countPlaceholders(input string) int {
	count := 0
	for i := 0; i < len(input)-1; i++ {
		if input[i] == '%' && input[i+1] == 's' {
			count++
		}
	}
	return count
}

// buildReplacements creates a slice of replacement values based on the count needed
func buildReplacements(count int, args StringReplacementArgs) []any {
	switch count {
	case 1:
		return []any{args.Placeholder1}
	case 2:
		return []any{args.Placeholder1, args.Placeholder2}
	case 3:
		return []any{args.Placeholder1, args.Placeholder2, args.Placeholder3}
	case 4:
		return []any{args.Placeholder1, args.Placeholder2, args.Placeholder3, args.Placeholder4}
	case 5:
		return []any{args.Placeholder1, args.Placeholder2, args.Placeholder3, args.Placeholder4, args.Placeholder5}
	default:
		return []any{}
	}
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
func ConvertJSONToCustomize(object []map[string]any) string {
	return convertJSONToCustomizeHelper(object, 0, "")
}

// convertJSONToCustomizeHelper is an internal helper with all parameters
func convertJSONToCustomizeHelper(object []map[string]any, level int, currentIndex string) string {
	var nodeBuilder strings.Builder
	for _, item := range object {
		chapters, ok := item["chapters"].([]interface{})
		if !ok {
			fmt.Println("Skipping item: not a chapter list")
			continue
		}
		for idx, chapter := range chapters {
			currentIndex := fmt.Sprintf("0%d.", idx+1)
			chapterMap, ok := chapter.(map[string]interface{})
			if !ok {
				fmt.Println("Skipping chapter: not a map")
				continue
			}
			nodeBuilder.WriteString(fmt.Sprintf(
				"%s%s %s (section Name -> %s)\n",
				repeatString("  ", level),
				currentIndex,
				chapterMap["title"],
				chapterMap["name"],
			))
		}
	}
	return nodeBuilder.String()
}

// repeatString returns a string with s repeated count times.
// Uses strings.Repeat for efficiency.
func repeatString(s string, count int) string {
	return strings.Repeat(s, count)
}

// // SearchElements performs a general query in the KnowledgeDB.
// //
// // The function returns the query results.
// //
// // Tags:
// //   - @displayName: Search Elements
// //
// // Parameters:
// //   - collectionName: the name of the collection to which the data objects will be added.
// //   - maxRetrievalCount: the maximum number of results to be retrieved.
// //   - queryString: the query string to be used for the query.
// //   - denseWeight: the weight for the dense vector. (default: 0.9)
// //   - sparseWeight: the weight for the sparse vector. (default: 0.1)
// //
// // Returns:
// //   - userResponse: the formatted user response string
// func SearchElements(collectionName string, maxRetrievalCount int, queryString string, denseWeight float64, sparseWeight float64) string {
// 	outputFields := []string{"name"}
// 	nodeType := "Method" // Specify the node type to filter by
// 	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, queryString, denseWeight, sparseWeight, nodeType)
// 	// Format results as requested
// 	elements := []string{}
// 	for _, scoredPoint := range scoredPoints {
// 		entry := scoredPoint.Payload
// 		name := entry["name"].GetStringValue()
// 		result := getElementByName(name, nodeType)
// 		for _, nentry := range result {
// 			entry := nentry["n"].(map[string]interface{})
// 			namePseudocode := entry["name_pseudocode"]
// 			completeName := entry["name"]
// 			remarks := entry["remarks"]
// 			summary := entry["summary"]
// 			parameters := entry["parameters"]
// 			element := fmt.Sprintf("%s ; %s ; %s ; Remarks: %s; Parameters: %s\n\n", namePseudocode, completeName, summary, remarks, parameters)
// 			elements = append(elements, element)
// 		}
// 	}
// 	userResponse := joinStrings(elements, "\n")
// 	return userResponse
// }

// joinStrings joins a slice of strings with the given separator.
// Uses strings.Join for efficiency.
func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

func getExampleReferences(baseSearchNodeComplete string, db string) (string, []interface{}) {
	var exampleNamesBuilder strings.Builder
	exampleReferencesInformation := []interface{}{}
	// Escape the string parameter properly for Cypher
	escapedName := strings.ReplaceAll(baseSearchNodeComplete, `"`, `\"`)
	query := fmt.Sprintf(`MATCH (root:Example {name: "%s"})-[r]-(neighbor) RETURN root.name AS rootName, label(r) AS relationshipType, r AS relationshipProps, neighbor.name AS neighborName, label(neighbor) AS neighborLabel, neighbor.parameters AS neighborParameters, neighbor.remarks AS neighborRemarks, neighbor.return_type AS neighborReturn, neighbor.summary AS neighborSummary`, escapedName)
	parameters := aali_graphdb.ParameterMap{}
	result := GeneralGraphDbQuery(query, parameters)
	for _, relationship := range result {
		element := relationship["neighborName"]
		elementType := relationship["neighborLabel"]
		if elementType == nil {
			elementType = "Unknown" // default value if not found
		}
		exampleNamesBuilder.WriteString(fmt.Sprintf("This example uses %s as a %s\n", element, elementType))
		referenceParameters := relationship["neighborParameters"]
		if referenceParameters == nil {
			referenceParameters = "No parameters available."
		}
		referenceRemarks := relationship["neighborRemarks"]
		if referenceRemarks == nil {
			referenceRemarks = "No remarks available."
		}
		referenceReturns := relationship["neighborReturn"]
		if referenceReturns == nil {
			referenceReturns = "No return available."
		}
		referenceSummary := relationship["neighborSummary"]
		if referenceSummary == nil {
			referenceSummary = "No summary available"
		}
		referencesInformation := map[string]any{
			"reference_name":       element,
			"reference_type":       elementType,
			"reference_parameters": referenceParameters,
			"reference_remarks":    referenceRemarks,
			"reference_returns":    referenceReturns,
			"reference_summary":    referenceSummary,
		}
		exampleReferencesInformation = append(exampleReferencesInformation, referencesInformation)
	}

	return exampleNamesBuilder.String(), exampleReferencesInformation
}

func getExampleNodesFromElement(baseSearchType string, baseSearchNodeComplete string, collectionName string) []map[string]interface{} {
	// Escape the string parameters properly for Cypher
	escapedNodeComplete := strings.ReplaceAll(baseSearchNodeComplete, `"`, `\"`)
	escapedType := strings.ReplaceAll(baseSearchType, `"`, `\"`)
	query := fmt.Sprintf(`MATCH (n:Element) <-[:Uses]- (example:Example)
			WHERE n.name = "%s" AND n.type = "%s"
			RETURN example
			`, escapedNodeComplete, escapedType)

	parameters := aali_graphdb.ParameterMap{}

	result := GeneralGraphDbQuery(query, parameters)
	preparedExample := []map[string]interface{}{}
	for _, relationship := range result {
		element := relationship["example"]
		if element == nil {
			logging.Log.Error(&logging.ContextMap{}, "Element is nil")
			continue
		}
		example := element.(map[string]interface{})
		exampleName := example["name"].(string)
		exampleText := ""
		orderedChunks := queryExample(exampleName, collectionName)
		for _, chunk := range orderedChunks {
			exampleText += chunk["text"].(string)
		}
		preparedExample = append(preparedExample, map[string]interface{}{
			"name": exampleName,
			"text": exampleText,
		})
	}
	return preparedExample
}

func queryExample(exampleName string, collectionName string) []map[string]interface{} {
	// search database
	client, err := qdrant_utils.QdrantClient()

	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error creating Qdrant client: %v", err)
		return []map[string]interface{}{}
	}
	resultCount := uint64(1000)
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadInclude([]string{"text", "document_name", "previous_chunk", "next_chunk", "guid"}...),
		Query:          nil,
		Limit:          &resultCount,
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatchKeyword("document_name", exampleName),
			},
		},
	}
	unorderedDictionary := map[string]interface{}{}
	firstChunk := map[string]interface{}{}
	newEntry := map[string]interface{}{}

	scoredPoints, err := client.Query(context.TODO(), &query)

	for _, scoredPoint := range scoredPoints {
		payload := scoredPoint.GetPayload()
		newEntry = map[string]interface{}{
			"text":           payload["text"].GetStringValue(),
			"document_name":  payload["document_name"].GetStringValue(),
			"previous_chunk": payload["previous_chunk"].GetStringValue(),
			"next_chunk":     payload["next_chunk"].GetStringValue(),
			"guid":           payload["guid"].GetStringValue(),
		}
		unorderedDictionary[payload["guid"].GetStringValue()] = newEntry

		if newEntry["previous_chunk"] == "" {
			firstChunk = newEntry
		}
	}

	nextEntryGUID := firstChunk["guid"].(string)

	output := []map[string]interface{}{firstChunk}
	nextEntry := map[string]interface{}{}

	if nextEntryGUID != "" || len(nextEntryGUID) > 0 {
		nextEntry = unorderedDictionary[nextEntryGUID].(map[string]interface{})
		output = append(output, nextEntry)
		nextEntryGUID = nextEntry["next_chunk"].(string)

		return output
	}

	return output

}

func queryUserGuideName(name string, resultCount uint64, collectionName string) []*qdrant.ScoredPoint {
	client, err := qdrant_utils.QdrantClient()
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload: qdrant.NewWithPayloadInclude([]string{"document_name",
			"section_name",
			"previous_chunk",
			"next_chunk",
			"text",
			"level",
			"parent_section_name",
			"guid"}...),
		Query: nil,
		Limit: &resultCount,
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatchKeyword("section_name", name),
			},
		},
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(&logging.ContextMap{}, "error in qdrant query: %q", err)
	}
	var results []*qdrant.ScoredPoint
	for _, scoredPoint := range scoredPoints {
		results = append(results, scoredPoint)
	}
	return results
}

func getDocumentation(baseSearchNodeComplete string, db string) (string, []interface{}) {
	var exampleNamesBuilder strings.Builder
	exampleReferencesInformation := []interface{}{}
	// Escape the string parameter properly for Cypher
	escapedName := strings.ReplaceAll(baseSearchNodeComplete, `"`, `\"`)
	query := fmt.Sprintf(`MATCH (root:Example {name: "%s"})-[r]-(neighbor) RETURN root.name AS rootName, label(r) AS relationshipType, r AS relationshipProps, neighbor.name AS neighborName, label(neighbor) AS neighborLabel, neighbor.parameters AS neighborParameters, neighbor.remarks AS neighborRemarks, neighbor.return_type AS neighborReturn, neighbor.summary AS neighborSummary`, escapedName)
	parameters := aali_graphdb.ParameterMap{}
	result := GeneralGraphDbQuery(query, parameters)
	for _, relationship := range result {
		element := relationship["neighborName"]
		elementType := relationship["neighborLabel"]
		if elementType == nil {
			elementType = "Unknown" // default value if not found
		}
		exampleNamesBuilder.WriteString(fmt.Sprintf("This example uses %s as a %s\n", element, elementType))
		referenceParameters := relationship["neighborParameters"]
		if referenceParameters == nil {
			referenceParameters = "No parameters available."
		}
		referenceRemarks := relationship["neighborRemarks"]
		if referenceRemarks == nil {
			referenceRemarks = "No remarks available."
		}
		referenceReturns := relationship["neighborReturn"]
		if referenceReturns == nil {
			referenceReturns = "No return available."
		}
		referenceSummary := relationship["neighborSummary"]
		if referenceSummary == nil {
			referenceSummary = "No summary available"
		}
		referencesInformation := map[string]any{
			"reference_name":       element,
			"reference_type":       elementType,
			"reference_parameters": referenceParameters,
			"reference_remarks":    referenceRemarks,
			"reference_returns":    referenceReturns,
			"reference_summary":    referenceSummary,
		}
		exampleReferencesInformation = append(exampleReferencesInformation, referencesInformation)
	}

	return exampleNamesBuilder.String(), exampleReferencesInformation
}

func PreprocessLLMJSON(s string) string {
	// Remove code fences and trim
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)

	// Extract only the first JSON object or array from the string
	start := strings.IndexAny(s, "{[")
	if start == -1 {
		return s // fallback, not found
	}

	// Find the matching closing bracket with proper nesting
	var end int
	if s[start] == '{' {
		end = findMatchingBrace(s, start)
	} else if s[start] == '[' {
		end = findMatchingBracket(s, start)
	}

	if end <= start {
		return s // fallback if no matching bracket found
	}

	jsonStr := s[start:end]

	// Clean up the JSON string step by step
	jsonStr = cleanupJSONString(jsonStr)

	return jsonStr
}

// findMatchingBrace finds the matching closing brace, handling strings properly
func findMatchingBrace(s string, start int) int {
	count := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		char := s[i]

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' && !escaped {
			inString = !inString
			continue
		}

		if !inString {
			if char == '{' {
				count++
			} else if char == '}' {
				count--
				if count == 0 {
					return i + 1
				}
			}
		}
	}
	return -1
}

// findMatchingBracket finds the matching closing bracket, handling strings properly
func findMatchingBracket(s string, start int) int {
	count := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		char := s[i]

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' && !escaped {
			inString = !inString
			continue
		}

		if !inString {
			if char == '[' {
				count++
			} else if char == ']' {
				count--
				if count == 0 {
					return i + 1
				}
			}
		}
	}
	return -1
}

// cleanupJSONString performs comprehensive cleanup of JSON string
func cleanupJSONString(jsonStr string) string {
	// Step 1: Fix single quotes in string values (but preserve them in Python code)
	jsonStr = fixSingleQuotes(jsonStr)

	// Step 2: Remove trailing commas before } and ]
	reTrailingComma := regexp.MustCompile(`,\s*([}\]])`)
	jsonStr = reTrailingComma.ReplaceAllString(jsonStr, "$1")

	// Step 3: Escape special characters in string values
	jsonStr = escapeStringValues(jsonStr)

	return jsonStr
}

// fixSingleQuotes replaces single quotes with double quotes only where appropriate
func fixSingleQuotes(s string) string {
	// Find JSON property values and fix single quotes only in the value part
	// Pattern: "property": 'value' -> "property": "value"
	re := regexp.MustCompile(`"([^"]+)":\s*'([^']*)'`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		parts := regexp.MustCompile(`"([^"]+)":\s*'([^']*)'`).FindStringSubmatch(match)
		if len(parts) == 3 {
			key := parts[1]
			value := parts[2]
			// Escape any double quotes in the value
			value = strings.ReplaceAll(value, `"`, `\"`)
			return fmt.Sprintf(`"%s": "%s"`, key, value)
		}
		return match
	})

	// Handle cases where property names also have single quotes
	// 'property': 'value' -> "property": "value"
	re2 := regexp.MustCompile(`'([^']+)':\s*'([^']*)'`)
	s = re2.ReplaceAllStringFunc(s, func(match string) string {
		parts := regexp.MustCompile(`'([^']+)':\s*'([^']*)'`).FindStringSubmatch(match)
		if len(parts) == 3 {
			key := parts[1]
			value := parts[2]
			// Escape any double quotes in the value
			value = strings.ReplaceAll(value, `"`, `\"`)
			return fmt.Sprintf(`"%s": "%s"`, key, value)
		}
		return match
	})

	return s
}

// escapeStringValues properly escapes string values in JSON
func escapeStringValues(s string) string {
	// Find all string values and escape them properly
	re := regexp.MustCompile(`"([^"]+)":\s*"([^"]*)"`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		parts := regexp.MustCompile(`"([^"]+)":\s*"([^"]*)"`).FindStringSubmatch(match)
		if len(parts) == 3 {
			key := parts[1]
			value := parts[2]

			// Escape backslashes first (they might be in file paths or Python code)
			value = strings.ReplaceAll(value, `\`, `\\`)

			// Don't double-escape already escaped quotes
			if !strings.Contains(value, `\"`) {
				// Escape unescaped quotes
				value = strings.ReplaceAll(value, `"`, `\"`)
			}

			return fmt.Sprintf(`"%s": "%s"`, key, value)
		}
		return match
	})

	return s
}

// Helper: Convert JSON string to map[string]interface{} in Go
func jsonStringToObject(jsonStr string) (map[string]interface{}, error) {
	clean := PreprocessLLMJSON(jsonStr)

	var obj map[string]interface{}
	err := json.Unmarshal([]byte(clean), &obj)

	if err != nil {
		// If first attempt fails, try additional cleanup
		logging.Log.Warnf(&logging.ContextMap{}, "First JSON parse failed, attempting additional cleanup: %v", err)

		// Try fixing common issues
		clean = fixCommonJSONIssues(clean)

		err = json.Unmarshal([]byte(clean), &obj)

		if err != nil {
			// Last resort: try to extract just the values manually
			logging.Log.Warnf(&logging.ContextMap{}, "Second JSON parse failed, attempting manual extraction: %v", err)
			obj = extractJSONManually(clean)
			if len(obj) > 0 {
				return obj, nil
			}
		}
	}

	return obj, err
}

// fixCommonJSONIssues attempts to fix additional common JSON formatting issues
func fixCommonJSONIssues(s string) string {
	// Fix unescaped newlines in string values
	re := regexp.MustCompile(`"([^"]+)":\s*"([^"]*\n[^"]*)"`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		parts := regexp.MustCompile(`"([^"]+)":\s*"([^"]*)"`).FindStringSubmatch(match)
		if len(parts) == 3 {
			key := parts[1]
			value := parts[2]
			value = strings.ReplaceAll(value, "\n", "\\n")
			value = strings.ReplaceAll(value, "\r", "\\r")
			value = strings.ReplaceAll(value, "\t", "\\t")
			return fmt.Sprintf(`"%s": "%s"`, key, value)
		}
		return match
	})

	// Fix boolean values that might be strings
	s = regexp.MustCompile(`"(true|false)"`).ReplaceAllString(s, "$1")

	// Fix number values that might be strings (but preserve actual string numbers)
	s = regexp.MustCompile(`:\s*"(\d+)"`).ReplaceAllString(s, `: $1`)
	s = regexp.MustCompile(`:\s*"(\d+\.\d+)"`).ReplaceAllString(s, `: $1`)

	return s
}

// extractJSONManually attempts to manually extract key-value pairs as a last resort
func extractJSONManually(s string) map[string]interface{} {
	obj := make(map[string]interface{})

	// Try to extract simple key-value pairs
	// Pattern: "key": "value" or "key": value
	re := regexp.MustCompile(`"([^"]+)":\s*(?:"([^"]*)"|([^,}\]]+))`)
	matches := re.FindAllStringSubmatch(s, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			key := match[1]
			var value interface{}

			if match[2] != "" {
				// String value
				value = match[2]
			} else if match[3] != "" {
				// Non-string value
				trimmed := strings.TrimSpace(match[3])
				switch trimmed {
				case "true":
					value = true
				case "false":
					value = false
				case "null":
					value = nil
				default:
					value = trimmed
				}
			}

			obj[key] = value
		}
	}

	return obj
}

func doHybridQuery(
	collectionName string,
	maxRetrievalCount int,
	outputFields []string,
	queryString string,
	denseWeight float64,
	sparseWeight float64,
	nodeType string) []*qdrant.ScoredPoint {

	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequestWithSparseDense(queryString, llmHandlerEndpoint, true, nil)
	defer close(responseChannel)

	// Process the first response and close the channel
	var embedding32 []float32
	var sparseVector []float32
	var indexVector []uint32

	var err error
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			if response.Error != nil && response.Error.Message != "" {
				panic(response.Error.Message)
			}
			panic("unknown error in embeddings response")
		}

		// Get embedded vector array (DENSE VECTOR)
		if response.EmbeddedData != nil {
			interfaceArray, ok := response.EmbeddedData.([]interface{})
			if !ok {
				errMessage := "error converting embedded data to interface array"
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
			embedding32, err = convertToFloat32Slice(interfaceArray)
			if err != nil {
				errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
		}

		// Get sparse vector
		if response.LexicalWeights != nil {
			sparseVectorInterface, ok := response.LexicalWeights.(map[string]interface{})
			if !ok {
				errMessage := "error converting lexical weights to interface array"
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
			sparseVector, indexVector, err = convertToSparseVector(sparseVectorInterface)
			if err != nil {
				errMessage := fmt.Sprintf("error converting sparse vector: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
		}

		// Mark that the first response has been received
		firstResponseReceived := true

		// Exit the loop after processing the first response
		if firstResponseReceived {
			break
		}
	}

	if len(embedding32) == 0 {
		logging.Log.Error(&logging.ContextMap{}, "No embeddings received from LLM handler")
		panic("No embeddings received from LLM handler")
	}

	if len(sparseVector) == 0 {
		logging.Log.Error(&logging.ContextMap{}, "No sparse vector received from LLM handler")
		panic("No sparse vector received from LLM handler")
	}

	if len(indexVector) == 0 {
		logging.Log.Error(&logging.ContextMap{}, "No index vector received from LLM handler")
		panic("No index vector received from LLM handler")
	}

	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	// perform the qdrant query
	limit := uint64(maxRetrievalCount)
	var filter *qdrant.Filter
	if nodeType != "" {
		filter = &qdrant.Filter{
			Must: []*qdrant.Condition{
				qdrant.NewMatchKeyword("type", nodeType),
			},
		}
	}

	using := "" // or "sparse_vector" based on the query type
	usingSparse := "sparse_vector"
	expression := qdrant.NewExpressionSum(&qdrant.SumExpression{
		Sum: []*qdrant.Expression{
			qdrant.NewExpressionMult(&qdrant.MultExpression{
				Mult: []*qdrant.Expression{
					qdrant.NewExpressionVariable("$score[0]"),  // dense score
					qdrant.NewExpressionConstant(float32(0.9)), // weight
				},
			}),

			// Another MultExpression: 0.25 * (tag match p,li)
			qdrant.NewExpressionMult(&qdrant.MultExpression{
				Mult: []*qdrant.Expression{
					qdrant.NewExpressionVariable("$score[1]"),   // sparse score
					qdrant.NewExpressionConstant(float32(0.12)), // weight
				},
			}),
		},
	})
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		Prefetch: []*qdrant.PrefetchQuery{
			{
				Limit:  &limit,
				Query:  qdrant.NewQueryDense(embedding32),
				Using:  &using,
				Filter: filter,
			},
			{
				Limit:  &limit,
				Query:  qdrant.NewQuerySparse(indexVector, sparseVector),
				Using:  &usingSparse,
				Filter: filter,
			},
		},
		WithVectors: qdrant.NewWithVectorsEnable(false),
		WithPayload: qdrant.NewWithPayloadInclude(outputFields...),
		Query: qdrant.NewQueryFormula(
			&qdrant.Formula{
				Expression: expression,
			},
		),
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}

	return scoredPoints
}

func getElementByName(nodeName string, nodeType string) []map[string]interface{} {
	// Escape the string parameters properly for Cypher
	escapedNodeName := strings.ReplaceAll(nodeName, `'`, `\'`)
	escapedNodeType := strings.ReplaceAll(nodeType, `'`, `\'`)
	query := fmt.Sprintf("MATCH (n:Element) WHERE n.name = '%s' AND n.type = '%s' RETURN n", escapedNodeName, escapedNodeType)
	result := GeneralGraphDbQuery(query, aali_graphdb.ParameterMap{})
	return result
}
