.. _function_registration:

Function registration
=====================

Register Go functions with Flowkit to make them available for AI workflow execution.

Registration process
--------------------

The function registration process follows these steps:

1. **Define Function** - Write your Go function with proper signatures
2. **Register with Flowkit** - Use the registration API to add your function
3. **Validate Signature** - Flowkit validates the function signature
4. **Store in Registry** - Valid functions are stored in the function registry
5. **Available for Execution** - Functions become available for AI workflow execution

Function structure
------------------

.. code-block:: text
   :caption: Function Components

   Function Definition
   ├── Name (unique identifier)
   ├── Input Schema
   │   ├── Parameter types
   │   └── Validation rules
   ├── Output Schema
   │   └── Return format
   ├── Implementation
   │   └── Go function code
   └── Metadata
       ├── Description
       └── Tags

Step 1: Define your function
----------------------------

.. container:: step-box

   │ **Function Definition**
   │ Create Go function with proper signature
   │ Add documentation and tags
   │
   └─ Follow Flowkit conventions

**Basic Function Template:**

.. code-block:: go
   :linenos:
   :caption: basic-function.go

   // ProcessData analyzes input data and returns processed results.
   // Tags:
   //   - @displayName: Process Data Analysis
   //   - @category: data-processing
   //   - @description: Processes raw data and applies analysis algorithms
   func ProcessData(ctx context.Context, input string) (string, error) {
       // Parse input JSON
       var params struct {
           Data     []interface{} `json:"data"`
           Method   string        `json:"method"`
           Options  map[string]interface{} `json:"options,omitempty"`
       }
       
       if err := json.Unmarshal([]byte(input), &params); err != nil {
           return "", fmt.Errorf("invalid input: %v", err)
       }
       
       // Process data based on method
       result := processWithMethod(params.Data, params.Method, params.Options)
       
       // Return JSON result
       output, _ := json.Marshal(map[string]interface{}{
           "processed_data": result,
           "method_used": params.Method,
           "timestamp": time.Now().Unix(),
       })
       
       return string(output), nil
   }

**Advanced Function with Validation:**

.. code-block:: go
   :linenos:
   :caption: advanced-function.go

   // AnalyzeDataset performs comprehensive dataset analysis.
   // Tags:
   //   - @displayName: Analyze Dataset
   //   - @category: analytics
   //   - @version: 2.1
   //   - @author: Data Team
   func AnalyzeDataset(ctx context.Context, input string) (string, error) {
       // Input validation
       var params struct {
           Dataset    []map[string]interface{} `json:"dataset" validate:"required,min=1"`
           AnalysisType string                `json:"analysis_type" validate:"required,oneof=basic advanced full"`
           OutputFormat string                `json:"output_format" validate:"oneof=json csv xml"`
           Filters      map[string]interface{} `json:"filters,omitempty"`
       }
       
       if err := json.Unmarshal([]byte(input), &params); err != nil {
           return "", fmt.Errorf("JSON parsing error: %v", err)
       }
       
       // Validate required fields
       if len(params.Dataset) == 0 {
           return "", fmt.Errorf("dataset cannot be empty")
       }
       
       if params.OutputFormat == "" {
           params.OutputFormat = "json" // Default format
       }
       
       // Perform analysis
       analysis := performAnalysis(params.Dataset, params.AnalysisType, params.Filters)
       
       // Format output
       output := formatOutput(analysis, params.OutputFormat)
       
       return output, nil
   }

Step 2: Register the Function
-----------------------------

.. container:: step-box

   │ **Function Registration**
   │ Add to ExternalFunctionsMap
   │ Make available via gRPC
   │
   └─ Function becomes callable

**Registration in externalfunctions.go:**

.. code-block:: go
   :linenos:
   :caption: externalfunctions.go

   package externalfunctions
   
   // ExternalFunctionsMap contains all functions available via gRPC
   var ExternalFunctionsMap = map[string]interface{}{
       // Data Processing Functions
       "ProcessData":          ProcessData,
       "AnalyzeDataset":       AnalyzeDataset,
       "TransformData":        TransformData,
       "ValidateData":         ValidateData,
       
       // Utility Functions
       "GenerateReport":       GenerateReport,
       "SendNotification":     SendNotification,
       "CacheResult":          CacheResult,
       
       // AI Integration Functions
       "CallLLM":              CallLLM,
       "ProcessNLP":           ProcessNLP,
       "ImageAnalysis":        ImageAnalysis,
       
       // System Functions
       "ResetSession":         ResetSession,
       "GetSystemStatus":      GetSystemStatus,
       "ExportData":           ExportData,
   }

**Function Categories:**

.. code-block:: go
   :linenos:
   :caption: function-categories.go

   // Organize functions by category for better management
   var FunctionCategories = map[string][]string{
       "data-processing": {
           "ProcessData",
           "AnalyzeDataset", 
           "TransformData",
           "ValidateData",
       },
       "utilities": {
           "GenerateReport",
           "SendNotification",
           "CacheResult",
       },
       "ai-integration": {
           "CallLLM",
           "ProcessNLP",
           "ImageAnalysis",
       },
       "system": {
           "ResetSession",
           "GetSystemStatus",
           "ExportData",
       },
   }

Step 3: Function metadata
-------------------------

**Documentation Tags:**

.. code-block:: go
   :linenos:
   :caption: metadata-example.go

   // ComplexAnalysis performs multi-stage data analysis with AI integration.
   // 
   // This function takes raw data, applies preprocessing, runs AI analysis,
   // and returns comprehensive insights with confidence scores.
   //
   // Tags:
   //   - @displayName: Complex AI Analysis
   //   - @category: ai-integration
   //   - @version: 3.0
   //   - @author: AI Team
   //   - @requires: ai-service,database
   //   - @performance: high-cpu
   //   - @timeout: 60s
   //   - @description: Multi-stage analysis combining traditional algorithms with AI
   //
   // Input Schema:
   //   {
   //     "data": [...],           // Required: Raw data array
   //     "model": "string",       // Required: AI model to use
   //     "confidence": 0.8,       // Optional: Minimum confidence threshold
   //     "preprocessing": {...}   // Optional: Preprocessing options
   //   }
   //
   // Output Schema:
   //   {
   //     "analysis": {...},       // Analysis results
   //     "confidence": 0.95,      // Overall confidence score
   //     "insights": [...],       // Key insights discovered
   //     "recommendations": [...] // Actionable recommendations
   //   }
   func ComplexAnalysis(ctx context.Context, input string) (string, error) {
       // Implementation...
   }

Function signatures
-------------------

**Standard Signatures:**

.. code-block:: go
   :linenos:
   :caption: function-signatures.go

   // Basic function (most common)
   func BasicFunction(ctx context.Context, input string) (string, error)
   
   // Function with memory access
   func MemoryFunction(ctx context.Context, input string, memory *Memory) (string, error)
   
   // Function with configuration
   func ConfigFunction(ctx context.Context, input string, config *Config) (string, error)
   
   // Streaming function
   func StreamingFunction(ctx context.Context, input string, stream Stream) error
   
   // Async function
   func AsyncFunction(ctx context.Context, input string) (*AsyncResult, error)

**Input/Output Examples:**

.. code-block:: go
   :linenos:
   :caption: io-examples.go

   // Simple string processing
   func ProcessText(ctx context.Context, input string) (string, error) {
       var params struct {
           Text string `json:"text"`
           Op   string `json:"operation"`
       }
       json.Unmarshal([]byte(input), &params)
       
       result := performOperation(params.Text, params.Op)
       return result, nil
   }
   
   // Structured data processing
   func ProcessStructuredData(ctx context.Context, input string) (string, error) {
       var params struct {
           Records []Record `json:"records"`
           Config  Config   `json:"config"`
       }
       json.Unmarshal([]byte(input), &params)
       
       output := map[string]interface{}{
           "processed": processRecords(params.Records, params.Config),
           "count":     len(params.Records),
           "timestamp": time.Now(),
       }
       
       result, _ := json.Marshal(output)
       return string(result), nil
   }

Error handling
--------------

**Best Practices:**

.. code-block:: go
   :linenos:
   :caption: error-handling.go

   func RobustFunction(ctx context.Context, input string) (string, error) {
       // 1. Input validation
       if input == "" {
           return "", fmt.Errorf("input cannot be empty")
       }
       
       var params RequestParams
       if err := json.Unmarshal([]byte(input), &params); err != nil {
           return "", fmt.Errorf("invalid JSON input: %v", err)
       }
       
       // 2. Parameter validation
       if err := validateParams(params); err != nil {
           return "", fmt.Errorf("parameter validation failed: %v", err)
       }
       
       // 3. Context timeout handling
       select {
       case <-ctx.Done():
           return "", fmt.Errorf("operation cancelled: %v", ctx.Err())
       default:
       }
       
       // 4. Business logic with error handling
       result, err := processWithRecovery(params)
       if err != nil {
           // Log error for debugging
           log.Error("Function execution failed", "error", err, "params", params)
           return "", fmt.Errorf("processing failed: %v", err)
       }
       
       // 5. Output formatting
       output, err := json.Marshal(result)
       if err != nil {
           return "", fmt.Errorf("output serialization failed: %v", err)
       }
       
       return string(output), nil
   }

Testing functions
-----------------

**Unit Test Template:**

.. code-block:: go
   :linenos:
   :caption: function_test.go

   func TestProcessData(t *testing.T) {
       tests := []struct {
           name     string
           input    string
           expected string
           wantErr  bool
       }{
           {
               name:     "valid input",
               input:    `{"data": [1,2,3], "method": "sum"}`,
               expected: `{"result": 6, "method": "sum"}`,
               wantErr:  false,
           },
           {
               name:    "invalid JSON",
               input:   `{invalid json}`,
               wantErr: true,
           },
           {
               name:    "missing method",
               input:   `{"data": [1,2,3]}`,
               wantErr: true,
           },
       }
       
       for _, tt := range tests {
           t.Run(tt.name, func(t *testing.T) {
               ctx := context.Background()
               result, err := ProcessData(ctx, tt.input)
               
               if tt.wantErr {
                   assert.Error(t, err)
                   return
               }
               
               assert.NoError(t, err)
               assert.JSONEq(t, tt.expected, result)
           })
       }
   }

**Integration Testing:**

.. code-block:: bash
   :linenos:
   :caption: test-integration.sh

   # Test function via gRPC
   grpc_cli call localhost:9090 \
     flowkit.FunctionService.CallFunction \
     '{
       "function": "ProcessData",
       "input": "{\"data\": [1,2,3,4,5], \"method\": \"average\"}"
     }'
   
   # Expected response
   # {
   #   "output": "{\"result\": 3, \"method\": \"average\"}",
   #   "status": "success",
   #   "execution_time_ms": 5
   # }

Deployment checklist
--------------------

.. container:: checklist

   **Before Deployment:**
   ☐ Function is properly documented
   ☐ Input/output schemas defined
   ☐ Error handling implemented
   ☐ Unit tests written and passing
   ☐ Integration tests successful
   ☐ Performance benchmarks acceptable
   ☐ Security review completed
   ☐ Function registered in map
   ☐ Flowkit service restarted

**Production Considerations:**

1. **Performance**: Functions should complete within reasonable time
2. **Memory**: Avoid memory leaks in long-running functions
3. **Concurrency**: Functions may be called concurrently
4. **Error Recovery**: Handle partial failures gracefully
5. **Logging**: Include appropriate logging for debugging
6. **Security**: Validate all inputs, sanitize outputs

Next steps
----------

.. container:: next-steps

   │ **Function Registered Successfully**
   ├─ :doc:`calling` - Execute your function
   └─ :doc:`agent_integration` - Connect with AALI agent