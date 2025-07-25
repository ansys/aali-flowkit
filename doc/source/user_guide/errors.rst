.. _errors:

Error codes
===========

Errors are returned as standard gRPC status errors, not embedded in the response message. When a function execution fails, the gRPC call returns an error with:

   - ``code`` *(int32)*: gRPC status code.
   - ``message`` *(string)*: Human-readable error message.
   - ``details`` *(google.protobuf.Any)*, *optional*: Additional error details.

Common error scenarios:

.. list-table:: FlowKit Error Codes
   :widths: 25 25 50
   :header-rows: 1

   * - Code
     - Status
     - Description
   * - 3
     - ``INVALID_ARGUMENT``
     - Function name is invalid or required parameters are missing.
   * - 5
     - ``NOT_FOUND``
     - Function does not exist in the registry.
   * - 7
     - ``PERMISSION_DENIED``
     - API key validation failed or insufficient permissions.
   * - 13
     - ``INTERNAL``
     - Function execution failed with an internal error.
   * - 14
     - ``UNAVAILABLE``
     - Required service (LLM, Qdrant, etc) is unavailable.
