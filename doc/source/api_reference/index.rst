.. _api_reference:

API Reference
=============

Complete reference for FlowKit packages and functions, automatically generated from source code.

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

.. toctree::
   :maxdepth: 2
   :hidden:
   :caption: Private Functions

   privatefunctions/codegeneration/index
   privatefunctions/generic/index
   privatefunctions/graphdb/index
   privatefunctions/qdrant/index
