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

FlowKit includes **over 180 functions** organized by category and continuously adds new functionality. Key categories include:

   - ``generic``: General purpose utility functions (REST API calls, UUID generation, JSON operations)
   - ``cast``: Type conversion functions for all Go primitive types
   - ``data_extraction``: File processing and content extraction
   - ``llm_handler``: LLM integration and AI operations
   - ``qdrant``: Vector database operations
   - ``knowledge_db``: Knowledge database and similarity search
   - ``auth``: Authentication and authorization
   - ``ansys_gpt``: Ansys GPT specific functionality
   - ``ansys_mesh_pilot``: Mesh generation and analysis tools
   - ``ansys_materials``: Materials database integration
   - ``mcp``: Model Context Protocol integration
   - ``rhsc``: Red Hat Service Catalog integration

Use the ``ListFunctions`` RPC method to get the complete list of available functions with their full signatures and documentation.
