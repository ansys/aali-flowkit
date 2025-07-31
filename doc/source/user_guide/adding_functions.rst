Adding Custom Functions
=======================

Extend FlowKit with your own functions by following three simple requirements.

Function Requirements
---------------------

1. **Add @displayName Tag**
~~~~~~~~~~~~~~~~~~~~~~~~~~~

Include a ``@displayName`` tag in comments so your function appears in the UI:

.. code-block:: go

   // ReverseString reverses the input string
   //
   // Tags:
   //   - @displayName: Reverse String
   //
   // Parameters:
   //   - input: The string to reverse
   //
   // Returns:
   //   - reversed: The reversed string
   func ReverseString(input string) (reversed string) {
       // Implementation
   }

2. **Name All Parameters and Returns**
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

FlowKit requires named parameters and return values:

.. code-block:: go

   // Good - named parameters and returns
   func ProcessData(input string, format string) (result string, success bool)

   // Bad - unnamed returns
   func ProcessData(input string, format string) (string, bool)

3. **Use Panic for Errors**
~~~~~~~~~~~~~~~~~~~~~~~~~~~

Handle errors by panicking with clear messages:

.. code-block:: go

   func ProcessData(input string) (result string) {
       if input == "" {
           panic("input cannot be empty")
       }
       // Process data...
       return result
   }

Step-by-Step Guide
------------------

**Step 1: Add Your Function**

Add to an existing category file in ``pkg/externalfunctions/`` (for example, ``generic.go``):

.. code-block:: go

   func YourFunction(param1 string, param2 string) (result string) {
       // Your implementation
       return result
   }

**Step 2: Register Your Function**

Add to ``ExternalFunctionsMap`` in ``pkg/externalfunctions/externalfunctions.go``:

.. code-block:: go

   var ExternalFunctionsMap = map[string]interface{}{
       // ... existing functions ...
       "YourFunction": YourFunction,
   }

**Step 3: Build and Test**

.. code-block:: bash

   # Generate required files
   go generate ./pkg/externalfunctions

   # Build
   go build ./...

   # Run and test
   go run main.go

Adding New Categories
---------------------

To create a new function category:

1. Create ``pkg/externalfunctions/yourcategory.go``
2. Add embed directive in ``main.go``:

   .. code-block:: go

      //go:embed pkg/externalfunctions/yourcategory.go
      var yourCategoryFile string

3. Register in the files map:

   .. code-block:: go

      files := map[string]string{
          // ... existing categories ...
          "your_category": yourCategoryFile,
      }

Adding Custom Types
-------------------

For complex data structures, define types in ``pkg/externalfunctions/types.go``:

.. code-block:: go

   type MyCustomType struct {
       Field1 string `json:"field1"`
       Field2 int    `json:"field2"`
   }

Use JSON strings to pass complex types:

.. code-block:: go

   func ProcessCustomType(data string) (result string) {
       var custom MyCustomType
       if err := json.Unmarshal([]byte(data), &custom); err != nil {
           panic(fmt.Sprintf("Invalid JSON: %v", err))
       }
       // Process...
       output, _ := json.Marshal(custom)
       return string(output)
   }

.. important::
   **Custom Types Registration**

   If you need custom input or output structs, they must be registered in the ``aali-sharedtypes`` repository:

   1. Add your type definition to ``aali-sharedtypes`` with proper JSON tags
   2. Implement type converters in ``aali-sharedtypes/pkg/typeconverters/typeconverters.go``:

      - ``ConvertStringToGivenType`` - converts string to your custom type
      - ``ConvertGivenTypeToString`` - converts your custom type to string

   3. Update the UI constants in ``aali-agent-configurator`` if the type should be available in the UI

   **After registering your custom types:**

   1. Merge your changes to the main branch in both ``aali-sharedtypes`` and ``aali-agent`` repositories
   2. Import the newest version of ``aali-sharedtypes`` in both FlowKit and the Agent
   3. The AALI Agent handles all type conversion and keeps track of variable values throughout workflow execution
   4. If a new variable type is not imported in the Agent, it cannot handle it properly

   The type converters are actually called by the Agent (not FlowKit directly) to convert between string representations and actual Go types. This is why both FlowKit and the Agent must have the updated shared-types imported.

   Without proper registration in ``aali-sharedtypes``, your custom types will not work correctly with the FlowKit type conversion system.

Tips
----

- All inputs and outputs must be strings
- Use JSON for complex data structures
- Check existing functions for patterns
- Test with ``grpcurl`` or client code
