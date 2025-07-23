.. _functions_dev:

Adding Custom Functions
=======================

Custom functions extend FlowKit's capabilities. Functions must be exported Go functions with specific documentation tags.

Function Structure
------------------

Functions must follow this pattern:

.. code-block:: go

   // FunctionName performs a specific operation.
   //
   // Tags:
   //   - @displayName: Display Name in UI
   //
   // Parameters:
   //   - param1: description of first parameter
   //   - param2: description of second parameter
   //
   // Returns:
   //   - result1: description of first return value
   func FunctionName(param1 string, param2 int) (result1 string) {
       // Implementation
       if err != nil {
           panic(fmt.Sprintf("Error: %v", err))
       }
       return "result"
   }

Requirements
------------

- Function name must be exported (capitalized)
- Include docstring with description
- Use ``@displayName`` tag for UI display
- Document all parameters and return values
- Error handling via panic (caught by gRPC server)

Registration
------------

Add the function to ``ExternalFunctionsMap`` in ``pkg/externalfunctions/externalfunctions.go``:

.. code-block:: go

   var ExternalFunctionsMap = map[string]interface{}{
       // Existing functions...
       "FunctionName": FunctionName,
   }

Enumerable Parameters
---------------------

Define enumerable parameters using custom string types:

.. code-block:: go

   type RequestMethod string

   const (
       GET    RequestMethod = "GET"
       POST   RequestMethod = "POST"
       PUT    RequestMethod = "PUT"
       DELETE RequestMethod = "DELETE"
   )

   func MakeRequest(method RequestMethod, url string) (response string) {
       // Implementation
   }

The UI automatically displays these as dropdown options.

Example Implementation
----------------------

.. code-block:: go

   // GenerateUUID generates a unique identifier.
   //
   // Tags:
   //   - @displayName: Generate UUID
   //
   // Returns:
   //   - uuid: a unique identifier string
   func GenerateUUID() string {
       return strings.Replace(uuid.New().String(), "-", "", -1)
   }

Add to ``ExternalFunctionsMap``:

.. code-block:: go

   "GenerateUUID": GenerateUUID,

Additional Files
----------------

- ``gen_cast.go`` - Generated file containing type conversion functions
- ``privatefunctions.go`` - Internal helper functions used by external functions