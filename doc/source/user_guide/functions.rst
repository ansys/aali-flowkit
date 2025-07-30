Functions Overview
==================

FlowKit is an external function repository for the AALI agent platform. It provides ready-to-use functions that can be dynamically discovered and executed through a gRPC interface.

.. warning::
   FlowKit is **not meant to be used standalone**. It should be used through the AALI API client. The direct usage examples provided in this documentation are for testing and understanding purposes only.

.. note::
   For installation instructions, see :doc:`../getting_started/installation`.

Functions are organized into categories, each serving a specific purpose in building workflows. Categories include various utility functions, integrations, and domain-specific operations.

.. important::
   Use ``ListFunctions`` to discover all available functions and their categories. The exact categories and their contents may vary based on your FlowKit configuration.

Discovering Functions
---------------------

Always use ``ListFunctions`` to discover available functions dynamically:

**Python Example:**

.. code-block:: python

   import grpc
   import flowkit_pb2
   import flowkit_pb2_grpc

   # Connect to FlowKit
   channel = grpc.insecure_channel('localhost:50051')
   stub = flowkit_pb2_grpc.FlowKitStub(channel)

   # List all functions
   response = stub.ListFunctions(flowkit_pb2.ListFunctionsRequest())

   # Browse by category
   for name, definition in response.functions.items():
       print(f"{definition.category}: {name}")

**Go Example:**

.. code-block:: go

   import pb "github.com/ansys/aali-sharedtypes/pkg/aaliflowkitgrpc"

   // List all functions
   response, err := client.ListFunctions(ctx, &pb.ListFunctionsRequest{})

   // Browse functions
   for name, def := range response.Functions {
       fmt.Printf("%s: %s\n", def.Category, name)
   }

Basic Usage Example
-------------------

Here's how to call a simple function:

.. code-block:: python

   # Call GenerateUUID function
   response = stub.RunFunction(flowkit_pb2.RunFunctionRequest(
       function_name="GenerateUUID",
       inputs={}
   ))

   # Get the result
   uuid = response.outputs["result"]
   print(f"Generated UUID: {uuid}")

.. note::
   All function inputs and outputs are strings. Complex data structures must be JSON-encoded.

Next Steps
----------

- Add your own functions → :doc:`adding_functions`
- Explore the API Reference for detailed function signatures
