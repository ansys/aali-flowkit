.. _llm_integration:

LLM Integration
===============

FlowKit provides comprehensive Large Language Model (LLM) integration capabilities, serving as the core AI engine for text generation, embeddings, and intelligent processing.

Overview
--------

The LLM Handler integration enables FlowKit to interact with various language models for:

- Text generation and chat completions
- Vector embeddings for semantic search
- Code generation and validation
- Keyword extraction
- Multi-model support with streaming capabilities

Architecture
------------

FlowKit's LLM integration follows a flexible architecture that supports multiple model providers:

.. code-block:: text

   FlowKit <---> LLM Handler Service <---> Model Providers (OpenAI, Azure, etc.)
   
The integration handles authentication, request formatting, response streaming, and error management transparently.

Core Capabilities
-----------------

Vector Embeddings
~~~~~~~~~~~~~~~~~

Generate embeddings for semantic search and similarity matching:

.. list-table:: Embedding Functions
   :header-rows: 1
   :widths: 30 70

   * - Function
     - Description
   * - ``PerformVectorEmbeddingRequest``
     - Generate embeddings for a single text input
   * - ``PerformBatchEmbeddingRequest``
     - Process multiple texts in parallel for efficiency
   * - ``PerformBatchHybridEmbeddingRequest``
     - Combine text and metadata for enhanced embeddings
   * - ``PerformVectorEmbeddingRequestWithTokenLimitCatch``
     - Handle token limits gracefully with automatic chunking

Text Generation
~~~~~~~~~~~~~~~

Flexible text generation with various models and configurations:

.. list-table:: Generation Functions
   :header-rows: 1
   :widths: 30 70

   * - Function
     - Description
   * - ``PerformGeneralRequest``
     - Standard chat completion with system and user prompts
   * - ``PerformGeneralRequestWithImages``
     - Multi-modal requests including image analysis
   * - ``PerformGeneralRequestSpecificModel``
     - Use a specific model by name
   * - ``PerformStreamGeneralRequest``
     - Real-time streaming responses for interactive applications

Specialized Processing
~~~~~~~~~~~~~~~~~~~~~~

Purpose-built functions for specific AI tasks:

.. list-table:: Specialized Functions
   :header-rows: 1
   :widths: 30 70

   * - Function
     - Description
   * - ``PerformCodeGenerationRequest``
     - Generate and validate code snippets
   * - ``PerformKeywordExtractionRequest``
     - Extract key terms and concepts from text
   * - ``PerformSummaryRequest``
     - Create concise summaries of long documents
   * - ``BuildLibraryContext``
     - Construct context from library documentation

Configuration
-------------

LLM integration is configured through FlowKit's main configuration:

.. code-block:: yaml

   # LLM Handler endpoint
   LLM_HANDLER_ENDPOINT: "ws://aali-llm:9003"
   
   # Model configuration (via environment or function parameters)
   LLM_MODEL: "gpt-4"
   LLM_TEMPERATURE: 0.7
   LLM_MAX_TOKENS: 2000

Usage Examples
--------------

**Example 1: Generate Text Embeddings**

.. code-block:: json

   {
     "name": "PerformVectorEmbeddingRequest",
     "inputs": {
       "input": "Understanding fluid dynamics in turbine design",
       "model": "text-embedding-ada-002"
     }
   }

**Example 2: Stream Chat Response**

.. code-block:: json

   {
     "name": "PerformStreamGeneralRequest",
     "inputs": {
       "systemPrompt": "You are an engineering assistant",
       "userPrompt": "Explain the finite element method",
       "temperature": 0.7,
       "maxTokens": 1000
     }
   }

**Example 3: Code Generation**

.. code-block:: json

   {
     "name": "PerformCodeGenerationRequest",
     "inputs": {
       "prompt": "Generate Python function to calculate stress tensor",
       "language": "python",
       "validateSyntax": true
     }
   }

Token Management
----------------

FlowKit provides comprehensive token counting and management:

- **Pre-request counting**: Estimate tokens before sending
- **Post-request tracking**: Monitor actual usage
- **Limit enforcement**: Prevent exceeding model limits
- **Automatic chunking**: Split large inputs automatically

Best Practices
--------------

1. **Model Selection**: Choose appropriate models for your use case
   - Embeddings: Use dedicated embedding models
   - Generation: Select based on quality/speed requirements
   
2. **Context Management**: Build effective prompts
   - Use system prompts for consistent behavior
   - Include relevant context in user prompts
   - Leverage message history for conversations

3. **Error Handling**: Implement robust error handling
   - Handle rate limits with exponential backoff
   - Catch token limit errors and chunk appropriately
   - Monitor model availability

4. **Performance Optimization**:
   - Use batch operations for multiple inputs
   - Enable streaming for long responses
   - Cache embeddings when possible

Integration with Other FlowKit Components
-----------------------------------------

LLM Handler integrates seamlessly with:

- **Knowledge DB**: Store embeddings for similarity search
- **Data Extraction**: Process documents with AI
- **Ansys GPT**: Provide domain-specific responses
- **MCP**: Enhance with external tool capabilities

Troubleshooting
---------------

Common issues and solutions:

**Token Limit Exceeded**
   - Use ``WithTokenLimitCatch`` variants
   - Reduce input size or chunk data
   - Select models with higher limits

**Model Not Available**
   - Check LLM_HANDLER_ENDPOINT configuration
   - Verify model name and availability
   - Ensure proper authentication

**Slow Response Times**
   - Enable streaming for better UX
   - Use batch operations when possible
   - Consider model performance tiers

**Inconsistent Outputs**
   - Adjust temperature settings
   - Improve prompt engineering
   - Use system prompts for consistency