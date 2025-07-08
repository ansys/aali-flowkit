.. _API-reference:

API reference
=============

Complete API documentation for AALI Flowkit, including all available functions and their parameters.

Quick navigation
----------------

.. grid:: 1 2 2 3
   :gutter: 2
   :class-container: sd-mb-3

   .. grid-item-card:: External Functions
      :link: autodoc/externalfunctions
      :link-type: doc
      :class-card: sd-shadow-sm sd-rounded-2 sd-p-2

      All functions available through gRPC interface

   .. grid-item-card:: Function Definitions
      :link: autodoc/functiondefinitions
      :link-type: doc
      :class-card: sd-shadow-sm sd-rounded-2 sd-p-2

      Core function structures and types

   .. grid-item-card:: gRPC Server
      :link: autodoc/grpcserver
      :link-type: doc
      :class-card: sd-shadow-sm sd-rounded-2 sd-p-2

      Server implementation details

   .. grid-item-card:: Internal States
      :link: autodoc/internalstates
      :link-type: doc
      :class-card: sd-shadow-sm sd-rounded-2 sd-p-2

      State management functions

   .. grid-item-card:: Private Functions
      :link: autodoc/privatefunctions_generic
      :link-type: doc
      :class-card: sd-shadow-sm sd-rounded-2 sd-p-2

      Internal helper functions

   .. grid-item-card:: Testing Functions
      :link: autodoc/functiontesting
      :link-type: doc
      :class-card: sd-shadow-sm sd-rounded-2 sd-p-2

      Test utilities and helpers

Overview
--------

Flowkit provides over 170 functions organized into the following categories:

**Core Functions** (``externalfunctions``)
   The main functions exposed through the gRPC interface for external use.

**Infrastructure** (``grpcserver``, ``functiondefinitions``)
   Server setup, function registration, and type definitions.

**Data Management** (``internalstates``)
   Memory management, state tracking, and session handling.

**Integrations** (``privatefunctions_*``)
   - ``privatefunctions_qdrant``: Vector database operations
   - ``privatefunctions_graphdb``: Graph database queries
   - ``privatefunctions_codegeneration``: AI code generation
   - ``privatefunctions_generic``: Utility functions

**Specialized** (``meshpilot_*``)
   - ``meshpilot_ampgraphdb``: AMP graph database integration
   - ``meshpilot_azure``: azure cloud services

Finding functions
-----------------

To find specific functions:

1. **By Category**: Use the cards that precede to browse by function type
2. **By Name**: Use the search feature (Ctrl+K) to find functions by name
3. **By Feature**: Check the categorized lists in each section

Common function patterns
------------------------

Most Flowkit functions follow these patterns:

**Input/Output Functions**
   - Accept specific typed parameters
   - Return structured results
   - Handle errors gracefully

**State-Aware Functions**
   - Can access session memory
   - Preserve context between calls
   - Support workflow continuity

**Integration Functions**
   - Connect to external services
   - Transform data formats
   - Handle authentication

Auto-generated documentation
----------------------------

All documentation in this section is auto-generated from Go source code using custom parsing tools. This ensures the documentation stays in sync with the actual implementation.
