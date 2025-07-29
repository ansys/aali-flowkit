.. _types_dev:

Adding Custom Types
===================

Define custom types for complex data structures in ``pkg/externalfunctions/types.go``.

Basic Type Definition
---------------------

.. code-block:: go

   type MyCustomType struct {
       Name  string  `json:"name"`
       Value float64 `json:"value"`
       Items []string `json:"items"`
   }

Using External Types
--------------------

Import types from shared packages:

.. code-block:: go

   import "github.com/ansys/aali-sharedtypes/pkg/sharedtypes"

   // Example usage in types.go:
   type similarityElement struct {
       Score float64                `json:"distance"`
       Data  sharedtypes.DbResponse `json:"data"`
   }

JSON Tags
---------

Add JSON tags for proper serialization. FlowKit supports extended tags:

.. code-block:: go

   // Example from types.go:
   type similaritySearchInput struct {
       CollectionName string   `json:"collection_name" description:"Name of the collection" required:"true"`
       MaxRetrievalCount int   `json:"max_retrieval_count" description:"Max results" required:"false"`
       OutputFields []string  `json:"output_fields" description:"Fields to include"`
   }

Type Usage
----------

Once defined, use your custom types in functions:

.. code-block:: go

   func MyFunction(input MyCustomType) string {
       // Implementation
       if err != nil {
           panic(fmt.Sprintf("Operation failed: %v", err))
       }
       return result
   }

Registering Functions with Custom Types
---------------------------------------

Don't forget to register functions that use custom types in ``ExternalFunctionsMap``:

.. code-block:: go

   var ExternalFunctionsMap = map[string]interface{}{
       "MyFunction": MyFunction,
       // ... other functions
   }

That's it. Types are automatically available after definition, but functions using them must still be registered.

.. seealso::
   - :ref:`functions_dev` for adding custom functions
   - :ref:`categories_dev` for organizing functions into categories
