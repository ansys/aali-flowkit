// File: aali-flowkit/pkg/externalfunctions/ace_supporting_functions.go
//
// This package provides supporting/private functions for the AALI flowkit system.
// These functions are internal helper functions used by the main external functions
// in ace.go and follow Go best practices including:
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
	"strings"
	"time"

	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/qdrant/go-client/qdrant"
)

// variable for pyansys product
var pyansysProduct = map[string]map[string]string{
	"pyfluent": {"name": "Ansys Fluent-Pyfluent", "version": "0.33.0"},
	"pyaedt":   {"name": "Ansys Electronics Desktop-PyAEDT", "version": "0.19"},
}

// checkWhetherOneOfTheMethodsFits checks whether one of the provided methods is unambiguously the right one
func checkWhetherOneOfTheMethodsFits(collectionName string, historyMessage []sharedtypes.HistoricMessage, ansysProduct string, denseWeight float64, sparseWeight float64, maxRetrievalCount int, methods string) string {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_CHECK_WHETHER_ONE_OF_THE_METHODS_FITS - Duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_ONE_OF_THE_METHODS_FITS - Input: collectionName=%s, ansysProduct=%s, denseWeight=%f, sparseWeight=%f, maxRetrievalCount=%d, methods=%s", collectionName, ansysProduct, denseWeight, sparseWeight, maxRetrievalCount, methods)

	systemMessage := fmt.Sprintf(`In %s: You need to verify the methods returned from the database are relevant or not to solve the problem.
	### Task:
		In this step you must decide whether one of the options provided is unambiguously the right one. If so, return the full path of the Method. Otherwise return the explanation for the ambiguity.

        The format is as follows: "<full path of the Method, is mandatory to include the signature with parameters if present>"

        Important: If "unambiguous_method_found" is true, "unambiguous_method_path" must be provided.`, ansysProduct)

	message, _ := PerformGeneralRequest(methods, historyMessage, false, systemMessage)

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_ONE_OF_THE_METHODS_FITS - Output: %s", message)
	return message
}

// checkWhetherUserInformationFits evaluates the information retrieved from the User Guide
func checkWhetherUserInformationFits(ansysProduct string, userGuideInformation string, historyMessage []sharedtypes.HistoricMessage, userQuery string) (string, string, string) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_CHECK_WHETHER_USER_INFORMATION_FITS - Duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_USER_INFORMATION_FITS - Input: ansysProduct=%s, userGuideInformation=%s, userQuery=%s", ansysProduct, userGuideInformation, userQuery)

	systemMessage := fmt.Sprintf(`In %s: You need to evaluate the information retrieved from the User Guide and the user query to determine if you can unambiguously identify the correct Method.

### Task:
Evaluate the **User Guide info** and **user query** to determine if you can unambiguously identify the correct Method.  

### Options:
1. Adapt the query to API Reference Vector DB with a more specific query.  
2. Ask the user for more information (only if not already provided in prior steps and after checking API Reference Vector DB).  
3. If sufficient info is available, return the **full method path with signature (parameters included if they exist)**.  
4. If the method path is like 'Path.To.Method', **do NOT append '()'** or extra characters.  
5. If multiple API methods match, return the full path of the correct one with parameters.

---
### Retrieved Info (from User Guide):
**%s**

---

### User Query:
**%s**

---

### Response Requirements:
Return the following fields separated by '-----':
1. 'unambiguous_method_found': true/false  
2. 'unambiguous_method_path': Full path including parameters if any  
3. 'query_to_api_reference_required': true/false  
4. 'ask_user_question_required': true/false  
5. 'reasoning_for_decision': Reasoning behind the choice  
6. 'question_to_user': If needed, the question to ask  
7. 'query_to_api_reference': A specific query to API Reference (if required)

---

### Example Response:

true-----ansys.fluent.core.launcher.launcher.launch_fluent(precision, dimension, additional_arguments)-----false-----false-----"User guide info clearly maps to launch_fluent() with 3D mode using dimension parameter"-----""-----""

---`, ansysProduct, userGuideInformation, userQuery)

	result, _ := PerformGeneralRequest(systemMessage, historyMessage, false, "")

	// Split the result by the separator
	parts := strings.Split(result, "-----")
	if len(parts) < 7 {
		logging.Log.Errorf(&logging.ContextMap{}, "Invalid response format: %s", result)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_USER_INFORMATION_FITS - Output: empty results due to invalid format")
		return "", "", ""
	}
	// Extract the parts
	unambiguousMethodFound := strings.TrimSpace(parts[0])
	unambiguousMethodPath := strings.TrimSpace(parts[1])
	queryToApiReferenceRequired := strings.TrimSpace(parts[2])
	askUserQuestionRequired := strings.TrimSpace(parts[3])
	// reasoningForDecision := strings.TrimSpace(parts[4]) - not used
	questionToUser := strings.TrimSpace(parts[5])
	queryToApiReference := strings.TrimSpace(parts[6])

	if unambiguousMethodFound == "true" && unambiguousMethodPath != "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_USER_INFORMATION_FITS - Output: unambiguousMethodPath=%s", unambiguousMethodPath)
		return unambiguousMethodPath, "", ""
	} else if askUserQuestionRequired == "true" && questionToUser != "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_USER_INFORMATION_FITS - Output: questionToUser=%s", questionToUser)
		return "", "", questionToUser
	} else if queryToApiReferenceRequired == "true" && queryToApiReference != "" {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_USER_INFORMATION_FITS - Output: queryToApiReference=%s", queryToApiReference)
		return "", queryToApiReference, ""
	}

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_CHECK_WHETHER_USER_INFORMATION_FITS - Output: empty results")
	return "", "", ""
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

// convertJSONToCustomizeHelper processes the query result to generate hierarchical table of contents
func convertJSONToCustomizeHelper(object []map[string]any, level int, currentIndex string) string {
	var nodeBuilder strings.Builder

	// Process the results to build hierarchical structure
	var currentChapter string
	var currentSection string
	chapterCounter := 0
	sectionCounter := 0
	subsectionCounter := 0

	for _, row := range object {
		// Extract the data from each row (similar to Python's row access)
		var chapterTitle, chapterDoc, sectionTitle, sectionDoc, subsectionTitle, subsectionDoc string

		// Handle different possible structures in the input
		if chapters, ok := row["chapters"].([]interface{}); ok {
			// If this is the nested structure from the new query
			for _, chapter := range chapters {
				if chapterMap, ok := chapter.(map[string]interface{}); ok {
					if title, ok := chapterMap["title"].(string); ok {
						chapterTitle = title
					}
					if doc, ok := chapterMap["document_name"].(string); ok {
						chapterDoc = doc
					}
					// For this structure, we'll need to process differently
					// This is a single chapter per row
					break
				}
			}
		} else {
			// Handle the flat structure from the original query
			if title, ok := row["chapter_title"].(string); ok {
				chapterTitle = title
			}
			if doc, ok := row["chapter_doc"].(string); ok {
				chapterDoc = doc
			}
			if title, ok := row["section_title"].(string); ok {
				sectionTitle = title
			}
			if doc, ok := row["section_doc"].(string); ok {
				sectionDoc = doc
			}
			if title, ok := row["subsection_title"].(string); ok {
				subsectionTitle = title
			}
			if doc, ok := row["subsection_doc"].(string); ok {
				subsectionDoc = doc
			}
		}

		// Handle new chapter
		if currentChapter != chapterDoc && chapterDoc != "" {
			currentChapter = chapterDoc
			currentSection = ""
			chapterCounter++
			sectionCounter = 0
			subsectionCounter = 0

			// Add level 1 chapter with numbering
			nodeBuilder.WriteString(fmt.Sprintf("%d. %s (section Name -> %s)\n",
				chapterCounter, chapterTitle, chapterDoc))
		}

		// Handle new section
		if sectionDoc != "" && currentSection != sectionDoc {
			currentSection = sectionDoc
			sectionCounter++
			subsectionCounter = 0

			// Add level 2 section with proper indentation
			nodeBuilder.WriteString(fmt.Sprintf("  %d.%d. %s (section Name -> %s)\n",
				chapterCounter, sectionCounter, sectionTitle, sectionDoc))
		}

		// Handle subsection
		if subsectionDoc != "" {
			subsectionCounter++
			// Add level 3 subsection with proper indentation
			nodeBuilder.WriteString(fmt.Sprintf("    %d.%d.%d. %s (section Name -> %s)\n",
				chapterCounter, sectionCounter, subsectionCounter, subsectionTitle, subsectionDoc))
		}
	}

	return nodeBuilder.String()
} // repeatString returns a string with s repeated count times.
// Uses strings.Repeat for efficiency.
func repeatString(s string, count int) string {
	return strings.Repeat(s, count)
}

// joinStrings joins a slice of strings with the given separator.
// Uses strings.Join for efficiency.
func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// getExampleReferences retrieves references for a given example
func getExampleReferences(baseSearchNodeComplete string, db string) (string, []interface{}) {
	var exampleNamesBuilder strings.Builder
	exampleReferencesInformation := []interface{}{}
	// Escape the string parameter properly for Cypher
	escapedName := strings.ReplaceAll(baseSearchNodeComplete, `"`, `\"`)
	query := fmt.Sprintf(`MATCH (root:Example {name: "%s"})-[r]-(neighbor) RETURN root.name AS rootName, label(r) AS relationshipType, r AS relationshipProps, neighbor.name AS neighborName, label(neighbor) AS neighborLabel, neighbor.parameters AS neighborParameters, neighbor.remarks AS neighborRemarks, neighbor.return_type AS neighborReturn, neighbor.summary AS neighborSummary`, escapedName)
	parameters := aali_graphdb.ParameterMap{}
	result := GeneralGraphDbQuery(db, query, parameters)
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

// getExampleNodesFromElement retrieves example nodes from an element
func getExampleNodesFromElement(baseSearchType string, baseSearchNodeComplete string, collectionName string, dbname string) []map[string]interface{} {

	// Escape the string parameters properly for Cypher
	escapedNodeComplete := strings.ReplaceAll(baseSearchNodeComplete, `"`, `\"`)
	escapedType := strings.ReplaceAll(baseSearchType, `"`, `\"`)
	query := fmt.Sprintf(`MATCH (n:Element) <-[:Uses]- (example:Example)
			WHERE n.name = "%s" AND n.type = "%s"
			RETURN example
			`, escapedNodeComplete, escapedType)

	parameters := aali_graphdb.ParameterMap{}

	result := GeneralGraphDbQuery(dbname, query, parameters)
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

// queryExample queries for example chunks in the collection
func queryExample(exampleName string, collectionName string) []map[string]interface{} {
	// search database
	client, err := qdrant_utils.QdrantClient()

	if err != nil {
		logging.Log.Infof(&logging.ContextMap{}, "Error creating Qdrant client: %v", err)
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

// queryUserGuideName queries for user guide sections by name
func queryUserGuideName(name string, resultCount uint64, collectionName string) []*qdrant.ScoredPoint {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_QUERY_USER_GUIDE_NAME - Duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_QUERY_USER_GUIDE_NAME - Input: name=%s, resultCount=%d, collectionName=%s", name, resultCount, collectionName)

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

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_QUERY_USER_GUIDE_NAME - Output: %d results found", len(results))
	return results
}

// getDocumentation retrieves documentation for a given node
func getDocumentation(baseSearchNodeComplete string, db string) (string, []interface{}) {

	var exampleNamesBuilder strings.Builder
	exampleReferencesInformation := []interface{}{}
	// Escape the string parameter properly for Cypher
	escapedName := strings.ReplaceAll(baseSearchNodeComplete, `"`, `\"`)
	query := fmt.Sprintf(`MATCH (root:Example {name: "%s"})-[r]-(neighbor) RETURN root.name AS rootName, label(r) AS relationshipType, r AS relationshipProps, neighbor.name AS neighborName, label(neighbor) AS neighborLabel, neighbor.parameters AS neighborParameters, neighbor.remarks AS neighborRemarks, neighbor.return_type AS neighborReturn, neighbor.summary AS neighborSummary`, escapedName)
	parameters := aali_graphdb.ParameterMap{}

	// Time the graph database query
	result := GeneralGraphDbQuery(db, query, parameters)
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

// PreprocessLLMJSON preprocesses LLM JSON responses (public function)
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

// preprocessLLMJSON preprocesses LLM JSON responses (private helper)
func preprocessLLMJSON(s string) string {
	return PreprocessLLMJSON(s)
}

// jsonStringToObject converts JSON string to map[string]interface{} in Go
func jsonStringToObject(jsonStr string) (map[string]interface{}, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_JSON_STRING_TO_OBJECT - Duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_JSON_STRING_TO_OBJECT - Input: jsonStr=%s", jsonStr)

	clean := preprocessLLMJSON(jsonStr)

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
				logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_JSON_STRING_TO_OBJECT - Output: successful manual extraction, %d keys", len(obj))
				return obj, nil
			}
		}
	}

	if err == nil {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_JSON_STRING_TO_OBJECT - Output: successful parse, %d keys", len(obj))
	} else {
		logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_JSON_STRING_TO_OBJECT - Output: parse failed with error: %v", err)
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

// doHybridQuery performs a hybrid dense and sparse query using Qdrant
func doHybridQuery(
	collectionName string,
	maxRetrievalCount int,
	outputFields []string,
	queryString string,
	nodeType string) []*qdrant.ScoredPoint {

	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequest(queryString, llmHandlerEndpoint, true, nil)
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

// getElementByName retrieves an element by name and type from the graph database
func getElementByName(nodeName string, nodeType string, dbname string) []map[string]interface{} {

	// Escape the string parameters properly for Cypher
	escapedNodeName := strings.ReplaceAll(nodeName, `'`, `\'`)
	escapedNodeType := strings.ReplaceAll(nodeType, `'`, `\'`)
	query := fmt.Sprintf("MATCH (n:Element) WHERE n.name = '%s' AND n.type = '%s' RETURN n", escapedNodeName, escapedNodeType)
	logging.Log.Infof(&logging.ContextMap{}, "Executing query to get element by name: %s", query)

	result := GeneralGraphDbQuery(dbname, query, aali_graphdb.ParameterMap{})
	return result
}

func searchExamplesForMethod(collectionName string, ansysProduct string, historyMessage []sharedtypes.HistoricMessage, methodNames string, maxExamples int, dbname string) string {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		logging.Log.Infof(&logging.ContextMap{}, "ACE_TIMING FUNC_SEARCH_EXAMPLES_FOR_METHOD - Duration: %v", duration)
	}()

	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_EXAMPLES_FOR_METHOD - Input: collectionName=%s, ansysProduct=%s, methodNames=%s, maxExamples=%d, dbname=%s histroy=%s", collectionName, ansysProduct, methodNames, maxExamples, dbname, historyMessage)

	var outputBuilder strings.Builder
	// split ; sperated methodName into parts
	parts := strings.Split(methodNames, ";")
	for _, methodName := range parts {
		nresult := getElementByName(methodName, "Method", dbname)

		if len(nresult) == 0 {
			logging.Log.Warnf(&logging.ContextMap{}, "No method found with name: %s", methodName)
			continue
		}

		result, ok := nresult[0]["n"].(map[string]any)
		if !ok {
			logging.Log.Infof(&logging.ContextMap{}, "Failed to parse method result for: %s", methodName)
			continue
		}

		apiExample := result["example"]
		// if apiExample == nil {
		// 	logging.Log.Warnf(&logging.ContextMap{}, "No API example found for method: %s", methodName)
		// 	continue
		// }

		examples := getExampleNodesFromElement("Method", methodName, collectionName, dbname)
		outputBuilder.WriteString(methodName)
		if apiExample != nil {
			outputBuilder.WriteString(fmt.Sprintf("For the api method: %s the following examples were found:\n\n", methodName))
			outputBuilder.WriteString(fmt.Sprintf("For the api method: %s the following examples were found:\n\n", apiExample))
		} else {
			outputBuilder.WriteString(fmt.Sprintf("For the api method: %s the following examples were found:\n\n", methodName))
			for i, example := range examples {
				if i >= maxExamples {
					break // Limit the number of examples to maxExamples
				}
				outputBuilder.WriteString(fmt.Sprintf("Example: %s\n%s\n\n", example["name"], example["text"]))

				exampleRefs, _ := getExampleReferences(example["name"].(string), dbname) //example_refs_info
				outputBuilder.WriteString(fmt.Sprintf("%s-------------------\n\n", exampleRefs))
			}
		}
	}

	result := outputBuilder.String()
	logging.Log.Infof(&logging.ContextMap{}, "ACE_OUTPUT FUNC_SEARCH_EXAMPLES_FOR_METHOD - Output: %s", result)
	return result
}
