.. _functions_dev:

Adding Custom Functions
=======================

Extend FlowKit by adding new Go functions that can be called via gRPC.

Requirements
------------

Your function must:

- Be exported (start with capital letter)
- Have a descriptive comment
- Include ``@displayName`` tag
- Document parameters and returns
- Use ``panic()`` for errors
- Be registered in ``ExternalFunctionsMap``

Example
-------

.. code-block:: go

   // GenerateUUID generates a new UUID (Universally Unique Identifier).
   //
   // Tags:
   //   - @displayName: Generate UUID
   //
   // Returns:
   //   - a string representation of the generated UUID
   func GenerateUUID() string {
       return strings.Replace(uuid.New().String(), "-", "", -1)
   }

Where to Add Functions
----------------------

Add your function to one of these files in ``pkg/externalfunctions/``:

- ``generic.go`` - Utility functions
- ``cast.go`` - Type conversions
- ``llmhandler.go`` - AI/LLM operations
- ``knowledgedb.go`` - Database operations
- ``dataextraction.go`` - File processing
- ``auth.go`` - Authentication
- ``mcp.go`` - MCP protocol functions
- ``ansysgpt.go`` - Ansys GPT specific functions
- ``ansysmaterials.go`` - Materials database functions
- ``ansysmeshpilot.go`` - Mesh generation functions
- ``qdrant.go`` - Qdrant vector database operations
- ``rhsc.go`` - Red Hat Service Catalog functions

Registering Your Function
-------------------------

**CRITICAL STEP:** after implementing your function, you must register it in ``pkg/externalfunctions/externalfunctions.go``:

.. code-block:: go

   var ExternalFunctionsMap = map[string]interface{}{
       // Add your function here
       "GenerateUUID": GenerateUUID,
       // ... other functions
   }

.. warning::
   Functions not added to ``ExternalFunctionsMap`` are discovered but **cannot be executed**.

Error Handling
--------------

Always use panic with descriptive messages:

.. code-block:: go

   if err != nil {
       panic(fmt.Sprintf("Failed to parse URL: %v", err))
   }

Complete Example
----------------

Here's the complete process:

1. **Implement** your function in a category file:

   .. code-block:: go

      // In pkg/externalfunctions/generic.go
      func MyNewFunction(input string) string {
          if input == "" {
              panic("Input cannot be empty")
          }
          return "Processed: " + input
      }

2. **Register** it in ``ExternalFunctionsMap``:

   .. code-block:: go

      // In pkg/externalfunctions/externalfunctions.go
      var ExternalFunctionsMap = map[string]interface{}{
          "MyNewFunction": MyNewFunction,
          // ... other functions
      }

That's it. Your function is now available via gRPC.

.. seealso::
   - :ref:`categories_dev` for creating new categories
   - :ref:`types_dev` for working with custom types
