Monitoring and logging
======================

Monitor function execution performance, track usage metrics, and analyze system patterns.

Metrics
-------

Currently, AALI Flowkit uses structured logging for observability rather than exposing HTTP metrics endpoints. Function execution metrics can be gathered from the app logs.

Logging
-------

The Flowkit uses structured JSON logging:

.. code-block:: json

   {
     "level": "info",
     "timestamp": "2024-01-15T10:30:00Z",
     "message": "Function execution completed",
     "function_name": "process_data",
     "execution_id": "exec123",
     "duration_ms": 500,
     "status": "success"
   }

Log levels
----------

- ``DEBUG``: Detailed debugging information
- ``INFO``: General operational messages
- ``WARN``: Warning conditions
- ``ERROR``: Error conditions

configuration
-------------

Configure logging via environment variables:

.. code-block:: bash

   export LOG_LEVEL=info
   export LOG_FORMAT=json
   export LOG_OUTPUT=stdout

Health monitoring
-----------------

AALI Flowkit runs as a gRPC service and does not expose HTTP health check endpoints. Monitor service health by:

- Checking if the gRPC server is accepting connections on the configured port
- Monitoring the service logs for startup and error messages
- Using gRPC health check probes if needed (requires implementation)

Alerting
--------

Set up alerts for:

- High error rates (>5% for 5 minutes)
- Long execution times (>30 seconds)
- High memory usage (>80%)
- Function execution failures

Dashboard
---------

For observability dashboards, consider monitoring:

- gRPC request rates and response times
- Application log error rates
- System resource usage (CPU, memory)
- Function execution patterns from log analysis
- Connection counts to backend services (LLM, databases)
