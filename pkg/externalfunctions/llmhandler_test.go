// Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package externalfunctions

import (
	"strings"
	"testing"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/stretchr/testify/assert"
)

func TestDbResponsePromptFormat(t *testing.T) {
	testCases := []struct {
		name          string
		dbResponse    sharedtypes.DbResponse
		expectedLines []string
	}{
		{
			"API Element",
			sharedtypes.DbResponse{
				Name:           "M:Namespace.Class.MyCusomMethod(System.String,System.Int32)",
				NamePseudocode: "MyCustomMethod",
				NameFormatted:  "My Custom Method",
				Type:           "Method",
				ParentClass:    "Namespace.Class",
				Metadata:       map[string]any{},
			},
			[]string{
				"=== START API ELEMENT #1 ===",
				"API: My Custom Method",
				"Function: MyCustomMethod",
				"Full Name: M:Namespace.Class.MyCusomMethod(System.String,System.Int32)",
				"Type: Method",
				"Parent: Namespace.Class",
				"=== END API ELEMENT #1 ===",
			},
		},
		{
			"Example",
			sharedtypes.DbResponse{
				DocumentName:           "examples/my_example.py",
				Text:                   "import random\n\ndef main():\n    print(random.randint(0, 10))\n\nif __name__ == '__main__':\n    main()\n",
				PreviousChunk:          "previous-chunk-id",
				NextChunk:              "next-chunk-id",
				Dependencies:           []any{"random"},
				DependencyEquivalences: map[string]any{"random": "random-equiv"},
			},
			[]string{
				"=== START EXAMPLE #1 ===",
				"Example File: examples/my_example.py",
				"Uses APIs: random",
				"Document: examples/my_example.py",
				"Code:",
				"```python",
				"import random",
				"",
				"def main():",
				"    print(random.randint(0, 10))",
				"",
				"if __name__ == '__main__':",
				"    main()",
				"",
				"```",
				"=== END EXAMPLE #1 ===",
			},
		},
		{
			"User Guide",
			sharedtypes.DbResponse{
				SectionName:       "Section",
				DocumentName:      "user_guide",
				Title:             "Title",
				ParentSectionName: "Parent",
				Level:             "2",
				Text:              "Here is the user\nguide content",
				PreviousChunk:     "prev-chunk",
				NextChunk:         "next-chunk",
			},
			[]string{
				"=== START USER GUIDE #1 ===",
				"Guide Title: Title",
				"Section: Section",
				"Parent Section: Parent",
				"Document: user_guide",
				"Content:",
				"Here is the user",
				"guide content",
				"=== END USER GUIDE #1 ===",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := dbResponsePromptFormat(tc.dbResponse, 1)
			expected := strings.Join(tc.expectedLines, "\n")
			assert.Equal(t, expected, result)
		})
	}

}
