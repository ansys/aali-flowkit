.. _materials:

Materials Intelligence Integration
==================================

FlowKit's Materials Intelligence integration provides AI-powered capabilities for materials data extraction, attribute identification, and intelligent materials selection for engineering applications.

Overview
--------

The Materials Intelligence integration enables:

- Automated extraction of material properties from text
- Intelligent attribute suggestion and validation
- Duplicate detection and filtering
- GUID-based material tracking
- Integration with materials databases
- Multi-model AI processing for accuracy

Architecture
------------

The materials processing pipeline:

.. code-block:: text

   Input Text → Attribute Extraction → Validation → Deduplication → Serialization
        ↓              ↓                   ↓             ↓              ↓
   Parse Query    AI Analysis      Check Database   Filter      Format Response

Core Functions
--------------

Attribute Management
~~~~~~~~~~~~~~~~~~~~

.. list-table:: Attribute Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``AddGuidsToAttributes``
     - Assign unique identifiers to material attributes
   * - ``FilterOutNonExistingAttributes``
     - Validate attributes against known database
   * - ``FilterOutDuplicateAttributes``
     - Remove duplicate attribute entries
   * - ``ExtractCriteriaSuggestions``
     - Extract material selection criteria from queries

Response Processing
~~~~~~~~~~~~~~~~~~~

.. list-table:: Response Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``SerializeResponse``
     - Format material data for API responses
   * - ``PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput``
     - Extract attributes using multiple AI models

Logging and Monitoring
~~~~~~~~~~~~~~~~~~~~~~

.. list-table:: Logging Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``LogRequestSuccess``
     - Log successful material queries
   * - ``LogRequestFailed``
     - Log failed material queries
   * - ``LogRequestFailedDebugWithMessage``
     - Log failures with detailed debug information

Configuration
-------------

Configure materials intelligence settings:

.. code-block:: yaml

   # Materials Database Configuration
   MATERIALS_DB_ENDPOINT: "https://materials-api.company.com"
   MATERIALS_DB_API_KEY: "your-api-key"
   
   # AI Processing Settings
   MATERIALS_AI_MODELS: ["gpt-4", "materials-bert"]
   CONFIDENCE_THRESHOLD: 0.85
   MAX_SUGGESTIONS: 10
   
   # Attribute Validation
   ENABLE_STRICT_VALIDATION: true
   CUSTOM_ATTRIBUTES_ALLOWED: false
   ATTRIBUTE_SCHEMA_VERSION: "2.0"

Usage Examples
--------------

**Example 1: Extract Material Attributes**

.. code-block:: json

   {
     "name": "PerformMultipleGeneralRequestsAndExtractAttributesWithOpenAiTokenOutput",
     "inputs": {
       "query": "Find aluminum alloys with tensile strength > 400 MPa and good corrosion resistance",
       "models": ["gpt-4", "materials-bert"],
       "extractionRules": {
         "material_type": "required",
         "mechanical_properties": "required",
         "environmental_properties": "optional"
       }
     }
   }

**Example 2: Add GUIDs to Attributes**

.. code-block:: json

   {
     "name": "AddGuidsToAttributes",
     "inputs": {
       "attributes": [
         {
           "name": "tensile_strength",
           "value": "450",
           "unit": "MPa",
           "confidence": 0.95
         },
         {
           "name": "density",
           "value": "2.7",
           "unit": "g/cm³",
           "confidence": 0.98
         }
       ]
     }
   }

**Example 3: Filter and Validate Attributes**

.. code-block:: json

   {
     "name": "FilterOutNonExistingAttributes",
     "inputs": {
       "attributes": ["tensile_strength", "yield_strength", "custom_property"],
       "validationMode": "strict",
       "allowCustom": false
     }
   }

Material Property Extraction
----------------------------

**Supported Property Types**:

1. **Mechanical Properties**:
   - Tensile strength
   - Yield strength
   - Elastic modulus
   - Hardness
   - Fatigue limit

2. **Physical Properties**:
   - Density
   - Melting point
   - Thermal conductivity
   - Electrical resistivity
   - Specific heat

3. **Chemical Properties**:
   - Composition
   - Corrosion resistance
   - Oxidation behavior
   - Chemical compatibility

**Extraction Process**:

.. code-block:: text

   "Need steel with high strength and good weldability"
                        ↓
   Extract Properties: 
   - Material: Steel
   - Tensile Strength: High (>500 MPa inferred)
   - Weldability: Good
                        ↓
   Validate & Suggest:
   - AISI 4140 (Heat Treated)
   - AISI 316L (Stainless)
   - A514 (High Strength Low Alloy)

Criteria Suggestions
--------------------

The system intelligently suggests selection criteria:

**Example Query**: "Material for high-temperature turbine blade"

**Suggested Criteria**:

.. code-block:: json

   {
     "primary_criteria": [
       {
         "property": "service_temperature",
         "operator": ">=",
         "value": "900°C"
       },
       {
         "property": "creep_resistance",
         "importance": "critical"
       }
     ],
     "secondary_criteria": [
       {
         "property": "oxidation_resistance",
         "importance": "high"
       },
       {
         "property": "thermal_fatigue",
         "importance": "high"
       }
     ],
     "suggested_materials": [
       "Inconel 718",
       "CMSX-4",
       "Rene N5"
     ]
   }

Attribute Validation
--------------------

**Validation Levels**:

1. **Strict Mode**: Only known attributes allowed
2. **Flexible Mode**: Custom attributes with warnings
3. **Learning Mode**: Track new attributes for review

**Validation Rules**:

- Check attribute names against schema
- Validate units and conversions
- Verify value ranges
- Ensure data type consistency

Deduplication Logic
-------------------

**Duplicate Detection**:

.. code-block:: python

   # Deduplication algorithm
   def filter_duplicates(attributes):
       unique = {}
       for attr in attributes:
           key = (attr.name, attr.unit)
           if key not in unique:
               unique[key] = attr
           else:
               # Keep the one with higher confidence
               if attr.confidence > unique[key].confidence:
                   unique[key] = attr
       return list(unique.values())

Response Serialization
----------------------

**Standard Response Format**:

.. code-block:: json

   {
     "query": "Original user query",
     "extracted_attributes": [
       {
         "guid": "attr_12345",
         "name": "tensile_strength",
         "value": 450,
         "unit": "MPa",
         "confidence": 0.95,
         "source": "gpt-4"
       }
     ],
     "suggested_materials": [
       {
         "name": "AA 7075-T6",
         "match_score": 0.92,
         "properties": {...}
       }
     ],
     "metadata": {
       "processing_time": "1.2s",
       "models_used": ["gpt-4", "materials-bert"],
       "tokens_used": 850
     }
   }

Multi-Model Processing
----------------------

**Model Ensemble Approach**:

1. Query multiple AI models in parallel
2. Compare and merge results
3. Calculate confidence scores
4. Resolve conflicts intelligently

**Conflict Resolution**:

.. code-block:: text

   Model A: "Tensile Strength = 450 MPa" (confidence: 0.9)
   Model B: "Tensile Strength = 480 MPa" (confidence: 0.8)
   
   Resolution Strategies:
   - Weighted average: 462 MPa
   - Highest confidence: 450 MPa
   - Consensus required: Flag for review

Best Practices
--------------

1. **Query Formulation**:
   - Include specific requirements
   - Use standard property names
   - Specify units when known
   - Provide application context

2. **Attribute Management**:
   - Maintain consistent naming
   - Use standard units
   - Track confidence scores
   - Document custom attributes

3. **Performance Optimization**:
   - Cache common queries
   - Batch similar requests
   - Use appropriate models
   - Monitor token usage

Integration Examples
--------------------

**Materials Selection Workflow**:

.. code-block:: python

   # 1. Extract requirements from query
   attributes = extract_criteria_suggestions(user_query)
   
   # 2. Add GUIDs for tracking
   tracked_attrs = add_guids_to_attributes(attributes)
   
   # 3. Validate against database
   valid_attrs = filter_non_existing_attributes(tracked_attrs)
   
   # 4. Remove duplicates
   unique_attrs = filter_duplicate_attributes(valid_attrs)
   
   # 5. Query materials database
   materials = search_materials(unique_attrs)
   
   # 6. Serialize response
   response = serialize_response(materials, unique_attrs)

Logging and Analytics
---------------------

**Request Tracking**:

.. code-block:: json

   {
     "request_id": "req_abc123",
     "timestamp": "2024-01-15T10:30:00Z",
     "user_id": "user_123",
     "query": "high strength aluminum",
     "status": "success",
     "attributes_extracted": 5,
     "materials_found": 12,
     "processing_time_ms": 1250,
     "tokens_used": 850
   }

**Analytics Insights**:

- Most searched properties
- Common material types
- Query success rates
- Model performance comparison

Troubleshooting
---------------

**No Attributes Extracted**:
   - Check query clarity
   - Verify model availability
   - Review extraction rules
   - Examine confidence thresholds

**Invalid Attributes**:
   - Update attribute schema
   - Check validation mode
   - Review naming conventions
   - Verify unit conversions

**Duplicate Issues**:
   - Review deduplication logic
   - Check attribute equality
   - Verify GUID generation
   - Monitor merge conflicts