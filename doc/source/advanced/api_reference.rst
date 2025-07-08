.. _API-reference:

API reference
=============

Complete API documentation for AALI Flowkit, including all available functions and their parameters.

.. note::
   The complete API reference documentation is auto-generated during the build process from the Go source code.

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

1. **By Name**: Use the search feature (Ctrl+K) to find functions by name
2. **By Feature**: Check the categorized lists below
3. **By Package**: Browse the auto-generated documentation sections

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
