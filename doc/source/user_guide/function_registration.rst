.. _function_registration:

Function Registration
=====================

Overview
--------

Flowkit provides 148 Go functions accessible through gRPC interface. Functions are dynamically registered at startup and exposed through the ``ListFunctions`` RPC method.

Function Categories
-------------------

**LLM Handler Functions**
   Vector embeddings, general LLM requests, message history management

**Knowledge Database Functions**
   Vector storage operations, similarity search, graph database queries

**Data Extraction Functions**
   File content processing, document parsing, collection management

**Generic Utility Functions**
   UUID generation, string operations, REST API calls, JSON processing

**ANSYS Service Functions**
   Integration with ANSYS GPT, Mesh Pilot, and Materials services

**Authentication Functions**
   API key validation, user management, token handling

**qdrant Functions**
   Vector database collection management and data insertion

**MCP Functions**
   Model Context Protocol resource and tool operations

Function Discovery
------------------

Functions are discovered through:

- Dynamic registration from embedded Go source files
- Reflection-based type extraction
- Automatic parameter and return type detection

Each function includes:

- Name and display name
- Input/output parameter definitions
- Type information and constraints
- Category classification

Next: :doc:`calling`
