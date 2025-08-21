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
	"os"
	"strings"
	"sync"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// similarityElement represents a single element in the similarity search result.
type similarityElement struct {
	Score float64                `json:"distance"`
	Data  sharedtypes.DbResponse `json:"data"`
}

// similaritySearchInput represents the input for the similarity search function.
type similaritySearchInput struct {
	CollectionName    string                `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
	EmbeddedVector    []float32             `json:"embedded_vector" description:"Embedded vector used for searching. Required for retrieval." required:"true"`
	MaxRetrievalCount int                   `json:"max_retrieval_count" description:"Maximum number of results to be retrieved. If it is not specified, the default value is milvus.MaxSearchRetrievalCount. Optional." required:"false"`
	OutputFields      []string              `json:"output_fields" description:"Fields to be included in the output. If not specified all fields will be retrieved.Optional" required:"false"`
	Filters           sharedtypes.DbFilters `json:"filters" description:"Filter for the query. Optional." required:"false"`
	MinScore          float64               `json:"min_score" description:"Filter objects with a score higher than the specified minimum. Optional." required:"false"`
	GetLeafNodes      bool                  `json:"get_leaf_nodes" description:"Flag to indicate whether to retrieve all the leaf nodes in the result node branch. Set to true to include the leaf nodes. Optional." required:"false"`
	GetSiblings       bool                  `json:"get_siblings" description:"Flag to indicate whether to retrieve the previous and next node to the result nodes. Set to true to include the siblings. Optional." required:"false"`
	GetParent         bool                  `json:"get_parent" description:"Flag to indicate whether to retrieve the parent object. Set to true to include the parent object. Optional." required:"false"`
	GetChildren       bool                  `json:"get_children" description:"Flag to indicate whether to retrieve the children objects. Set to true to include the children objects. Optional." required:"false"`
}

// similaritySearchOutput represents the output for the similarity search function.
type similaritySearchOutput struct {
	SimilarityResult []similarityElement `json:"similarity_result" description:"Similarity Result"`
}

// queryInput represents the input for the query function.
type queryInput struct {
	CollectionName    string                `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
	MaxRetrievalCount int                   `json:"max_retrieval_count" description:"Maximum number of results to be retrieved. If not specified, the default value is retrieve all database. If the number of results is too big for the database, the request will be cancelled. Optional." required:"false"`
	OutputFields      []string              `json:"output_fields" description:"Fields to be included in the output. If not specified all fields will be retrieved.Optional" required:"false"`
	Filters           sharedtypes.DbFilters `json:"filters" description:"Filter for the query. At least one filter must be defined." required:"true"`
}

// queryOutput represents the output for the query function.
type queryOutput struct {
	QueryResult []sharedtypes.DbResponse `json:"queryResult" description:"Returns the results of the query."`
}

// retrieveDependenciesInput represents the input for the retrieveDependencies function.
type retrieveDependenciesInput struct {
	CollectionName        string                    `json:"collection_name" description:"Name of the collection to which the data objects will be added. Required for adding data." required:"true"`
	RelationshipName      string                    `json:"relationship_name" description:"Name of the relationship to retrieve dependencies for. Required for retrieving dependencies." required:"true"`
	RelationshipDirection string                    `json:"relationship_direction" description:"Direction of the relationship to retrieve dependencies for. It can be either 'in', 'out' or 'both'. Required for retrieving dependencies." required:"true"`
	SourceDocumentId      string                    `json:"source_document_id" description:"Document ID of the source node. Required for retrieving dependencies." required:"true"`
	NodeTypesFilter       sharedtypes.DbArrayFilter `json:"node_types_filter" description:"Filter based on node types. Use MilvusArrayFilter for specifying node type filtering criteria. Optional." required:"false"`
	DocumentTypesFilter   []string                  `json:"document_types_filter" description:"Filter based on document types. Use MilvusArrayFilter for specifying document type filtering criteria. Optional." required:"false"`
	MaxHopsNumber         int                       `json:"max_hops_number" description:"Maximum number of hops to traverse. Optional." required:"true"`
}

// retrieveDependenciesOutput represents the output for the retrieveDependencies function.
type retrieveDependenciesOutput struct {
	Success         bool     `json:"success" description:"Returns true if the collections were listed successfully. Returns false or an error if not."`
	DependenciesIds []string `json:"dependencies_ids" description:"A list of document IDs that are dependencies of the specified source node."`
}

// summaryCounters represents the summary counters structure for the Neo4j query.
type summaryCounters struct {
	NodesCreated         int `json:"nodes_created"`
	NodesDeleted         int `json:"nodes_deleted"`
	RelationshipsCreated int `json:"relationships_created"`
	RelationshipsDeleted int `json:"relationships_deleted"`
	PropertiesSet        int `json:"properties_set"`
	LabelsAdded          int `json:"labels_added"`
	LabelsRemoved        int `json:"labels_removed"`
	IndexesAdded         int `json:"indexes_added"`
	IndexesRemoved       int `json:"indexes_removed"`
	ConstraintsAdded     int `json:"constraints_added"`
	ConstraintsRemoved   int `json:"constraints_removed"`
}

// ACSVectorQuery represents the vector query structure for the Azure Cognitive Search.
type ACSVectorQuery struct {
	Kind   string    `json:"kind"`
	K      int       `json:"k"`
	Vector []float32 `json:"vector"`
	Fields string    `json:"fields"`
}

// ACSRequest represents the request structure for the Azure Cognitive Search.
type ACSSearchRequest struct {
	Search                string           `json:"search"`
	VectorQueries         []ACSVectorQuery `json:"vectorQueries"`
	VectorFilterMode      string           `json:"vectorFilterMode"`
	Filter                string           `json:"filter"`
	QueryType             string           `json:"queryType"`
	SemanticConfiguration string           `json:"semanticConfiguration"`
	Top                   int              `json:"top"`
	Select                string           `json:"select"`
	Count                 bool             `json:"count"`
}

// ACSSearchResponseStruct represents the response structure for the Azure Cognitive Search for granular-ansysgpt, ansysgpt-documentation-2023r2, scade-documentation-2023r2, ansys-dot-com-marketing.
type ACSSearchResponseStruct struct {
	OdataContext string                          `json:"@odata.context"`
	OdataCount   int                             `json:"@odata.count"`
	Value        []sharedtypes.ACSSearchResponse `json:"value"`
}

// ACSSearchResponseStruct represents the response structure for the Azure Cognitive Search for ansysgpt-alh & ansysgpt-scbu.
type ACSSearchResponseStructALH struct {
	OdataContext string                 `json:"@odata.context"`
	OdataCount   int                    `json:"@odata.count"`
	Value        []ACSSearchResponseALH `json:"value"`
}

// ACSSearchResponse represents the response structure for the Azure Cognitive Search for ansysgpt-alh & ansysgpt-scbu.
type ACSSearchResponseALH struct {
	SourcetitleSAP      string  `json:"sourcetitleSAP"`
	SourceURLSAP        string  `json:"sourceURLSAP"`
	SourcetitleDCB      string  `json:"sourcetitleDCB"`
	SourceURLDCB        string  `json:"sourceURLDCB"`
	Content             string  `json:"content"`
	TypeOFasset         string  `json:"typeOFasset"`
	Physics             string  `json:"physics"`
	Product             string  `json:"product"`
	Version             string  `json:"version"`
	Weight              float64 `json:"weight"`
	TokenSize           int     `json:"token_size"`
	SearchScore         float64 `json:"@search.score"`
	SearchRerankerScore float64 `json:"@search.rerankerScore"`
	IndexName           string  `json:"indexName"`
}

// ACSSearchResponseStruct represents the response structure for the Azure Cognitive Search for lsdyna-documentation-r14.
type ACSSearchResponseStructLSdyna struct {
	OdataContext string                    `json:"@odata.context"`
	OdataCount   int                       `json:"@odata.count"`
	Value        []ACSSearchResponseLSdyna `json:"value"`
}

// ACSSearchResponse represents the response structure for the Azure Cognitive Search for lsdyna-documentation-r14.
type ACSSearchResponseLSdyna struct {
	Title               string  `json:"title"`
	Url                 string  `json:"url"`
	Content             string  `json:"content"`
	TypeOFasset         string  `json:"typeOFasset"`
	Physics             string  `json:"physics"`
	Product             string  `json:"product"`
	TokenSize           int     `json:"token_size"`
	SearchScore         float64 `json:"@search.score"`
	SearchRerankerScore float64 `json:"@search.rerankerScore"`
	IndexName           string  `json:"indexName"`
}

// ACSSearchResponseStructCrtech represents the response structure for the Azure Cognitive Search for external-crtech-thermal-desktop.
type ACSSearchResponseStructCrtech struct {
	OdataContext string                    `json:"@odata.context"`
	OdataCount   int                       `json:"@odata.count"`
	Value        []ACSSearchResponseCrtech `json:"value"`
}

// ACSSearchResponseCrtech represents the response structure for the Azure Cognitive Search for external-crtech-thermal-desktop.
type ACSSearchResponseCrtech struct {
	Physics             string  `json:"physics"`
	SourceTitleLvl3     string  `json:"sourceTitle_lvl3"`
	SourceURLLvl3       string  `json:"sourceURL_lvl3"`
	TokenSize           int     `json:"token_size"`
	SourceTitleLvl2     string  `json:"sourceTitle_lvl2"`
	Weight              float64 `json:"weight"`
	SourceURLLvl2       string  `json:"sourceURL_lvl2"`
	Product             string  `json:"product"`
	Content             string  `json:"content"`
	TypeOFasset         string  `json:"typeOFasset"`
	Version             string  `json:"version"`
	BridgeId            string  `json:"bridge_id"`
	SearchScore         float64 `json:"@search.score"`
	SearchRerankerScore float64 `json:"@search.rerankerScore"`
	IndexName           string  `json:"indexName"`
}

// TransportType represents the MCP transport protocol options as requested by Gautam
// This enables dropdown/checkbox selection in the workflow configurator
type TransportType string

// MCP transport protocol constants
const (
	TransportWebSocket TransportType = "websocket" // Default - for remote MCP servers over WebSocket
	TransportSSE       TransportType = "sse"       // Server-Side Events for HTTP-based streaming
	TransportSTDIO     TransportType = "stdio"     // Standard I/O for local MCP servers
)

// MCPConfig represents the configuration for MCP (Model Context Protocol) connections.
// It includes transport selection, authentication, and connection options.
type MCPConfig struct {
	// ServerURL is the URL of the MCP server (required)
	// Examples: "ws://localhost:3000", "wss://mcp.example.com", "http://localhost:8080" (for SSE)
	ServerURL string `json:"serverURL" description:"The URL of the MCP server" required:"true"`

	// Transport specifies the transport protocol to use
	// Options: "websocket" (default), "sse", "stdio"
	// Currently only "websocket" is implemented
	Transport string `json:"transport" description:"Transport protocol: websocket, sse, or stdio" required:"false"`

	// AuthToken is an optional authentication token (as requested by Gautam)
	// Can be a direct value or environment variable reference like ${MCP_TOKEN}
	// When provided, it will be sent as a Bearer token in the Authorization header
	AuthToken string `json:"authToken" description:"Optional authentication token" required:"false"`

	// Timeout specifies the connection timeout in seconds
	// Default: 30 seconds if not specified
	Timeout int `json:"timeout" description:"Connection timeout in seconds (default: 30)" required:"false"`
}

// GetAuthToken returns the authentication token, resolving environment variables if needed.
// If the token starts with ${ and ends with }, it's treated as an environment variable reference.
// Example: ${MCP_TOKEN} will return the value of the MCP_TOKEN environment variable.
func (config *MCPConfig) GetAuthToken() string {
	if len(config.AuthToken) > 3 &&
		strings.HasPrefix(config.AuthToken, "${") &&
		strings.HasSuffix(config.AuthToken, "}") {
		envVar := config.AuthToken[2 : len(config.AuthToken)-1]
		return os.Getenv(envVar)
	}
	return config.AuthToken
}

// AnsysGPTRetrieverModuleRequest represents the request structure for the Ansys GPT Retriever Module.
type AnsysGPTRetrieverModuleRequest struct {
	UserInput     string `json:"user_input"`
	DataSource    string `json:"data_source"`
	FilterPhysics string `json:"filter_physics"`
	FilterVersion string `json:"filter_version,omitempty"`
	FilterProduct string `json:"filter_product,omitempty"`
	NumDocs       int    `json:"num_docs"`
	Platform      string `json:"platform"`
}

// DataExtractionBranch represents the branch structure for the data extraction.
type DataExtractionBranch struct {
	Text             string
	ChildDataObjects []*sharedtypes.DbData
	ChildDataIds     []uuid.UUID
}

// DataExtractionLLMInputChannelItem represents the input channel item for the data extraction llm handler workers.
type DataExtractionLLMInputChannelItem struct {
	Data                *sharedtypes.DbData
	Adapter             string
	ChatRequestType     string
	MaxNumberOfKeywords uint32

	InstructionSequenceWaitGroup *sync.WaitGroup
	Lock                         *sync.Mutex

	EmbeddingVector []float32
	Summary         string
	Keywords        []string
	CollectionName  string
}

type DataExtractionSplitterServiceRequest struct {
	DocumentContent []byte `json:"document_content"`
	ChunkSize       int    `json:"chunk_size"`
	ChunkOverlap    int    `json:"chunk_overlap"`
}

type DataExtractionSplitterServiceResponse struct {
	Chunks []string `json:"chunks"`
}

type TokenCountUpdateRequest struct {
	InputToken  int    `json:"input_token"`
	OutputToken int    `json:"output_token"`
	Platform    string `json:"platform"`
}

type GeneralDataExtractionDocument struct {
	DocumentName  string           `json:"document_name"`
	Guid          string           `json:"guid"`
	PreviousChunk string           `json:"previous_chunk"`
	NextChunk     string           `json:"next_chunk"`
	DenseVector   []float32        `json:"dense_vector"`
	SparseVector  map[uint]float32 `json:"sparse_vector"`
	Text          string           `json:"text"`
}

// MongoDbContext is the structure for the mongodb client
type MongoDbContext struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
}

type MongoDbCustomerObject struct {
	ApiKey          string `bson:"api_key"`
	CustomerName    string `bson:"customer_name"`
	AccessDenied    bool   `bson:"access_denied"`
	TotalTokenCount int    `bson:"total_token_usage"`
	TokenLimit      int    `bson:"token_limit"`
	WarningSent     bool   `bson:"warning_sent"`
}

type MongoDbCustomerObjectDisco struct {
	UserId                string                                  `bson:"user_id"`
	AccessDenied          bool                                    `bson:"access_denied"`
	ModelIdTokenCountDict map[string]MongoDbTokenCountObjectDisco `bson:"model_id_token_count_dict"`
	WarningSent           bool                                    `bson:"warning_sent"`
}

type MongoDbTokenCountObjectDisco struct {
	InputTokenCount     int   `bson:"input_token_count"`
	OutputTokenCount    int   `bson:"output_token_count"`
	TokenLimit          int   `bson:"token_limit"`
	TokenLimitTimestamp int64 `bson:"token_limit_timestamp"`
}

// EmailRequest represents the structure of the POST request body
type EmailRequest struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Content string `json:"content"`
}

type copilotGenerateRequest struct {
	Query     string                 `json:"query"`
	SessionID string                 `json:"session_id"`
	Mode      string                 `json:"mode"`
	Timeout   int                    `json:"timeout"`
	Priority  int                    `json:"priority"`
	Options   copilotGenerateOptions `json:"options"`
}

type copilotGenerateOptions struct {
	AgentPreference  string `json:"agent_preference"`
	SaveIntermediate bool   `json:"save_intermediate"`
	SimilarityTopK   int    `json:"similarity_top_k"`
	NoCritique       bool   `json:"no_critique"`
	MaxIterations    int    `json:"max_iterations"`
	ForceAzure       bool   `json:"force_azure"`
}

// kvdbSingleResponse is a struct to hold the response from the KVDB for a single entry.
type kvdbSingleResponse struct {
	Value string `json:"value"`
}

// kvdbErrorResponse is a struct to hold the error response from the KVDB.
type kvdbErrorResponse struct {
	Error string `json:"error"`
}

type materialsCustomerObject struct {
	ApiKey          string `json:"api_key"`
	CustomerName    string `json:"customer_name"`
	AccessDenied    bool   `json:"access_denied"`
	TotalTokenCount int    `json:"total_token_usage"`
	TokenLimit      int    `json:"token_limit"`
	WarningSent     bool   `json:"warning_sent"`
}
