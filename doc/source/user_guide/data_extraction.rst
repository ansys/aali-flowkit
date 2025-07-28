.. _data_extraction:

Data Extraction Pipeline
========================

FlowKit's Data Extraction Pipeline provides comprehensive capabilities for extracting, processing, and indexing content from various sources to build searchable knowledge bases.

Overview
--------

The Data Extraction Pipeline enables FlowKit to:

- Extract content from GitHub repositories and local files
- Process multiple document formats (code, documentation, PDFs, etc.)
- Split documents intelligently for optimal indexing
- Generate hierarchical document structures
- Store processed content in vector and graph databases
- Support incremental updates and versioning

Architecture
------------

The extraction pipeline follows a modular architecture:

.. code-block:: text

   Source Discovery → Content Retrieval → Document Processing → Chunking → Storage
         ↓                   ↓                    ↓              ↓          ↓
   GitHub/Local      Download/Read         Type Detection    Smart Split  Vector/Graph DB

Core Functions
--------------

Source Discovery
~~~~~~~~~~~~~~~~

.. list-table:: Discovery Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``GetGithubFilesToExtract``
     - Discover files in GitHub repositories with filtering
   * - ``GetLocalFilesToExtract``
     - Scan local directories for extractable content
   * - ``AppendStringSlices``
     - Combine multiple file lists for batch processing

Content Retrieval
~~~~~~~~~~~~~~~~~

.. list-table:: Retrieval Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``DownloadGithubFileContent``
     - Fetch content from GitHub with authentication
   * - ``GetLocalFileContent``
     - Read local files with encoding detection
   * - ``GetDocumentType``
     - Identify document format and structure

Document Processing
~~~~~~~~~~~~~~~~~~~

.. list-table:: Processing Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``LangchainSplitter``
     - Intelligent document chunking with overlap
   * - ``GenerateDocumentTree``
     - Create hierarchical document structure
   * - ``CreateGeneralDataExtractionDocumentObjects``
     - Generate standardized document objects

Storage Operations
~~~~~~~~~~~~~~~~~~

.. list-table:: Storage Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``AddDataRequest``
     - Add processed documents to storage
   * - ``CreateCollectionRequest``
     - Create new document collections
   * - Integration functions
     - Store in vector/graph databases

Configuration
-------------

Configure the extraction pipeline through FlowKit's settings:

.. code-block:: yaml

   # Data Extraction Configuration
   EXTRACTION_BATCH_SIZE: 100
   CHUNK_SIZE: 1000
   CHUNK_OVERLAP: 200
   
   # GitHub Settings
   GITHUB_TOKEN: "your-github-token"
   GITHUB_API_TIMEOUT: 30
   
   # Document Processing
   SUPPORTED_EXTENSIONS: [".py", ".go", ".js", ".md", ".rst", ".pdf"]
   MAX_FILE_SIZE_MB: 50
   ENCODING_DETECTION: true

Usage Examples
--------------

**Example 1: Extract GitHub Repository**

.. code-block:: json

   {
     "name": "GetGithubFilesToExtract",
     "inputs": {
       "owner": "ansys",
       "repo": "pyansys",
       "branch": "main",
       "path": "src",
       "extensions": [".py", ".md"],
       "excludePaths": ["tests", "__pycache__"]
     }
   }

**Example 2: Process Local Documentation**

.. code-block:: json

   {
     "name": "GetLocalFilesToExtract",
     "inputs": {
       "rootPath": "/docs/user_guides",
       "patterns": ["*.rst", "*.md"],
       "recursive": true,
       "ignorePatterns": ["_build/*", "*.tmp"]
     }
   }

**Example 3: Smart Document Splitting**

.. code-block:: json

   {
     "name": "LangchainSplitter",
     "inputs": {
       "content": "Long technical document content...",
       "chunkSize": 1000,
       "chunkOverlap": 200,
       "separators": ["\n\n", "\n", ". ", " "],
       "keepSeparator": true
     }
   }

Document Types
--------------

The pipeline automatically handles various document types:

**Code Files**:
   - Language-aware splitting
   - Syntax preservation
   - Comment extraction
   - Function/class detection

**Documentation**:
   - Markdown/RST parsing
   - Section preservation
   - Code block handling
   - Cross-reference tracking

**PDFs**:
   - Text extraction
   - Layout analysis
   - Image handling
   - Metadata preservation

Chunking Strategies
-------------------

Intelligent chunking ensures optimal retrieval:

1. **Semantic Chunking**: Preserve meaning and context
2. **Size Optimization**: Balance between completeness and retrieval
3. **Overlap Management**: Maintain context across chunks
4. **Hierarchy Preservation**: Keep document structure

.. code-block:: python

   # Example chunk metadata
   {
     "chunk_id": "doc_001_chunk_003",
     "document_id": "user_guide_mesh_generation",
     "position": 3,
     "total_chunks": 15,
     "overlap_previous": 200,
     "overlap_next": 200,
     "semantic_tags": ["meshing", "boundary_conditions"]
   }

Document Tree Structure
-----------------------

The pipeline generates hierarchical document structures:

.. code-block:: text

   Repository
   ├── Documentation
   │   ├── User Guides
   │   │   ├── Getting Started
   │   │   └── Advanced Topics
   │   └── API Reference
   │       ├── Core Functions
   │       └── Utilities
   └── Source Code
       ├── Core Modules
       └── Examples

Collection Management
---------------------

Organize extracted content into collections:

**Creating Collections**:

.. code-block:: json

   {
     "name": "CreateCollectionRequest",
     "inputs": {
       "collectionName": "ansys_fluent_docs",
       "description": "Fluent user documentation",
       "metadata": {
         "version": "2024R1",
         "language": "en",
         "type": "user_guide"
       }
     }
   }

**Adding Documents**:

.. code-block:: json

   {
     "name": "AddDataRequest",
     "inputs": {
       "collection": "ansys_fluent_docs",
       "documents": [...],
       "embeddings": [...],
       "updateMode": "merge"
     }
   }

Extraction Workflows
--------------------

**Complete Repository Extraction**:

1. Discover files in repository
2. Filter by type and path
3. Download content in batches
4. Process and chunk documents
5. Generate embeddings
6. Store in databases

**Incremental Updates**:

1. Track last extraction timestamp
2. Identify changed files
3. Re-process modified content
4. Update affected chunks
5. Maintain version history

Best Practices
--------------

1. **Source Selection**:
   - Choose relevant file types
   - Exclude generated/temporary files
   - Consider file size limits

2. **Chunking Configuration**:
   - Adjust size based on content type
   - Use appropriate overlap for context
   - Preserve semantic boundaries

3. **Metadata Management**:
   - Tag documents comprehensively
   - Track source information
   - Maintain update timestamps

4. **Performance Optimization**:
   - Process in parallel batches
   - Cache frequently accessed content
   - Use incremental updates

Integration with Other Components
---------------------------------

Data Extraction integrates with:

- **Knowledge DB**: Store processed documents
- **LLM Handler**: Generate embeddings
- **Vector Search**: Enable similarity queries
- **Graph Database**: Store relationships

Error Handling
--------------

The pipeline includes robust error handling:

**File Access Errors**:
   - Retry with backoff
   - Skip and log inaccessible files
   - Continue with remaining files

**Processing Errors**:
   - Fallback to simple splitting
   - Log problematic content
   - Mark documents for review

**Storage Errors**:
   - Queue for retry
   - Use local cache
   - Alert on persistent failures

Monitoring and Metrics
----------------------

Track extraction pipeline performance:

.. code-block:: text

   Extraction Summary:
   - Files Discovered: 1,250
   - Files Processed: 1,245
   - Files Skipped: 5
   - Total Chunks: 15,420
   - Processing Time: 12m 35s
   - Average Chunk Size: 850 tokens
   - Storage Used: 125 MB

Advanced Features
-----------------

**Custom Extractors**:
   - Plugin architecture for new formats
   - Domain-specific processing
   - Custom chunking strategies

**Quality Validation**:
   - Chunk coherence scoring
   - Duplicate detection
   - Content verification

**Relationship Extraction**:
   - Cross-document references
   - Dependency mapping
   - Knowledge graph building