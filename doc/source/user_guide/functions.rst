.. _functions:

Working with Functions
======================

FlowKit provides numerous ready-to-use functions that connect your databases, call APIs, process data, and more.

Function Categories
~~~~~~~~~~~~~~~~~~~

**data_extraction**
  - ``GetLocalFilesToExtract`` - Find files by pattern
  - ``GetLocalFileContent`` - Read file contents
  - ``LangchainSplitter`` - Split documents into chunks

**generic**
  - ``SendRestAPICall`` - Make HTTP requests
  - ``GenerateUUID`` - Create unique identifiers

**cast**
  - ``CastAnyToString`` - Convert any type to string
  - ``CastAnyToFloat64`` - Convert any type to float64
  - ``CastAnyToInt`` - Convert any type to integer

**llm_handler**
  - ``PerformVectorEmbeddingRequest`` - Generate embeddings
  - ``PerformBatchEmbeddingRequest`` - Batch embeddings
  - ``PerformGeneralRequest`` - LLM completions

**ansys_gpt**
  - ``AnsysGPTCheckProhibitedWords`` - Content validation
  - ``AnsysGPTPerformLLMRequest`` - Ansys-specific LLM

**knowledge_db**
  - ``SendVectorsToKnowledgeDB`` - Store vectors with relationships
  - ``SimilaritySearch`` - Vector and graph similarity search

**qdrant**
  - ``StoreElementsInVectorDatabase`` - Store vectors
  - ``QdrantCreateCollection`` - Create collections
  - ``QdrantInsertData`` - Insert vectors

**Additional categories**: ansys_mesh_pilot, ansys_materials, auth, mcp, rhsc

See :doc:`categories` for detailed descriptions of all categories.

Discovering Available Functions
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

See what functions you can use:

.. code-block:: python

    from aali_client import FlowKitClient

    # Connect to FlowKit
    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )

    try:
        # Get all available functions
        functions = client.list_functions()

        # Organize by category for easier browsing
        by_category = {}
        for name, info in functions.items():
            category = info.category
            if category not in by_category:
                by_category[category] = []
            by_category[category].append({
                'name': name,
                'display': info.display_name,
                'description': info.description
            })

        # Display functions by category
        for category, funcs in sorted(by_category.items()):
            print(f"\n{category.upper()} Functions:")
            for func in funcs:
                print(f"  - {func['name']}: {func['display']}")

    except Exception as e:
        print(f"Failed to list functions: {e}")

Understanding Functions
~~~~~~~~~~~~~~~~~~~~~~~

Each function has:
- **A name** you use in code: ``SendRestAPICall``
- **A friendly name** for documentation: "REST Call"
- **Parameters** that tell it what to do
- **Return values** with your results

Calling Functions
~~~~~~~~~~~~~~~~~

.. code-block:: python

    from aali_client import FlowKitClient
    import time

    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )

    # Example: process and store technical documents
    documents = [
        {
            "id": f"ansys_doc_{i}",
            "title": f"Ansys Fluent Tutorial {i}",
            "content": "Step-by-step guide for setting up turbulent flow simulation...",
            "author": "Ansys Documentation Team",
            "created": time.time(),
            "tags": ["fluent", "cfd", "tutorial"]
        }
        for i in range(1, 4)
    ]

    try:
        # Store documents with automatic embedding generation
        result = client.run_function(
            "StoreElementsInVectorDatabase",
            {
                "elements": documents,
                "collection": "ansys_tutorials",
                "embeddingField": "content",  # Generate embeddings from this field
                "metadataFields": ["title", "author", "tags"]  # Store as metadata
            }
        )

        # Check results
        if result["success"]:
            print(f"✓ Stored {result['stored_count']} documents")
            print(f"✓ Collection: {result['collection']}")
            print(f"✓ Embedding dimension: {result['vector_dimension']}")
        else:
            print(f"✗ Storage failed: {result['error']}")

    except Exception as e:
        print(f"Error calling function: {e}")

Common Use Cases
~~~~~~~~~~~~~~~~

**API Integration**
    - ``SendRestAPICall`` - Call any REST API
    - ``GenerateUUID`` - Create unique identifiers

**Data Processing**
    - ``GetLocalFilesToExtract`` - Find files to process
    - ``LangchainSplitter`` - Split documents intelligently
    - ``StoreElementsInVectorDatabase`` - Enable semantic search

**AI/ML Operations**
    - ``PerformVectorEmbeddingRequest`` - Convert text to vectors
    - ``PerformGeneralRequest`` - Get AI completions
    - ``SimilaritySearch`` - Find similar content

Example: similarity search
~~~~~~~~~~~~~~~~~~~~~~~~~~

Find similar content using vector search:

.. code-block:: python

    from aali_client import FlowKitClient

    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )

    # Example: find similar engineering problems and solutions
    try:
        # Search for similar content
        results = client.run_function("SimilaritySearch", {
            "query": "How to model turbulent flow in a pipe with heat transfer?",
            "collection": "engineering_docs",
            "top_k": 5,  # Return top 5 most similar documents
            "score_threshold": 0.7,  # Minimum similarity score (0-1)
            "include_metadata": True,  # Include document metadata
            "filters": {
                "category": ["CFD", "Heat Transfer"],  # Filter by categories
                "tags": {"$contains": "turbulence"}  # Must contain this tag
            }
        })

        # Process search results
        if results["success"]:
            print(f"Found {len(results['matches'])} similar documents:\n")

            for i, match in enumerate(results['matches'], 1):
                print(f"{i}. {match['metadata']['title']}")
                print(f"   Score: {match['score']:.3f}")
                print(f"   Category: {match['metadata']['category']}")
                print(f"   Preview: {match['content'][:100]}...")
                print(f"   ID: {match['id']}\n")
        else:
            print(f"Search failed: {results['error']}")

    except Exception as e:
        print(f"Error performing search: {e}")

Example: generate embeddings
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Convert text to vectors for AI operations:

.. code-block:: python

    from aali_client import FlowKitClient

    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )

    # Example: generate embeddings for technical queries
    queries = [
        "stress analysis of composite materials",
        "thermal simulation in electronic components",
        "fluid-structure interaction in turbomachinery"
    ]

    try:
        # Generate embeddings for multiple texts
        result = client.run_function("PerformVectorEmbeddingRequest", {
            "texts": queries,
            "model": "ansys-engineering-v1",  # Engineering-specific model
            "normalize": True,  # Normalize vectors for cosine similarity
            "return_tokens": True  # Include token count for billing
        })

        if result["success"]:
            embeddings = result["embeddings"]
            print(f"Generated {len(embeddings)} embeddings")
            print(f"Embedding dimension: {len(embeddings[0])}")
            print(f"Total tokens used: {result['total_tokens']}")

            # Example: calculate similarity between first two queries
            import numpy as np

            vec1 = np.array(embeddings[0])
            vec2 = np.array(embeddings[1])
            similarity = np.dot(vec1, vec2)  # Cosine similarity (normalized)

            print(f"\nSimilarity between:")
            print(f"  '{queries[0]}'")
            print(f"  '{queries[1]}'")
            print(f"  Score: {similarity:.3f}")

        else:
            print(f"Embedding generation failed: {result['error']}")

    except Exception as e:
        print(f"Error generating embeddings: {e}")

Next Steps
~~~~~~~~~~

- :doc:`categories` - Browse functions by category
- :doc:`types` - Understand data handling
