.. _agent_integration:

Agent Integration
=================

Flowkit provides gRPC functions for AI agent integration.

Integration Architecture
------------------------

AI agents connect to Flowkit through:

1. **gRPC Client** - Standard gRPC client implementation
2. **Function Discovery** - ListFunctions call to enumerate capabilities
3. **Function Execution** - RunFunction calls with typed parameters
4. **Error Handling** - Standard gRPC status codes

Function Discovery
------------------

Agents can dynamically discover available functions:

.. code-block:: protobuf

   rpc ListFunctions(ListFunctionsRequest) returns (ListFunctionsResponse)

Response includes function definitions with:

- Function name and description
- Input parameter types and constraints
- Output parameter types
- Function category

Function Execution
------------------

Agents execute functions through:

.. code-block:: protobuf

   rpc RunFunction(FunctionInputs) returns (FunctionOutputs)

Parameters include:

- Function name (string)
- Typed input parameters
- Authentication metadata

Authentication
--------------

API key authentication via gRPC metadata:

.. code-block:: text

   x-api-key: <api-key-value>

Error Codes
-----------

Standard gRPC status codes:

- ``UNAUTHENTICATED`` - Missing or invalid API key
- ``NOT_FOUND`` - Function not found
- ``INVALID_ARGUMENT`` - Invalid input parameters
- ``INTERNAL`` - Function execution error

Next: :doc:`../advanced/index`
