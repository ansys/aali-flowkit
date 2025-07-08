.. _installation:

Installation
============

* Go 1.24 or later
* Git

Build from source
-----------------

.. code-block:: bash

   git clone https://github.com/ansys/aali-flowkit.git
   cd aali-flowkit
   go mod tidy
   go build -o flowkit main.go

Docker
------

.. code-block:: bash

   docker build -t aali-flowkit .
   docker run -p 50051:50051 aali-flowkit

Run
---

.. code-block:: bash

   ./flowkit

The server listens on port 50051 by default. See :doc:`configuration` for setup options.
