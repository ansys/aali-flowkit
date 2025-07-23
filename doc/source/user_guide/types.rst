.. _types:

Working with Data
=================

FlowKit handles data types automatically, so you can focus on building your workflows.

Common Data Types
~~~~~~~~~~~~~~~~~

**Text and Numbers**
    - Strings for text, URLs, file paths
    - Numbers for counts, measurements, IDs  
    - Booleans for true/false conditions

**Collections**
    - Lists for multiple items (documents, results, IDs)
    - Dictionaries for structured data (API responses, configurations)
    - Binary data for files and images

**Special Types**
    - Search results from databases
    - Chat conversations for AI interactions
    - Generated code blocks
    - Material properties for engineering data

Type Conversions Made Easy
~~~~~~~~~~~~~~~~~~~~~~~~~~

FlowKit provides automatic conversions when needed:

.. code-block:: python

    from aali_client import FlowKitClient
    
    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )
    
    # Example: convert mixed data types for processing
    simulation_data = {
        "temperature": 273.15,
        "pressure": 101325,
        "convergence": True,
        "iterations": 1000,
        "residuals": [1e-3, 1e-4, 1e-5]
    }
    
    try:
        # Convert complex data to string for logging
        text_result = client.run_function("CastAnyToString", {
            "input": simulation_data
        })
        print(f"Simulation summary: {text_result['output']}")
        
        # Convert string input to number for calculations
        user_input = "350.75"  # Temperature in Kelvin from UI
        
        number_result = client.run_function("CastAnyToFloat64", {
            "input": user_input
        })
        
        if number_result["success"]:
            temp_kelvin = number_result["output"]
            temp_celsius = temp_kelvin - 273.15
            print(f"Temperature: {temp_kelvin}K = {temp_celsius:.2f}°C")
        else:
            print(f"Conversion failed: {number_result['error']}")
            
    except Exception as e:
        print(f"Type conversion error: {e}")

Real-Time Streaming
~~~~~~~~~~~~~~~~~~~

Get responses as they happen for better user experience:

.. code-block:: python

    from aali_client import FlowKitClient
    import sys
    
    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )
    
    # Example: stream engineering analysis explanations
    try:
        # Request streaming response for complex technical query
        response = client.run_function("AnsysGPTPerformLLMRequest", {
            "message": "Explain the finite element method for structural analysis, including mesh considerations and convergence criteria",
            "stream": True,
            "max_tokens": 500,
            "temperature": 0.7,  # Balance creativity and accuracy
            "context": "User is a mechanical engineer familiar with CAD but new to FEA"
        })
        
        # Stream response in real-time
        print("Assistant: ", end='')
        total_tokens = 0
        
        for chunk in response["stream"]:
            if chunk["type"] == "content":
                print(chunk["text"], end='', flush=True)
            elif chunk["type"] == "token_count":
                total_tokens = chunk["count"]
            elif chunk["type"] == "error":
                print(f"\nError: {chunk['message']}")
                break
                
        print(f"\n\nTokens used: {total_tokens}")
        
    except Exception as e:
        print(f"Streaming error: {e}")

Flexible Data Handling
~~~~~~~~~~~~~~~~~~~~~~

FlowKit accepts various data structures without rigid schemas:

.. code-block:: python

    from aali_client import FlowKitClient
    from datetime import datetime
    
    client = FlowKitClient(
        address="localhost:50051",
        api_key="your-api-key"
    )
    
    # Example: store heterogeneous engineering data
    try:
        # Different document structures in same collection
        engineering_data = [
            {
                # Simulation result format
                "type": "simulation",
                "solver": "Fluent",
                "case": "pipe_flow_heat_transfer",
                "results": {
                    "max_velocity": 2.45,
                    "pressure_drop": 1250.5,
                    "outlet_temp": 348.2
                },
                "convergence": True,
                "runtime_seconds": 3600
            },
            {
                # Material property format
                "type": "material",
                "name": "Aluminum 6061-T6",
                "properties": {
                    "density": 2700,
                    "youngs_modulus": 68.9e9,
                    "yield_strength": 276e6,
                    "thermal_conductivity": 167
                },
                "temperature_range": [233, 473],
                "source": "MatWeb"
            },
            {
                # Design specification format
                "type": "specification",
                "component": "Heat Exchanger HX-101",
                "requirements": [
                    "Heat duty: 500 kW",
                    "Max pressure: 10 bar",
                    "Fluid compatibility: Water/Glycol"
                ],
                "created": datetime.now().isoformat(),
                "status": "approved"
            }
        ]
        
        # Store all different formats together
        result = client.run_function("StoreElementsInVectorDatabase", {
            "collection": "engineering_knowledge",
            "elements": engineering_data,
            "embeddingField": "auto",  # FlowKit determines best field
            "preserveStructure": True   # Keep original structure intact
        })
        
        if result["success"]:
            print(f"✓ Stored {result['stored_count']} documents")
            print(f"✓ Handled {result['unique_schemas']} different formats")
        else:
            print(f"✗ Storage failed: {result['error']}")
            
    except Exception as e:
        print(f"Data handling error: {e}")

Best Practices
~~~~~~~~~~~~~~

- Let FlowKit handle type conversions when possible
- Use descriptive names for your data fields
- Test with sample data before full automation
- Check function documentation for expected types

Next Steps
~~~~~~~~~~

- :doc:`functions` - See data types each function expects
- :doc:`categories` - Find functions for your data type