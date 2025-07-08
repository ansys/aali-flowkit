.. _calling_functions:

Calling registered functions
============================

How to invoke registered functions in Flowkit using the gRPC API.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Function Invocation Methods
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Registered functions can be called using the gRPC API. Three primary methods are provided:

      * `ListFunctions`: Lists all available functions
      * `RunFunction`: Executes a function and returns a single response
      * `StreamFunction`: Executes a function and returns a stream of responses

      Example usage with `grpcurl`:

      .. code-block:: bash

         # List all available functions
         grpcurl -plaintext localhost:50051 proto.ExternalFunctions/ListFunctions

         # Call a function
         grpcurl -plaintext -d '{
           "name": "AssignStringToString",
           "inputs": [{"name": "inputString", "value": "Hello Flowkit"}]
         }' \
         localhost:50051 proto.ExternalFunctions/RunFunction

   .. grid-item-card:: Request and Response Format
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      **RunFunction Request (FunctionInputs):**

      * `name`: Name of the registered function (string)
      * `inputs`: Array of FunctionInput objects with name, value, and type

      **RunFunction Response (FunctionOutputs):**

      * `name`: Name of the executed function
      * `outputs`: Array of FunctionOutput objects with name, value, and type

      **ListFunctions Response (ListFunctionsResponse):**

      * `functions`: Map of function names to their definitions

      If you call a function that is not registered, you will receive a gRPC error response.
