.. _installation:

Installation
============

Build and run the Flowkit gRPC server.

Requirements
------------

- **Go 1.24.2** or later
- **Git** for cloning the repository

Install from source
-------------------

Clone and build the server:

.. code-block:: bash

   # Clone the repository
   git clone https://github.com/ansys/aali-flowkit.git
   cd aali-flowkit

   # Download dependencies
   go mod tidy

   # Build the executable
   go build -o flowkit main.go

Start the service
-----------------

Run the server:

.. code-block:: bash

   ./flowkit

Output:

.. code-block:: text

   Aali FlowKit started successfully; gRPC server listening on address '0.0.0.0:50051'...

Next: :doc:`quickstart`
