.. _calling_functions:

Function Calls
==============

Flowkit exposes Go functions through gRPC interface calls.

gRPC Interface
--------------

The server provides two main RPC methods:

1. **ListFunctions** - Returns available function definitions
2. **RunFunction** - Executes a function with provided inputs
3. **StreamFunction** - Executes streaming functions

Request Format
--------------

Function execution requires:

- Function name (string)
- Input parameters (typed according to function definition)
- Optional metadata for authentication

Response Format
---------------

Function responses include:

- Output parameters (typed results)
- Error information if execution fails
- Execution metadata

Error Handling
--------------

The server returns gRPC status codes for:

- ``NOT_FOUND`` - Function does not exist
- ``UNAUTHENTICATED`` - Invalid or missing API key
- ``INTERNAL`` - Function execution error

Client Implementation
---------------------

Flowkit supports any gRPC-compatible client implementation.

Refer to the gRPC documentation for language-specific client generation.

Next: :doc:`agent_integration`
