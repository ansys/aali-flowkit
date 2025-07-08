.. _agent_integration:

Agent integration
=================

How the AALI agent communicates with Flowkit over gRPC and how to configure the integration.

.. grid:: 1
   :gutter: 2

   .. grid-item-card:: Communication Overview
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      The AALI agent connects to Flowkit via gRPC to execute workflow steps. For each step, the agent can invoke one or more registered functions.

      **Call sequence:**

      1. agent selects a function to invoke
      2. Sends a gRPC request with input arguments
      3. Flowkit executes the function and returns a response
      4. agent processes the result and continues the workflow

      Both synchronous and streaming (multi-response) functions are supported.

   .. grid-item-card:: Available Methods
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      All available methods are defined in `pkg/externalfunctions/externalfunctions.go`.

      The gRPC service interface is defined in the shared proto files:

      .. code-block:: text

         service ExternalFunctionService {
             rpc CallFunction (FunctionCallRequest) returns (FunctionCallResponse);
         }

   .. grid-item-card:: Agent configuration
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      To connect the agent to a running Flowkit instance, set the Flowkit gRPC endpoint in the agent's configuration.

      Example config (YAML):

      .. code-block:: yaml

         flowkit:
           address: "localhost:50051"

      Or set the environment variable:

      .. code-block:: bash

         export FLOWKIT_GRPC_ADDR=localhost:50051

   .. grid-item-card:: Direct gRPC Access
      :class-card: sd-shadow-sm sd-rounded-md
      :text-align: left

      Any client can interact with Flowkit using its gRPC API (not just the agent):

      * Use `grpcurl`, custom scripts, or test clients for debugging and development
      * Example (list available services):

        .. code-block:: bash

           grpcurl -plaintext localhost:50051 list

      For troubleshooting connection errors, check that:
      * The server is running and reachable on the configured port
      * The proto file matches the running Flowkit server version
