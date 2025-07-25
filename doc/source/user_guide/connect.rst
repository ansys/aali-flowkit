.. _connect:

Interact with the Application
=============================

Once the application is running, it launches a gRPC server which is waiting for new connections on the port specified in the config. The server implements the ``ExternalFunctions`` service defined in the proto file.

Connection Examples
-------------------

**Go Client Example**

.. code-block:: go

   import (
       "context"
       "google.golang.org/grpc"
       "google.golang.org/grpc/credentials/insecure"
       "google.golang.org/grpc/metadata"
       pb "github.com/ansys/aali-sharedtypes/pkg/aaliflowkitgrpc"
   )

   // Create connection options
   var opts []grpc.DialOption

   // For non-SSL connections (development)
   opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

   // For SSL connections (production)
   // creds, _ := credentials.NewClientTLSFromFile("path/to/ca-cert.pem", "")
   // opts = append(opts, grpc.WithTransportCredentials(creds))

   // Connect to FlowKit
   conn, err := grpc.Dial("localhost:50051", opts...)
   if err != nil {
       log.Fatalf("Failed to connect: %v", err)
   }
   defer conn.Close()

   client := pb.NewExternalFunctionsClient(conn)

   // Add authentication header
   ctx := metadata.AppendToOutgoingContext(context.Background(), "x-api-key", "your-api-key")

   // List available functions
   resp, err := client.ListFunctions(ctx, &pb.Empty{})

.. note::
   The gRPC client code and protocol buffer definitions are provided by the
   ``aali-sharedtypes`` dependency. The proto definitions can be found at:

   - Repository: ``github.com/ansys/aali-sharedtypes``
   - Package: ``pkg/aaliflowkitgrpc``

   This package provides the generated Go client code for interacting with
   FlowKit's gRPC service, including the ``ExternalFunctionsClient`` and all
   message types.

Authentication
--------------

Authentication uses API keys passed in gRPC metadata with the header ``x-api-key``. Include this header in all requests:

- **Header name**: ``x-api-key``
- **Header value**: Your FlowKit API key (from ``FLOWKIT_API_KEY`` configuration)

The server validates the API key on each request using an authentication interceptor.
