.. _functions_dev:

Adding Custom Functions
=======================

Custom functions extend FlowKit's capabilities. Functions intended for external use must be exported (capitalized) Go functions with specific documentation tags.

Function Structure
------------------

Functions must follow this pattern:

- Start with a comment describing the function
- Include Tags section with @displayName
- Document all Parameters
- Document all Returns
- Use panic for error handling

Requirements
------------

- Function name must be exported (capitalized)
- Include docstring with description
- Use ``@displayName`` tag for UI display
- Document all parameters and return values
- Error handling via panic (caught by gRPC server)

Function Discovery
------------------

FlowKit automatically discovers functions from embedded Go files. Functions are embedded in ``main.go`` using ``//go:embed`` directives and extracted at startup.

No manual registration is required - simply create your function in the appropriate category file.

Function Implementation Examples
-----------------------------------

Simple Function
~~~~~~~~~~~~~~~

.. code-block:: go

   // GenerateUUID generates a new UUID (Universally Unique Identifier).
   //
   // Tags:
   //   - @displayName: Generate UUID
   //
   // Returns:
   //   - a string representation of the generated UUID
   func GenerateUUID() string {
       return strings.Replace(uuid.New().String(), "-", "", -1)
   }

Function with Parameters
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

   // SendRestAPICall sends an API call to the specified URL with headers and query parameters.
   //
   // Tags:
   //   - @displayName: REST Call
   //
   // Parameters:
   //   - requestType: the type of the request (GET, POST, PUT, PATCH, DELETE)
   //   - endpoint: the URL to send the request to
   //   - header: the headers to include in the request
   //   - query: the query parameters to include in the request
   //   - jsonBody: the body of the request as a JSON string
   //
   // Returns:
   //   - success: a boolean indicating whether the request was successful
   //   - returnJsonBody: the JSON body of the response as a string
   func SendRestAPICall(requestType string, endpoint string, header map[string]string, 
                        query map[string]string, jsonBody string) (success bool, returnJsonBody string) {
       // verify correct request type
       if requestType != "GET" && requestType != "POST" && requestType != "PUT" && requestType != "PATCH" && requestType != "DELETE" {
           panic(fmt.Sprintf("Invalid request type: %v", requestType))
       }

       // Parse the URL and add query parameters
       parsedURL, err := url.Parse(endpoint)
       if err != nil {
           panic(fmt.Sprintf("Error parsing URL: %v", err))
       }

       q := parsedURL.Query()
       for key, value := range query {
           q.Add(key, value)
       }
       parsedURL.RawQuery = q.Encode()

       // Create the HTTP request
       var req *http.Request
       if jsonBody != "" {
           req, err = http.NewRequest(requestType, parsedURL.String(), bytes.NewBuffer([]byte(jsonBody)))
       } else {
           req, err = http.NewRequest(requestType, parsedURL.String(), nil)
       }
       if err != nil {
           panic(fmt.Sprintf("Error creating request: %v", err))
       }

       // Add headers
       for key, value := range header {
           req.Header.Add(key, value)
       }

       // Execute the request
       client := &http.Client{}
       resp, err := client.Do(req)
       if err != nil {
           panic(fmt.Sprintf("Error executing request: %v", err))
       }
       defer resp.Body.Close()

       // Read the response body
       body, err := io.ReadAll(resp.Body)
       if err != nil {
           panic(fmt.Sprintf("Error reading response body: %v", err))
       }

       // Check if the response code is successful (2xx)
       success = resp.StatusCode >= 200 && resp.StatusCode < 300

       return success, string(body)
   }

Complex Function with Custom Types
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

   // SendVectorsToKnowledgeDB sends the given vector to the KnowledgeDB and
   // returns the most relevant data.
   //
   // Tags:
   //   - @displayName: Similarity Search
   //
   // Parameters:
   //   - vector: the vector to be sent to the KnowledgeDB
   //   - keywords: the keywords to filter the results
   //   - keywordsSearch: flag to enable keywords search
   //   - collection: the collection name
   //   - similaritySearchResults: number of results to return
   //   - similaritySearchMinScore: minimum score for results
   //
   // Returns:
   //   - databaseResponse: array of the most relevant data
   func SendVectorsToKnowledgeDB(vector []float32, keywords []string, keywordsSearch bool, 
                                 collection string, similaritySearchResults int, 
                                 similaritySearchMinScore float64) (databaseResponse []sharedtypes.DbResponse) {
       client, err := createKnowledgeDBClient()
       if err != nil {
           logPanic(nil, "unable to create knowledge DB client: %q", err)
       }
       
       searchRequest := &pb.SearchRequest{
           Vector:     vector,
           Keywords:   keywords,
           Collection: collection,
           Limit:      int32(similaritySearchResults),
           MinScore:   float32(similaritySearchMinScore),
       }
       
       response, err := client.Search(context.Background(), searchRequest)
       if err != nil {
           logPanic(nil, "search request failed: %q", err)
       }
       
       return convertToDbResponse(response.Results)
   }

MCP Function Example
~~~~~~~~~~~~~~~~~~~~

Here's an example of an MCP integration function:

.. code-block:: go

   // ListTools lists all available tools from an MCP server.
   //
   // Tags:
   //   - @displayName: List MCP Tools
   //
   // Parameters:
   //   - serverURL: the URL of the MCP server
   //
   // Returns:
   //   - tools: list of available MCP tools
   //   - error: any error that occurred
   func ListTools(serverURL string) ([]mcptypes.MCPTool, error) {
       client, err := mcpclient.GetClient(serverURL)
       if err != nil {
           panic(fmt.Sprintf("Failed to get MCP client: %v", err))
       }
       
       tools, err := client.ListTools()
       if err != nil {
           panic(fmt.Sprintf("Failed to list tools: %v", err))
       }
       
       return tools, nil
   }

Error Handling Pattern
~~~~~~~~~~~~~~~~~~~~~~

FlowKit functions use ``panic`` with formatted error messages for error handling:

.. code-block:: go

   parsedURL, err := url.Parse(endpoint)
   if err != nil {
       panic(fmt.Sprintf("Error parsing URL: %v", err))
   }

   // Or with validation
   if requestType != "GET" && requestType != "POST" {
       panic(fmt.Sprintf("Invalid request type: %v", requestType))
   }

Function Discovery Process
---------------------------

FlowKit uses an automatic function discovery system:

1. **File Embedding**: Functions are embedded at compile time using ``//go:embed`` directives in ``main.go``
2. **Category Mapping**: Each category file is mapped to a category name
3. **Automatic Extraction**: At startup, ``functiondefinitions.ExtractFunctionDefinitionsFromPackage()`` parses each embedded file
4. **Function Registration**: Discovered functions are automatically added to ``internalstates.AvailableFunctions``

.. code-block:: go

   // In main.go - embedded files
   //go:embed pkg/externalfunctions/generic.go
   var genericFile string
   
   //go:embed pkg/externalfunctions/llmhandler.go
   var llmHandlerFile string
   
   // File mapping
   files := map[string]string{
       "generic":      genericFile,
       "llm_handler":  llmHandlerFile,
       "knowledge_db": knowledgeDBFile,
       // ... other categories
   }
   
   // Automatic discovery
   for category, file := range files {
       err := functiondefinitions.ExtractFunctionDefinitionsFromPackage(file, category)
   }

The ``ExternalFunctionsMap`` in ``pkg/externalfunctions/externalfunctions.go`` serves as a runtime function registry, but registration happens automatically through discovery.

Category Files Structure
------------------------

Functions are organized in category files within ``pkg/externalfunctions/``:

- ``generic.go`` - General purpose utility functions
- ``cast.go`` - Type conversion functions
- ``llmhandler.go`` - LLM integration and AI functions
- ``knowledgedb.go`` - Database operations and vector search
- ``dataextraction.go`` - File processing and content extraction
- ``auth.go`` - Authentication and authorization functions
- ``ansysgpt.go`` - Ansys GPT specific functions
- ``ansysmeshpilot.go`` - Ansys Mesh Pilot functions
- ``ansysmaterials.go`` - Ansys Materials functions
- ``qdrant.go`` - Qdrant vector database functions
- ``mcp.go`` - Model Context Protocol functions
- ``rhsc.go`` - RHSC specific functions

Private functions are organized in ``pkg/privatefunctions/`` subdirectories:

- ``pkg/privatefunctions/generic/`` - Generic helper functions
- ``pkg/privatefunctions/codegeneration/`` - Code generation utilities
- ``pkg/privatefunctions/graphdb/`` - Graph database helpers
- ``pkg/privatefunctions/qdrant/`` - Qdrant database helpers
