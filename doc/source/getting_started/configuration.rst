.. _configuration:

Configuration
=============

FlowKit works out of the box with zero configuration. For most use cases, you only need to set a few parameters.

Quick Start
-----------

**Minimal Configuration**

FlowKit only needs one setting to run:

.. code-block:: yaml

   # config.yaml
   FLOWKIT_ADDRESS: "localhost:50051"

FlowKit starts with sensible defaults for everything else.

**With External Services**

If you're using functions that need external services, add their endpoints:

.. code-block:: yaml

   # config.yaml
   FLOWKIT_ADDRESS: "localhost:50051"
   FLOWKIT_API_KEY: "your-secure-key"  # Optional authentication

   # Add only what you need:
   LLM_HANDLER_ENDPOINT: "ws://aali-llm:9003"      # For LLM functions
   GRAPHDB_ADDRESS: "http://aali-graphdb:8080"     # For knowledge DB functions
   QDRANT_HOST: "qdrant"                           # For vector search functions
   QDRANT_PORT: 6334

Configuration Examples
----------------------

**Local Development**

.. code-block:: yaml

   # config.yaml for local development
   FLOWKIT_ADDRESS: "localhost:50051"
   LOG_LEVEL: "debug"  # Optional: see more logs during development

**Docker Deployment**

.. code-block:: yaml

   # config.yaml for Docker
   FLOWKIT_ADDRESS: "0.0.0.0:50051"  # Listen on all interfaces
   FLOWKIT_API_KEY: "production-key"  # Secure your endpoint

   # Service names from docker-compose:
   LLM_HANDLER_ENDPOINT: "ws://aali-llm:9003"
   GRAPHDB_ADDRESS: "http://aali-graphdb:8080"
   QDRANT_HOST: "qdrant"


.. important::
   Some configuration options expect the **names of environment variables** that contain the actual values, not the values directly. For example:

   .. code-block:: bash

      # First, set the actual value in an environment variable
      export MY_API_KEY="actual-api-key-value"

      # Then reference that variable name in the config
      export FLOWKIT_API_KEY="MY_API_KEY"

Where to Place config.yaml
--------------------------

FlowKit looks for ``config.yaml`` in these locations (in order):

1. Project root directory
2. ``configs/`` directory
3. Path specified by ``AALI_CONFIG_PATH`` environment variable

.. code-block:: bash

   # Example: Use custom config location
   export AALI_CONFIG_PATH="/path/to/your/config.yaml"
   go run main.go

Advanced Configuration
----------------------

For specialized use cases, FlowKit supports additional configuration options.

**Logging Configuration**

.. code-block:: yaml

   LOG_LEVEL: "info"              # debug, info, warning, error, fatal
   LOCAL_LOGS: true               # Write logs to file
   LOCAL_LOGS_LOCATION: "app.log" # Log file path
   ERROR_FILE_LOCATION: "errors.log"

**SSL/TLS Configuration**

.. code-block:: yaml

   USE_SSL: true
   SSL_CERT_PUBLIC_KEY_FILE: "/path/to/cert.pem"
   SSL_CERT_PRIVATE_KEY_FILE: "/path/to/key.pem"

**Azure Key Vault Integration**

For enterprise deployments, configuration can be loaded from Azure Key Vault:

.. code-block:: yaml

   EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT: true
   AZURE_KEY_VAULT_NAME: "VAULT_NAME_ENV_VAR"
   AZURE_MANAGED_IDENTITY_ID: "IDENTITY_ENV_VAR"

.. note::
   The Azure Key Vault settings follow the same pattern as other environment variables - they expect the **names** of environment variables that contain the actual values:

   .. code-block:: bash

      # Set the actual values in environment variables
      export MY_VAULT_NAME="my-actual-keyvault"
      export MY_IDENTITY="00000000-0000-0000-0000-000000000000"

      # Then reference those variable names in the config
      export AZURE_KEY_VAULT_NAME="MY_VAULT_NAME"
      export AZURE_MANAGED_IDENTITY_ID="MY_IDENTITY"

**All Configuration Options**

For a complete list of all available configuration options, see the `Full Configuration Reference`_ below.

.. note::
   Most users don't need to configure anything beyond ``FLOWKIT_ADDRESS`` and the endpoints for services they actually use. FlowKit provides sensible defaults for everything else.

Tips
----

- Start with minimal configuration and add settings only as needed
- Use environment variables for sensitive values like API keys
- In production, always set ``FLOWKIT_API_KEY`` to secure your endpoint
- Service endpoints are only needed if you use those specific functions

Full Configuration Reference
----------------------------

**General Settings**

- ``FLOWKIT_ADDRESS`` (required): Where FlowKit listens (for example, ``localhost:50051``)
- ``FLOWKIT_API_KEY``: API key for authentication
- ``STAGE``: Environment stage (default: ``DEV``)
- ``VERSION``: Service version (default: ``1.0.0``)
- ``SERVICE_NAME``: Service name for logging (default: ``aali``)

**Service Endpoints**

- ``LLM_HANDLER_ENDPOINT``: AALI LLM service (default: ``ws://aali-llm:9003``)
- ``GRAPHDB_ADDRESS``: Graph database (default: ``http://aali-graphdb:8080``)
- ``QDRANT_HOST``: Qdrant hostname (default: ``qdrant``)
- ``QDRANT_PORT``: Qdrant port (default: ``6334``)

**Logging Settings**

- ``LOG_LEVEL``: Logging level (default: ``info``)
- ``LOCAL_LOGS``: Enable file logging (default: ``true``)
- ``LOCAL_LOGS_LOCATION``: Log path (default: ``logs.log``)
- ``ERROR_FILE_LOCATION``: Error log path (default: ``error.log``)
- ``DATADOG_LOGS``: Enable Datadog logging (default: ``false``)
- ``LOGGING_URL``: Datadog logs endpoint
- ``LOGGING_API_KEY``: Datadog API key
- ``DATADOG_SOURCE``: Datadog source tag
- ``DATADOG_METRICS``: Enable Datadog metrics
- ``METRICS_URL``: Datadog metrics endpoint

**SSL/TLS Settings**

- ``USE_SSL``: Enable SSL/TLS (default: ``false``)
- ``SSL_CERT_PUBLIC_KEY_FILE``: Path to certificate
- ``SSL_CERT_PRIVATE_KEY_FILE``: Path to private key

**Azure Key Vault Settings**

- ``EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT``: Use Azure Key Vault (default: ``false``)
- ``AZURE_KEY_VAULT_NAME``: Key Vault name
- ``AZURE_MANAGED_IDENTITY_ID``: Managed Identity for auth

**Legacy Settings**

- ``EXTERNALFUNCTIONS_GRPC_PORT``: Deprecated, use port in ``FLOWKIT_ADDRESS``
