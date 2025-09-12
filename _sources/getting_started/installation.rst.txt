Installation
============

Get FlowKit up and running in minutes with these simple installation methods.

Quick Start with Docker (Recommended)
-------------------------------------

The fastest way to get started with FlowKit:

.. code:: bash

   docker pull ghcr.io/ansys/aali-flowkit:latest
   docker run -p 50051:50051 ghcr.io/ansys/aali-flowkit:latest

Install from Source
-------------------

For local development or customization:

.. code:: bash

   # Clone the repository
   git clone https://github.com/ansys/aali-flowkit.git
   cd aali-flowkit

   # Run FlowKit
   go run main.go

.. note::
   Make sure you have Go 1.21+ installed.

Basic Configuration
-------------------

FlowKit works out of the box with default settings. For custom configuration:

1. Copy the example configuration:

   .. tab-set::

      .. tab-item:: Linux/Mac

         .. code:: bash

            cp configs/config.yaml config.yaml

      .. tab-item:: Windows

         .. code:: batch

            copy configs\config.yaml config.yaml

2. Modify only what you need. Most users only need to set:

   .. code:: yaml

      FLOWKIT_API_KEY: "your-api-key"  # Optional, for authentication

For advanced configuration options, see :doc:`configuration`.

Verify Installation
-------------------

FlowKit is running successfully when you see:

.. code:: text

   {"level":"info","msg":"Aali FlowKit started successfully; gRPC server listening on address 'localhost:50051.'"}

Quick Test
----------

Test FlowKit is working by listing available functions using grpcurl:

.. tab-set::

   .. tab-item:: Mac

      .. code:: bash

         # Install grpcurl
         brew install grpcurl

         # List available functions
         grpcurl -plaintext localhost:50051 list

   .. tab-item:: Linux

      .. code:: bash

         # Install grpcurl
         curl -sSL https://github.com/fullstorydev/grpcurl/releases/download/v1.8.7/grpcurl_1.8.7_linux_x86_64.tar.gz | tar xz
         sudo mv grpcurl /usr/local/bin/

         # List available functions
         grpcurl -plaintext localhost:50051 list

   .. tab-item:: Windows

      .. code:: batch

         :: Download grpcurl from GitHub releases:
         :: https://github.com/fullstorydev/grpcurl/releases
         :: Or use WSL with Linux instructions

         :: List available functions
         grpcurl -plaintext localhost:50051 list

Troubleshooting
---------------

If FlowKit fails to start:

1. Check ``error.log`` in the root folder
2. Verify all required configuration values are set
3. Ensure the port 50051 is not in use
4. Check service endpoint connectivity

Next Steps
----------

- Explore available functions in the :doc:`../user_guide/index`
- Learn how to :doc:`../user_guide/adding_functions`
- Check the :doc:`../api_reference/externalfunctions/index` for all available functions

.. button-ref:: index
    :ref-type: doc
    :color: primary
    :shadow:
    :expand:

    Back to Getting started
