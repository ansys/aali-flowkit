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

   - ``filter`` *(string)*, *optional*: Filter functions by category or name pattern.

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

Functions are organized by category:

   - ``generic``: SendRestAPICall, GenerateUUID, ExtractJSONStringField
   - ``cast``: CastAnyToString, CastAnyToFloat64, CastAnyToInt
   - ``data_extraction``: GetLocalFilesToExtract, GetLocalFileContent, LangchainSplitter
   - ``llm_handler``: PerformVectorEmbeddingRequest, PerformGeneralRequest
   - ``qdrant``: StoreElementsInVectorDatabase, QdrantCreateCollection
   - ``knowledge_db``: SimilaritySearch, SendVectorsToKnowledgeDB
   - ``auth``: CheckApiKeyAuthMongoDb, UpdateTotalTokenCountForCustomerMongoDb
   - ``ansys_gpt``: AnsysGPTPerformLLMRequest, AnsysGPTCheckProhibitedWords
   - ``ansys_mesh_pilot``: SimilartitySearchOnPathDescriptions
   - ``ansys_materials``: ExtractCriteriaSuggestions, SerializeResponse
