.. _configuration:

Configuration
=============

This documentation provides a comprehensive guide to the configuration settings for **FlowKit** service.

FlowKit uses a YAML configuration file named ``config.yaml``. By default, FlowKit looks for this file in the project root directory, but it can also be placed in the ``configs/`` directory. All parameters are read at startup.

General settings
----------------

.. list-table::
   :header-rows: 1
   :widths: 30 15 55 15 15

   * - Argument
     - Type
     - Description
     - Required
     - Default

   * - STAGE
     - string
     - Specifies the environment stage.
     - ``False``
     - ``DEV``

   * - VERSION
     - string
     - Specifies the version of the service.
     - ``False``
     - ``1.0.0``

   * - SERVICE_NAME
     - string
     - Defines the name of the service.
     - ``False``
     - ``aali``

   * - FLOWKIT_ADDRESS
     - string
     - Address where FlowKit listens for incoming gRPC requests.
     - ``True``
     - ``''``

   * - FLOWKIT_API_KEY
     - string
     - API key used to authenticate with FlowKit.
     - ``True``
     - ``''``

   * - USE_SSL
     - bool
     - Whether to use SSL for securing the endpoints.
     - ``False``
     - ``false``

   * - SSL_CERT_PUBLIC_KEY_FILE
     - string
     - Path to the public key file for SSL.
     - ``False``
     - ``''``

   * - SSL_CERT_PRIVATE_KEY_FILE
     - string
     - Path to the private key file for SSL.
     - ``False``
     - ``''``

.. note::
   **Legacy parameter**: ``EXTERNALFUNCTIONS_GRPC_PORT`` - This parameter was used in older versions to specify a separate port for the gRPC service. It is now deprecated, and the port should be included in the ``FLOWKIT_ADDRESS`` parameter (for example, ``localhost:50051``). If both are specified, FlowKit handles the legacy port definition for backward compatibility.

Service endpoints
-----------------

.. list-table::
   :header-rows: 1
   :widths: 30 15 55 15

   * - Argument
     - Type
     - Description
     - Default

   * - LLM_HANDLER_ENDPOINT
     - string
     - Endpoint where FlowKit connects to aali-llm.
     - ``ws://aali-llm:9003``

   * - GRAPHDB_ADDRESS
     - string
     - Address of the aali-graphdb service.
     - ``http://aali-graphdb:8080``

   * - QDRANT_HOST
     - string
     - Hostname of the Qdrant vector database.
     - ``qdrant``

   * - QDRANT_PORT
     - int
     - Port of the Qdrant vector database.
     - ``6334``

Workflow configuration
----------------------

.. list-table::
   :header-rows: 1
   :widths: 30 15 55 15

   * - Argument
     - Type
     - Description
     - Default

   * - WORKFLOW_CONFIG_VARIABLES
     - map
     - **[OPTIONAL]** Key-value pairs for workflow-specific configuration. Used primarily by ansysmeshpilot functions for tool names, collection names, database queries, and prompt templates. This section is commented out in the default config.
     - ``{}`` (commented out by default)

Logging settings
----------------

.. list-table::
   :header-rows: 1
   :widths: 30 15 55 15

   * - Argument
     - Type
     - Description
     - Default

   * - LOG_LEVEL
     - string
     - Specifies the logging level. Valid values: ``debug``, ``info``, ``warning``, ``error``, ``fatal``
     - ``info``

   * - ERROR_FILE_LOCATION
     - string
     - Location where fatal errors are logged.
     - ``error.log``

   * - LOCAL_LOGS
     - bool
     - If true, a local log file is created.
     - ``true``

   * - LOCAL_LOGS_LOCATION
     - string
     - Location of the local log file.
     - ``logs.log``

   * - DATADOG_LOGS
     - bool
     - If true, logs are sent to Datadog.
     - ``false``

   * - LOGGING_URL
     - string
     - Datadog URL where logs are sent.
     - ``https://http-intake.logs.datadoghq.eu/api/v2/logs``

   * - LOGGING_API_KEY
     - string
     - Datadog API key for authentication.
     - ``''``

   * - DATADOG_SOURCE
     - string
     - Datadog source identifier.
     - ``nginx``

   * - DATADOG_METRICS
     - bool
     - If true, metrics are sent to Datadog.
     - ``false``

   * - METRICS_URL
     - string
     - Datadog URL where metrics are sent.
     - ``''``

Function-specific configuration
-------------------------------

Some FlowKit functions require additional parameters that are not part of the core FlowKit configuration. These are typically set as environment variables or passed directly to functions:

.. list-table::
   :header-rows: 1
   :widths: 30 15 55

   * - Parameter
     - Used By
     - Description

   * - ANSYS_AUTHORIZATION_URL
     - ansysmeshpilot functions
     - Authorization URL for Ansys services

   * - LLM_API_KEY
     - llmhandler functions
     - API key for LLM service authentication

   * - FLOWKIT_PYTHON_ENDPOINT
     - specific functions
     - Endpoint for Python service integration

   * - FLOWKIT_PYTHON_API_KEY
     - specific functions
     - API key for Python service authentication

.. note::
   These parameters are **not** part of the main ``config.yaml`` file. They are used by specific functions and should be configured as environment variables or function parameters as needed.

Azure Key Vault settings
------------------------

.. list-table::
   :header-rows: 1
   :widths: 30 15 55 15

   * - Argument
     - Type
     - Description
     - Default

   * - EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT
     - bool
     - If true, configuration is extracted from Azure Key Vault.
     - ``false``

   * - AZURE_KEY_VAULT_NAME
     - string
     - Name of the Azure Key Vault.
     - ``''``

   * - AZURE_MANAGED_IDENTITY_ID
     - string
     - Azure Managed Identity ID for authentication.
     - ``''``

Configuration examples
----------------------

Create a ``config.yaml`` file with your settings. You can place it either in the project root directory or in the ``configs/`` directory.

**Local development configuration**

.. code-block:: yaml

   # General settings
   STAGE: "DEV"
   VERSION: "1.0.0"
   SERVICE_NAME: "aali"

   # FlowKit settings
   FLOWKIT_ADDRESS: "localhost:50051"
   FLOWKIT_API_KEY: "dev-api-key"

   # Service endpoints
   LLM_HANDLER_ENDPOINT: "ws://localhost:9003"
   GRAPHDB_ADDRESS: "http://localhost:8080"
   QDRANT_HOST: "localhost"
   QDRANT_PORT: 6334

   # Logging settings
   LOG_LEVEL: "debug"
   ERROR_FILE_LOCATION: "error.log"
   LOCAL_LOGS: true
   LOCAL_LOGS_LOCATION: "logs.log"
   DATADOG_LOGS: false

**Docker configuration**

.. code-block:: yaml

   # General settings
   STAGE: "PROD"
   VERSION: "1.0.0"
   SERVICE_NAME: "aali-flowkit"

   # FlowKit settings
   FLOWKIT_ADDRESS: "0.0.0.0:50051"
   FLOWKIT_API_KEY: "your-secure-api-key"

   # Service endpoints
   LLM_HANDLER_ENDPOINT: "ws://aali-llm:9003"
   GRAPHDB_ADDRESS: "http://aali-graphdb:8080"
   QDRANT_HOST: "qdrant"
   QDRANT_PORT: 6334

   # SSL settings
   USE_SSL: true
   SSL_CERT_PUBLIC_KEY_FILE: "/certs/flowkit.crt"
   SSL_CERT_PRIVATE_KEY_FILE: "/certs/flowkit.key"

   # Logging settings
   LOG_LEVEL: "info"
   ERROR_FILE_LOCATION: "error.log"
   LOCAL_LOGS: false
   DATADOG_LOGS: false

   # WORKFLOW_CONFIG_VARIABLES:
   #   MESHPILOT_DB_ENDPOINT: "http://meshpilot-db:8080"
   #   APP_TOOL_1_NAME: "MeshGenerator"
   #   COLLECTION_1_NAME: "mesh_collection"

   # Note: Function-specific parameters like ANSYS_AUTHORIZATION_URL,
   # LLM_API_KEY, etc. are used by individual functions and are not
   # part of the core FlowKit configuration. These should be set as
   # environment variables or passed directly to functions that need them.
