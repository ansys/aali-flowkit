.. _knowledge_db:

Knowledge Database Integration
==============================

FlowKit's Knowledge Database integration provides powerful vector and graph database capabilities for storing, searching, and retrieving structured knowledge for AI-powered applications.

Overview
--------

The Knowledge Database integration enables:

- Vector storage for semantic similarity search
- Graph database for relationship mapping
- Hybrid search combining vector and metadata
- Collection management for organized storage
- Real-time indexing and retrieval
- Support for multiple database backends

Architecture
------------

The Knowledge DB follows a dual-storage architecture:

.. code-block:: text

   FlowKit Application
         ├── Vector Database (Embeddings & Similarity)
         │     └── Collections → Documents → Chunks
         └── Graph Database (Relationships & Structure)
               └── Nodes → Edges → Properties

Core Functions
--------------

Vector Operations
~~~~~~~~~~~~~~~~~

.. list-table:: Vector Database Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``SendVectorsToKnowledgeDB``
     - Store embeddings with metadata in vector database
   * - ``SimilaritySearch``
     - Find similar documents using vector similarity
   * - ``GetListCollections``
     - List all available vector collections
   * - ``GeneralQuery``
     - Execute general queries across collections

Graph Operations
~~~~~~~~~~~~~~~~

.. list-table:: Graph Database Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``GeneralGraphDbQuery``
     - Execute Cypher queries on graph database
   * - ``AddGraphDbParameter``
     - Add parameters to graph queries safely
   * - ``RetrieveDependencies``
     - Find connected nodes and relationships

Filtering and Search
~~~~~~~~~~~~~~~~~~~~

.. list-table:: Filter Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``CreateKeywordsDbFilter``
     - Filter by keyword matches
   * - ``CreateTagsDbFilter``
     - Filter by document tags
   * - ``CreateMetadataDbFilter``
     - Filter by custom metadata fields
   * - ``CreateDbFilter``
     - Combine multiple filter criteria

Configuration
-------------

Configure Knowledge DB connections and settings:

.. code-block:: yaml

   # Vector Database Configuration
   VECTOR_DB_TYPE: "qdrant"
   VECTOR_DB_HOST: "qdrant-service"
   VECTOR_DB_PORT: 6333
   VECTOR_DB_COLLECTION_PREFIX: "flowkit_"
   
   # Graph Database Configuration
   GRAPH_DB_TYPE: "neo4j"
   GRAPH_DB_URI: "bolt://neo4j:7687"
   GRAPH_DB_USER: "neo4j"
   GRAPH_DB_PASSWORD: "your-password"
   
   # Search Settings
   DEFAULT_SEARCH_LIMIT: 10
   SIMILARITY_THRESHOLD: 0.7
   ENABLE_HYBRID_SEARCH: true

Usage Examples
--------------

**Example 1: Store Document Embeddings**

.. code-block:: json

   {
     "name": "SendVectorsToKnowledgeDB",
     "inputs": {
       "collection": "technical_docs",
       "documents": [
         {
           "id": "doc_001",
           "content": "Finite element analysis guide",
           "embedding": [0.1, 0.2, ...],
           "metadata": {
             "type": "user_guide",
             "product": "Mechanical",
             "version": "2024R1"
           }
         }
       ]
     }
   }

**Example 2: Similarity Search with Filters**

.. code-block:: json

   {
     "name": "SimilaritySearch",
     "inputs": {
       "collection": "technical_docs",
       "queryEmbedding": [0.15, 0.25, ...],
       "limit": 5,
       "filters": {
         "product": "Mechanical",
         "type": "user_guide"
       },
       "scoreThreshold": 0.8
     }
   }

**Example 3: Graph Query for Dependencies**

.. code-block:: json

   {
     "name": "GeneralGraphDbQuery",
     "inputs": {
       "query": "MATCH (d:Document)-[:REFERENCES]->(r:Document) WHERE d.id = $docId RETURN r",
       "parameters": {
         "docId": "mesh_guide_001"
       }
     }
   }

Collection Management
---------------------

Organize knowledge into logical collections:

**Collection Structure**:

.. code-block:: text

   Knowledge Base
   ├── technical_docs
   │   ├── user_guides
   │   ├── api_references
   │   └── tutorials
   ├── code_examples
   │   ├── python
   │   ├── ansys_scripts
   │   └── workflows
   └── support_articles
       ├── troubleshooting
       └── best_practices

**Collection Operations**:

- Create collections with schema
- Configure indexing strategies
- Set retention policies
- Manage access permissions

Vector Search Capabilities
--------------------------

**Semantic Search**:
   Find documents based on meaning rather than keywords

**Hybrid Search**:
   Combine vector similarity with metadata filtering

**Multi-Vector Search**:
   Search using multiple query vectors simultaneously

**Range Queries**:
   Find documents within similarity thresholds

Advanced Filtering
------------------

Create complex filters for precise retrieval:

**Keyword Filter Example**:

.. code-block:: json

   {
     "name": "CreateKeywordsDbFilter",
     "inputs": {
       "keywords": ["mesh", "boundary conditions", "CFD"],
       "operator": "OR",
       "field": "content"
     }
   }

**Metadata Filter Example**:

.. code-block:: json

   {
     "name": "CreateMetadataDbFilter",
     "inputs": {
       "filters": [
         {"field": "product", "value": "Fluent", "operator": "eq"},
         {"field": "version", "value": "2023R1", "operator": "gte"},
         {"field": "language", "value": ["en", "de"], "operator": "in"}
       ],
       "combineOperator": "AND"
     }
   }

Graph Database Features
-----------------------

**Relationship Modeling**:

.. code-block:: cypher

   // Example graph structure
   (Document)-[:REFERENCES]->(Document)
   (Document)-[:PART_OF]->(Collection)
   (Document)-[:AUTHORED_BY]->(User)
   (Document)-[:TAGGED_WITH]->(Tag)

**Dependency Tracking**:

.. code-block:: json

   {
     "name": "RetrieveDependencies",
     "inputs": {
       "nodeId": "workflow_123",
       "relationshipTypes": ["DEPENDS_ON", "IMPORTS", "REFERENCES"],
       "maxDepth": 3
     }
   }

Best Practices
--------------

1. **Data Organization**:
   - Use meaningful collection names
   - Apply consistent metadata schemas
   - Document collection purposes

2. **Embedding Quality**:
   - Use appropriate embedding models
   - Normalize vectors if needed
   - Consider dimensionality

3. **Search Optimization**:
   - Index frequently queried fields
   - Use filters to reduce search space
   - Cache common queries

4. **Graph Design**:
   - Keep relationships simple
   - Use meaningful relationship types
   - Avoid deep nesting

Performance Optimization
------------------------

**Indexing Strategies**:
   - Create indexes on filtered fields
   - Use appropriate vector index types
   - Balance speed vs accuracy

**Batch Operations**:
   - Insert documents in batches
   - Use bulk update operations
   - Parallelize when possible

**Query Optimization**:
   - Limit result sets appropriately
   - Use early filtering
   - Avoid full collection scans

Integration Examples
--------------------

**RAG (Retrieval Augmented Generation)**:

.. code-block:: python

   # 1. Generate query embedding
   embedding = perform_embedding(user_query)
   
   # 2. Search similar documents
   results = similarity_search(
       collection="knowledge_base",
       embedding=embedding,
       limit=5
   )
   
   # 3. Build context from results
   context = build_context(results)
   
   # 4. Generate response with LLM
   response = generate_with_context(query, context)

**Knowledge Graph Navigation**:

.. code-block:: python

   # 1. Find starting node
   start = find_document(title="Getting Started")
   
   # 2. Explore related documents
   related = graph_query(
       "MATCH (d:Document {id: $id})-[:RELATED_TO]-(r:Document) "
       "RETURN r ORDER BY r.relevance DESC LIMIT 10",
       {"id": start.id}
   )

Monitoring and Maintenance
--------------------------

**Collection Statistics**:

.. code-block:: text

   Collection: technical_docs
   - Total Documents: 15,420
   - Total Vectors: 15,420
   - Average Vector Dimension: 1536
   - Index Type: HNSW
   - Index Build Time: 2m 15s
   - Average Query Time: 15ms

**Health Checks**:
   - Monitor query performance
   - Track storage usage
   - Check index statistics
   - Validate data integrity

Troubleshooting
---------------

**Slow Queries**:
   - Check index configuration
   - Reduce result limit
   - Optimize filter usage
   - Consider sharding

**Poor Search Results**:
   - Verify embedding quality
   - Check similarity threshold
   - Review metadata accuracy
   - Validate query formation

**Storage Issues**:
   - Monitor disk usage
   - Implement retention policies
   - Archive old data
   - Optimize vector dimensions