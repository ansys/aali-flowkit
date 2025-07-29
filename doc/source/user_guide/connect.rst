.. _connect:

Connect to FlowKit
==================

Once the app is running, it launches a gRPC server which is waiting for new connections on the address specified in the ``FLOWKIT_ADDRESS`` configuration parameter. The server implements the ``ExternalFunctions`` service, which provides methods to list, execute, and manage functions dynamically.

Example Connection
------------------

**Go Client**

.. code-block:: go

   import (
       "context"
       "log"
       "google.golang.org/grpc"
       "google.golang.org/grpc/credentials/insecure"
       "google.golang.org/grpc/metadata"
       pb "github.com/ansys/aali-sharedtypes/pkg/aaliflowkitgrpc"
   )

   // Connect to FlowKit
   conn, err := grpc.Dial("localhost:50051",
       grpc.WithTransportCredentials(insecure.NewCredentials()))
   if err != nil {
       log.Fatalf("Failed to connect: %v", err)
   }
   defer conn.Close()

   client := pb.NewExternalFunctionsClient(conn)

   // Add authentication header
   ctx := metadata.AppendToOutgoingContext(context.Background(), "x-api-key", "your-api-key")

   // List available functions
   resp, err := client.ListFunctions(ctx, &pb.ListFunctionsRequest{})
   if err != nil {
       log.Fatalf("Failed to list functions: %v", err)
   }

**Other Languages**

For Python, Java, C# and other languages, generate the client code from the proto file using the appropriate protoc compiler. The proto file is available in the `aali-sharedtypes <https://github.com/ansys/aali-sharedtypes>`_ repository.

.. note::
   The gRPC client code and protocol buffer definitions are provided by the
   ``aali-sharedtypes`` dependency. The proto definitions can be found at:

   - Repository: `github.com/ansys/aali-sharedtypes <https://github.com/ansys/aali-sharedtypes>`_
   - Go package: ``pkg/aaliflowkitgrpc``

   This package provides the generated client code for interacting with
   FlowKit's gRPC service, including the ``ExternalFunctionsClient`` and all
   message types. For other languages, you'll need to generate the client
   code from the proto file using the appropriate protoc compiler.

Authentication
--------------

Use the ``x-api-key`` header with the value from ``FLOWKIT_API_KEY`` configuration in all requests.

SSL/TLS Configuration
---------------------

For production environments, enable SSL/TLS encryption in your ``config.yaml``:

.. code-block:: yaml

   USE_SSL: true
   SSL_CERT_PUBLIC_KEY_FILE: "/path/to/server.crt"
   SSL_CERT_PRIVATE_KEY_FILE: "/path/to/server.key"

When SSL is enabled, use standard gRPC secure connection methods for your client language. For Go, use ``grpc.WithTransportCredentials()`` with appropriate credentials.

Troubleshooting
---------------

**Connection Refused**
   - Verify FlowKit is running (check logs for "gRPC server listening on address")
   - Check the ``FLOWKIT_ADDRESS`` configuration

**Authentication Failed**
   - Verify the API key matches ``FLOWKIT_API_KEY`` in configuration
   - Ensure the ``x-api-key`` header is included in requests

**SSL Certificate Errors**
   - Verify certificate paths in configuration
   - Ensure certificates are valid and not expired

**Debug Logging**
   - Set ``LOG_LEVEL: "debug"`` in configuration
   - Check logs in the file specified by ``LOCAL_LOGS_LOCATION`` (default: "logs.log")
   - Fatal errors are logged to ``ERROR_FILE_LOCATION`` (default: "error.log")
