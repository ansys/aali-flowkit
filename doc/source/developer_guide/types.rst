.. _types_dev:

Adding Custom Types
===================

Custom types enable complex data structures in function parameters and return values.

Type Definition
---------------

Define types in ``pkg/externalfunctions/types.go``:

.. code-block:: go

   type CustomData struct {
       Name    string `json:"name"`
       Value   int    `json:"value"`
       Active  bool   `json:"active"`
   }

Type Registration
-----------------

Types must be registered in multiple locations:

1. **aali-sharedtypes Repository**

   Add type to appropriate package:

   .. code-block:: go

      // In aali-sharedtypes/pkg/sharedtypes/
      type CustomData struct {
          Name    string `json:"name"`
          Value   int    `json:"value"`
          Active  bool   `json:"active"`
      }

2. **Type Converters**

   Update ``aali-sharedtypes/pkg/typeconverters/typeconverters.go``:

   .. code-block:: go

      // In ConvertStringToGivenType
      case "CustomData":
          output := sharedtypes.CustomData{}
          err := json.Unmarshal([]byte(value), &output)
          if err != nil {
              return nil, err
          }
          return output, nil

      // In ConvertGivenTypeToString
      case "CustomData":
          output, err := json.Marshal(value.(sharedtypes.CustomData))
          if err != nil {
              return "", err
          }
          return string(output), nil

3. **UI Configuration**

   Update ``aali-agent-configurator/src/app/constants/constants.ts``:

   .. code-block:: typescript

      export const goTypes: string[] = [
          // Existing types...
          "CustomData",
          "[]CustomData",
      ]

Array Types
-----------

Array types are automatically supported:

- Single type: ``CustomData``
- Array type: ``[]CustomData``

Both must be added to the UI configuration.

Usage in Functions
------------------

.. code-block:: go

   func ProcessCustomData(data CustomData) (result string) {
       return fmt.Sprintf("Processing %s with value %d", data.Name, data.Value)
   }
