.. _configuration:

Configuration
=============

Flowkit reads configuration from ``config.yaml`` or environment variables.

Basic configuration
-------------------

.. code-block:: yaml

   # config.yaml
   LOG_LEVEL: "info"
   FLOWKIT_ADDRESS: "0.0.0.0:50051"
   FLOWKIT_API_KEY: "your-api-key"

Environment variables
---------------------

.. code-block:: bash

   export LOG_LEVEL=info
   export FLOWKIT_ADDRESS=0.0.0.0:50051
   export FLOWKIT_API_KEY=your-api-key

Common options
--------------

* **LOG_LEVEL**: debug, info, warning, error
* **FLOWKIT_ADDRESS**: Server bind address (default: 0.0.0.0:50051)
* **FLOWKIT_API_KEY**: API key for authentication
* **USE_SSL**: true/false for SSL encryption
