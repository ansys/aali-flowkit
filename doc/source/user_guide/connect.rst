.. _connect:

Connecting to FlowKit
=====================

FlowKit helps you connect and automate your databases, APIs, and services. Here's how to get started.

Quick Start
~~~~~~~~~~~

Connect to FlowKit in seconds:

1. Install the client library
2. Get your API key from your administrator  
3. Start automating

Python Client
~~~~~~~~~~~~~

.. code-block:: python

    from aali_client import FlowKitClient
    
    # Connect to FlowKit
    client = FlowKitClient(api_key="your-api-key")
    
    # See what functions are available
    functions = client.list_functions()
    print(f"Found {len(functions)} functions to work with")
    
    # Example: generate a unique ID
    result = client.run_function("GenerateUUID", {})
    print(f"Generated ID: {result}")

Common Connection Scenarios  
~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Connect to External APIs**

.. code-block:: python

    from aali_client import FlowKitClient
    
    # Initialize client with your credentials
    client = FlowKitClient(
        address="localhost:50051",  # or your FlowKit server address
        api_key="your-api-key"
    )
    
    # Example: fetch engineering simulation results from Ansys Cloud
    try:
        response = client.run_function("SendRestAPICall", {
            "url": "https://api.ansys.com/v1/simulations/12345/results",
            "method": "GET",
            "headers": {
                "Authorization": f"Bearer {your_ansys_token}",
                "Accept": "application/json"
            }
        })
        
        # Handle the response
        if response["status_code"] == 200:
            results = response["body"]
            print(f"Simulation completed: {results['status']}")
            print(f"Max stress: {results['max_stress_mpa']} MPa")
        else:
            print(f"Error: {response['status_code']} - {response['body']}")
            
    except Exception as e:
        print(f"Failed to fetch results: {e}")

**Work with Databases**

.. code-block:: python

    from aali_client import FlowKitClient
    import json
    
    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )
    
    # Example: store engineering documentation for semantic search
    documents = [
        {
            "id": "doc_001",
            "title": "Fluent User Guide - Turbulence Models",
            "content": "The k-epsilon model is suitable for fully turbulent flows...",
            "category": "CFD",
            "tags": ["turbulence", "k-epsilon", "fluent"]
        },
        {
            "id": "doc_002", 
            "title": "Mechanical APDL - Contact Analysis",
            "content": "Contact elements allow modeling of two surfaces coming into contact...",
            "category": "FEA",
            "tags": ["contact", "nonlinear", "mechanical"]
        }
    ]
    
    try:
        # Store documents with automatic embedding generation
        result = client.run_function("StoreElementsInVectorDatabase", {
            "elements": documents,
            "collection": "engineering_docs",
            "embeddingField": "content",  # Field to generate embeddings from
            "metadataFields": ["title", "category", "tags"]  # Additional searchable fields
        })
        
        print(f"Successfully stored {result['stored_count']} documents")
        print(f"Collection: {result['collection']}")
        
    except Exception as e:
        print(f"Failed to store documents: {e}")

Configuration Options
~~~~~~~~~~~~~~~~~~~~~

**Custom Server Address**

.. code-block:: python

    # Connect to remote FlowKit server
    client = FlowKitClient(
        address="flowkit.yourcompany.com:50051",
        api_key="your-api-key"
    )

**Secure Connections**

.. code-block:: python

    # Enable TLS for production
    client = FlowKitClient(
        address="flowkit.yourcompany.com:50051",
        api_key="your-api-key",
        use_tls=True
    )

**Server Configuration**

For complete FlowKit server configuration including advanced workflow settings, see :doc:`../getting_started/configuration`.

Error Handling
~~~~~~~~~~~~~~

FlowKit provides consistent error responses to help you handle failures gracefully in your applications.

Error Response Format
^^^^^^^^^^^^^^^^^^^^^

All error responses from FlowKit follow a consistent structure:

.. code-block:: go

   type ErrorResponse struct {
       Error   string `json:"error"`    // Human-readable error message
       Code    string `json:"code"`     // Machine-readable error code
       Details string `json:"details"`  // Optional additional context
   }

Example error responses:

.. code-block:: json

   {
       "error": "Failed to execute workflow",
       "code": "WORKFLOW_EXECUTION_ERROR",
       "details": "Node 'data-extraction' timeout after 30s"
   }

.. code-block:: json

   {
       "error": "Invalid configuration",
       "code": "CONFIG_VALIDATION_ERROR",
       "details": "Missing required field: WORKFLOW_ENGINE_ADDRESS"
   }

Common Error Codes
^^^^^^^^^^^^^^^^^^

FlowKit uses standardized error codes for programmatic error handling:

- ``WORKFLOW_EXECUTION_ERROR`` - Workflow failed during execution
- ``CONFIG_VALIDATION_ERROR`` - Configuration is invalid
- ``NODE_TIMEOUT_ERROR`` - A node exceeded its timeout
- ``FUNCTION_NOT_FOUND`` - Referenced function doesn't exist
- ``AUTH_ERROR`` - Authentication/authorization failure
- ``CONNECTION_ERROR`` - Failed to connect to external service
- ``INVALID_INPUT`` - Function parameters are invalid
- ``RESOURCE_UNAVAILABLE`` - Required resource is not available

Handling Errors in Code
^^^^^^^^^^^^^^^^^^^^^^^^

Implement robust error handling in your FlowKit applications:

.. code-block:: python

   from aali_client import FlowKitClient
   import json
   
   client = FlowKitClient(
       address="localhost:50051",
       api_key="your-api-key"
   )
   
   try:
       result = client.run_function("SendRestAPICall", {
           "url": "https://api.example.com/data",
           "method": "GET"
       })
       
       # Handle successful response
       if result.get("status_code") == 200:
           data = result["body"]
           print(f"Success: {data}")
       else:
           # Handle HTTP errors
           print(f"HTTP Error {result['status_code']}: {result['body']}")
           
   except Exception as e:
       # Parse FlowKit error response
       try:
           error_data = json.loads(str(e))
           error_code = error_data.get("code", "UNKNOWN_ERROR")
           error_msg = error_data.get("error", str(e))
           error_details = error_data.get("details", "")
           
           print(f"FlowKit Error [{error_code}]: {error_msg}")
           if error_details:
               print(f"Details: {error_details}")
               
           # Handle specific error types
           if error_code == "FUNCTION_NOT_FOUND":
               print("Try client.list_functions() to see available functions")
           elif error_code == "CONFIG_VALIDATION_ERROR":
               print("Check your configuration settings")
           elif error_code == "AUTH_ERROR":
               print("Verify your API key is correct")
               
       except (json.JSONDecodeError, AttributeError):
           # Fallback for non-FlowKit errors
           print(f"Unexpected error: {e}")

Next Steps
~~~~~~~~~~

- :doc:`functions` - Learn about available functions
- :doc:`categories` - Explore function categories  
- Build your first automation workflow