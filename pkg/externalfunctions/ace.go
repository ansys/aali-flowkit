// File: aali-flowkit/pkg/externalfunctions/data_extraction.go
package externalfunctions

import (
	"context"
	"fmt"

	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
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

// ConvertJsonToCustomize convert json to customize format
//
// Tags:
//   - @displayName: Convert json to customize format
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
func ConvertJsonToCustomize(object []map[string]any) string {
	return convertJsonToCustomizeHelper(object, 0, "")
}

// Internal helper with all parameters
func convertJsonToCustomizeHelper(object []map[string]any, level int, currentIndex string) string {
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

// HybridQuery performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Hybrid Query
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//   - queryString: the query string to be used for the query.
//   - denseWeight: the weight for the dense vector. (default: 0.9)
//   - sparseWeight: the weight for the sparse vector. (default: 0.1)
//
// Returns:
//   - databaseResponse: the query results
func HybridQuery(collectionName string, maxRetrievalCount int, outputFields []string, filters sharedtypes.DbFilters, queryString string, denseWeight float64, sparseWeight float64) (databaseResponse []sharedtypes.DbResponse) {

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
		// logging.Log.Infof(&logging.ContextMap{}, "testing Received response from LLM handler: %s", response)
		// Check if the response is an error
		// logging.Log.Infof(&logging.ContextMap{}, "Processing response from LLM handler: %s", response)
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

		if err != nil {
			errMessage := fmt.Sprintf("error converting lexical weights to sparse vector: %v", err)
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

	logging.Log.Infof(&logging.ContextMap{}, "Received embedding vector: %s", embedding32)
	logging.Log.Infof(&logging.ContextMap{}, "Received sparse vector: %s", sparseVector)
	logging.Log.Infof(&logging.ContextMap{}, "Received index vector: %s", indexVector)

	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	// perform the qdrant query
	limit := uint64(maxRetrievalCount)
	filter := qdrant_utils.DbFiltersAsQdrant(filters)
	using := "" // or "sparse_vector" based on the query type
	usingSparse := "sparse_vector"
	logging.Log.Infof(logCtx, "mighty Using collection %s with limit %d", collectionName, limit)
	expression := qdrant.NewExpressionSum(&qdrant.SumExpression{
		Sum: []*qdrant.Expression{
			// qdrant.NewExpressionVariable("$score"),
			// MultExpression: 0.5 * (tag match h1,h2...)
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

	// FormulaQuery(
	//         formula=SumExpression(
	//             sum=[
	//                 MultExpression(mult=["$score[0]", dense_weight]),
	//                 MultExpression(mult=["$score[1]", sparse_weight]),
	//             ]
	//         )
	// 	)
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		Prefetch: []*qdrant.PrefetchQuery{
			{
				Limit:  &limit,                            // Set to a pointer to uint64 if needed
				Query:  qdrant.NewQueryDense(embedding32), // Or: &qdrant.Query{}
				Using:  &using,                            // Or: pointer to string
				Filter: filter,                            // Or: &qdrant.Filter{}
			},
			{
				Limit:  &limit,                                           // Set to a pointer to uint64 if needed
				Query:  qdrant.NewQuerySparse(indexVector, sparseVector), // Or: &qdrant.Query{}
				Using:  &usingSparse,                                     // Or: pointer to string
				Filter: filter,                                           // Or: &qdrant.Filter{}
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
	logging.Log.Infof(logCtx, "Got %d points from qdrant query", len(scoredPoints))

	// convert to aali type
	databaseResponse = make([]sharedtypes.DbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {

		dbResponse, err := qdrant_utils.QdrantPayloadToType[sharedtypes.DbResponse](scoredPoint.Payload)
		if err != nil {
			logPanic(logCtx, "error converting qdrant payload to dbResponse: %q", err)
		}
		databaseResponse[i] = dbResponse
	}
	logging.Log.Infof(logCtx, "Converted %s points to aali type", databaseResponse)
	return databaseResponse
}
