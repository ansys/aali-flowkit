Functions
=========

FlowKit provides dynamic function execution through a gRPC interface. Functions are organized by category and can be discovered at runtime.

Function Discovery
------------------

Use ``ListFunctions`` to discover available functions dynamically:

.. code-block:: go

   import (
       "context"
       "fmt"
       "google.golang.org/grpc"
       "google.golang.org/grpc/credentials/insecure"
       "google.golang.org/grpc/metadata"
       pb "github.com/ansys/aali-sharedtypes/pkg/aaliflowkitgrpc"
   )

   // Connect to FlowKit
   conn, _ := grpc.Dial("localhost:50051",
       grpc.WithTransportCredentials(insecure.NewCredentials()))
   client := pb.NewExternalFunctionsClient(conn)

   // Add authentication
   ctx := metadata.AppendToOutgoingContext(context.Background(), "x-api-key", "your-api-key")

   // List all available functions
   response, _ := client.ListFunctions(ctx, &pb.ListFunctionsRequest{})

   // Print function names
   for name, definition := range response.Functions {
       fmt.Printf("Function: %s - %s\n", name, definition.Description)
   }

This approach prevents hardcoded function lists and ensures you always have access to the latest available functions.

Function Categories
-------------------

FlowKit organizes functions by their purpose:

**llm handler**
   Execute language model requests and stream responses. Includes functions for embeddings, general chat, and code generation.

**knowledge db**
   Interact with vector and graph databases for similarity search, data retrieval, and dependency tracking.

**ansys gpt**
   Specialized functions for ANSYS GPT integration including query processing, semantic search, and context building.

**data extraction**
   Process and extract data from various sources including GitHub repositories, local files, and documents.

**generic**
   Provide common operations like UUID generation, string manipulation, JSON processing, and REST API calls.

**code generation**
   Support code generation workflows with element loading, example management, and user guide integration.

**ansys mesh pilot**
   Enable mesh-related operations including path descriptions, action synthesis, and problem-solving workflows.

**qdrant**
   Direct integration with Qdrant vector database for collection management and data insertion.

**auth**
   Handle authentication flows, API key validation, token management, and user access control.

**MCP**
   Enable Model Context Protocol features for dynamic tool discovery and AI agent integration.

**materials**
   Process material-related requests with attribute extraction and response serialization.

**rhsc**
   Support Red Hat Service Catalog operations with request body generation.

Each category contains multiple specialized functions designed for specific workflows within AALI.

Calling Functions
-----------------

Execute functions using ``RunFunction`` with the function name and inputs:

.. code-block:: go

   // Execute a simple function
   response, err := client.RunFunction(ctx, &pb.FunctionInputs{
       Name: "GenerateUUID",
       Inputs: []*pb.FunctionInput{},
   })
   if err != nil {
       log.Fatal(err)
   }

   // Print the UUID (without hyphens)
   for _, output := range response.Outputs {
       fmt.Printf("UUID: %s\n", output.Value)
       // Output: 550e8400e29b41d4a716446655440000
   }

For functions requiring parameters:

.. code-block:: go

   // Execute function with inputs
   response, err := client.RunFunction(ctx, &pb.FunctionInputs{
       Name: "StringConcat",
       Inputs: []*pb.FunctionInput{
           {Name: "a", Value: "Hello"},
           {Name: "b", Value: "World"},
           {Name: "separator", Value: " "},
       },
   })

   // Output: "Hello World"

**Streaming Functions**

For real-time responses, use ``StreamFunction``:

.. code-block:: go

   // Stream function for real-time responses
   stream, err := client.StreamFunction(ctx, &pb.FunctionInputs{
       Name: "PerformGeneralRequest",
       Inputs: []*pb.FunctionInput{
           {Name: "prompt", Value: "Explain quantum computing"},
           {Name: "stream", Value: "true"},
       },
   })

   // Handle streaming responses
   for {
       response, err := stream.Recv()
       if err == io.EOF {
           break
       }
       // Process each response chunk
       fmt.Print(response.Outputs[0].Value)
   }

This is particularly useful for LLM responses and long-running operations.

What's Next?
------------

- **Get started immediately** → :doc:`quickstart`
- **Learn about FlowKit integration** → :doc:`integration`
- **Detailed client examples** → :doc:`connect`
