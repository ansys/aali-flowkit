.. _prerequisites:

Prerequisites
=============

FlowKit requires the following tools and services.

Platform
~~~~~~~~

* **Linux**: Ubuntu 20.04 LTS or later
* **macOS**: macOS 11 or later
* **Windows**: Windows 10/11 with WSL2

Required for Running FlowKit
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

These tools are essential to run FlowKit.

Go 1.21+
--------

**Installation time:** ~5 minutes

.. code-block:: bash

    # Verify Go installation
    go version

    # Install Go (Linux)
    wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin

Git
---

**Installation time:** ~2 minutes

.. code-block:: bash

    # Ubuntu/Debian
    sudo apt-get update && sudo apt-get install git
    
    # macOS (with Homebrew)
    brew install git

Make (Build Tool)
-----------------

**Installation time:** ~2 minutes

.. code-block:: bash

    # Ubuntu/Debian
    sudo apt-get install build-essential
    
    # macOS - typically pre-installed with Xcode Command Line Tools
    xcode-select --install

Required for Development Only
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

These tools are only needed if you plan to modify gRPC definitions or regenerate code.

Protobuf Compiler
-----------------

**Installation time:** ~5 minutes

.. code-block:: bash

    # Ubuntu/Debian
    sudo apt-get install protobuf-compiler

    # macOS
    brew install protobuf

Go gRPC Code Generation Tools
-----------------------------

**Installation time:** ~3 minutes

.. code-block:: bash

    # Install protoc plugins for Go
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

Optional tools
~~~~~~~~~~~~~~

Docker
------

.. code-block:: bash

    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh

The Docker image includes Python 3.10+ and PyAnsys libraries.


Service dependencies
~~~~~~~~~~~~~~~~~~~~

Required credentials for:

* MongoDB (authentication)
* Qdrant (vector storage)
* Azure Key Vault (secrets)
* LLM endpoints

Environment setup
~~~~~~~~~~~~~~~~~

.. code-block:: bash

    export SERVICE_NAME="aali-flowkit"
    export LOG_LEVEL="info"
    export GRPC_PORT="50051"

    # Service endpoints
    export LLM_HANDLER_ENDPOINT="https://your-llm-service"
    export KNOWLEDGE_DB_ENDPOINT="https://your-knowledge-db"
    export QDRANT_ENDPOINT="https://your-qdrant-instance"



Next steps
~~~~~~~~~~

* :doc:`github` - Clone the repository
* :doc:`configuration` - Configure FlowKit