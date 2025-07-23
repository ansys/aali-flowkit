.. _troubleshooting:

Troubleshooting
===============

Common issues and solutions for setting up FlowKit.

Installation Issues
~~~~~~~~~~~~~~~~~~~

Go version errors
-----------------

**Problem:** ``go: command not found``

**Solution:**

.. code-block:: bash

    # Ensure Go is in your PATH
    export PATH=$PATH:/usr/local/go/bin

    # Add to your shell profile for persistence
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    source ~/.bashrc

**Problem:** Go version is too old (< 1.21)

**Solution:** Download and install the latest Go version from https://go.dev/dl/

Missing protoc (Development Only)
---------------------------------

**Problem:** ``protoc: command not found``

**Solution:** This is only needed for development. If you're just running FlowKit, you can ignore this error.

Build Issues
~~~~~~~~~~~~

Dependency download hangs
-------------------------

**Problem:** ``go mod download`` appears stuck

**Solution:** This is normal. The command downloads ~500MB of dependencies and shows no progress. Wait 5-10 minutes.

If it's truly stuck:

.. code-block:: bash

    # Check your Go proxy settings
    go env GOPROXY

    # Set a different proxy if needed
    export GOPROXY=https://proxy.golang.org,direct

Build fails with permission errors
-----------------------------------

**Problem:** Permission denied when building

**Solution:**

.. code-block:: bash

    # Ensure you own the directory
    sudo chown -R $USER:$USER .

    # Or build in your home directory
    cd ~
    git clone https://github.com/ansys/aali-flowkit.git
    cd aali-flowkit

Configuration Issues
~~~~~~~~~~~~~~~~~~~~

Config file not found
---------------------

**Problem:** ``configs/config.yaml.example`` not found

**Solution:** Use the existing ``configs/config.yaml`` as your example:

.. code-block:: bash

    # The repository includes a documented config.yaml
    cp configs/config.yaml configs/my-config.yaml
    # Edit my-config.yaml with your settings

Invalid configuration values
----------------------------

**Problem:** FlowKit fails to start with config errors

**Solution:** Check these common settings:

- ``FLOWKIT_ADDRESS``: Use ``localhost:50051`` for local, ``0.0.0.0:50051`` for Docker
- ``LLM_HANDLER_ENDPOINT``: Must be a valid WebSocket URL (ws:// or wss://)
- Port numbers must be integers, not strings

Runtime Issues
~~~~~~~~~~~~~~

Port already in use
-------------------

**Problem:** ``bind: address already in use``

**Solution:**

.. code-block:: bash

    # Find what's using the port
    lsof -i :50051

    # Use a different port
    export FLOWKIT_ADDRESS=localhost:50052

    # Or in config.yaml
    FLOWKIT_ADDRESS: "localhost:50052"

Cannot connect to dependencies
------------------------------

**Problem:** Cannot connect to GraphDB, Qdrant, or LLM handler

**Solution:** These are optional services. FlowKit starts without them but some features won't work.

To run with full features:

.. code-block:: bash

    # Use Docker Compose (if available)
    docker-compose up -d

    # Or disable in config
    EXTRACT_CONFIG_FROM_AZURE_KEY_VAULT: false
    DATADOG_LOGS: false
    DATADOG_METRICS: false

Binary won't start
------------------

**Problem:** ``./aali-flowkit`` gives permission denied

**Solution:**

.. code-block:: bash

    # Make it executable
    chmod +x aali-flowkit

    # Run it
    ./aali-flowkit

Docker Issues
~~~~~~~~~~~~~

Docker build fails
------------------

**Problem:** Docker build errors

**Solution:** Ensure Docker daemon is running and you have sufficient disk space:

.. code-block:: bash

    # Check Docker
    docker info

    # Clean up old images
    docker system prune -a

    # Build with no cache
    docker build --no-cache -f docker/Dockerfile -t aali-flowkit:latest .

Network Issues
~~~~~~~~~~~~~~

Cannot clone repository
-----------------------

**Problem:** ``fatal: unable to access 'https://github.com/ansys/aali-flowkit.git'``

**Solution:**

.. code-block:: bash

    # Check network connectivity
    ping github.com

    # Try SSH instead of HTTPS
    git clone git@github.com:ansys/aali-flowkit.git

    # Or use a proxy if behind firewall
    export https_proxy=http://your-proxy:port
    git clone https://github.com/ansys/aali-flowkit.git

Getting Help
~~~~~~~~~~~~

If you encounter issues not covered here:

1. Check the `GitHub Issues <https://github.com/ansys/aali-flowkit/issues>`_
2. Review the complete logs: ``cat logs.log`` or ``cat error.log``
3. Run with debug logging: ``export LOG_LEVEL=debug``
4. Create a new issue with:

   - Your OS and Go version
   - Complete error messages
   - Steps to reproduce
