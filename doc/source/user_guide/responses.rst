.. _responses:

FlowKit Responses
=================

In this section the responses a client can expect to receive via gRPC after executing a function is explained in detail.

Response Format
~~~~~~~~~~~~~~~

The system returns ``FunctionOutputs`` for successful function execution:

**FunctionOutputs**

   - ``name`` *(string)*: The name of the executed function.
   - ``outputs`` *(array<FunctionOutput>)*: Array of output values from the function.

**FunctionOutput Type**

Each output value in the array contains:

   - ``Name`` *(string)*: The name of the output parameter.
   - ``Value`` *(string)*: The output value as a string.
   - ``GoType`` *(string)*: The Go type of the output value.

Here is an example successful response:

.. code-block:: json

   {
       "name": "GenerateUUID",
       "outputs": [
           {
               "Name": "uuid",
               "Value": "550e8400e29b41d4a716446655440000",
               "GoType": "string"
           }
       ]
   }

Streaming Responses
~~~~~~~~~~~~~~~~~~~

Functions that support streaming return ``StreamOutput`` messages:

**StreamOutput**

   - ``MessageCounter`` *(int)*: Counter for the message in the stream.
   - ``IsLast`` *(bool)*: Indicates if this is the final message.
   - ``Value`` *(string)*: The streamed output value.

Stream characteristics:
   - Initial messages contain partial results with ``IsLast: false``
   - Final message has ``IsLast: true``
   - Errors end the stream immediately
