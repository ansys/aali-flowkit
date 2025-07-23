Prerequisites
=============

1. Install GO
-------------

.. tab-set::

   .. tab-item:: macOS

    Follow these steps to install Go on macOS.

    1. Install Homebrew if you haven't already. You can find instructions on the `Homebrew website <https://brew.sh/>`_.
    2. Open a Terminal and run the following commands:

     .. code:: bash

        brew update
        brew install go

   .. tab-item:: Ubuntu

      .. code:: bash

         sudo apt update
         sudo apt install snapd
         sudo snap install go --classic

   .. tab-item:: Windows

      Follow these steps to install Go on Windows:

      1. Download the Go installer from the official website: `Go Downloads <https://golang.org/dl/>`_.
      2. Run the installer executable and follow the on-screen instructions.
      3. Once the installation is complete, open a new Command Prompt and verify the installation by running the command:

      .. code:: bash

        go version

2. Install Git
--------------

.. tab-set::

   .. tab-item:: macOS

      .. code:: bash

        brew install git

   .. tab-item:: Ubuntu

      .. code:: bash

         sudo apt update
         sudo apt install git

   .. tab-item:: Windows

      1. Download the Git installer from the official website: `Git Downloads <https://git-scm.com/download/win>`_.
      2. Run the installer executable and follow the on-screen instructions.
      3. Once the installation is complete, open a new Command Prompt and verify the installation by running the command:

      .. code:: bash

        git --version


.. button-ref:: index
    :ref-type: doc
    :color: primary
    :shadow:
    :expand:

    Go to Getting started
