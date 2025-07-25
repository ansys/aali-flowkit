.. _troubleshooting_dev:

Troubleshooting
===============

Common issues when developing FlowKit extensions and their solutions.

Function Registration Issues
----------------------------

**Problem**: Function not appearing in FlowKit UI

**Solutions**:

1. **Check Function Export**: Ensure function name starts with capital letter:

   .. code-block:: go

      // Correct - exported function
      func GenerateUUID() string { ... }

      // Incorrect - not exported
      func generateUUID() string { ... }

2. **Verify File Embedding**: Ensure the category file is embedded in ``main.go``:

   .. code-block:: go

      //go:embed pkg/externalfunctions/generic.go
      var genericFile string
      
      // And included in files map
      files := map[string]string{
          "generic": genericFile,
      }

3. **Check Function Registry**: Add function to ``ExternalFunctionsMap``:

   .. code-block:: go

      var ExternalFunctionsMap = map[string]interface{}{
          "GenerateUUID": GenerateUUID,  // Add this line
      }

4. **Check Documentation Format**: Ensure proper docstring format:

   .. code-block:: go

      // GenerateUUID generates a new UUID.  // ← Description required
      //
      // Tags:
      //   - @displayName: Generate UUID      // ← @displayName required
      //
      // Returns:
      //   - uuid: string representation      // ← Document returns
      func GenerateUUID() string { ... }

5. **Verify Discovery Process**: Check that function discovery completed successfully in startup logs:

   .. code-block:: bash

      # Look for successful extraction messages
      grep "ExtractFunctionDefinitionsFromPackage" logs.log

Type Conversion Errors
----------------------

**Problem**: ``panic: interface conversion: interface {} is <type>, not <expected_type>``

**Solutions**:

1. **Use Proper Cast Functions**: Use existing cast functions from ``cast.go``:

   .. code-block:: go

      // Use cast functions instead of direct type assertion
      str := CastAnyToString(data)  // Safe
      // str := data.(string)       // Unsafe - can panic

2. **Check Input Types**: Verify function parameter types match expected inputs:

   .. code-block:: go

      func ProcessData(input string) {  // Expects string
          // If FlowKit passes int, use cast function:
          // input = CastIntToString(inputInt)
      }

Error Handling Issues
---------------------

**Problem**: Functions crashing FlowKit server

**Solutions**:

1. **Use logPanic for Errors**: Always use ``logPanic`` instead of regular ``panic``:

   .. code-block:: go

      client, err := createClient()
      if err != nil {
          logPanic(nil, "unable to create client: %q", err)  // Correct
          // panic("client creation failed")                 // Incorrect
      }

2. **Validate Input Parameters**: Check parameters before use:

   .. code-block:: go

      func ProcessFile(filename string) {
          if filename == "" {
              logPanic(nil, "filename cannot be empty")
          }
          // Continue processing...
      }

3. **Handle External Dependencies**: Check external service availability:

   .. code-block:: go

      func CallExternalAPI(url string) {
          if url == "" {
              logPanic(nil, "URL cannot be empty")
          }
          
          resp, err := http.Get(url)
          if err != nil {
              logPanic(nil, "failed to call API: %q", err)
          }
          defer resp.Body.Close()
      }

Build and Compilation Issues
----------------------------

**Problem**: ``undefined: <function_name>`` during build

**Solutions**:

1. **Check Package Declaration**: Ensure all files use same package:

   .. code-block:: go

      package externalfunctions  // Must match across all files

2. **Import Required Packages**: Add necessary imports:

   .. code-block:: go

      import (
          "context"
          "strings"
          "github.com/google/uuid"
      )

3. **Verify File Location**: Ensure files are in ``pkg/externalfunctions/`` directory.

Performance Issues
------------------

**Problem**: Function execution is slow

**Solutions**:

1. **Optimize Database Queries**: Use proper indexing and limit results:

   .. code-block:: go

      func SearchDatabase(query string) []Result {
          // Limit results to reasonable number
          results := db.Search(query).Limit(100)
          return results
      }

2. **Cache Expensive Operations**: Cache results of expensive computations:

   .. code-block:: go

      var cache = make(map[string]string)

      func ExpensiveOperation(input string) string {
          if result, exists := cache[input]; exists {
              return result
          }
          
          result := doExpensiveWork(input)
          cache[input] = result
          return result
      }

3. **Use Context for Timeouts**: Implement timeouts for long-running operations:

   .. code-block:: go

      func LongRunningTask() {
          ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
          defer cancel()
          
          // Use ctx in operations that support it
          result, err := client.Operation(ctx, params)
          if err != nil {
              logPanic(nil, "operation timed out: %q", err)
          }
      }

Memory Issues
-------------

**Problem**: High memory usage or memory leaks

**Solutions**:

1. **Close Resources**: Always close files, connections, and HTTP responses:

   .. code-block:: go

      func ReadFile(filename string) string {
          file, err := os.Open(filename)
          if err != nil {
              logPanic(nil, "cannot open file: %q", err)
          }
          defer file.Close()  // Important: close resource
          
          // Read file content...
      }

2. **Limit Data Processing**: Process large datasets in chunks:

   .. code-block:: go

      func ProcessLargeData(data []string) {
          const chunkSize = 1000
          for i := 0; i < len(data); i += chunkSize {
              end := i + chunkSize
              if end > len(data) {
                  end = len(data)
              }
              processChunk(data[i:end])
          }
      }

JSON Marshaling Issues
----------------------

**Problem**: JSON serialization/deserialization errors

**Solutions**:

1. **Use Proper JSON Tags**: Add JSON tags to struct fields:

   .. code-block:: go

      type Response struct {
          ID      string `json:"id"`
          Message string `json:"message"`
          Data    []Item `json:"data"`
      }

2. **Handle Nested Structures**: Properly define nested types:

   .. code-block:: go

      type ComplexResponse struct {
          Status   string            `json:"status"`
          Metadata map[string]any    `json:"metadata"`
          Items    []ResponseItem    `json:"items"`
      }

      type ResponseItem struct {
          ID    string `json:"id"`
          Value string `json:"value"`
      }

Configuration Issues
--------------------

**Problem**: Function can't access configuration values

**Solutions**:

1. **Use Config Package**: Access configuration through proper channels:

   .. code-block:: go

      // Access configuration values properly
      dbURL := config.GetDatabaseURL()
      apiKey := config.GetAPIKey()

2. **Environment Variables**: Use environment variables for sensitive data:

   .. code-block:: go

      apiKey := os.Getenv("API_KEY")
      if apiKey == "" {
          logPanic(nil, "API_KEY environment variable not set")
      }

Debug Tips
----------

1. **Add Debug Logging**: Use proper logging for debugging:

   .. code-block:: go

      log.Printf("Function called with params: %+v", params)

2. **Test Functions Independently**: Create test functions to verify behavior:

   .. code-block:: go

      func TestGenerateUUID() {
          uuid := GenerateUUID()
          if len(uuid) != 32 {  // UUID without dashes
              panic("Invalid UUID format")
          }
      }

3. **Use Go's Built-in Tools**: Use ``go vet`` and ``go fmt`` to catch issues:

   .. code-block:: bash

      go vet ./pkg/externalfunctions/
      go fmt ./pkg/externalfunctions/

Getting Help
------------

If issues persist:

1. Check the FlowKit logs for detailed error messages
2. Verify that similar existing functions work correctly
3. Review the ``ExternalFunctionsMap`` for registration patterns
4. Test functions with simple inputs first, then increase complexity
5. Use the troubleshooting patterns from existing working functions