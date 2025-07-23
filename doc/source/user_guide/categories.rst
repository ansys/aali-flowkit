.. _categories:

FlowKit Function Categories
===========================

FlowKit provides functions for engineering data processing and AI integration within the AALI ecosystem. Each category represents actual implemented capabilities.

AALI LLM Handler Functions
--------------------------

FlowKit connects to AALI's language model service via WebSocket at ``ws://aali-llm:9003``:

- ``PerformVectorEmbeddingRequest`` - Convert engineering text to vector embeddings
- ``PerformBatchEmbeddingRequest`` - Batch process multiple documents for embeddings
- ``PerformGeneralRequest`` - Stream LLM completions with token tracking
- ``PerformCodeLLMRequest`` - Generate code for engineering workflows
- ``BuildLibraryContext`` - Construct context from documentation libraries

**Real Use Case**: Process Ansys documentation through AALI-LLM to generate embeddings for semantic search. The WebSocket connection handles streaming responses and tracks token usage per request.

Qdrant Vector Search Integration
--------------------------------

FlowKit provides native Qdrant integration at ``qdrant:6334``:

- ``QdrantCreateCollection`` - Set up collections with custom vector dimensions
- ``QdrantInsertData`` - Store embeddings with metadata

**Real Use Case**: Store embeddings from Ansys technical documentation in Qdrant collections. Engineers search for similar problems using natural language queries against the vector database.

Graph Database Operations (KuzuDB)
----------------------------------

FlowKit integrates with AALI's graph database at ``aali-graphdb:8080``:

- ``GeneralGraphDbQuery`` - Execute Cypher queries on the knowledge graph
- ``SendVectorsToKnowledgeDB`` - Store vectors with graph relationships
- ``RetrieveDependencies`` - Find connected documentation and code
- ``SimilaritySearch`` - Hybrid vector-graph search
- ``CreateDbFilter`` - Build filters for metadata and keywords

**Real Use Case**: Build knowledge graphs from engineering specifications where nodes represent components and edges represent dependencies. Combine with vector search for context-aware retrieval.

AnsysGPT Domain-Specific Functions
----------------------------------

Engineering-focused AI capabilities with Ansys domain knowledge:

- ``AnsysGPTCheckProhibitedWords`` - Validate queries against compliance rules
- ``AnsysGPTExtractFieldsFromQuery`` - Parse engineering parameters from natural language
- ``AnsysGPTPerformLLMRequest`` - Stream responses with engineering context
- ``AnsysGPTACSSemanticHybridSearchs`` - Search Azure Cognitive Search indices
- ``AnsysGPTReorderSearchResponseAndReturnOnlyTopK`` - Rank results by relevance

**Real Use Case**: Engineers query "fatigue analysis for titanium at 500°C." FlowKit extracts material (titanium), property (fatigue), and condition (500°C) to search Ansys knowledge bases with proper citations.

Ansys Mesh Pilot Functions
--------------------------

Mesh generation and optimization for simulations:

- ``SimilartitySearchOnPathDescriptions`` - Find similar mesh configurations
- ``AppendMeshPilotHistory`` - Track mesh generation iterations

**Real Use Case**: Given a CAD geometry, find similar past meshing solutions in Qdrant and generate an optimized mesh configuration based on simulation requirements.

**Configuration**: Meshpilot functions require advanced workflow configuration. See :ref:`advanced-configuration` for tool setup, database connections, and prompt templates.

Ansys Materials Functions
-------------------------

Material property extraction and search:

- ``ExtractCriteriaSuggestions`` - Parse material requirements from queries
- ``AddGuidsToAttributes`` - Link material properties to database IDs
- ``FilterOutNonExistingAttributes`` - Validate against material database
- ``SerializeResponse`` - Format results for engineering applications

**Real Use Case**: Extract material criteria from "high-temperature resistant alloy with yield strength > 800 MPa" and query the Ansys Granta materials database.

Data Extraction Pipeline
------------------------

Process engineering documentation and code:

- ``GetLocalFilesToExtract`` - Scan directories for technical documents
- ``GetGithubFilesToExtract`` - Extract from Ansys GitHub repositories
- ``LangchainSplitter`` - Split documents into semantic chunks
- ``CreateGeneralDataExtractionDocumentObjects`` - Structure for vector storage
- ``StoreElementsInVectorDatabase`` - Pipeline to Qdrant with embeddings

**Real Use Case**: Ingest Ansys PyFluent documentation from GitHub, split into chunks, generate embeddings via AALI-LLM, and store in Qdrant for retrieval-augmented generation.

Authentication Functions
------------------------

API key validation using MongoDB (auth-specific, not general database):

- ``CheckApiKeyAuthMongoDb`` - Validate API keys against MongoDB
- ``UpdateTotalTokenCountForCustomerMongoDb`` - Track LLM token usage
- ``CheckTokenLimitReached`` - Enforce usage quotas
- ``DenyCustomerAccessAndSendWarningMongoDb`` - Handle limit violations

**Real Use Case**: Each FlowKit request validates the API key, increments token count in MongoDB, and enforces per-customer limits for AALI-LLM usage.

Core Utility Functions
----------------------

Essential operations for workflow automation:

- ``SendRestAPICall`` - Call external APIs with authentication
- ``GenerateUUID`` - Create unique identifiers for tracking
- ``ExtractJSONStringField`` - Parse API responses
- ``CastAnyToString`` and numerous other cast functions - Type conversions

**Real Use Case**: Chain API calls to Ansys solvers, extract results from JSON responses, and convert between data types for downstream processing.

## Verified Capabilities

Based on code analysis in:
- ``pkg/externalfunctions/`` - Function implementations
- ``pkg/privatefunctions/qdrant/`` - Qdrant client
- ``pkg/privatefunctions/graphdb/`` - Graph database client
- ``configs/config.yaml`` - Service endpoints

**NOT Found**: PostgreSQL, Redis, CRM integrations, invoice processing, or generic business automation features. FlowKit is specifically built for engineering data processing within AALI.
