.. _grpc:

gRPC Protocol
=============

AALI Flowkit uses gRPC (Google Remote Procedure Call) as its primary communication protocol. This section provides detailed information about the gRPC implementation, service definitions, and client integration.

Overview
--------

gRPC provides several advantages for Flowkit:

- **High Performance**: Binary protocol with efficient serialization
- **Type Safety**: Strong typing through Protocol Buffers
- **Streaming Support**: Bidirectional streaming for real-time communication
- **Language Agnostic**: Clients can be implemented in multiple languages
- **Built-in Authentication**: Support for various authentication mechanisms

Service architecture
--------------------

Flowkit gRPC services
~~~~~~~~~~~~~~~~~~~~~

Flowkit exposes the following gRPC services:

**FlowkitService**
   Main service for function execution and system interaction

**FunctionRegistryService**
   Service for function registration and discovery

**MemoryService**
   Service for memory management operations

**MonitoringService**
   Service for system monitoring and metrics

**HealthService**
   Service for health checks and status monitoring

Service endpoints
-----------------

Main Flowkit service
~~~~~~~~~~~~~~~~~~~~

.. code-block:: protobuf

   service FlowkitService {
       // Execute a registered function
       rpc ExecuteFunction(ExecuteFunctionRequest) returns (ExecuteFunctionResponse);

       // Execute function with streaming response
       rpc ExecuteFunctionStream(ExecuteFunctionRequest) returns (stream StreamResponse);

       // List available functions
       rpc ListFunctions(ListFunctionsRequest) returns (ListFunctionsResponse);

       // Get function details
       rpc GetFunctionInfo(GetFunctionInfoRequest) returns (GetFunctionInfoResponse);

       // Get system information
       rpc GetSystemInfo(GetSystemInfoRequest) returns (GetSystemInfoResponse);
   }

Function registry service
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: protobuf

   service FunctionRegistryService {
       // Register a new function
       rpc RegisterFunction(RegisterFunctionRequest) returns (RegisterFunctionResponse);

       // Unregister a function
       rpc UnregisterFunction(UnregisterFunctionRequest) returns (UnregisterFunctionResponse);

       // Update function metadata
       rpc UpdateFunction(UpdateFunctionRequest) returns (UpdateFunctionResponse);

       // Search functions by criteria
       rpc SearchFunctions(SearchFunctionsRequest) returns (SearchFunctionsResponse);
   }

Memory service
~~~~~~~~~~~~~~

.. code-block:: protobuf

   service MemoryService {
       // Store data in memory
       rpc Store(StoreRequest) returns (StoreResponse);

       // Retrieve data from memory
       rpc Retrieve(RetrieveRequest) returns (RetrieveResponse);

       // List memory stores
       rpc ListStores(ListStoresRequest) returns (ListStoresResponse);

       // Clear memory store
       rpc ClearStore(ClearStoreRequest) returns (ClearStoreResponse);
   }

Message types
-------------

Core message types
~~~~~~~~~~~~~~~~~~

**ExecuteFunctionRequest**

.. code-block:: protobuf

   message ExecuteFunctionRequest {
       string function_name = 1;
       map<string, google.protobuf.Any> parameters = 2;
       ExecutionContext context = 3;
       ExecutionOptions options = 4;
   }

**ExecuteFunctionResponse**

.. code-block:: protobuf

   message ExecuteFunctionResponse {
       bool success = 1;
       google.protobuf.Any result = 2;
       string error_message = 3;
       ExecutionMetadata metadata = 4;
   }

**FunctionDefinition**

.. code-block:: protobuf

   message FunctionDefinition {
       string name = 1;
       string description = 2;
       string category = 3;
       repeated ParameterDefinition parameters = 4;
       ReturnValueDefinition return_value = 5;
       FunctionMetadata metadata = 6;
   }

Authentication and security
---------------------------

API key authentication
~~~~~~~~~~~~~~~~~~~~~~

Flowkit supports API key authentication for gRPC calls:

.. code-block:: go

   import (
       "google.golang.org/grpc"
       "google.golang.org/grpc/metadata"
   )

   func createAuthenticatedConnection() (*grpc.ClientConn, error) {
       conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
       if err != nil {
           return nil, err
       }

       return conn, nil
   }

   func addAPIKeyToContext(ctx context.Context, apiKey string) context.Context {
       md := metadata.Pairs("authorization", "Bearer "+apiKey)
       return metadata.NewOutgoingContext(ctx, md)
   }

TLS configuration
~~~~~~~~~~~~~~~~~

For production deployments, enable TLS:

.. code-block:: go

   import (
       "crypto/tls"
       "google.golang.org/grpc"
       "google.golang.org/grpc/credentials"
   )

   func createSecureConnection() (*grpc.ClientConn, error) {
       config := &tls.Config{
           ServerName: "flowkit.example.com",
       }
       creds := credentials.NewTLS(config)

       conn, err := grpc.Dial("flowkit.example.com:50051",
           grpc.WithTransportCredentials(creds))
       if err != nil {
           return nil, err
       }

       return conn, nil
   }

Client Implementation
---------------------

Go client example
~~~~~~~~~~~~~~~~~

.. code-block:: go

   package main

   import (
       "context"
       "log"
       "time"

       "google.golang.org/grpc"
       pb "aali-flowkit/proto/flowkit"
   )

   func main() {
       // Establish connection
       conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
       if err != nil {
           log.Fatalf("Failed to connect: %v", err)
       }
       defer conn.Close()

       // Create client
       client := pb.NewFlowkitServiceClient(conn)

       // Execute function
       ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
       defer cancel()

       request := &pb.ExecuteFunctionRequest{
           FunctionName: "PerformGeneralRequest",
           Parameters: map[string]*any.Any{
               "input": mustMarshalAny("Hello, Flowkit!"),
               "isStream": mustMarshalAny(false),
           },
       }

       response, err := client.ExecuteFunction(ctx, request)
       if err != nil {
           log.Fatalf("Function execution failed: %v", err)
       }

       log.Printf("Result: %v", response.Result)
   }

Python client example
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: python

   import grpc
   from google.protobuf import any_pb2
   import flowkit_pb2
   import flowkit_pb2_grpc

   def main():
       # Create channel
       channel = grpc.insecure_channel('localhost:50051')
       stub = flowkit_pb2_grpc.FlowkitServiceStub(channel)

       # Prepare request
       request = flowkit_pb2.ExecuteFunctionRequest()
       request.function_name = "PerformGeneralRequest"

       # Add parameters
       input_param = any_pb2.Any()
       input_param.Pack(flowkit_pb2.StringValue(value="Hello, Flowkit!"))
       request.parameters["input"].CopyFrom(input_param)

       # Execute function
       try:
           response = stub.ExecuteFunction(request)
           print(f"Success: {response.success}")
           print(f"Result: {response.result}")
       except grpc.RpcError as e:
           print(f"Error: {e.code()} - {e.details()}")

   if __name__ == '__main__':
       main()

JavaScript/Node.js client example
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: javascript

   const grpc = require('@grpc/grpc-js');
   const protoLoader = require('@grpc/proto-loader');

   // Load protocol buffer
   const packageDefinition = protoLoader.loadSync('flowkit.proto', {
       keepCase: true,
       longs: String,
       enums: String,
       defaults: true,
       oneofs: true
   });

   const flowkit = grpc.loadPackageDefinition(packageDefinition).flowkit;

   function main() {
       // Create client
       const client = new flowkit.FlowkitService('localhost:50051',
           grpc.credentials.createInsecure());

       // Prepare request
       const request = {
           functionName: 'PerformGeneralRequest',
           parameters: {
               input: { value: 'Hello, Flowkit!' },
               isStream: { value: false }
           }
       };

       // Execute function
       client.executeFunction(request, (error, response) => {
           if (error) {
               console.error('Error:', error);
               return;
           }

           console.log('Success:', response.success);
           console.log('Result:', response.result);
       });
   }

   main();

Streaming support
-----------------

Server streaming
~~~~~~~~~~~~~~~~

For functions that return streaming responses:

.. code-block:: go

   func streamingExample(client pb.FlowkitServiceClient) {
       ctx := context.Background()

       request := &pb.ExecuteFunctionRequest{
           FunctionName: "StreamingFunction",
           Parameters: map[string]*any.Any{
               "input": mustMarshalAny("stream this"),
           },
       }

       stream, err := client.ExecuteFunctionStream(ctx, request)
       if err != nil {
           log.Fatalf("Streaming failed: %v", err)
       }

       for {
           response, err := stream.Recv()
           if err == io.EOF {
               break
           }
           if err != nil {
               log.Fatalf("Stream error: %v", err)
           }

           log.Printf("Streamed: %v", response.Data)
       }
   }

Client streaming
~~~~~~~~~~~~~~~~

For functions that accept streaming input:

.. code-block:: go

   func clientStreamingExample(client pb.FlowkitServiceClient) {
       ctx := context.Background()

       stream, err := client.ExecuteFunctionClientStream(ctx)
       if err != nil {
           log.Fatalf("Client streaming failed: %v", err)
       }

       // Send multiple requests
       for i := 0; i < 10; i++ {
           request := &pb.StreamRequest{
               Data: fmt.Sprintf("Message %d", i),
           }

           if err := stream.Send(request); err != nil {
               log.Fatalf("Send error: %v", err)
           }
       }

       response, err := stream.CloseAndRecv()
       if err != nil {
           log.Fatalf("Close error: %v", err)
       }

       log.Printf("Final result: %v", response)
   }

Error handling
--------------

gRPC status codes
~~~~~~~~~~~~~~~~~

Flowkit uses standard gRPC status codes:

- ``OK`` (0): Success
- ``CANCELLED`` (1): Operation cancelled
- ``UNKNOWN`` (2): Unknown error
- ``INVALID_ARGUMENT`` (3): Invalid parameters
- ``DEADLINE_EXCEEDED`` (4): Timeout
- ``NOT_FOUND`` (5): Function not found
- ``ALREADY_EXISTS`` (6): Function already registered
- ``PERMISSION_DENIED`` (7): Authentication failed
- ``RESOURCE_EXHAUSTED`` (8): Rate limit exceeded
- ``FAILED_PRECONDITION`` (9): Precondition failed
- ``ABORTED`` (10): Operation aborted
- ``OUT_OF_RANGE`` (11): Parameter out of range
- ``UNIMPLEMENTED`` (12): Function not implemented
- ``INTERNAL`` (13): Internal server error
- ``UNAVAILABLE`` (14): Service unavailable
- ``DATA_LOSS`` (15): Data corruption
- ``UNAUTHENTICATED`` (16): Missing authentication

Error handling example
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

   import (
       "google.golang.org/grpc/codes"
       "google.golang.org/grpc/status"
   )

   func handleError(err error) {
       if err != nil {
           st, ok := status.FromError(err)
           if ok {
               switch st.Code() {
               case codes.NotFound:
                   log.Printf("Function not found: %s", st.Message())
               case codes.InvalidArgument:
                   log.Printf("Invalid parameters: %s", st.Message())
               case codes.Unauthenticated:
                   log.Printf("Authentication required: %s", st.Message())
               default:
                   log.Printf("gRPC error [%s]: %s", st.Code(), st.Message())
               }
           } else {
               log.Printf("Non-gRPC error: %v", err)
           }
       }
   }

Performance optimization
------------------------

Connection pooling
~~~~~~~~~~~~~~~~~~

For high-throughput applications, use connection pooling:

.. code-block:: go

   type ConnectionPool struct {
       connections []*grpc.ClientConn
       current     int32
   }

   func NewConnectionPool(address string, size int) (*ConnectionPool, error) {
       pool := &ConnectionPool{
           connections: make([]*grpc.ClientConn, size),
       }

       for i := 0; i < size; i++ {
           conn, err := grpc.Dial(address, grpc.WithInsecure())
           if err != nil {
               return nil, err
           }
           pool.connections[i] = conn
       }

       return pool, nil
   }

   func (p *ConnectionPool) GetConnection() *grpc.ClientConn {
       idx := atomic.AddInt32(&p.current, 1) % int32(len(p.connections))
       return p.connections[idx]
   }

Request compression
~~~~~~~~~~~~~~~~~~~

Enable compression for large payloads:

.. code-block:: go

   import "google.golang.org/grpc/encoding/gzip"

   // Enable compression
   conn, err := grpc.Dial("localhost:50051",
       grpc.WithInsecure(),
       grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))

Timeouts and deadlines
~~~~~~~~~~~~~~~~~~~~~~

Set appropriate timeouts:

.. code-block:: go

   import "google.golang.org/grpc/codes"

   func executeWithTimeout(client pb.FlowkitServiceClient) {
       ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
       defer cancel()

       response, err := client.ExecuteFunction(ctx, request)
       if err != nil {
           st, _ := status.FromError(err)
           if st.Code() == codes.DeadlineExceeded {
               log.Println("Request timed out")
           }
           return
       }

       // Process response
   }

Best practices
--------------

1. **Use connection pooling**: For high-throughput applications
2. **Set appropriate timeouts**: Prevent hanging requests
3. **Handle errors gracefully**: Implement proper error handling
4. **Enable compression**: For large message payloads
5. **Use TLS in production**: Secure communication channels
6. **Implement retries**: With exponential backoff for transient errors
7. **Monitor performance**: Track latency and error rates
