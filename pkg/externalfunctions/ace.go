// File: aali-flowkit/pkg/externalfunctions/ace.go
package externalfunctions

import (
	"context"
	"encoding/json"
	"fmt"

	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/qdrant/go-client/qdrant"
)

// StringReplaceWithArray replace a string with an array of strings
//
// Tags:
//   - @displayName: Replace a string with an array of strings
//
// Parameters:
//   - input: the input string
//   - replacements: the array of strings to replace
//
// Returns:
//   - the input string with the replacements applied
func StringReplaceWithArray(input string, placeholder1 string, placeholder2 string, placeholder3 string, placeholder4 string, placeholder5 string) string {

	// Count the number of placeholders in the input string
	placeholderCount := 0
	for i := 0; i < len(input)-1; i++ {
		if input[i] == '%' && input[i+1] == 's' {
			placeholderCount++
		}
	}

	// Perform the replacement
	var replacements []any
	if placeholderCount == 1 {
		replacements = []any{placeholder1}
	}
	if placeholderCount == 2 {
		replacements = []any{placeholder1, placeholder2}
	}
	if placeholderCount == 3 {
		replacements = []any{placeholder1, placeholder2, placeholder3}
	}
	if placeholderCount == 4 {
		replacements = []any{placeholder1, placeholder2, placeholder3, placeholder4}
	}
	if placeholderCount == 5 {
		replacements = []any{placeholder1, placeholder2, placeholder3, placeholder4, placeholder5}
	}
	input = fmt.Sprintf(input, replacements...)

	return input
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
	var nodeString string
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
			nodeString += fmt.Sprintf(
				"%s%s %s (section Name -> %s)\n",
				repeatString("  ", level),
				currentIndex,
				chapterMap["title"],
				chapterMap["name"],
			)
		}
	}
	return nodeString
}

func repeatString(s string, count int) string {
	var result string
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// WhetherOneOfTheMethodsFits performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Whether One Of The Methods Fits
//
// Parameters:
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - historyMessage: the history of messages to be used in the query
//
// Returns:
//   - updated historyMessage: the updated history message with the response from the LLM
//   - Method found: a boolean indicating whether an method was found
//   - method path: the full path of the Method if found
func WhetherOneOfTheMethodsFits(ansysProduct string, historyMessage []sharedtypes.HistoricMessage, optionsString string) ([]sharedtypes.HistoricMessage, bool, string) {
	// Define the system message guiding the LLM to evaluate the methods
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
        In this step you must decide whether one of the options provided is unambiguously the right one. If so, return the full path of the MethodW. Otherwise return the explanation for the ambiguity.

        The format is as follows:
        {{
            "unambiguous_method_found": true/false,
            "unambiguous_method_path": "<full path of the Method, is mandatory to include the signature with parameters if present>",
            "explanation": "empty / explanation for the ambiguity"
        }}

        Important: If "unambiguous_method_found" is true, "unambiguous_method_path" must be provided.
        `, ansysProduct)
	historyMessage = append(historyMessage, sharedtypes.HistoricMessage{Role: "user", Content: systemMessage})
	// Send the query to the LLM handler
	message, _ := PerformGeneralRequest("found the following Methods in the DB. Every line describes a new method\n"+optionsString, historyMessage, false, systemMessage)

	messageJSON, err := jsonStringToObject(message)

	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
	}

	unambiguousMethodFound, ok := messageJSON["unambiguous_method_found"].(bool)
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "unambiguous_method_found is not a boolean")
		panic("unambiguous_method_found is not a boolean")
	}

	methodPath, ok := messageJSON["unambiguous_method_path"].(string)
	if !ok && unambiguousMethodFound {
		logging.Log.Error(&logging.ContextMap{}, "unambiguous_method_path is not a string")
		panic("unambiguous_method_path is not a string")
	}

	logging.Log.Infof(&logging.ContextMap{}, "unambiguousMethodFound: %v, methodPath: %s", unambiguousMethodFound, methodPath)

	return historyMessage, unambiguousMethodFound, methodPath

}

// SearchElements performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Elements
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - queryString: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//
// Returns:
//   - userResponse: the formatted user response string
//   - elements: the formatted element strings as a slice
func SearchElements(collectionName string, maxRetrievalCount int, queryString string, denseWeight float64, sparseWeight float64) string {
	outputFields := []string{"name"}
	nodeType := "Method" // Specify the node type to filter by
	logging.Log.Infof(&logging.ContextMap{}, "Searching for elements with query: %s", queryString)
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, queryString, denseWeight, sparseWeight, nodeType)
	// Format results as requested
	elements := []string{}
	for _, scoredPoint := range scoredPoints {
		entry := scoredPoint.Payload
		name := entry["name"].GetStringValue()
		result := getElementByName(name, nodeType)
		for _, nentry := range result {
			entry := nentry["n"].(map[string]interface{})
			logging.Log.Infof(&logging.ContextMap{}, "GraphDB query result: %v", entry)
			namePseudocode := entry["name_pseudocode"]
			completeName := entry["name"]
			remarks := entry["remarks"]
			summary := entry["summary"]
			parameters := entry["parameters"]
			element := fmt.Sprintf("%s ; %s ; %s ; Remarks: %s; Parameters: %s\n\n", namePseudocode, completeName, summary, remarks, parameters)
			elements = append(elements, element)
		}
	}
	userResponse := joinStrings(elements, "\n")
	return userResponse
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

// SearchExamples performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Examples
//
// Parameters:
//   - tableOfContents: the table of contents to be used in the system message
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - queryString: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//
// Returns:
//   - elements: the formatted element strings
//   - userResponse: the formatted user response string
func SearchExamples(tableOfContents string, ansysProduct string, collectionName string, maxRetrievalCount int, queryString string, denseWeight float64, sparseWeight float64, userQuery string) string {
	system_message := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided. As a first step, you need to search the Ansys API Reference Vector DB to find the relevant Method. Return the optimal search query to search the %s API Reference vector database. Make sure that you do not remove any relevant information from the original query. The format in the following JSON format, do not add anything else (no extra keys, no extra texts, or formatting (including no code fences)): {{ 'response': 'optimal vector db search query' }}`, ansysProduct, ansysProduct)

	multiple_candidates_system_message := fmt.Sprintf(`In Ansys Fluent-Pyfluent you must create a script to efficiently execute the instructions provided. Propose multiple candidate queries (up to 5 highly relevant variations) that will be useful to completing the user's instruction. If available, scour the User Guide table of contents that can help you generate domain-relevant queries: %s IMPORTANT: - Do not remove any critical domain terms from the user's query. - NO FILLER WORDS OR PHRASES. - Localize to the user's intent if possible (e.g., structural or thermal context). - Keep your answer under 5 meaningful variations max. Return them in valid JSON with the following structure exactly (no extra keys, no extra texts, or formatting (including no code fences)): {{ "candidate_queries": [ "query_variant_1", "query_variant_2", "... up to 5" ] }} `, tableOfContents)
	historyMessage := []sharedtypes.HistoricMessage{}

	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, multiple_candidates_system_message)

	messageJSON, err := jsonStringToObject(result)

	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
	}

	candidateQueries, ok := messageJSON["candidate_queries"].([]interface{})
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "candidate_queries is not a slice")
		panic("candidate_queries is not a slice")
	}

	// If no candidate queries were found, use the original query
	best_query := ""
	if len(candidateQueries) == 0 {
		logging.Log.Infof(&logging.ContextMap{}, "No candidate queries found, using original query: %s", queryString)
		candidateQueries = []interface{}{queryString}
	} else {

		ranking_system_message := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided. You have proposed multiple potential queries. Now, please rank these queries in terms of likely effectiveness for searching the %s API to fulfill the user's intent. Then return only the best overall query in JSON (no extra keys, no extra texts, or formatting (including no code fences)). Format: {{ 'response': 'the single best query to scour the API reference to generate code'}} Consider which query would retrieve the most relevant methods or functionalities.`, ansysProduct, ansysProduct)

		candidateQueriesString := ""
		for i, query := range candidateQueries {
			if i > 0 {
				candidateQueriesString += "\n"
			}
			candidateQueriesString += fmt.Sprintf(`"- %s"`, query)
		}

		result, _ := PerformGeneralRequest("Candidate queries:\n"+candidateQueriesString, historyMessage, false, ranking_system_message)

		messageJSON, err = jsonStringToObject(result)

		if err != nil {
			logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
			panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
		}

		best_query, ok = messageJSON["response"].(string)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "response is not a string")
			panic("response is not a string")
		}
	}
	if best_query == "" {
		logging.Log.Error(&logging.ContextMap{}, "Best query is empty")
		result, _ := PerformGeneralRequest(userQuery, historyMessage, false, system_message)
		messageJSON, err = jsonStringToObject(result)

		if err != nil {
			logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
			panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
		}
		best_query, ok = messageJSON["response"].(string)
	}

	outputFields := []string{"text", "document_name", "previous_chunk", "next_chunk", "guid"}
	scoredPoints := doHybridQuery(collectionName, maxRetrievalCount, outputFields, best_query, denseWeight, sparseWeight, "")
	// Format results as requested
	examplesString := ""
	for _, scoredPoint := range scoredPoints {
		entry := scoredPoint.Payload
		name := entry["document_name"].GetStringValue()
		exampleRefs, _ := getExampleReferences(name, "aali") //example_refs_info

		// Format the examples as a string
		examplesString += fmt.Sprintf("Example: {%s}\n{%s}\n\n", entry["document_name"], entry["text"])
		examplesString += fmt.Sprintf("Example {%s} References: {%s}\n\n", entry["document_name"], exampleRefs)
	}
	return examplesString
}

func getExampleReferences(baseSearchNodeComplete string, db string) (string, []interface{}) {
	exampleNames := ""
	exampleReferencesInformation := []interface{}{}
	query := fmt.Sprintf(`MATCH (root:Example {name: "%s"})-[r]-(neighbor) RETURN root.name AS rootName, label(r) AS relationshipType, r AS relationshipProps, neighbor.name AS neighborName, label(neighbor) AS neighborLabel, neighbor.parameters AS neighborParameters, neighbor.remarks AS neighborRemarks, neighbor.return_type AS neighborReturn, neighbor.summary AS neighborSummary`, baseSearchNodeComplete)
	parameters := aali_graphdb.ParameterMap{}
	result := GeneralGraphDbQuery(query, parameters)
	for _, relationship := range result {
		element := relationship["neighborName"]
		elementType := relationship["neighborLabel"]
		if elementType == nil {
			elementType = "Unknown" // default value if not found
		}
		exampleNames += fmt.Sprintf("This example uses %s as a %s\n", element, elementType)
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

	return exampleNames, exampleReferencesInformation
}

// CheckWhetherPossibleToSelectMethod performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Check Whether Possible To Select Method
//
// Parameters:
//   - historyMessage: the history of messages to be used in the query
//
// Returns:
//   - historyMessage: the updated history message with the response from the LLM
//   - methodFound: a boolean indicating whether a method was found
//   - methodPath: the full path of the Method if found
//   - API reference query results
//   - Whether the section requires references
//   - The question to ask the user
func CheckWhetherPossibleToSelectMethod(apiCollectionName string, historyMessage []sharedtypes.HistoricMessage, ansysProduct string) ([]sharedtypes.HistoricMessage, bool, string, string, bool, string) {
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
            {{
                "index": "18.5.1",
                "sub_chapter_name": "Structural Results",
                "section_name": "ds_using_select_results_structural_types.xml::Deformation",
                "get_references": true
            }},
            {{
                "index": "17.10.1",
                "sub_chapter_name": "Using Result Trackers",
                "section_name": "ds_using_solve.xml::Structural Result Trackers",
                "get_references": false
            }},
            {{
                "index": "17.8",
                "sub_chapter_name": "Solving",
                "section_name": "ds_using_solve.xml::Specifying Solution Information",
                "get_references": true
            }}
        ]`, ansysProduct)
	message, _ := PerformGeneralRequest("", historyMessage, false, systemMessage)

	messageJSON, err := jsonStringToObject(message)

	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
	}

	unambiguousMethodFound := messageJSON["unambiguous_method_found"].(bool)
	unambiguousMethodPath := messageJSON["unambiguous_method_path"].(string)
	queryToAPIReferenceRequired := messageJSON["query_to_api_reference_required"].(bool)
	askUserQuestionRequired := messageJSON["ask_user_question_required"].(bool)
	// reasoning_for_decision := messageJSON["reasoning_for_decision"].(string)
	questionToUser := messageJSON["question_to_user"].(string)
	queryToAPIReference := messageJSON["query_to_api_reference"].(string)

	if unambiguousMethodFound {
		logging.Log.Infof(&logging.ContextMap{}, "Unambiguous method found: %s", unambiguousMethodPath)
		return historyMessage, true, unambiguousMethodPath, "", false, ""
	}

	if queryToAPIReferenceRequired {
		logging.Log.Infof(&logging.ContextMap{}, "Query to API reference required: %s", queryToAPIReference)
		// query api reference
		scoredPoints := doHybridQuery(apiCollectionName, 10, []string{"name"}, queryToAPIReference, 0.9, 0.1, "Method")

		// Format results as requested
		elements := []string{}
		for _, scoredPoint := range scoredPoints {
			entry := scoredPoint.Payload
			completeName := entry["name"].GetStringValue()
			namePseudocode := entry["name_pseudocode"].GetStringValue()
			remarks := entry["remarks"].GetStringValue()
			summary := entry["summary"].GetStringValue()
			parameters := entry["parameters"].GetStringValue()

			element := fmt.Sprintf("%s ; %s ; %s ; Remarks: %s; Parameters: %s\n\n", namePseudocode, completeName, summary, remarks, parameters)

			elements = append(elements, element)

		}

		options := joinStrings(elements, "\n")
		logging.Log.Infof(&logging.ContextMap{}, "API reference query results: %s", options)
		return historyMessage, false, "", options, false, ""
	}

	if askUserQuestionRequired {
		return historyMessage, false, "", "", true, questionToUser
	}

	return historyMessage, false, "", "", false, ""

}

// SearchExamplesForMethod performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Examples for Method
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - queryString: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//
// Returns:
//   - elements: the formatted element strings
//   - userResponse: the formatted user response string
func SearchExamplesForMethod(exampleCollectionName string, maxRetrievalCount int, queryString string, denseWeight float64, sparseWeight float64, ansysProduct string, historyMessage []sharedtypes.HistoricMessage, tableOfContentsString string, methodName string, maxExamples int) ([]sharedtypes.HistoricMessage, string) {
	nresult := getElementByName(methodName, "Method")
	result := nresult[0]["n"].(map[string]interface{})
	apiExample := result["example"]
	examples := getExampleNodesFromElement("Method", methodName, exampleCollectionName)

	if apiExample == nil || apiExample == "" || len(examples) == 0 {
		logging.Log.Infof(&logging.ContextMap{}, "No examples found for method: %s", methodName)
		return historyMessage, ""
	}

	exampleString := fmt.Sprintf("For the api method: %s the following examples were found:\n\n", methodName)

	if apiExample != nil {
		exampleString += fmt.Sprintf("API Example: %s\n\n", apiExample)
	}

	for i, example := range examples {
		if i >= maxExamples {
			break // Limit the number of examples to maxExamples
		}
		exampleString += fmt.Sprintf("Example: %s\n%s\n\n", example["name"], example["text"])

		exampleRefs, _ := getExampleReferences(example["name"].(string), "aali") //example_refs_info
		exampleString += fmt.Sprintf("%s-------------------\n\n", exampleRefs)

		return historyMessage, exampleString
	}

	return historyMessage, fmt.Sprintf("No examples found for method: %s", methodName)

}

// CheckForUserInformation performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Check For User Information
//
// Parameters:
//   - ansysProduct: the name of the Ansys product to be used in the system message
//   - userquery: the user query to be used for the query
//   - histroyMessage: the history of messages to be used in the query
//
// Returns:
//   - historyMessage: the updated history message with the response from the LLM
//   - bool indicating whether method was found
//   - methodPath: the full path of the Method if found
//   - Question to ask the user
func CheckForUserInformation(ansysProduct string, userQuery string, historyMessage []sharedtypes.HistoricMessage) ([]sharedtypes.HistoricMessage, bool, string, string) {
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
            In this step, you must evaluate the information retrieved from the User Guide and decide whether you have enough information to unambiguously identify the correct Method or whether you need some user input.
            If you need user input, return the query you would like to ask the user. If you have enough information, return the full path of the Method.
            Don't ask the user for information that is already provided in the query %s.
            
            Respond with the following JSON format:
            {{
                "unambiguous_method_found": true/false,
                "question_to_user": "question_to_user",
                "unambiguous_method_path": "<full path of the Method, is mandatory to include the signature with parameters if present>"
            }}
            Important: If "unambiguous_method_found" is true, "unambiguous_method_path" must be provided.    
            `, ansysProduct, userQuery)
	result, _ := PerformGeneralRequest(userQuery, historyMessage, false, systemMessage)

	messageJSON, err := jsonStringToObject(result)

	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
	}
	unambiguousMethodFound, ok := messageJSON["unambiguous_method_found"].(bool)
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "unambiguous_method_found is not a boolean")
		panic("unambiguous_method_found is not a boolean")
	}
	methodPath, ok := messageJSON["unambiguous_method_path"].(string)
	if !ok && unambiguousMethodFound {
		logging.Log.Error(&logging.ContextMap{}, "unambiguous_method_path is not a string")
		panic("unambiguous_method_path is not a string")
	}
	questionToUser, ok := messageJSON["question_to_user"].(string)
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "question_to_user is not a string")
		panic("question_to_user is not a string")
	}

	if unambiguousMethodFound {
		return historyMessage, unambiguousMethodFound, methodPath, ""
	}

	// If the method was not found unambiguously, append the question to the user to the history
	if questionToUser != "" {
		// If the method was not found unambiguously, append the question to the user to the history
		historyMessage = append(historyMessage, sharedtypes.HistoricMessage{
			Role:    "assistant",
			Content: questionToUser,
		})

		return historyMessage, false, "", questionToUser
	} else {
		return historyMessage, unambiguousMethodFound, methodPath, ""
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
//
// Returns:
//   - historyMessage: the updated history message with the response from the LLM
func GenerateCode(ansysProduct string, methodName string, examples string, historyMessages []sharedtypes.HistoricMessage) []sharedtypes.HistoricMessage {
	systemMessage := fmt.Sprintf(`In %s: You need to create a script to execute the instructions provided.
        Use the API definition and the related APIs found. Do you best to generate the code based on the information available.

            - This is the final step, **DO NOT** use any of the previous response formats.
            - **ONLY** return code that solves the user query.
            - If an unambiguous method was found, use the method to generate the code.

        Respond with the following JSON format, do not add anything else (no extra keys, no extra texts, or formatting (including no code fences)):
        {{
            "code": "only the generated code"
        }}`, ansysProduct)

	if methodName != "" {
		historyMessages = append(historyMessages, sharedtypes.HistoricMessage{
			Role:    "user",
			Content: fmt.Sprintf("Unambiguous Method Signature: %s", methodName),
		})
	}
	historyMessages = append(historyMessages, sharedtypes.HistoricMessage{
		Role:    "user",
		Content: fmt.Sprintf("%s", examples),
	})
	result, _ := PerformGeneralRequest("", historyMessages, false, systemMessage)

	messageJSON, err := jsonStringToObject(result)

	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
	}
	code, ok := messageJSON["code"].(string)
	if !ok {
		logging.Log.Error(&logging.ContextMap{}, "code is not a string")
		panic("code is not a string")
	}
	historyMessages = append(historyMessages, sharedtypes.HistoricMessage{
		Role:    "assistant",
		Content: code,
	})

	return historyMessages

}

func getExampleNodesFromElement(baseSearchType string, baseSearchNodeComplete string, collectionName string) []map[string]interface{} {
	query := fmt.Sprintf(`MATCH (n:Element) <-[:Uses]- (example:Example)
            WHERE n.name = %s AND n.type = %s
            RETURN example
            `, baseSearchNodeComplete, baseSearchType)

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
		panic(fmt.Sprintf("Error creating Qdrant client: %v", err))
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

// SearchDocumentation performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Search Documentation
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - queryString: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//
// Returns:
//   - elements: the formatted element strings
//   - userResponse: the formatted user response string
func SearchDocumentation(collectionName string, maxRetrievalCount int, queryString string, denseWeight float64, sparseWeight float64, ansysProduct string, historyMessage []sharedtypes.HistoricMessage, tableOfContentsString string) []sharedtypes.HistoricMessage {
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
            {{
                "index": "18.5.1",
                "sub_chapter_name": "Structural Results",
                "section_name": "ds_using_select_results_structural_types.xml::Deformation",
                "get_references": true
            }},
            {{
                "index": "17.10.1",
                "sub_chapter_name": "Using Result Trackers",
                "section_name": "ds_using_solve.xml::Structural Result Trackers",
                "get_references": false
            }},
            {{
                "index": "17.8",
                "sub_chapter_name": "Solving",
                "section_name": "ds_using_solve.xml::Specifying Solution Information",
                "get_references": true
            }}
        ]`, ansysProduct)
	message, _ := PerformGeneralRequest("User Guide Table of Contents:\n"+tableOfContentsString, historyMessage, false, systemMessage)
	_, err := jsonStringToObject(message)
	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON object: %v", err)
		panic(fmt.Sprintf("Error converting message to JSON object: %v", err))
	}
	// messageJSON is expected to be a slice of map[string]interface{} (JSON array)
	var chapters []map[string]interface{}
	err = json.Unmarshal([]byte(message), &chapters)
	if err != nil {
		logging.Log.Error(&logging.ContextMap{}, "Error converting message to JSON array: %v", err)
		panic(fmt.Sprintf("Error converting message to JSON array: %v", err))
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
	var retrievedUserGuideSections = ""
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
		retrievedUserGuideSections += fmt.Sprintf("Index: %s, Title: %s, Section Name: %s\n", sectionName, subChapterName, index)
		getReferences, ok := item["get_references"].(bool)
		if !ok {
			logging.Log.Error(&logging.ContextMap{}, "get_references is not a boolean")
			continue
		}
		userResponse := ""
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

			query := fmt.Sprintf("MATCH (n:UserGuide {name: %s})-[:References]->(reference) RETURN reference.name AS section_name", realSectionName)
			parameters := aali_graphdb.ParameterMap{}
			result := GeneralGraphDbQuery(query, parameters)
			if len(result) == 0 {
				logging.Log.Warnf(&logging.ContextMap{}, "No references found for section: %s", sectionName)
				continue
			}
			// Append section name from result
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
			logging.Log.Infof(&logging.ContextMap{}, "Skipping references for section: %s", sectionName)
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
		retrievedUserGuideSections += userResponse
		retrievedUserGuideSections += "\n\n\n-------------------\n\n\n"
	}

	historyMessage = append(historyMessage, sharedtypes.HistoricMessage{Role: "user", Content: "Retrieved information from user guide:\n\n\n" + retrievedUserGuideSections})

	return historyMessage
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
	exampleNames := ""
	exampleReferencesInformation := []interface{}{}
	query := fmt.Sprintf(`MATCH (root:Example {name: "%s"})-[r]-(neighbor) RETURN root.name AS rootName, label(r) AS relationshipType, r AS relationshipProps, neighbor.name AS neighborName, label(neighbor) AS neighborLabel, neighbor.parameters AS neighborParameters, neighbor.remarks AS neighborRemarks, neighbor.return_type AS neighborReturn, neighbor.summary AS neighborSummary`, baseSearchNodeComplete)
	parameters := aali_graphdb.ParameterMap{}
	result := GeneralGraphDbQuery(query, parameters)
	for _, relationship := range result {
		element := relationship["neighborName"]
		elementType := relationship["neighborLabel"]
		if elementType == nil {
			elementType = "Unknown" // default value if not found
		}
		exampleNames += fmt.Sprintf("This example uses %s as a %s\n", element, elementType)
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

	return exampleNames, exampleReferencesInformation
}

// Helper: Convert JSON string to map[string]interface{} in Go
func jsonStringToObject(jsonStr string) (map[string]interface{}, error) {
	var obj map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &obj)
	return obj, err
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
			panic(response.Error)
		}

		// Get embedded vector array (DENSE VECTOR)
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
		// Get sparse vector
		sparseVectorInterface, ok := response.LexicalWeights.(map[string]interface{})
		if !ok {
			errMessage := "error converting lexical weights to interface array"
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
		sparseVector, indexVector, err = convertToSparseVector(sparseVectorInterface)
		embedding32, err = convertToFloat32Slice(interfaceArray)
		if err != nil {
			errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
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
	query := fmt.Sprintf("MATCH (n:Element) WHERE n.name = '%s' AND n.type = '%s' RETURN n", nodeName, nodeType)
	result := GeneralGraphDbQuery(query, aali_graphdb.ParameterMap{})
	return result
}
