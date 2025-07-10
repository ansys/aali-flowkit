.. _quickstart:

Quick Start
===========

.. code-block:: bash

   # Quick build and run
   git clone https://github.com/ansys/aali-flowkit.git
   cd aali-flowkit
   go build -o flowkit
   ./flowkit

The server starts on port **50051**.

Test the connection
-------------------

Install **grpcurl** to interact with Flowkit:

.. code-block:: bash

   go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

List available services:

.. code-block:: bash

   grpcurl -plaintext localhost:50051 list

You should see:

.. code-block:: text

   aaliflowkitgrpc.ExternalFunctions
   grpc.reflection.v1alpha.ServerReflection

Your first function call
------------------------

Generate a UUID using a built-in function:

.. code-block:: bash

   grpcurl -plaintext -d '{"name": "GenerateUUID", "inputs": []}' \
     localhost:50051 aaliflowkitgrpc.ExternalFunctions/RunFunction

Response:

.. code-block:: json

   {
     "outputs": [{
       "name": "uuid",
       "value": "\"f47ac10b-58cc-4372-a567-0e02b2c3d479\""
     }]
   }

What's happening?
-----------------

1. **Flowkit** exposes Go functions through gRPC
2. **AI workflows** can call these functions remotely
3. **Functions** process data and return results
4. **No Go knowledge needed** to use the functions

Next: :doc:`configuration`
