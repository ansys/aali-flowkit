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
	//"fmt"
	"strings"

	"github.com/ansys/aali-flowkit/pkg/privatefunctions/graphdb"
	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

// PyaedtGetElementContextFromGraphDb  graph database.
//
// Tags:
//   - @displayName: Pyaedt Get context from graph db
//
// Parameters:
//   - elementName - string
//   - elementType - string
// TODO: Move to pyaedt.go file 
func PyaedtGetElementContextFromGraphDb(dbResponse sharedtypes.ApiDbResponse) (exampleNames []string , parameters []string, returnTypes []string, pyaedtGroupCaller string){
	ctx := &logging.ContextMap{}
	err := graphdb.Initialize(config.GlobalConfig.GRAPHDB_ADDRESS)
	if err != nil {
		logPanic(nil, "error initializing graphdb: %v", err)
	}
	// kapatil : instead of element names, can we use GUID ?
        // Assuming this is a single entry point 
	//if len(dbResponse) > 0 {
		elementType := dbResponse.Type
		elementName := dbResponse.Name
		logging.Log.Debugf(ctx, "reading entry points %s of type %s", elementName, elementType)


		// Get pyaedt group
		// if element type = method -> go to class -> check isPyaedtGroup
		// if element type = class -> check isPyaedtGroup
		pGroup, err := graphdb.GraphDbDriver.GetPyaedtGroupFromCodeGenerationElement(elementType, elementName)
		if err != nil {
			logPanic(ctx, "error Getting pyaedtGroup from code generation element: %v", err)
		} else {

			pyaedtGroupCaller, err := graphdb.GraphDbDriver.GetPyaedtGroupCaller(pGroup)
		        if err != nil {
				logPanic(ctx, "error Getting pyaedtGroup Caller from code generation element: %v", err)
			}
			logging.Log.Debugf(ctx, "PyaedtCaller Type %s", pyaedtGroupCaller)

		}


		//exampleNames, err := graphdb.GraphDbDriver.GetExamplesFromCodeGenerationElement(elementType, elementName)
		//if err != nil {
		//	logPanic(ctx, "error Getting examples from code generation element: %v", err)
		//}
		if len(exampleNames) > 0 {
			for ex, _ := range exampleNames {
				logging.Log.Debugf(ctx, "Reading examples %s", ex)
	        	}   
	        } else {
			logging.Log.Debugf(ctx, "GetExamplesFromGraphDB: No db points for this entry point")
		}
		// Get Parameters
		parameters, err = graphdb.GraphDbDriver.GetParametersFromCodeGenerationElement(elementType, elementName)
		if err != nil {
			logPanic(ctx, "error Getting parameters from code generation element: %v", err)
		}
		if len(parameters) > 0 {
			for ex, _ := range parameters {
				logging.Log.Debugf(ctx, "Reading parameters %s", ex)
	        	}   
	        } else {
			logging.Log.Debugf(ctx, "GetParameters: No db points for this entry point")
		}
	
		// Get Return types
		returnTypes, err = graphdb.GraphDbDriver.GetReturnTypeFromCodeGenerationElement(elementType, elementName)
		if err != nil {
			logPanic(ctx, "error Getting return types  from code generation element: %v", err)
		}
		if len(returnTypes) > 0 {
			for ex, _ := range returnTypes {
				logging.Log.Debugf(ctx, "Reading returnTypes %s", ex)
	        	}   
	        } else {
			logging.Log.Debugf(ctx, "GetReturnTypes: No db points for this entry point")
		}
	
	//} else {
	//	logging.Log.Debugf(ctx, "No entry point to graph DB found, Will skip context addition")
	//}
	return exampleNames, parameters, returnTypes, pyaedtGroupCaller
}

// PyaedtBatchGetElementContextFromGraphDb  graph database.
//
// Tags:
//   - @displayName: Pyaedt Get element context from graph db
//
// Parameters:
//   - dbResponse - input response from API methods
// TODO: Move to pyaedt.go file 
func PyaedtBatchGetElementContextFromGraphDb(dbResponses []sharedtypes.ApiDbResponse) (elementContexts []string) {
	ctx := &logging.ContextMap{}
	err := graphdb.Initialize(config.GlobalConfig.GRAPHDB_ADDRESS)
	if err != nil {
		logPanic(nil, "error initializing graphdb: %v", err)
	}
        elementParams := make(map[string][]string)
	elementReturns := make(map[string][]string)
	// kapatil : instead of element names, can we use GUID ?
        // Assuming this is a single entry point 
	for _, dbResponse := range dbResponses {
		elementType := dbResponse.Type
		elementName := dbResponse.Name
		logging.Log.Debugf(ctx, "Reading entry points %s of type %s", elementName, elementType)
		tempPrompt := "For "
		tempPrompt += elementName
			
		_, parameters, returnTypes, pyaedtGroupCallerType := PyaedtGetElementContextFromGraphDb(dbResponses[0])
		if len(parameters) > 0 {
			elementParams[elementName] = parameters
			tempPrompt += elementParams[elementName][0]
			}
		if len(returnTypes) > 0 {
			elementReturns[elementName] = returnTypes
			tempPrompt += elementReturns[elementName][0]
		
		}
		if pyaedtGroupCallerType != "Pyaedt_Application" {
			tempPrompt += "Additionally this function "+elementName+ " needs an pyaedt application or solver object like HFSS,Maxwell, Circuit,Q3d, etc as an argument." 
		}
		logging.Log.Debugf(ctx, "kapatil: Create Context prompt draft: %s", tempPrompt)
		elementContexts = append(elementContexts, tempPrompt)

	}
	if len(dbResponses) == 0 {
		logging.Log.Debugf(ctx, "No function context found")
	}
	// Create prompt for each element information
	return elementContexts
}


// SendVectorsToUserGuide sends the given vector to the user guide  and returns the most relevant data
//
// Tags:
//   - @displayName: User guide Similarity Search
//
// Parameters:
//   - vector: the vector to be sent to the KnowledgeDB
//   - keywords: the keywords to be used to filter the results
//   - keywordsSearch: the flag to enable the keywords search
//   - collection: the collection name
//   - similaritySearchResults: the number of results to be returned
//   - similaritySearchMinScore: the minimum score for the results
//   - sparseVector: optional sparse vector for hybrid search (pass empty map for dense-only search)
//
// Returns:
//   - citations: an array of the most relevant user guide urls
func SendVectorsToUserGuide(vector []float32, keywords []string, keywordsSearch bool, collection string, similaritySearchResults int, similaritySearchMinScore float64, sparseVector map[uint]float32) (citations []string ) {//databaseResponse []sharedtypes.DbResponse) {
	// Use the provided sparse vector directly (will be empty map if not provided)
	sparse := sparseVector
        collection = "user_guide"
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}
	// Pure vector similarity search across all collection types
	filter := qdrant.Filter{}
	// Note: Keyword search disabled for now to ensure broad compatibility

	limit := uint64(similaritySearchResults)
	scoreThreshold := float32(similaritySearchMinScore)

	var query qdrant.QueryPoints

	// Use fusion if both dense and sparse vectors are available
	if sparse != nil && len(sparse) > 0 {
		// Create prefetch queries for hybrid search using RRF (Reciprocal Rank Fusion)
		prefetchQueries := []*qdrant.PrefetchQuery{
			// Dense vector search prefetch
			{
				Query:  qdrant.NewQueryDense(vector),
				Using:  nil, // Use default (unnamed) vector
				Filter: &filter,
				Limit:  &limit,
			},
			// Sparse vector search prefetch
			{
				Query:  createSparseQuery(sparse),
				Using:  qdrant.PtrOf("sparse_vector"), // Use sparse vector field
				Filter: &filter,
				Limit:  &limit,
			},
		}

		query = qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryFusion(qdrant.Fusion_RRF), // Use Reciprocal Rank Fusion
			Prefetch:       prefetchQueries,
			Limit:          &limit,
			ScoreThreshold: &scoreThreshold,
			Filter:         &filter,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
		}
	} else {
		// DENSE-ONLY SEARCH: Simplified approach
		query = qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryDense(vector),
			Limit:          &limit,
			ScoreThreshold: &scoreThreshold,
			Filter:         &filter,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
		}
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}

	// Transform results
	logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Got %d points from qdrant query", len(scoredPoints))
	citations = make([]string, len(scoredPoints))
	dbResponses := make([]sharedtypes.DbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {
		logging.Log.Debugf(&logging.ContextMap{}, "Result #%d:", i)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity score: %v", scoredPoint.Score)
		dbResponse, err := qdrant_utils.QdrantPayloadToType[sharedtypes.DbResponse](scoredPoint.GetPayload())
                // Add the result to the list
		if err != nil {
		}
		dbResponses[i] = dbResponse
		/// Form the user guide url: <base url> + document name
		citations[i] = "https://github.com/ansys/pyaedt/tree/main/doc/source/User_guide/" + dbResponse.DocumentName
		// TODO: Add url validation
		//logging.Log.Debugf(&logging.ContextMap{}, "Similarity doc title: %v", dbResponse.Title)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity doc path relative: %v", dbResponse.DocumentName)

	}
	
	return citations

}


// SendVectorsToKnowledgeDB sends the given vector to the KnowledgeDB and returns the most relevant data
//
// Tags:
//   - @displayName: Similarity Search
//
// Parameters:
//   - vector: the vector to be sent to the KnowledgeDB
//   - keywords: the keywords to be used to filter the results
//   - keywordsSearch: the flag to enable the keywords search
//   - collection: the collection name
//   - similaritySearchResults: the number of results to be returned
//   - similaritySearchMinScore: the minimum score for the results
//   - sparseVector: optional sparse vector for hybrid search (pass empty map for dense-only search)
//
// Returns:
//   - databaseResponse: an array of the most relevant data
func SendVectorsToKnowledgeDB(vector []float32, keywords []string, keywordsSearch bool, collection string, similaritySearchResults int, similaritySearchMinScore float64, sparseVector map[uint]float32) (databaseResponse []sharedtypes.ApiDbResponse) {
	// Use the provided sparse vector directly (will be empty map if not provided)
	sparse := sparseVector

	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}
	

	// Example inputs 1: Create HFSS design / launch hfss
	// Example keywords: ["hfss design", "launch hfss"]
	// Example inputs 2: Using existing desktop session, create HFSS instance with new project "abc" and aedt version 2025 R1 in non-graphical mode
	// Example keywords: ["ansys.aedt.core","HFSS","Project","aedt version","non-graphical mode"]

	// keywords = []string{"ansys.aedt.core", "HFSS", "project", "aedt version", "non-graphical mode"}
	// keywords = []string{"hfss design", "launch hfss"}	
	
	// Pure vector similarity search across all collection types
	filter := qdrant.Filter{}
        //filter := qdrant.Filter{
	//	Should: []*qdrant.Condition{},
	//}

	// perform the qdrant query: Phrase match on keywords
	//if keywordsSearch && len(keywords) > 0 {
		// TODO: Check if keyword extraction failed:
		// The provided text does not mention anything about ansys.aedt.core Launch AEDT or Desktop objects so I cannot extract keywords from it.

		//exclude_keywords := map[string]struct{}{
		//	"ansys":           {},
		//	"aedt":            {},
	//		"core":            {},
		//	"ansys.aedt.core": {},
		//	"ansys.aedt":      {},
		//	"aedt.core":       {},
		//}

		//for _, keyword := range keywords {
			// Exclude keywords that will always in the library context, including it in the 'Should' clause
			// would lead to all results being returned, which is not desired.
			//if _, found := exclude_keywords[strings.ToLower(keyword)]; !found {
			//	filter.Should = append(filter.Should, qdrant.NewMatchText("name", keyword))
			//}
		//}
	//}

	//logging.Log.Debugf(logCtx, "********* Filter %v *********\n", filter)
	// Note: Keyword search disabled for now to ensure broad compatibility

	limit := uint64(similaritySearchResults)
	scoreThreshold := float32(similaritySearchMinScore)

	var query qdrant.QueryPoints
	
	// Use fusion if both dense and sparse vectors are available
	if sparse != nil && len(sparse) > 0 {
		// Create prefetch queries for hybrid search using RRF (Reciprocal Rank Fusion)
		prefetchQueries := []*qdrant.PrefetchQuery{
			// Dense vector search prefetch
			{
				Query:  qdrant.NewQueryDense(vector),
				Using:  nil, // Use default (unnamed) vector
				Filter: &filter,
				Limit:  &limit,
			},
			// Sparse vector search prefetch
			{
				Query:  createSparseQuery(sparse),
				Using:  qdrant.PtrOf("sparse_vector"), // Use sparse vector field
				Filter: &filter,
				Limit:  &limit,
			},
		}

		query = qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryFusion(qdrant.Fusion_RRF), // Use Reciprocal Rank Fusion
			Prefetch:       prefetchQueries,
			Limit:          &limit,
			ScoreThreshold: &scoreThreshold,
			Filter:         &filter,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
		}
	} else {
		// DENSE-ONLY SEARCH: Simplified approach
		query = qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryDense(vector),
			Limit:          &limit,
			ScoreThreshold: &scoreThreshold,
			Filter:         &filter,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
		}
	}

	// perform the qdrant query
	



	logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Similarity search Query to Qdrant %s", query)
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}

	// Transform results
	logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Got %f points from qdrant query", len(scoredPoints))
	
	dbResponses := make([]sharedtypes.ApiDbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {
		logging.Log.Debugf(&logging.ContextMap{}, "Result #%d:", i)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity score: %v", scoredPoint.Score)
		dbResponse, err := qdrant_utils.QdrantPayloadToType[sharedtypes.ApiDbResponse](scoredPoint.GetPayload())
                // Add the result to the list
		if err != nil {
		}
		dbResponses[i] = dbResponse
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity element name: %v", dbResponse.Name)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity pyaedt_group: %v", dbResponse.PyaedtGroup)
	}
        
        //logging.Log.Debugf(&logging.ContextMap{}, "examples: %d", len(exampledbResponse))

	return dbResponses

}


// SendVectorToExamplesDb sends the given vector to the KnowledgeDB and returns the most relevant data
//
// Tags:
//   - @displayName: Example Similarity Search 
//
// Parameters:
//   - vector: the vector to be sent to the KnowledgeDB
//   - keywords: the keywords to be used to filter the results
//   - keywordsSearch: the flag to enable the keywords search
//   - collection: the collection name
//   - similaritySearchResults: the number of results to be returned
//   - similaritySearchMinScore: the minimum score for the results
//   - sparseVector: optional sparse vector for hybrid search (pass empty map for dense-only search)
//
// Returns:
//   - databaseResponse: an array of the most relevant data
func SendVectorsToExampleDB(vector []float32, keywords []string, keywordsSearch bool, collection string, similaritySearchResults int, similaritySearchMinScore float64, sparseVector map[uint]float32) (databaseResponse []sharedtypes.ExampleDbResponse) {
	// Use the provided sparse vector directly (will be empty map if not provided)
	sparse := sparseVector
	qclient, err := qdrant_utils.QdrantClient()
	collExists, err := qclient.CollectionExists(context.TODO(), "examples") //your collection name here
	if collExists {
		logging.Log.Debugf(&logging.ContextMap{}, "Found example collection")
	}
	collection = "examples" // TODO: Your examples collection name  here
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}
	// Pure vector similarity search across all collection types
	filter := qdrant.Filter{}
	// Note: Keyword search disabled for now to ensure broad compatibility

	limit := uint64(similaritySearchResults)
	scoreThreshold := float32(similaritySearchMinScore)

	var query qdrant.QueryPoints
	
	logging.Log.Debugf(&logging.ContextMap{}, "*********** dense %v ***********", vector)
	logging.Log.Debugf(&logging.ContextMap{}, "*********** sparse %v, %d***********", sparse, len(sparse))

	// Use fusion if both dense and sparse vectors are available
	if sparse != nil && len(sparse) > 0 {
		logging.Log.Debugf(&logging.ContextMap{}, "*********** Hybrid: Sparse + dense query ***********")
		// Create prefetch queries for hybrid search using RRF (Reciprocal Rank Fusion)
		prefetchQueries := []*qdrant.PrefetchQuery{
			// Dense vector search prefetch
			{
				Query:  qdrant.NewQueryDense(vector),
				Using:  nil, // Use default (unnamed) vector
				Filter: &filter,
				Limit:  &limit,
			},
			// Sparse vector search prefetch
			{
				Query:  createSparseQuery(sparse),
				Using:  qdrant.PtrOf("sparse_vector"), // Use sparse vector field
				Filter: &filter,
				Limit:  &limit,
			},
		}

		query = qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryFusion(qdrant.Fusion_RRF), // Use Reciprocal Rank Fusion
			Prefetch:       prefetchQueries,
			Limit:          &limit,
			ScoreThreshold: &scoreThreshold,
			Filter:         &filter,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
		}
	} else {
		// DENSE-ONLY SEARCH: Simplified approach
		query = qdrant.QueryPoints{
			CollectionName: collection,
			Query:          qdrant.NewQueryDense(vector),
			Limit:          &limit,
			ScoreThreshold: &scoreThreshold,
			Filter:         &filter,
			WithVectors:    qdrant.NewWithVectorsEnable(false),
			WithPayload:    qdrant.NewWithPayloadEnable(true),
		}
	}

	logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Example Similarity search Query to Qdrant %s", query)
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}

	// Transform results
	logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Got %f points from qdrant query", len(scoredPoints))
	
	dbResponses := make([]sharedtypes.ExampleDbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {
		logging.Log.Debugf(&logging.ContextMap{}, "Result #%d:", i)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity score: %v", scoredPoint.Score)
		dbResponse, err := qdrant_utils.QdrantPayloadToType[sharedtypes.ExampleDbResponse](scoredPoint.GetPayload())
                // Add the result to the list
		if err != nil {
		}
		dbResponses[i] = dbResponse
		//logging.Log.Debugf(&logging.ContextMap{}, "Similarity element name: %v", dbResponse.Name)
		logging.Log.Debugf(&logging.ContextMap{}, "Similarity text: %v", dbResponse.Text)
	}
	return dbResponses
}



// Helper function to create sparse query from map[uint]float32
func createSparseQuery(sparseVector map[uint]float32) *qdrant.Query {
	if len(sparseVector) == 0 {
		return nil
	}

	indices := make([]uint32, 0, len(sparseVector))
	values := make([]float32, 0, len(sparseVector))

	for idx, val := range sparseVector {
		indices = append(indices, uint32(idx))
		values = append(values, val)
	}

	return qdrant.NewQuerySparse(indices, values)
}

// GetListCollections retrieves the list of collections from the KnowledgeDB.
//
// Tags:
//   - @displayName: List Collections
//
// The function returns the list of collections.
//
// Parameters:
//   - knowledgeDbEndpoint: the KnowledgeDB endpoint
//
// Returns:
//   - collectionsList: the list of collections
func GetListCollections() (collectionsList []string) {
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	collectionsList, err = client.ListCollections(context.TODO())
	if err != nil {
		logPanic(logCtx, "unable to list qdrant collections: %q", err)
	}
	return collectionsList
}

// RetrieveDependencies retrieves the dependencies of the specified source node.
//
// The function returns the list of dependencies.
//
// Tags:
//   - @displayName: Retrieve Dependencies
//
// Parameters:
//   - relationshipName: the name of the relationship to retrieve dependencies for.
//   - relationshipDirection: the direction of the relationship to retrieve dependencies for.
//   - sourceDocumentId: the document ID of the source node.
//   - nodeTypesFilter: filter based on node types.
//   - maxHopsNumber: maximum number of hops to traverse.
//
// Returns:
//   - dependenciesIds: the list of dependencies
func RetrieveDependencies(
	relationshipName string,
	relationshipDirection string,
	sourceDocumentId string,
	nodeTypesFilter sharedtypes.DbArrayFilter,
	maxHopsNumber int) (dependenciesIds []string) {
	ctx := &logging.ContextMap{}
	dependenciesIds, err := graphdb.GraphDbDriver.RetrieveDependencies(
		ctx,
		relationshipName,
		relationshipDirection,
		sourceDocumentId,
		nodeTypesFilter,
		[]string{},
		maxHopsNumber,
	)
	if err != nil {
		logPanic(nil, "unable to retrieve dependencies: %q", err)
	}
	return dependenciesIds
}

// AddGraphDbParameter adds a new GraphDbParameter to a map[string]GraphDbParameter
//
// Tags:
//   - @displayName: Add Graph DB Parameter
//
// Parameters:
//   - parameters: the existing collection of parameters
//   - name: the name of the new parameter
//   - value: the value of the new parameter
//   - paramType: the type of the new parameter
//
// Returns:
//   - The original parameters with the new one added
func AddGraphDbParameter(parameters aali_graphdb.ParameterMap, name string, value string, paramType string) aali_graphdb.ParameterMap {
	valType := sharedtypes.GraphDbValueType(strings.ToLower(paramType))
	val, err := valType.Parse(value)
	if err != nil {
		logPanic(nil, "could not build graph DB parameter: %v", err)
	}
	parameters[name] = val
	return parameters
}

// GeneralGraphDbQuery executes the given Cypher query and returns the response.
//
// The function returns the graph db response.
//
// Tags:
//   - @displayName: General Graph DB Query
//
// Parameters:
//   - query: the Cypher query to be executed.
//   - parameters: parameters to pass to the query during execution
//
// Returns:
//   - databaseResponse: the graph db response
func GeneralGraphDbQuery(query string, parameters aali_graphdb.ParameterMap) []map[string]any {
	// Initialize the graph database.
	err := graphdb.Initialize(config.GlobalConfig.GRAPHDB_ADDRESS)
	if err != nil {
		logPanic(nil, "error initializing graphdb: %v", err)
	}
	res, err := graphdb.GraphDbDriver.WriteCypherQuery(query, parameters)
	if err != nil {
		logPanic(nil, "error executing cypher query: %q", err)
	}
	return res
}

// GeneralQuery performs a general query in the KnowledgeDB.
//
// The function returns the query results.
//
// Tags:
//   - @displayName: Query
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//
// Returns:
//   - databaseResponse: the query results
func GeneralQuery(collectionName string, maxRetrievalCount int, outputFields []string, filters sharedtypes.DbFilters) (databaseResponse []sharedtypes.DbResponse) {
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	// perform the qdrant query
	limit := uint64(maxRetrievalCount)
	filter := qdrant_utils.DbFiltersAsQdrant(filters)
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		Limit:          &limit,
		Filter:         filter,
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadInclude(outputFields...),
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}
	logging.Log.Debugf(logCtx, "Got %d points from qdrant query", len(scoredPoints))

	// convert to aali type
	databaseResponse = make([]sharedtypes.DbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {

		dbResponse, err := qdrant_utils.QdrantPayloadToType[sharedtypes.DbResponse](scoredPoint.Payload)
		if err != nil {
			logPanic(logCtx, "error converting qdrant payload to dbResponse: %q", err)
		}
		databaseResponse[i] = dbResponse
	}
	return databaseResponse
}

// SimilaritySearch performs a similarity search in the KnowledgeDB.
//
// The function returns the similarity search results.
//
// Tags:
//   - @displayName: Similarity Search (Filtered)
//
// Parameters:
//   - collectionName: the name of the collection to which the data objects will be added.
//   - embeddedVector: the embedded vector used for searching.
//   - maxRetrievalCount: the maximum number of results to be retrieved.
//   - outputFields: the fields to be included in the output.
//   - filters: the filter for the query.
//   - minScore: the minimum score filter.
//   - getLeafNodes: flag to indicate whether to retrieve all the leaf nodes in the result node branch.
//   - getSiblings: flag to indicate whether to retrieve the previous and next node to the result nodes.
//   - getParent: flag to indicate whether to retrieve the parent object.
//   - getChildren: flag to indicate whether to retrieve the children objects.
//
// Returns:
//   - databaseResponse: the similarity search results
func SimilaritySearch(
	collectionName string,
	embeddedVector []float32,
	maxRetrievalCount int,
	filters sharedtypes.DbFilters,
	minScore float64,
	getLeafNodes bool,
	getSiblings bool,
	getParent bool,
	getChildren bool) (databaseResponse []sharedtypes.DbResponse) {
	logCtx := &logging.ContextMap{}
	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	// perform the qdrant query
	limit := uint64(maxRetrievalCount)
	scoreThreshold := float32(minScore)
	query := qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          qdrant.NewQueryDense(embeddedVector),
		Limit:          &limit,
		ScoreThreshold: &scoreThreshold,
		Filter:         qdrant_utils.DbFiltersAsQdrant(filters),
		WithVectors:    qdrant.NewWithVectorsEnable(false),
		WithPayload:    qdrant.NewWithPayloadEnable(true),
	}
	scoredPoints, err := client.Query(context.TODO(), &query)
	if err != nil {
		logPanic(logCtx, "error in qdrant query: %q", err)
	}
	logging.Log.Debugf(logCtx, "Got %d points from qdrant query", len(scoredPoints))

	// convert to aali type
	databaseResponse = make([]sharedtypes.DbResponse, len(scoredPoints))
	for i, scoredPoint := range scoredPoints {

		dbResponse, err := qdrant_utils.QdrantPayloadToType[sharedtypes.DbResponse](scoredPoint.Payload)
		if err != nil {
			logPanic(logCtx, "error converting qdrant payload to dbResponse: %q", err)
		}
		id, err := uuid.Parse(scoredPoint.Id.GetUuid())
		if err != nil {
			logPanic(logCtx, "point ID is not parseable as a UUID: %v", err)
		}
		dbResponse.Guid = id
		databaseResponse[i] = dbResponse
	}

	// get related nodes if requested
	if getLeafNodes {
		logging.Log.Debugf(logCtx, "getting leaf nodes")
		err := qdrant_utils.RetrieveLeafNodes(logCtx, client, collectionName, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting leaf nodes: %q", err)
		}
	}
	if getSiblings {
		logging.Log.Debugf(logCtx, "getting sibling nodes")
		err := qdrant_utils.RetrieveDirectSiblingNodes(logCtx, client, collectionName, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting sibling nodes: %q", err)
		}
	}
	if getParent {
		logging.Log.Debugf(logCtx, "getting parent nodes")
		err := qdrant_utils.RetrieveParentNodes(logCtx, client, collectionName, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting parent nodes: %q", err)
		}
	}
	if getChildren {
		logging.Log.Debugf(logCtx, "getting child nodes")
		err := qdrant_utils.RetrieveChildNodes(logCtx, client, collectionName, &databaseResponse)
		if err != nil {
			logPanic(logCtx, "error getting child nodes: %q", err)
		}
	}
	return databaseResponse
}

// CreateKeywordsDbFilter creates a keywords filter for the KnowledgeDB.
//
// The function returns the keywords filter.
//
// Tags:
//   - @displayName: Keywords Filter
//
// Parameters:
//   - keywords: the keywords to be used for the filter
//   - needAll: flag to indicate whether all keywords are needed
//
// Returns:
//   - databaseFilter: the keywords filter
func CreateKeywordsDbFilter(keywords []string, needAll bool) (databaseFilter sharedtypes.DbArrayFilter) {
	var keywordsFilters sharedtypes.DbArrayFilter

	// -- Add the keywords filter if needed
	if len(keywords) > 0 {
		keywordsFilters = createDbArrayFilter(keywords, needAll)
	}

	return keywordsFilters
}

// CreateTagsDbFilter creates a tags filter for the KnowledgeDB.
//
// The function returns the tags filter.
//
// Tags:
//   - @displayName: Tags Filter
//
// Parameters:
//   - tags: the tags to be used for the filter
//   - needAll: flag to indicate whether all tags are needed
//
// Returns:
//   - databaseFilter: the tags filter
func CreateTagsDbFilter(tags []string, needAll bool) (databaseFilter sharedtypes.DbArrayFilter) {
	var tagsFilters sharedtypes.DbArrayFilter

	// -- Add the tags filter if needed
	if len(tags) > 0 {
		tagsFilters = createDbArrayFilter(tags, needAll)
	}

	return tagsFilters
}

// CreateMetadataDbFilter creates a metadata filter for the KnowledgeDB.
//
// The function returns the metadata filter.
//
// Tags:
//   - @displayName: Metadata Filter
//
// Parameters:
//   - fieldName: the name of the field
//   - fieldType: the type of the field
//   - filterData: the filter data
//   - needAll: flag to indicate whether all data is needed
//
// Returns:
//   - databaseFilter: the metadata filter
func CreateMetadataDbFilter(fieldName string, fieldType string, filterData []string, needAll bool) (databaseFilter sharedtypes.DbJsonFilter) {
	return createDbJsonFilter(fieldName, fieldType, filterData, needAll)
}

// CreateDbFilter creates a filter for the KnowledgeDB.
//
// The function returns the filter.
//
// Tags:
//   - @displayName: Create Filter
//
// Parameters:
//   - guid: the guid filter
//   - documentId: the document ID filter
//   - documentName: the document name filter
//   - level: the level filter
//   - tags: the tags filter
//   - keywords: the keywords filter
//   - metadata: the metadata filter
//
// Returns:
//   - databaseFilter: the filter
func CreateDbFilter(
	guid []string,
	documentId []string,
	documentName []string,
	level []string,
	tags sharedtypes.DbArrayFilter,
	keywords sharedtypes.DbArrayFilter,
	metadata []sharedtypes.DbJsonFilter) (databaseFilter sharedtypes.DbFilters) {
	var filters sharedtypes.DbFilters

	// -- Add the guid filter if needed
	if len(guid) > 0 {
		filters.GuidFilter = guid
	}

	// -- Add the document ID filter if needed
	if len(documentId) > 0 {
		filters.DocumentIdFilter = documentId
	}

	// -- Add the document name filter if needed
	if len(documentName) > 0 {
		filters.DocumentNameFilter = documentName
	}

	// -- Add the level filter if needed
	if len(level) > 0 {
		filters.LevelFilter = level
	}

	// -- Add the tags filter if needed
	if len(tags.FilterData) > 0 {
		filters.TagsFilter = tags
	}

	// -- Add the keywords filter if needed
	if len(keywords.FilterData) > 0 {
		filters.KeywordsFilter = keywords
	}

	// -- Add the metadata filter if needed
	if len(metadata) > 0 {
		filters.MetadataFilter = metadata
	}

	return filters
}

// AddDataRequest sends a request to the add_data endpoint.
//
// Tags:
//   - @displayName: Add Data
//
// Parameters:
//   - collectionName: name of the collection the request is sent to.
//   - data: the data to add.
func AddDataRequest(collectionName string, documentData []sharedtypes.DbData) {
	points := make([]*qdrant.PointStruct, len(documentData))
	for i, doc := range documentData {
		id := qdrant.NewIDUUID(doc.Guid.String())
		vector := qdrant.NewVectorsDense(doc.Embedding)
		payload, err := qdrant_utils.ToQdrantPayload(doc)
		if err != nil {
			logPanic(nil, "unable to transform document data to json: %q", err)
		}
		delete(payload, "guid")
		delete(payload, "embedding")
		points[i] = &qdrant.PointStruct{
			Id:      id,
			Vectors: vector,
			Payload: payload,
		}
	}

	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(nil, "unable to create qdrant client: %q", err)
	}

	ctx := context.TODO()

	resp, err := client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
		Wait:           qdrant.PtrOf(true),
	})
	if err != nil {
		logPanic(nil, "failed to insert data: %q", err)
	}
	logging.Log.Debugf(&logging.ContextMap{}, "successfully upserted %d points into qdrant collection %q: %q", len(points), collectionName, resp.GetStatus())
}

// CreateCollectionRequest sends a request to the collection endpoint.
//
// Tags:
//   - @displayName: Create Collection
//
// Parameters:
//   - collectionName: the name of the collection to create.
//   - vectorSize: the length of the vector S
//   - vectorDistance: the vector similarity distance algorithm to use for the vector index (cosine, dot, euclid, manhattan)
func CreateCollectionRequest(collectionName string, vectorSize uint64, vectorDistance string) {
	logCtx := &logging.ContextMap{}

	client, err := qdrant_utils.QdrantClient()
	if err != nil {
		logPanic(logCtx, "unable to create qdrant client: %q", err)
	}

	ctx := context.TODO()

	// check if collection already exists
	collectionExists, err := client.CollectionExists(ctx, collectionName)
	if err != nil {
		logPanic(logCtx, "unable to determine if collection already exists: %v", err)
	}
	if collectionExists {
		logging.Log.Debugf(logCtx, "collection %q already exists, skipping creation", collectionName)
		return
	}

	// create the collection
	err = client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant_utils.VectorDistance(vectorDistance),
		}),
	})
	if err != nil {
		logPanic(logCtx, "failed to create collection: %q", err)
	}
	logging.Log.Debugf(logCtx, "Created collection: %s", collectionName)

	// now create the default indexes (these are the things that other knowledgedb functions filter/search on)
	// does ID need to be indexed?
	indexes := []struct {
		name      string
		fieldType qdrant.FieldType
	}{
		{"level", qdrant.FieldType_FieldTypeKeyword},
		{"keywords", qdrant.FieldType_FieldTypeKeyword},
		{"document_id", qdrant.FieldType_FieldTypeKeyword},
		{"tags", qdrant.FieldType_FieldTypeKeyword},
	}
	for _, index := range indexes {
		request := qdrant.CreateFieldIndexCollection{
			CollectionName: collectionName,
			FieldName:      index.name,
			FieldType:      &index.fieldType,
		}
		res, err := client.CreateFieldIndex(ctx, &request)
		if err != nil {
			logPanic(logCtx, "error creating payload index on %q: %v", index.name, err)
		}
		logging.Log.Debugf(logCtx, "created payload index on %q: %q", index.name, res.Status)
	}
}
