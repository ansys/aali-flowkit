.. _mesh_pilot:

Ansys Mesh Pilot Integration
============================

Ansys Mesh Pilot is FlowKit's intelligent mesh generation assistant that provides AI-driven guidance for creating optimal computational meshes in engineering simulations.

Overview
--------

The Mesh Pilot integration transforms mesh generation from a manual, expertise-dependent process into an intelligent, guided workflow by:

- Analyzing user requirements and suggesting appropriate meshing strategies
- Providing step-by-step guidance through complex meshing operations
- Learning from past meshing decisions and outcomes
- Synthesizing optimal action sequences for specific geometries
- Offering solutions to common meshing problems

Core Components
---------------

Path Description Analysis
~~~~~~~~~~~~~~~~~~~~~~~~~

Mesh Pilot uses sophisticated path analysis to understand meshing workflows:

.. code-block:: text

   User Intent → Path Search → Property Extraction → Action Synthesis → Result Validation
                     ↓              ↓                    ↓                ↓
                 Find Similar   Extract Mesh      Generate Steps    Check Quality
                 Workflows      Parameters        Sequence

Key Functions
-------------

Workflow Discovery
~~~~~~~~~~~~~~~~~~

.. list-table:: Path and Workflow Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``SimilartitySearchOnPathDescriptions``
     - Find similar meshing workflows based on user requirements
   * - ``FindRelevantPathDescription``
     - Identify the most relevant workflow for current task
   * - ``FetchPropertiesFromPathDescription``
     - Extract mesh properties and parameters from workflows
   * - ``FetchNodeDescriptionsFromPathDescription``
     - Get detailed node information for workflow steps
   * - ``FetchActionsPathFromPathDescription``
     - Retrieve action sequences from workflow paths

Action Generation
~~~~~~~~~~~~~~~~~

.. list-table:: Action Synthesis Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``SynthesizeActions``
     - Generate optimal action sequence for meshing task
   * - ``SynthesizeActionsTool2``
     - Specialized synthesis for surface meshing
   * - ``SynthesizeActionsTool11``
     - Volume mesh generation actions
   * - ``SynthesizeActionsTool12``
     - Boundary layer mesh actions
   * - ``SynthesizeActionsTool17``
     - Advanced mesh refinement actions

Problem Solving
~~~~~~~~~~~~~~~

.. list-table:: Solution Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``GetSolutionsToFixProblem``
     - Identify solutions for mesh quality issues
   * - ``GetSelectedSolution``
     - Choose optimal solution from alternatives
   * - ``SelectedSolution``
     - Apply selected solution to mesh
   * - ``ProcessMainAgentOutput``
     - Process and validate mesh generation results

History and Context
~~~~~~~~~~~~~~~~~~~

.. list-table:: Context Management Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``AppendToolHistory``
     - Track tool usage in workflow
   * - ``AppendMeshPilotHistory``
     - Maintain mesh pilot interaction history
   * - ``ParseHistory``
     - Extract insights from historical data
   * - ``ParseHistoryToHistoricMessages``
     - Convert history to training examples

Configuration
-------------

Mesh Pilot requires configuration for action templates and workflow definitions:

.. code-block:: yaml

   # Mesh Pilot Configuration
   MESH_PILOT_CONFIG_PATH: "/config/mesh_pilot_actions.yaml"
   MESH_PILOT_DB_ENDPOINT: "ws://qdrant:6333"
   
   # Workflow Settings
   MAX_WORKFLOW_STEPS: 50
   DEFAULT_MESH_QUALITY_THRESHOLD: 0.3
   ENABLE_AUTO_REFINEMENT: true

Workflow Examples
-----------------

**Example 1: Basic Surface Mesh Generation**

.. code-block:: json

   {
     "name": "SynthesizeActions",
     "inputs": {
       "userQuery": "Generate surface mesh for turbine blade",
       "geometry": "blade.scdoc",
       "requirements": {
         "elementSize": "5mm",
         "quality": "high",
         "curvature": "fine"
       }
     }
   }

**Example 2: Complex Volume Meshing with Boundary Layers**

.. code-block:: json

   {
     "name": "GenerateActionsSubWorkflowPrompt",
     "inputs": {
       "task": "Create volume mesh with inflation layers",
       "geometry": "combustor.scdoc",
       "specifications": {
         "bulkSize": "10mm",
         "inflationLayers": 15,
         "firstLayerHeight": "0.1mm",
         "growthRate": 1.2
       }
     }
   }

**Example 3: Mesh Quality Problem Resolution**

.. code-block:: json

   {
     "name": "GetSolutionsToFixProblem",
     "inputs": {
       "problem": "Poor element quality in sharp corners",
       "meshStats": {
         "minQuality": 0.15,
         "maxSkewness": 0.85,
         "problemElements": 1250
       },
       "context": "CFD analysis of flow separation"
     }
   }

Action Synthesis Process
------------------------

Mesh Pilot synthesizes actions through a multi-step process:

1. **Intent Analysis**: Understanding user requirements and constraints
2. **Path Matching**: Finding similar successful workflows
3. **Parameter Extraction**: Identifying relevant mesh parameters
4. **Action Generation**: Creating step-by-step instructions
5. **Validation**: Ensuring generated actions meet quality criteria

Subworkflow Management
----------------------

Complex meshing tasks are broken into manageable subworkflows:

.. code-block:: text

   Main Workflow
   ├── Geometry Preparation
   │   ├── Import and Clean
   │   └── Define Named Selections
   ├── Surface Meshing
   │   ├── Global Sizing
   │   └── Local Refinements
   └── Volume Meshing
       ├── Tet/Hex Generation
       └── Boundary Layer Insertion

Prompt Generation
-----------------

Mesh Pilot generates context-aware prompts for guidance:

.. list-table:: Prompt Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``GenerateUserPrompt``
     - Basic prompt for mesh operations
   * - ``GenerateUserPromptWithList``
     - Prompt with action list options
   * - ``GenerateUserPromptWithContext``
     - Context-enriched prompts
   * - ``GenerateHelperSubWorkflowPrompt``
     - Prompts for workflow assistance

Integration with Qdrant
-----------------------

Mesh Pilot uses Qdrant vector database for workflow similarity search:

.. code-block:: json

   {
     "name": "SimilartitySearchOnPathDescriptionsQdrant",
     "inputs": {
       "query": "hex mesh for structural analysis",
       "collection": "mesh_workflows",
       "limit": 10,
       "scoreThreshold": 0.8
     }
   }

Best Practices
--------------

1. **Geometry Preparation**:
   - Clean geometry before meshing
   - Define clear named selections
   - Verify water-tight volumes

2. **Mesh Strategy**:
   - Start with global sizing
   - Add local refinements progressively
   - Validate quality at each step

3. **Problem Resolution**:
   - Use diagnostic tools first
   - Apply targeted fixes
   - Verify improvements

4. **Workflow Management**:
   - Save successful workflows
   - Document parameter choices
   - Build reusable templates

Output Formatting
-----------------

Mesh Pilot provides structured outputs:

.. code-block:: text

   Mesh Generation Complete:
   - Total Elements: 2,500,000
   - Quality Metrics:
     * Min Orthogonal Quality: 0.35
     * Max Skewness: 0.65
     * Aspect Ratio: < 5:1
   - Workflow Steps: 12
   - Execution Time: 5m 23s

The integration includes utilities for formatting:

- ``MarkdownToHTML``: Convert guidance to HTML
- ``FinalizeResult``: Format final mesh statistics
- ``FinalizeMessage``: Prepare user-friendly messages

Troubleshooting
---------------

Common issues and solutions:

**Poor Mesh Quality**
   - Review geometry for small features
   - Adjust sizing functions
   - Consider different element types

**Failed Mesh Generation**
   - Check geometry validity
   - Simplify complex regions
   - Use decomposition strategies

**Performance Issues**
   - Optimize element count
   - Use parallel meshing
   - Enable GPU acceleration

**Workflow Not Found**
   - Broaden search criteria
   - Create custom workflow
   - Consult mesh examples