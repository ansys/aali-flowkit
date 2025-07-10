.. _configuration:

Configuration
=============

Set up Flowkit for your environment.

Configuration file
------------------

Flowkit uses ``configs/config.yaml`` for settings. Key options:

**Server settings:**

.. code-block:: yaml

   # Where Flowkit listens for connections
   FLOWKIT_ADDRESS: "0.0.0.0:50051"

   # API key for security
   FLOWKIT_API_KEY: "flowkit-api-key"

**Logging:**

.. code-block:: yaml

   # Log detail level
   LOG_LEVEL: "info"
   ERROR_FILE_LOCATION: "error.log"

   # Save logs to file
   LOCAL_LOGS: true
   LOCAL_LOGS_LOCATION: "logs.log"

   # Send logs to Datadog (optional)
   DATADOG_LOGS: false
   LOGGING_URL: "https://http-intake.logs.datadoghq.eu/api/v2/logs"
   LOGGING_API_KEY: ""
   STAGE: "DEV"
   VERSION: "1.0.0"
   SERVICE_NAME: "aali"

**SSL Security:**

.. code-block:: yaml

   USE_SSL: false
   SSL_CERT_PUBLIC_KEY_FILE: ""
   SSL_CERT_PRIVATE_KEY_FILE: ""

**Azure Key Vault (optional):**

.. code-block:: yaml

   EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT: false
   AZURE_KEY_VAULT_NAME: ""
   AZURE_MANAGED_IDENTITY_ID: ""

**External services:**

.. code-block:: yaml

   # Connect to other AALI services
   LLM_HANDLER_ENDPOINT: "ws://aali-llm:9003"
   GRAPHDB_ADDRESS: "aali-graphdb:8080"
   QDRANT_HOST: "qdrant"
   QDRANT_PORT: 6334

Environment variables
---------------------

Override any setting with environment variables:

.. code-block:: bash

   export FLOWKIT_ADDRESS=localhost:50051
   export FLOWKIT_API_KEY=my-secret-key
   export LOG_LEVEL=debug

Next steps
----------

- Learn how to :doc:`../user_guide/index`
- Explore :doc:`../advanced/index` for production setup
