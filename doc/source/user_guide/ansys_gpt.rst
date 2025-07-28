.. _ansys_gpt:

Ansys GPT Integration
=====================

Ansys GPT is FlowKit's specialized integration for engineering-domain AI capabilities, providing intelligent responses tailored for Ansys product users and engineering workflows.

Overview
--------

The Ansys GPT integration transforms general LLM capabilities into engineering-specific intelligence by:

- Integrating with Azure Cognitive Search for technical documentation
- Filtering prohibited terms and ensuring appropriate responses
- Extracting engineering parameters from natural language queries
- Building context from Ansys documentation and knowledge bases
- Providing citations and references for answers

Core Components
---------------

Query Processing Pipeline
~~~~~~~~~~~~~~~~~~~~~~~~~

Ansys GPT processes user queries through a sophisticated pipeline:

.. code-block:: text

   User Query → Field Extraction → Rephrasing → Search → Context Building → LLM Response
                       ↓                           ↓            ↓
                 Extract Physics,          Azure Cognitive   Filter & Rank
                 Products, Version             Search         Results

Key Functions
-------------

Query Understanding
~~~~~~~~~~~~~~~~~~~

.. list-table:: Query Processing Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``AnsysGPTExtractFieldsFromQuery``
     - Extract physics types, products, and versions from natural language
   * - ``AnsysGPTPerformLLMRephraseRequest``
     - Rephrase queries for better search results
   * - ``AnsysGPTCheckProhibitedWords``
     - Ensure queries don't contain inappropriate terms
   * - ``AnsysGPTGetSystemPrompt``
     - Generate context-aware system prompts

Search and Retrieval
~~~~~~~~~~~~~~~~~~~~

.. list-table:: Search Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``AnsysGPTACSSemanticHybridSearchs``
     - Perform hybrid search across documentation
   * - ``AnsysGPTReturnIndexList``
     - Get relevant search indices based on query
   * - ``AnsysGPTReorderSearchResponseAndReturnOnlyTopK``
     - Rank and filter search results by relevance
   * - ``AnsysGPTRemoveNoneCitationsFromSearchResponse``
     - Clean up results without valid citations

Response Generation
~~~~~~~~~~~~~~~~~~~

.. list-table:: Response Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``AnsysGPTBuildFinalQuery``
     - Construct the final prompt with context
   * - ``AnsysGPTPerformLLMRequest``
     - Generate the AI response with streaming support
   * - ``AnsysGPTExtractCitations``
     - Extract and format citations from responses
   * - ``AnsysGPTReturnFinalMessage``
     - Format the complete response with references

Configuration
-------------

Ansys GPT requires specific configuration for Azure Cognitive Search integration:

.. code-block:: yaml

   # Azure Cognitive Search Configuration
   ACS_ENDPOINT: "https://your-search.search.windows.net"
   ACS_API_KEY: "your-api-key"
   ACS_API_VERSION: "2021-04-30-Preview"
   
   # Ansys GPT Specific Settings
   ANSYS_GPT_PHYSICS_TYPES: ["CFD", "Structural", "Electromagnetic"]
   ANSYS_GPT_PRODUCTS: ["Fluent", "Mechanical", "HFSS"]
   ANSYS_GPT_DEFAULT_VERSION: "2024R1"

Field Extraction
----------------

Ansys GPT automatically extracts structured information from queries:

**Example Query**: "How do you set up boundary conditions in Fluent for CFD analysis?"

**Extracted Fields**:

.. code-block:: json

   {
     "physics": "CFD",
     "product": "Fluent",
     "topic": "boundary conditions",
     "version": "latest"
   }

Search Integration
------------------

The integration uses Azure Cognitive Search for retrieving relevant documentation:

**Search Process**:

1. **Index Selection**: Choose appropriate indices based on extracted fields
2. **Hybrid Search**: Combine semantic and keyword search
3. **Result Filtering**: Apply physics and product filters
4. **Ranking**: Score results based on relevance
5. **Context Building**: Construct context from top results

Usage Examples
--------------

**Example 1: Engineering Query with Context**

.. code-block:: json

   {
     "name": "AnsysGPTPerformLLMRequest",
     "inputs": {
       "query": "How to perform modal analysis in Mechanical?",
       "history": [],
       "physics": ["Structural"],
       "products": ["Mechanical"],
       "includeContext": true
     }
   }

**Example 2: Extract Fields from Query**

.. code-block:: json

   {
     "name": "AnsysGPTExtractFieldsFromQuery",
     "inputs": {
       "query": "CFD simulation setup in Fluent 2024R1",
       "fieldValues": {
         "physics": ["CFD", "Thermal", "Structural"],
         "products": ["Fluent", "CFX", "Mechanical"],
         "versions": ["2024R1", "2023R2", "2023R1"]
       }
     }
   }

**Example 3: Semantic Search with Filters**

.. code-block:: json

   {
     "name": "AnsysGPTACSSemanticHybridSearchs",
     "inputs": {
       "query": "turbulence modeling best practices",
       "embeddedQuery": [0.1, 0.2, ...],
       "indexList": ["fluent-docs", "cfd-theory"],
       "filter": {
         "physics": "CFD",
         "product": "Fluent"
       },
       "topK": 10
     }
   }

Prohibited Words Handling
-------------------------

Ansys GPT includes sophisticated filtering to ensure appropriate responses:

- **Pre-query Filtering**: Check queries before processing
- **Custom Error Messages**: Provide helpful alternatives
- **Logging**: Track filtered queries for improvement

Citation Management
-------------------

All Ansys GPT responses include proper citations:

.. code-block:: text

   Response: "To set up boundary conditions in Fluent, navigate to..."
   
   References:
   [1] Fluent User Guide - Chapter 7: Boundary Conditions
   [2] CFD Best Practices Guide - Section 3.2
   [3] Tutorial: Basic Flow Setup - Step 4

Best Practices
--------------

1. **Query Formulation**:
   - Include product names when known
   - Specify version if relevant
   - Use engineering terminology

2. **Context Management**:
   - Provide conversation history for follow-ups
   - Include relevant project details
   - Specify physics type explicitly

3. **Search Optimization**:
   - Use specific technical terms
   - Include relevant keywords
   - Leverage filters effectively

4. **Response Quality**:
   - Verify citations are relevant
   - Check technical accuracy
   - Ensure completeness

Integration with Other Components
---------------------------------

Ansys GPT works seamlessly with:

- **LLM Handler**: For AI response generation
- **Knowledge DB**: For custom documentation storage
- **Azure Services**: For search and retrieval
- **Auth System**: For user access control

Troubleshooting
---------------

Common issues and solutions:

**No Search Results Found**
   - Verify index configuration
   - Check search service connectivity
   - Broaden search terms

**Incorrect Field Extraction**
   - Review field value mappings
   - Update default field configurations
   - Provide more context in queries

**Missing Citations**
   - Ensure search results have metadata
   - Verify citation extraction logic
   - Check document processing pipeline

**Performance Issues**
   - Optimize search index size
   - Use appropriate top-K values
   - Enable result caching