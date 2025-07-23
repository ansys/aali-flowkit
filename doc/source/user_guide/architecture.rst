.. _architecture:

How FlowKit Works
=================

FlowKit simplifies connecting your databases, APIs, and services into automated workflows. Here's what you need to know.

Core Concepts
~~~~~~~~~~~~~

**Functions**: Pre-built components that handle specific tasks
    - Database operations (MongoDB for authentication only)
    - API integrations (REST, webhooks)
    - Data transformations
    - AI/ML operations

**Categories**: Functions are organized by what they do
    - **generic**: Common tasks like HTTP requests, file operations
    - **data_extraction**: Process documents and extract data
    - **llm_handler**: AI and language model integrations
    - **qdrant**: Vector database for semantic search
    - See :doc:`categories` for the complete list

**Workflows**: Chain functions together to build automations
    - Connect data sources to processing steps
    - Transform data between services
    - Trigger actions based on conditions

Building Automations
~~~~~~~~~~~~~~~~~~~~

FlowKit handles the complexity so you can focus on your business logic:

1. **Connect** your data sources and services
2. **Process** data using built-in functions
3. **Automate** repetitive tasks and workflows

Complete Workflow Example:

.. code-block:: python

    from aali_client import FlowKitClient
    import os
    import json
    from datetime import datetime

    # Initialize FlowKit client
    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )

    def process_engineering_documentation():
        """
        Complete workflow: extract, process, store, and search engineering docs
        """

        # Configuration
        doc_directory = "/path/to/ansys/documentation"
        collection_name = "ansys_knowledge_base"
        processed_files = []
        errors = []

        try:
            # Step 1: Find all documentation files
            print("Step 1: Discovering documentation files...")

            file_result = client.run_function("GetLocalFilesToExtract", {
                "directory": doc_directory,
                "pattern": "*.pdf,*.docx,*.html",
                "recursive": True,
                "exclude_patterns": ["*draft*", "*temp*"]
            })

            if not file_result["success"]:
                raise Exception(f"Failed to find files: {file_result['error']}")

            files = file_result["files"]
            print(f"Found {len(files)} documentation files")

            # Step 2: Process each file
            for i, file_path in enumerate(files):
                print(f"\nStep 2.{i+1}: Processing {os.path.basename(file_path)}...")

                try:
                    # Extract content
                    content_result = client.run_function("GetLocalFileContent", {
                        "path": file_path,
                        "extract_metadata": True,
                        "preserve_formatting": True
                    })

                    if not content_result["success"]:
                        errors.append(f"Failed to read {file_path}: {content_result['error']}")
                        continue

                    # Split into semantic chunks
                    split_result = client.run_function("LangchainSplitter", {
                        "content": content_result["content"],
                        "chunk_size": 1000,
                        "chunk_overlap": 200,
                        "separator": "\n\n",  # Split on paragraphs
                        "keep_separator": True
                    })

                    # Prepare chunks with metadata
                    chunks = []
                    for j, chunk_text in enumerate(split_result["chunks"]):
                        chunks.append({
                            "id": f"{os.path.basename(file_path)}_chunk_{j}",
                            "content": chunk_text,
                            "metadata": {
                                "source_file": file_path,
                                "chunk_index": j,
                                "total_chunks": len(split_result["chunks"]),
                                "file_type": content_result["metadata"].get("type", "unknown"),
                                "created_date": content_result["metadata"].get("created", ""),
                                "processed_date": datetime.now().isoformat()
                            }
                        })

                    # Store chunks with embeddings
                    store_result = client.run_function("StoreElementsInVectorDatabase", {
                        "elements": chunks,
                        "collection": collection_name,
                        "embeddingField": "content",
                        "metadataFields": ["metadata"],
                        "batch_size": 100  # Process in batches for large files
                    })

                    if store_result["success"]:
                        processed_files.append(file_path)
                        print(f"✓ Stored {store_result['stored_count']} chunks")
                    else:
                        errors.append(f"Failed to store {file_path}: {store_result['error']}")

                except Exception as e:
                    errors.append(f"Error processing {file_path}: {str(e)}")
                    continue

            # Step 3: Test the search functionality
            print("\nStep 3: Testing search functionality...")

            test_queries = [
                "How to set up turbulence models in Fluent?",
                "Contact analysis best practices",
                "Mesh refinement techniques"
            ]

            for query in test_queries:
                search_result = client.run_function("SimilaritySearch", {
                    "query": query,
                    "collection": collection_name,
                    "top_k": 3,
                    "include_metadata": True
                })

                if search_result["success"]:
                    print(f"\nQuery: '{query}'")
                    print(f"Found {len(search_result['matches'])} relevant chunks:")
                    for match in search_result['matches']:
                        print(f"  - Score: {match['score']:.3f} | Source: {match['metadata']['source_file']}")

            # Summary
            print(f"\n{'='*60}")
            print("WORKFLOW COMPLETE")
            print(f"{'='*60}")
            print(f"✓ Processed files: {len(processed_files)}")
            print(f"✗ Errors: {len(errors)}")
            print(f"✓ Collection: {collection_name}")

            if errors:
                print("\nErrors encountered:")
                for error in errors[:5]:  # Show first 5 errors
                    print(f"  - {error}")

            return {
                "success": True,
                "processed": processed_files,
                "errors": errors
            }

        except Exception as e:
            print(f"\nWorkflow failed: {e}")
            return {
                "success": False,
                "error": str(e),
                "processed": processed_files,
                "errors": errors
            }

    # Run the workflow
    if __name__ == "__main__":
        result = process_engineering_documentation()

        # Save results for audit
        with open("workflow_results.json", "w") as f:
            json.dump(result, f, indent=2)

Why FlowKit?
~~~~~~~~~~~~

- **No infrastructure setup**: Focus on your automation logic
- **Pre-built integrations**: Connect popular databases and APIs instantly
- **Scalable**: Handle everything from simple scripts to complex pipelines
- **Developer-friendly**: Use Python, Go, or any gRPC-compatible language

Next Steps
~~~~~~~~~~

- :doc:`connect` - Get connected in minutes
- :doc:`functions` - Explore available functions
- :doc:`categories` - Browse by use case
