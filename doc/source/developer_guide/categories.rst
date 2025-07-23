.. _categories_dev:

Adding Function Categories
==========================

Categories organize related functions. Each category corresponds to a Go file in ``pkg/externalfunctions/``.

Creating a Category
-------------------

1. **Create Category File**

   Create a new file in ``pkg/externalfunctions/``:

   .. code-block:: go

      // File: pkg/externalfunctions/mycategory.go
      package externalfunctions

      // Functions for this category...

2. **Update Main File**

   Add embed directive and file mapping in ``main.go``:

   .. code-block:: go

      //go:embed pkg/externalfunctions/mycategory.go
      var myCategoryFile string

      // In main function
      files := map[string]string{
          // Existing categories...
          "mycategory": myCategoryFile,
      }

3. **Update UI Configuration**

   Add to ``aali-agent-configurator/src/app/constants/constants.ts``:

   .. code-block:: typescript

      export const functionCategories = {
          // Existing categories...
          "mycategory": "My Category",
      }

Category Naming
---------------

- File name: lowercase (``ansysgpt.go``, ``dataextraction.go``)
- Category key: lowercase with underscores (``ansys_gpt``, ``data_extraction``)
- Display name: Title case (``Ansys GPT``, ``Data Extraction``)

Functions in Category
---------------------

All exported functions in the category file are automatically discovered:

.. code-block:: go

   // File: pkg/externalfunctions/mycategory.go

   // CategoryFunction1 does something specific.
   //
   // Tags:
   //   - @displayName: Category Function 1
   func CategoryFunction1() string {
       return "result"
   }

   // CategoryFunction2 does something else.
   //
   // Tags:
   //   - @displayName: Category Function 2
   func CategoryFunction2(input string) string {
       return input
   }

Register all functions in ``ExternalFunctionsMap``:

.. code-block:: go

   "CategoryFunction1": CategoryFunction1,
   "CategoryFunction2": CategoryFunction2,
