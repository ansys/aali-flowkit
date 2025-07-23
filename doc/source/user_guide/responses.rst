.. _responses:

FlowKit Responses
=================

In this section the responses a client can expect to receive via gRPC after executing a function is explained in detail.

Response Format
~~~~~~~~~~~~~~~

The system produces responses as a stream of ``ExecuteFunctionResponse`` messages:

**ExecuteFunctionResponse**

   - ``outputs`` *(map<string, Value>)*: Function output values as key-value pairs.
   - ``error`` *(Error)*, *optional*: Error information if execution failed.
   - ``metadata`` *(map<string, string>)*, *optional*: Additional metadata about the execution.
   - ``is_final`` *(bool)*: Indicates if this is the final response message.

**Value Type**

Values in FlowKit use a flexible type system:

   - ``string_value`` *(string)*: String representation.
   - ``number_value`` *(double)*: Numeric value.
   - ``bool_value`` *(bool)*: Boolean value.
   - ``struct_value`` *(google.protobuf.Struct)*: Complex structured data.
   - ``list_value`` *(google.protobuf.ListValue)*: Array of values.

Here is an example successful response:

.. code-block:: json

   {
       "outputs": {
           "uuid": {
               "string_value": "550e8400-e29b-41d4-a716-446655440000"
           }
       },
       "is_final": true
   }

Streaming Responses
~~~~~~~~~~~~~~~~~~~

Functions that support streaming send multiple response messages:

   - Initial messages contain partial results with ``is_final: false``
   - Final message contains complete results with ``is_final: true``
   - Errors end the stream immediately