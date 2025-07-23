.. _github:

Installing from GitHub
======================

Clone and build FlowKit from source.

Quick Start Alternative
~~~~~~~~~~~~~~~~~~~~~~~

For automated setup, use the quick start script:

.. code-block:: bash

    curl -sSL https://raw.githubusercontent.com/ansys/aali-flowkit/main/scripts/quick_start.sh | bash

Or if you've already cloned the repository:

.. code-block:: bash

    ./scripts/quick_start.sh

The script checks prerequisites, downloads dependencies, and builds FlowKit automatically.

Clone repository
~~~~~~~~~~~~~~~~

.. code-block:: bash

    git clone https://github.com/ansys/aali-flowkit.git
    cd aali-flowkit

Repository structure
~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    ls -la

Key files:

* ``main.go`` - Entry point
* ``pkg/`` - Go packages
* ``configs/`` - Configuration
* ``docker/`` - Docker files
* ``go.mod`` - Dependencies

Install dependencies
~~~~~~~~~~~~~~~~~~~~

.. note::
   The ``go mod download`` step may take 5-10 minutes on first run
   depending on your internet connection. No output during download is normal.

.. code-block:: bash

    go mod download  # This downloads ~500MB of dependencies
    go mod verify

Build
~~~~~

Development build:

.. code-block:: bash

    go build -o aali-flowkit main.go  # Takes 30-60 seconds

Production build:

.. code-block:: bash

    go build -ldflags="-s -w" -o aali-flowkit main.go  # Takes 30-60 seconds

Docker:

.. note::
   Docker build can take 5-10 minutes on first run as it downloads
   base images and installs Python dependencies.

.. code-block:: bash

    docker build -f docker/Dockerfile -t aali-flowkit:latest .

Verify
~~~~~~

.. code-block:: bash

    # Check if binary was built successfully
    ls -la aali-flowkit
    
    # Check version from VERSION file
    cat VERSION

    # Docker - verify image was built
    docker images | grep aali-flowkit

Development
~~~~~~~~~~~

Hot reload:

.. code-block:: bash

    go install github.com/cosmtrek/air@latest
    air

Run tests:

.. code-block:: bash

    go test ./...

Generate code:

.. code-block:: bash

    go generate ./...

Next steps
~~~~~~~~~~

* :doc:`configuration` - Configure FlowKit
* :doc:`running` - Start the service