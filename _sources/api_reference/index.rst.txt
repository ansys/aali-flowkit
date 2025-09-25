.. _api_reference:

API Reference
=============

Technical documentation for advanced users and developers extending FlowKit.

This section contains detailed API documentation for FlowKit's internal packages and implementation details.

.. grid:: 2
   :gutter: 3 3 4 4

   .. grid-item-card:: gRPC Server
      :link: grpcserver/index
      :link-type: doc

      Core gRPC service implementation with RunFunction
      and ListFunctions RPCs.

   .. grid-item-card:: External Functions
      :link: externalfunctions/index
      :link-type: doc

      Over 180 functions across multiple categories available
      through the gRPC interface.

   .. grid-item-card:: Function Definitions
      :link: functiondefinitions/index
      :link-type: doc

      AST parsing and function discovery system that
      extracts metadata from Go source.

   .. grid-item-card:: Internal States
      :link: internalstates/index
      :link-type: doc

      Function registry and state management for
      thread-safe operation.

   .. grid-item-card:: Function Testing
      :link: functiontesting/index
      :link-type: doc

      Testing utilities and helpers for verifying
      function implementations.

Private Functions
-----------------

.. grid:: 2
   :gutter: 3 3 4 4

   .. grid-item-card:: Code Generation
      :link: privatefunctions/codegeneration/index
      :link-type: doc

      Utilities for generating Go code including
      cast functions and type conversions.

   .. grid-item-card:: Generic Utilities
      :link: privatefunctions/generic/index
      :link-type: doc

      Common utility functions for string manipulation,
      JSON handling, and general operations.

   .. grid-item-card:: Graph Database
      :link: privatefunctions/graphdb/index
      :link-type: doc

      MongoDB-based graph database operations for
      workflow and knowledge management.

   .. grid-item-card:: Qdrant Integration
      :link: privatefunctions/qdrant/index
      :link-type: doc

      Vector database operations for semantic search
      and embeddings storage.

MeshPilot Integration
---------------------

.. grid:: 2
   :gutter: 3 3 4 4

   .. grid-item-card:: AMP Graph Database
      :link: meshpilot/ampgraphdb/index
      :link-type: doc

      ANSYS MeshPilot graph database integration
      for mesh analysis workflows.

   .. grid-item-card:: Azure Integration
      :link: meshpilot/azure/index
      :link-type: doc

      Azure-specific functions for MeshPilot
      cloud deployments.

.. toctree::
   :maxdepth: 2
   :hidden:
   :caption: Core Packages

   grpcserver/index
   functiondefinitions/index
   internalstates/index

.. toctree::
   :maxdepth: 2
   :hidden:
   :caption: Function Packages

   externalfunctions/index
   functiontesting/index

.. toctree::
   :maxdepth: 2
   :hidden:
   :caption: Private Functions

   privatefunctions/codegeneration/index
   privatefunctions/generic/index
   privatefunctions/graphdb/index
   privatefunctions/qdrant/index

.. toctree::
   :maxdepth: 2
   :hidden:
   :caption: MeshPilot

   meshpilot/ampgraphdb/index
   meshpilot/azure/index
