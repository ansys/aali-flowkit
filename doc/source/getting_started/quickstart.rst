.. _quickstart:

Quickstart
==========


.. code-block:: bash

   git clone https://github.com/ansys/aali-flowkit.git
   cd aali-flowkit
   go build -o flowkit
   ./flowkit

Expected output::

   Starting Flowkit gRPC server on :50051
   Server ready to accept connections

Test connection
---------------

Install grpcurl to test the server:

.. code-block:: bash

   go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

List available services:

.. code-block:: bash

   grpcurl -plaintext localhost:50051 list

Expected output::

   aaliflowkitgrpc.ExternalFunctions
   grpc.reflection.v1alpha.ServerReflection

Call a function
---------------

List all functions:

.. code-block:: bash

   grpcurl -plaintext localhost:50051 aaliflowkitgrpc.ExternalFunctions/ListFunctions

Call the UUID generator:

.. code-block:: bash

   grpcurl -plaintext -d '{"name": "GenerateUUID", "inputs": []}' \
     localhost:50051 aaliflowkitgrpc.ExternalFunctions/RunFunction

Response:

.. code-block:: json

   {
     "outputs": ["f47ac10b-58cc-4372-a567-0e02b2c3d479"],
     "status": "success"
   }

Next steps
----------

See the :doc:`../user_guide/index` for function registration and usage.
