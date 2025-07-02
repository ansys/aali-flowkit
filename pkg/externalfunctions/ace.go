// File: aali-flowkit/pkg/externalfunctions/data_extraction.go
package externalfunctions

import (
	"fmt"

	"github.com/ansys/aali-sharedtypes/pkg/logging"
)

// ConvertJsonToCustomize convert json to customize format
//
// Tags:
//   - @displayName: Convert json to customize format
//
// Parameters:
//   - object: the object
//
// Returns:
//   - the value of the field as a string
//
// Example output:
// 01. Getting started (section Name -> getting_started\\getting_started_contents.md)
// 02. User guide (section Name -> user_guide\\user_guide_contents.md)
// 03. API reference (section Name -> api\\api_contents.md)
// 04. Contributing to PyFluent (section Name -> contributing\\contributing_contents.md)
// 05. Release notes (section Name -> changelog.md)
func ConvertJsonToCustomize(object []map[string]any) string {
	return convertJsonToCustomizeHelper(object, 0, "")
}

// Internal helper with all parameters
func convertJsonToCustomizeHelper(object []map[string]any, level int, currentIndex string) string {
	var nodeString string

	for _, item := range object {
		chapters, ok := item["chapters"].([]interface{})
		if !ok {
			fmt.Println("Skipping item: not a chapter list")
			continue
		}

		for idx, chapter := range chapters {
			logging.Log.Infof(&logging.ContextMap{}, "chapter: %v", chapter)
			currentIndex := fmt.Sprintf("0%d.", idx+1)
			chapterMap, ok := chapter.(map[string]interface{})
			if !ok {
				fmt.Println("Skipping chapter: not a map")
				continue
			}
			nodeString += fmt.Sprintf(
				"%s%s %s (section Name -> %s)\n",
				repeatString("  ", level),
				currentIndex,
				chapterMap["title"],
				chapterMap["name"],
			)

		}
	}

	logging.Log.Infof(&logging.ContextMap{}, "Output of convertJsonToCustomizeHelper: %s", nodeString)

	return nodeString
}

func repeatString(s string, count int) string {
	var result string
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
