.. _categories_dev:

Adding Function Categories
==========================

Categories organize related functions. Each category corresponds to a Go file in ``pkg/externalfunctions/``.

Creating a Category
-------------------

1. **Create Category File**

   Create a new file in ``pkg/externalfunctions/`` with your category name (e.g., ``mycategory.go``).

2. **Add Embed Directive**

   Add the ``//go:embed`` directive in ``main.go``:

   .. code-block:: go

      //go:embed pkg/externalfunctions/mycategory.go
      var myCategoryFile string

3. **Update File Mapping**

   Add your category to the files map in ``main.go``:

   .. code-block:: go

      files := map[string]string{
          "my_category": myCategoryFile,
          // ... existing categories
      }

4. **Add Functions to Registry**

   Include your functions in ``ExternalFunctionsMap`` in ``pkg/externalfunctions/externalfunctions.go``:

   .. code-block:: go

      var ExternalFunctionsMap = map[string]interface{}{
          "MyFunction": MyFunction,
          // ... other functions
      }

Category Naming
---------------

- File name: lowercase (``ansysgpt.go``, ``dataextraction.go``)
- Category key: lowercase with underscores (``ansys_gpt``, ``data_extraction``)

Current Categories with Examples
---------------------------------

The following categories exist in FlowKit with example functions from each:

Generic Functions (``generic.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

General purpose utility functions:

.. code-block:: go

   // Example functions in this category:
   "AssignStringToString":  AssignStringToString,
   "SendRestAPICall":       SendRestAPICall,
   "GenerateUUID":          GenerateUUID,
   "StringConcat":          StringConcat,
   "ExtractJSONStringField": ExtractJSONStringField,

LLM Handler Functions (``llmhandler.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

LLM integration and AI functions:

.. code-block:: go

   // Example functions in this category:
   "PerformVectorEmbeddingRequest":    PerformVectorEmbeddingRequest,
   "PerformGeneralRequest":            PerformGeneralRequest,
   "BuildLibraryContext":              BuildLibraryContext,
   "CreateEmbeddings":                 CreateEmbeddings,
   "FinalizeMessage":                  FinalizeMessage,

Knowledge Database Functions (``knowledgedb.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Database operations and vector search:

.. code-block:: go

   // Example functions in this category:
   "SendVectorsToKnowledgeDB": SendVectorsToKnowledgeDB,
   "GetListCollections":       GetListCollections,
   "GeneralGraphDbQuery":      GeneralGraphDbQuery,
   "SimilaritySearch":         SimilaritySearch,
   "CreateDbFilter":           CreateDbFilter,

Data Extraction Functions (``dataextraction.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

File processing and content extraction:

.. code-block:: go

   // Example functions in this category:
   "GetGithubFilesToExtract":   GetGithubFilesToExtract,
   "GetLocalFilesToExtract":    GetLocalFilesToExtract,
   "AddDataRequest":            AddDataRequest,
   "CreateCollectionRequest":   CreateCollectionRequest,
   "GetDocumentType":           GetDocumentType,

Cast Functions (``cast.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Type conversion utilities:

.. code-block:: go

   // Example functions in this category:
   "CastAnyToString":      CastAnyToString,
   "CastStringToAny":      CastStringToAny,
   "CastAnyToInt":         CastAnyToInt,
   "CastAnyToFloat64":     CastAnyToFloat64,
   "CastIntToAny":         CastIntToAny,

Authentication Functions (``auth.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

User authentication and authorization:

.. code-block:: go

   // Example functions in this category:
   "CheckApiKeyAuthKvDb":      CheckApiKeyAuthKvDb,
   "CheckApiKeyAuthMongoDb":   CheckApiKeyAuthMongoDb,
   "CheckCreateUserIdMongoDb": CheckCreateUserIdMongoDb,
   "CheckTokenLimitReached":   CheckTokenLimitReached,

MCP Protocol Functions (``mcp.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Model Control Protocol integration:

.. code-block:: go

   // Example functions in this category:
   "ExecuteTool":          ExecuteTool,
   "GetResource":          GetResource,
   "ListAllTools":         ListAllTools,

Qdrant Vector Database Functions (``qdrant.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Qdrant-specific vector operations:

.. code-block:: go

   // Example functions in this category:
   "QdrantCreateCollection": QdrantCreateCollection,
   "QdrantInsertData":       QdrantInsertData,

Ansys Materials Functions (``ansysmaterials.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Materials database and analysis functions:

.. code-block:: go

   // Example functions in this category:
   "SerializeResponse":              SerializeResponse,
   "AddGuidsToAttributes":           AddGuidsToAttributes,
   "FilterOutNonExistingAttributes": FilterOutNonExistingAttributes,
   "ExtractCriteriaSuggestions":     ExtractCriteriaSuggestions,
   "LogRequestSuccess":              LogRequestSuccess,

MCP Protocol Functions (``mcp.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Model Control Protocol integration:

.. code-block:: go

   // Example functions in this category:
   "ListAll":         ListAll,
   "ExecuteTool":     ExecuteTool,
   "GetResource":     GetResource,
   "GetSystemPrompt": GetSystemPrompt,

RHSC Functions (``rhsc.go``)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Red Hat Service Catalog functions:

.. code-block:: go

   // Example functions in this category:
   "SetCopilotGenerateRequestJsonBody": SetCopilotGenerateRequestJsonBody,

Category Implementation Example
-------------------------------

Here's how a complete category file is structured (``generic.go``):

.. code-block:: go

   package externalfunctions

   import (
       "strings"
       "github.com/google/uuid"
   )

   // GenerateUUID generates a new UUID
   func GenerateUUID() string {
       return strings.Replace(uuid.New().String(), "-", "", -1)
   }

   // StringConcat concatenates multiple strings
   func StringConcat(strings ...string) string {
       return strings.Join(strings, "")
   }

Functions in Category
---------------------

All exported functions in the category file are automatically discovered during startup. Functions must:

1. **Be exported** (start with capital letter)
2. **Have proper documentation** with ``@displayName`` tag
3. **Be included** in ``ExternalFunctionsMap`` for runtime access

The discovery process extracts function signatures and documentation from embedded files, while ``ExternalFunctionsMap`` provides the actual function implementations for execution.

Complete Category Implementation
---------------------------------

Here's the complete process for the ``generic`` category:

**1. File embedding in main.go:**

.. code-block:: go

   //go:embed pkg/externalfunctions/generic.go
   var genericFile string

**2. Category mapping:**

.. code-block:: go

   files := map[string]string{
       "generic": genericFile,
       // ... other categories
   }

**3. Function implementations in ExternalFunctionsMap:**

.. code-block:: go

   var ExternalFunctionsMap = map[string]interface{}{
       // Generic functions
       "AssignStringToString":   AssignStringToString,
       "SendRestAPICall":        SendRestAPICall,
       "GenerateUUID":           GenerateUUID,
       "StringConcat":           StringConcat,
       "ExtractJSONStringField": ExtractJSONStringField,
       
       // LLM handler functions
       "PerformVectorEmbeddingRequest": PerformVectorEmbeddingRequest,
       "PerformGeneralRequest":         PerformGeneralRequest,
       "BuildLibraryContext":           BuildLibraryContext,
       
       // ... 180+ total functions across all categories
   }

The system currently includes **185 functions** across **12 categories**.
