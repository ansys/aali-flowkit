.. _running:

Running FlowKit
===============

Start FlowKit in various environments.

Local development
~~~~~~~~~~~~~~~~~

Direct execution
----------------

.. code-block:: bash

    ./aali-flowkit

Server starts on configured port (default 50051).

Go run
------

.. code-block:: bash

    go run main.go

Docker
~~~~~~

Run container
-------------

.. code-block:: bash

    docker run -d \
      --name aali-flowkit \
      -p 50051:50051 \
      -v $(pwd)/configs:/app/configs \
      -e LOG_LEVEL=info \
      aali-flowkit:latest

Docker Compose
--------------

.. code-block:: yaml

    version: '3.8'
    services:
      flowkit:
        image: aali-flowkit:latest
        ports:
          - "50051:50051"
        environment:
          - LOG_LEVEL=info
          - FLOWKIT_ADDRESS=0.0.0.0:50051
          - LLM_HANDLER_ENDPOINT=ws://llm:9003
          - GRAPHDB_ADDRESS=graphdb:8080
          - QDRANT_HOST=qdrant
        volumes:
          - ./configs:/app/configs

Kubernetes
~~~~~~~~~~

.. code-block:: yaml

    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: aali-flowkit
    spec:
      replicas: 3
      selector:
        matchLabels:
          app: aali-flowkit
      template:
        metadata:
          labels:
            app: aali-flowkit
        spec:
          containers:
          - name: flowkit
            image: aali-flowkit:latest
            ports:
            - containerPort: 50051
              name: grpc
            env:
            - name: FLOWKIT_ADDRESS
              value: "0.0.0.0:50051"
            livenessProbe:
              grpc:
                port: 50051
              initialDelaySeconds: 5
            readinessProbe:
              grpc:
                port: 50051
              initialDelaySeconds: 5


Systemd
~~~~~~~

``/etc/systemd/system/aali-flowkit.service``:

.. code-block:: ini

    [Unit]
    Description=AALI FlowKit gRPC Service
    After=network.target

    [Service]
    Type=simple
    User=aali
    WorkingDirectory=/opt/aali-flowkit
    ExecStart=/opt/aali-flowkit/aali-flowkit
    Restart=always
    RestartSec=5

    [Install]
    WantedBy=multi-user.target

.. code-block:: bash

    sudo systemctl enable aali-flowkit
    sudo systemctl start aali-flowkit

Monitoring
~~~~~~~~~~

.. code-block:: bash

    # Local
    tail -f logs.log

    # Docker
    docker logs -f aali-flowkit

    # Kubernetes
    kubectl logs -f deployment/aali-flowkit

Shutdown
~~~~~~~~

.. code-block:: bash

    # Process
    kill -SIGTERM $(pgrep aali-flowkit)

    # Docker
    docker stop aali-flowkit

    # Kubernetes
    kubectl scale deployment aali-flowkit --replicas=0

Troubleshooting
~~~~~~~~~~~~~~~

**Port in use**

.. code-block:: bash

    lsof -i :50051
    export FLOWKIT_ADDRESS=localhost:50052

**Connection refused**

* Check firewall
* Verify service running
* Confirm client config

**Function not found**

* Function must be exported
* Check embedded files in main.go
* Review startup logs

Next steps
~~~~~~~~~~

* :doc:`../user_guide/connect` - Connect to FlowKit