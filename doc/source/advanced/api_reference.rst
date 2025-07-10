.. _API-reference:

API reference
=============

API documentation for AALI Flowkit functions and gRPC interface.

Overview
--------

Flowkit provides 148 functions organized into the following categories:

**External Functions** (``externalfunctions``)
   Functions exposed through the gRPC interface.

**Function Registration** (``functiondefinitions``)
   Function registration and type definitions.

**gRPC Server** (``grpcserver``)
   Server implementation and request handling.

**Internal States** (``internalstates``)
   State management for function registration.

Function Categories
-------------------

**LLM Handler Functions**
   Vector embeddings, general requests, message history management.

**Knowledge Database Functions**
   Vector storage, similarity search, graph database queries.

**Data Extraction Functions**
   File content processing, document parsing, collection management.

**Generic Utility Functions**
   UUID generation, string operations, REST API calls.

**ANSYS Service Functions**
   Integration with ANSYS GPT, Mesh Pilot, and Materials services.

**Authentication Functions**
   API key validation, user management, token handling.

gRPC Interface
--------------

The server exposes two main RPC methods:

- ``ListFunctions``: Returns all available functions
- ``RunFunction``: Executes a function with provided inputs
