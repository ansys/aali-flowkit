.. _types_dev:

Adding Custom Types
===================

Custom types enable complex data structures in function parameters and return values.

Type Definition Examples
-------------------------

Define types in ``pkg/externalfunctions/types.go``. Here are real examples from the codebase:

Database Response Type
~~~~~~~~~~~~~~~~~~~~~~

FlowKit uses the external ``sharedtypes.DbResponse`` struct from the shared types package:

.. code-block:: go

   import "github.com/ansys/aali-sharedtypes"
   
   // DbResponse is imported from sharedtypes and includes fields like:
   // - DocumentId, DocumentName, Summary, Guid
   // Used in similarity search results and vector operations
   func ProcessSearchResults() []sharedtypes.DbResponse {
       // Function implementation
   }

Complex Input Type with JSON Tags
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

   // similaritySearchInput represents the input for the similarity search function.
   type similaritySearchInput struct {
       CollectionName    string    `json:"collection_name" description:"Name of the collection to search in. Required." required:"true"`
       EmbeddedVector    []float32 `json:"embedded_vector" description:"Embedded vector used for searching. Required for retrieval." required:"true"`
       MaxRetrievalCount int       `json:"max_retrieval_count" description:"Maximum number of results to retrieve. Optional." required:"false"`
       MinScore          float64   `json:"min_score" description:"Minimum similarity score for results. Optional." required:"false"`
       Keywords          []string  `json:"keywords" description:"Keywords to filter results. Optional." required:"false"`
       UseKeywords       bool      `json:"use_keywords" description:"Whether to use keyword filtering. Optional." required:"false"`
   }

Internal Type Example
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

   // Variables struct used in function processing
   type variables struct {
       variableName string
       metadata     map[string]interface{}
   }

Type Registration and Usage
---------------------------

Custom types are automatically available once defined in ``types.go``. They can be used in function parameters and return values:

.. code-block:: go

   // Function using custom types
   func ProcessSearchResults(input similaritySearchInput) ([]DbResponse, error) {
       // Function implementation
   }

Type Conversion Functions
-------------------------

The ``cast.go`` file contains type conversion functions. Example:

.. code-block:: go

   // CastAnyToString casts data of type any to string
   //
   // Tags:
   //   - @displayName: Cast Any to String
   //
   // Parameters:
   //   - data (any)
   //
   // Returns:
   //   - string
   func CastAnyToString(data any) string {
       return data.(string)
   }

   // CastStringToAny casts a string to any type
   //
   // Tags:
   //   - @displayName: Cast String to Any
   //
   // Parameters:
   //   - data (string)
   //
   // Returns:
   //   - any
   func CastStringToAny(data string) any {
       return data
   }

   // CastAnyToInt casts data of type any to int
   //
   // Tags:
   //   - @displayName: Cast Any to Int
   //
   // Parameters:
   //   - data (any)
   //
   // Returns:
   //   - int
   func CastAnyToInt(data any) int {
       return data.(int)
   }

Array Types
-----------

Array types are automatically supported. Both single types and array types (prefixed with ``[]``) can be used:

.. code-block:: go

   // Function accepting array parameters
   func ProcessMultipleItems(items []string, scores []float64) []DbResponse {
       // Implementation
   }

Best Practices for Type Definitions
-----------------------------------

1. **Use JSON tags** for proper serialization:
   
   .. code-block:: go
   
      type MyType struct {
          Field1 string `json:"field1"`
          Field2 int    `json:"field2"`
      }

2. **Include descriptions** for complex types:
   
   .. code-block:: go
   
      type ComplexType struct {
          Field string `json:"field" description:"Description of the field" required:"true"`
      }

3. **Use proper Go naming conventions** - exported types start with capital letters.

4. **Include error handling** in cast functions using ``logPanic``.
