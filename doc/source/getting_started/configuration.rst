.. _configuration:

Configuration
=============

FlowKit uses YAML configuration and environment variables. Supports local files and Azure Key Vault.

Configuration file
~~~~~~~~~~~~~~~~~~

Copy the example configuration:

.. code-block:: bash

    cp configs/config.yaml.example configs/config.yaml

Core settings
~~~~~~~~~~~~~

Service
-------

.. code-block:: yaml

    FLOWKIT_ADDRESS: "0.0.0.0:50051"  # gRPC server
    FLOWKIT_API_KEY: "your-api-key"   # Authentication

* ``localhost:50051`` for local development
* ``0.0.0.0:50051`` for Docker

Logging
-------

.. code-block:: yaml

    LOG_LEVEL: "info"                 # debug, info, warning, error, fatal
    ERROR_FILE_LOCATION: "error.log"

    # Local logs
    LOCAL_LOGS: true
    LOCAL_LOGS_LOCATION: "logs.log"

    # Datadog (optional)
    DATADOG_LOGS: false
    LOGGING_URL: "https://http-intake.logs.datadoghq.eu/api/v2/logs"
    LOGGING_API_KEY: ""

    # Metadata
    STAGE: "DEV"
    VERSION: "1.0.0"
    SERVICE_NAME: "aali-flowkit"

External services
~~~~~~~~~~~~~~~~~

LLM handler
-----------

.. code-block:: yaml

    LLM_HANDLER_ENDPOINT: "ws://aali-llm:9003"

Databases
---------

.. code-block:: yaml

    # Graph database
    GRAPHDB_ADDRESS: "aali-graphdb:8080"

    # Qdrant
    QDRANT_HOST: "qdrant"
    QDRANT_PORT: 6334

Security
~~~~~~~~

SSL/TLS
-------

.. code-block:: yaml

    USE_SSL: false
    SSL_CERT_PUBLIC_KEY_FILE: "/path/to/cert.pem"
    SSL_CERT_PRIVATE_KEY_FILE: "/path/to/key.pem"

Azure Key Vault
---------------

.. code-block:: yaml

    EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT: true
    AZURE_KEY_VAULT_NAME: "your-vault-name"
    AZURE_MANAGED_IDENTITY_ID: "your-identity-id"

Environment variables
~~~~~~~~~~~~~~~~~~~~~

Configuration file location
----------------------------

By default, FlowKit looks for ``configs/config.yaml`` in the working directory.
Override with ``AALI_CONFIG_PATH``:

.. code-block:: bash

    export AALI_CONFIG_PATH=/custom/path/to/config.yaml
    export AALI_CONFIG_PATH=/etc/aali/flowkit.yaml  # System-wide config
    export AALI_CONFIG_PATH=./env/production.yaml   # Environment-specific

Override configuration values
-----------------------------

Override any config value:

.. code-block:: bash

    export LOG_LEVEL=debug
    export FLOWKIT_ADDRESS=localhost:50051
    export LLM_HANDLER_ENDPOINT=ws://localhost:9003

Environment variables take precedence over ``config.yaml``.

.. _advanced-configuration:

Advanced Configuration Options
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

For production deployments and specialized use cases, FlowKit supports additional configuration options.

Workflow Engine Settings
-------------------------

Configure dynamic workflow execution and tool management:

.. code-block:: yaml

    # Meshpilot Database
    MESHPILOT_DB_ENDPOINT: "your-meshpilot-endpoint"

    # Collection Configuration
    COLLECTION_1_NAME: "primary_collection"
    COLLECTION_2_NAME: "secondary_collection"
    COLLECTION_3_NAME: "tertiary_collection"
    COLLECTION_4_NAME: "workflow_collection"
    COLLECTION_5_NAME: "cache_collection"
    COLLECTION_6_NAME: "metadata_collection"

Tool Configuration
------------------

Dynamic tool and action configuration:

.. code-block:: yaml

    # Tool Counts (determines how many tools are loaded)
    APP_ACTIONS_TOOL_TOTAL_AMOUNT: 17
    APP_HELPER_TOOL_TOTAL_AMOUNT: 6

    # Tool Definitions (examples for first few tools)
    APP_TOOL_1_NAME: "primary_tool"
    APP_TOOL_3_NAME: "analysis_tool"

    # Action Definitions
    APP_TOOL_ACTION_1_NAME: "initialize_workflow"
    APP_TOOL_ACTION_2_NAME: "process_data"
    APP_TOOL_ACTION_3_NAME: "validate_results"
    APP_TOOL_ACTION_4_NAME: "generate_report"
    APP_TOOL_ACTION_5_NAME: "cleanup_resources"
    APP_TOOL_ACTION_6_NAME: "archive_results"

    # Additional actions (11, 12, 14, 15, 17 are commonly used)
    APP_TOOL_ACTION_11_NAME: "mesh_analysis"
    APP_TOOL_ACTION_12_NAME: "optimization"
    APP_TOOL_ACTION_14_NAME: "quality_check"
    APP_TOOL_ACTION_15_NAME: "post_processing"
    APP_TOOL_ACTION_17_NAME: "final_validation"

Database Query Templates
------------------------

Customize database interaction patterns:

.. code-block:: yaml

    # Node and Property Queries
    APP_DATABASE_GET_PROPERTIES_QUERY: "MATCH (n) RETURN properties(n)"
    APP_DATABASE_GET_STATE_NODE_QUERY: "MATCH (s:State) RETURN s"

    # Path Queries
    APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_LABEL_1: "primary_nodes"
    APP_DATABASE_FETCH_PATH_NODES_QUERY_NODE_LABEL_2: "secondary_nodes"

    # Action Queries
    APP_DATABASE_GET_ACTIONS_QUERY_LABEL_1: "workflow_actions"
    APP_DATABASE_GET_ACTIONS_QUERY_LABEL_2: "system_actions"
    APP_DATABASE_GET_SOLUTIONS_QUERY: "MATCH (sol:Solution) RETURN sol"

Prompt Templates
----------------

Configure AI prompt templates for workflow synthesis:

.. code-block:: yaml

    # Action Synthesis
    APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_FIND_KEY: "search_pattern"
    APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_TOOL2_VALUE: "analysis_prompt"
    APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_REPLACE_KEY_1: "replace_pattern_1"
    APP_PROMPT_TEMPLATE_SYNTHESIZE_ACTION_REPLACE_KEY_2: "replace_pattern_2"

    # Output Synthesis
    APP_PROMPT_TEMPLATE_SYNTHESIZE_OUTPUT_KEY_1: "output_format_1"
    APP_PROMPT_TEMPLATE_SYNTHESIZE_OUTPUT_KEY_2: "output_format_2"

    # Workflow Identification
    APP_SUBWORKFLOW_IDENTIFICATION_SYSTEM_PROMPT: "You are a workflow analyzer..."
    APP_SUBWORKFLOW_IDENTIFICATION_USER_PROMPT: "Identify the workflow pattern..."

External Service Integration
----------------------------

Additional service endpoints beyond basic configuration:

.. code-block:: yaml

    # Extended LLM Services
    LLM_API_KEY: "your-llm-api-key"

    # Python Integration
    FLOWKIT_PYTHON_ENDPOINT: "http://python-service:8000"
    FLOWKIT_PYTHON_API_KEY: "python-service-key"

    # Authorization Services
    ANSYS_AUTHORIZATION_URL: "https://auth.ansys.com/oauth/token"

    # Additional gRPC Ports
    EXTERNALFUNCTIONS_GRPC_PORT: 50052

Success and Status Messages
---------------------------

Customize user-facing messages for different tool outcomes:

.. code-block:: yaml

    # Success Messages
    APP_ACTION_TOOL_1_SUCCESS_MESSAGE: "Primary tool executed successfully"
    APP_ACTION_TOOL_2_SUCCESS_MESSAGE: "Secondary tool completed"
    APP_ACTION_TOOL_3_SUCCESS_MESSAGE: "Analysis tool finished"
    APP_ACTION_TOOL_4_SUCCESS_MESSAGE: "Report generated"
    APP_ACTION_TOOL_5_SUCCESS_MESSAGE: "Cleanup completed"
    APP_ACTION_TOOL_6_SUCCESS_MESSAGE: "Results archived"

    # No Action Messages
    APP_ACTION_TOOL_1_NO_ACTION_MESSAGE: "Primary tool: no action required"
    APP_ACTION_TOOL_2_NO_ACTION_MESSAGE: "Secondary tool: skipped"
    APP_ACTION_TOOL_3_NO_ACTION_MESSAGE: "Analysis tool: no changes needed"

    # Extended Action Messages
    APP_ACTION_TOOL_11_SUCCESS_MESSAGE: "Mesh analysis completed"
    APP_ACTION_TOOL_12_SUCCESS_MESSAGE: "Optimization finished"
    APP_ACTION_TOOL_14_SUCCESS_MESSAGE: "Quality check passed"
    APP_ACTION_TOOL_15_SUCCESS_MESSAGE: "Post-processing done"
    APP_ACTION_TOOL_17_SUCCESS_MESSAGE: "Final validation successful"

    APP_ACTION_TOOL_14_NO_ACTION_MESSAGE: "Quality check: no issues found"
    APP_ACTION_TOOL_15_NO_ACTION_MESSAGE: "Post-processing: not required"

Tool Action Configuration
-------------------------

Configure tool behavior and targeting:

.. code-block:: yaml

    # Action Keys and Targets
    APP_TOOL_ACTIONS_KEY_1: "primary_action_set"
    APP_TOOL_ACTIONS_KEY_2: "secondary_action_set"
    APP_TOOL_ACTIONS_TARGET_1: "workflow_target"

.. note::

   **When to use Advanced Configuration:**

   * **Workflow Engine Settings**: When deploying custom Meshpilot workflows
   * **Tool Configuration**: For dynamic tool loading and custom actions
   * **Prompt Templates**: When customizing AI-driven workflow synthesis
   * **External Services**: For additional API integrations beyond basic setup

   Most users only need the preceding :ref:`Core settings <configuration>` to get started.

Defaults
~~~~~~~~

.. code-block:: go

    // Code fragment - part of larger implementation
    config.InitConfig([]string{}, map[string]interface{}{
        "SERVICE_NAME":        "aali-flowkit",
        "VERSION":             "1.0",
        "STAGE":               "PROD",
        "LOG_LEVEL":           "error",
        "ERROR_FILE_LOCATION": "error.log",
        "LOCAL_LOGS_LOCATION": "logs.log",
        "DATADOG_SOURCE":      "nginx",
    })

Validation
~~~~~~~~~~

Check connectivity:

.. code-block:: bash

    curl -X GET http://aali-graphdb:8080/health
    curl -X GET http://qdrant:6333/health

Next steps
~~~~~~~~~~

* :doc:`running` - Start FlowKit
