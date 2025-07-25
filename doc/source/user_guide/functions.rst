.. _functions:

FlowKit Functions
=================

In this section the functions available through FlowKit and how to execute them is explained in detail.

Function Discovery
~~~~~~~~~~~~~~~~~~

The system provides a ``ListFunctions`` RPC method that returns all available functions:

.. code-block:: protobuf

   rpc ListFunctions(ListFunctionsRequest) returns (ListFunctionsResponse);

**ListFunctionsRequest**

   Empty request - no parameters required.

**ListFunctionsResponse**

   - ``functions`` *(map<string, FunctionDefinition>)*: Map of function names to their definitions.

Function Execution
~~~~~~~~~~~~~~~~~~

Execute functions using the ``RunFunction`` RPC method:

.. code-block:: protobuf

   rpc RunFunction(FunctionInputs) returns (FunctionOutputs);

**FunctionInputs**

   - ``name`` *(string)*, *mandatory*: Function name to execute.
   - ``inputs`` *(map<string, Value>)*, *mandatory*: Input parameters as key-value pairs.

Here is an example request:

.. code-block:: json

   {
       "name": "GenerateUUID",
       "inputs": {}
   }

Stream Function Execution
~~~~~~~~~~~~~~~~~~~~~~~~~

For functions that produce streaming outputs, use the ``StreamFunction`` RPC method:

.. code-block:: protobuf

   rpc StreamFunction(FunctionInputs) returns (stream FunctionOutputs);

This method accepts the same ``FunctionInputs`` as ``RunFunction`` but returns a stream of ``FunctionOutputs``, allowing functions to send multiple responses over time. This is useful for:

- Long-running operations with progress updates
- Functions that generate data incrementally
- Real-time data processing pipelines

Available Functions
~~~~~~~~~~~~~~~~~~~

FlowKit includes **185 functions** organized by category. Here are some key functions from each category:

   - ``generic`` (8 functions): SendRestAPICall, GenerateUUID, ExtractJSONStringField, StringConcat
   - ``cast`` (multiple functions): CastAnyToString, CastAnyToFloat64, CastAnyToInt, CastStringToAny
   - ``data_extraction`` (10 functions): GetLocalFilesToExtract, GetLocalFileContent, LangchainSplitter, AddDataRequest
   - ``llm_handler`` (22 functions): PerformVectorEmbeddingRequest, PerformGeneralRequest, BuildLibraryContext, AppendMessageHistory
   - ``qdrant`` (2 functions): QdrantCreateCollection, QdrantInsertData
   - ``knowledge_db`` (11 functions): SimilaritySearch, SendVectorsToKnowledgeDB, GetListCollections, GeneralGraphDbQuery
   - ``auth`` (11 functions): CheckApiKeyAuthMongoDb, UpdateTotalTokenCountForCustomerMongoDb, SendLogicAppNotificationEmail
   - ``ansys_gpt`` (20 functions): AnsysGPTPerformLLMRequest, AnsysGPTCheckProhibitedWords, AnsysGPTReturnIndexList
   - ``ansys_mesh_pilot`` (25 functions): SimilartitySearchOnPathDescriptions, FetchPropertiesFromPathDescription, SynthesizeActions
   - ``ansys_materials`` (12 functions): ExtractCriteriaSuggestions, SerializeResponse, FilterOutDuplicateAttributes
   - ``mcp`` (4 functions): ListAll, ExecuteTool, GetResource, GetSystemPrompt
   - ``rhsc`` (1 function): SetCopilotGenerateRequestJsonBody

Use the ``ListFunctions`` RPC method to get the complete list of available functions with their full signatures and documentation.
