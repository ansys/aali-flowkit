
package externalfunctions

import (
	"context"
	"fmt"
	"strings"

	"github.com/ansys/aali-flowkit/pkg/privatefunctions/graphdb"
	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/ansys/aali-sharedtypes/pkg/aali_graphdb"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

// LoadExamplesForGivenElement stores elements in the graph database.
//
// Tags:
//   - @displayName: Load Examples from Graph DB for element
//
// Parameters:
//   - elementName - string
//   - elementType - string
func StoreElementsInGraphDatabase(elements []sharedtypes.CodeGenerationElement) {
	ctx := &logging.ContextMap{}

	err = graphdb.GraphDbDriver.GetExamplesFromCodeGenerationElement(elementType, elementName)
	if err != nil {
		logPanic(ctx, "error Getting examples from code generation element: %v", err)
	}

	err = graphdb.GraphDbDriver.CreateCodeGenerationRelationships(elements)
	if err != nil {
		errMsg := fmt.Sprintf("error adding code gen relationships to graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}
}
p

