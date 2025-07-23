Installing FlowKit
==================

Install from GitHub
-------------------

To install the FlowKit repository from GitHub, execute the following commands:

.. code:: bash

   git clone https://github.com/ansys/aali-flowkit.git
   cd aali-flowkit

Create configuration file
~~~~~~~~~~~~~~~~~~~~~~~~~

Create ``config.yaml`` to manage configuration settings.
You can use the example configuration file in the ``configs`` folder.

Required values
~~~~~~~~~~~~~~~

.. code:: bash

   FLOWKIT_ADDRESS: "localhost:50051"  # Use 0.0.0.0:50051 for Docker
   FLOWKIT_API_KEY: "your-api-key"

Optional values
~~~~~~~~~~~~~~~

.. code:: bash

   # General settings
   STAGE: PROD
   VERSION: 1.0
   SERVICE_NAME: aali-flowkit

   # Logging settings
   LOG_LEVEL: info
   ERROR_FILE_LOCATION: error.log
   LOCAL_LOGS: false
   LOCAL_LOGS_LOCATION: logs.log

   # Datadog (optional)
   DATADOG_LOGS: false
   LOGGING_URL: https://http-intake.logs.datadoghq.eu/api/v2/logs
   LOGGING_API_KEY: ""
   DATADOG_SOURCE: nginx

   # Service endpoints
   LLM_HANDLER_ENDPOINT: ws://aali-llm:9003
   GRAPHDB_ADDRESS: http://aali-graphdb:8080
   QDRANT_HOST: qdrant
   QDRANT_PORT: 6334

   # SSL (optional)
   USE_SSL: false
   SSL_CERT_PUBLIC_KEY_FILE: ""
   SSL_CERT_PRIVATE_KEY_FILE: ""

The configuration file serves as a central repository for managing FlowKit settings.
To understand each parameter's purpose, refer to the configuration documentation.

Start the app
---------------------

Start FlowKit by running the following command from the root folder:

.. code:: bash

   go run main.go

Starting the app
~~~~~~~~~~~~~~~~~~~~~~~~

.. code:: bash

   {"level":"info","msg":"Aali FlowKit started successfully; gRPC server listening on address 'localhost:50051.'"}

Handling errors during startup
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

If there is an error starting the app, FlowKit generates a log file named ``error.log`` in the root folder.

By following these logs, you can monitor FlowKit status and troubleshoot any issues during runtime.

.. button-ref:: index
    :ref-type: doc
    :color: primary
    :shadow:
    :expand:

    Go to Getting started
