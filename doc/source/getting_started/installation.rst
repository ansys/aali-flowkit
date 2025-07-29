Installation
============

This guide covers different methods to install and set up FlowKit.

Install from source
-------------------

To install FlowKit from source, clone the repository from GitHub:

.. code:: bash

   git clone https://github.com/ansys/aali-flowkit.git
   cd aali-flowkit

Install with Docker
-------------------

To run FlowKit using Docker:

.. code:: bash

   docker pull ghcr.io/ansys/aali-flowkit:latest
   docker run -p 50051:50051 ghcr.io/ansys/aali-flowkit:latest

For Docker Compose setup, see the Docker Compose documentation.

Configuration
-------------

FlowKit requires a configuration file to run. 

1. Copy the example configuration:

   .. code:: bash

      cp configs/config.yaml config.yaml

2. Edit ``config.yaml`` with your settings:

   **Required settings:**

   .. code:: yaml

      FLOWKIT_ADDRESS: "localhost:50051"  # Use 0.0.0.0:50051 for Docker
      FLOWKIT_API_KEY: "your-api-key"

   For full configuration options, see :doc:`configuration`.

Running FlowKit
---------------

From source
~~~~~~~~~~~

Start FlowKit by running:

.. code:: bash

   go run main.go

With Docker
~~~~~~~~~~~

.. code:: bash

   docker run -p 50051:50051 -v $(pwd)/config.yaml:/app/config.yaml ghcr.io/ansys/aali-flowkit:latest

Verify installation
-------------------

Successful startup shows:

.. code:: bash

   {"level":"info","msg":"Aali FlowKit started successfully; gRPC server listening on address 'localhost:50051.'"}

Troubleshooting
---------------

If FlowKit fails to start:

1. Check ``error.log`` in the root folder
2. Verify all required configuration values are set
3. Ensure the port 50051 is not in use
4. Check service endpoint connectivity

Next steps
----------

- Review the :doc:`configuration` guide for all available options
- See the :doc:`../user_guide/quickstart` for your first FlowKit function
- Explore :doc:`../user_guide/functions` for available functions

.. button-ref:: index
    :ref-type: doc
    :color: primary
    :shadow:
    :expand:

    Back to Getting started