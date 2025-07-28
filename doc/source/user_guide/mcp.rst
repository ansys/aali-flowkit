.. _mcp:

MCP Integration
===============

Model Context Protocol (MCP) integration enables FlowKit to connect with external AI tools and services through a standardized protocol.

Overview
--------

The Model Context Protocol (MCP) is an open protocol that enables seamless integration between AI applications and external data sources or tools. FlowKit's MCP integration allows you to:

- Connect to MCP-compliant servers
- Execute tools from external MCP providers  
- Access resources from connected MCP servers
- Retrieve system prompts for AI model configuration

MCP Architecture
----------------

FlowKit acts as an MCP client, connecting to external MCP servers that provide tools, resources, and prompts. The integration follows the standard MCP client-server architecture:

.. code-block:: text

   FlowKit (MCP Client) <---> MCP Server <---> External Tools/Resources
   
The MCP client in FlowKit manages connections, handles protocol negotiation, and provides a unified interface for accessing MCP capabilities through gRPC.

Available Functions
-------------------

FlowKit provides seven MCP functions that fully implement the MCP standard:

.. list-table:: MCP Functions
   :header-rows: 1
   :widths: 25 75

   * - Function
     - Description
   * - ``ListTools``
     - Lists all available tools from an MCP server
   * - ``ListResources``
     - Lists all available resources from an MCP server
   * - ``ListPrompts``
     - Lists all available prompts from an MCP server
   * - ``CallTool``
     - Calls/executes a specific tool with provided arguments
   * - ``ReadResource``
     - Reads a specific resource from an MCP server
   * - ``GetPrompt``
     - Gets a specific prompt with arguments from an MCP server
   * - ``MCPClient``
     - Generic wrapper function for all MCP operations

Function Details
~~~~~~~~~~~~~~~~

**ListTools Function**

Lists all available tools from an MCP server.

.. code-block:: go

   // List all tools available on the MCP server
   tools, err := ListTools(serverURL)
   // Returns: []MCPTool with name, description, and schema

**ListResources Function**

Lists all available resources from an MCP server.

.. code-block:: go

   // List all resources available on the MCP server
   resources, err := ListResources(serverURL)
   // Returns: []MCPResource with uri, name, and description

**ListPrompts Function**

Lists all available prompts from an MCP server.

.. code-block:: go

   // List all prompts available on the MCP server
   prompts, err := ListPrompts(serverURL)
   // Returns: []MCPPrompt with name, description, and arguments

**CallTool Function**

Executes an MCP tool with the specified parameters.

.. code-block:: go

   // Call a specific tool with arguments
   result, err := CallTool(serverURL, toolName, arguments)
   // Returns: MCPToolResult with output data

**ReadResource Function**

Reads a specific resource from an MCP server.

.. code-block:: go

   // Read a specific resource by URI
   content, err := ReadResource(serverURL, resourceURI)
   // Returns: MCPResourceContent with data and metadata

**GetPrompt Function**

Gets a specific prompt with arguments from an MCP server.

.. code-block:: go

   // Get a specific prompt with arguments
   prompt, err := GetPrompt(serverURL, promptName, arguments)
   // Returns: MCPPromptResult with expanded prompt text

**MCPClient Function**

Generic wrapper for all MCP operations.

.. code-block:: go

   // Generic MCP client for any operation
   result, err := MCPClient(command, serverURL, arguments)
   // command: "list-tools", "call-tool", etc.
   // Returns: interface{} based on the command

Configuration
-------------

MCP integration requires configuration of MCP servers in FlowKit. Add MCP server endpoints to your ``config.yaml``:

.. code-block:: yaml

   # MCP Configuration (example)
   MCP_SERVERS:
     - name: "example-mcp-server"
       endpoint: "http://mcp-server:3000"
       api_key: "your-api-key"

Usage Examples
--------------

**Example 1: Listing Available Tools**

.. code-block:: json

   {
     "name": "ListTools",
     "inputs": {
       "serverURL": "http://mcp-server:3000"
     }
   }

This returns all available tools from the MCP server.

**Example 2: Calling an MCP Tool**

.. code-block:: json

   {
     "name": "CallTool", 
     "inputs": {
       "serverURL": "http://mcp-server:3000",
       "toolName": "data-analyzer",
       "arguments": {
         "data": "sample data",
         "format": "json"
       }
     }
   }

**Example 3: Reading a Resource**

.. code-block:: json

   {
     "name": "ReadResource",
     "inputs": {
       "serverURL": "http://mcp-server:3000",
       "resourceURI": "resource://data/config.json"
     }
   }

**Example 4: Getting a Prompt**

.. code-block:: json

   {
     "name": "GetPrompt",
     "inputs": {
       "serverURL": "http://mcp-server:3000",
       "promptName": "code-review",
       "arguments": {
         "language": "go",
         "context": "function implementation"
       }
     }
   }

Integration with Other FlowKit Functions
----------------------------------------

MCP functions can be combined with other FlowKit capabilities:

1. **With LLM Functions**: Use MCP tools to enhance LLM responses with external data
2. **With Data Extraction**: Process MCP resources through FlowKit's data extraction pipeline
3. **With Knowledge DB**: Store MCP-retrieved data in vector databases


Troubleshooting
---------------

Common issues and solutions:

**MCP Server Connection Failed**
   - Verify server endpoint is correct
   - Check network connectivity
   - Ensure API keys are valid

**Tool Execution Timeout**
   - Check if the MCP server is responding
   - Verify tool arguments are correct
   - Consider increasing timeout values

**Resource Not Found**
   - Verify resource URI format
   - Check server permissions
   - Ensure resource exists on the MCP server

