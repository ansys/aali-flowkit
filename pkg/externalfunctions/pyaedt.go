package externalfunctions


import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ansys/aali-flowkit/pkg/privatefunctions/codegeneration"
	"github.com/ansys/aali-flowkit/pkg/privatefunctions/graphdb"
	"github.com/qdrant/go-client/qdrant"

	qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	"github.com/google/uuid"
	"github.com/pandodao/tokenizer-go"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// GetElementMethodContextFromGraphDb  graph database.
//
// Tags:
//   - @displayName: Get context information from Graph DB for method element
//
// Parameters:
//   - elementName - string
// 
func GetElementMethodContextFromGraphDb(element sharedtypes.CodeGenerationElement) {
	ctx := &logging.ContextMap{}

	// kapatil : instead of element names, can we use GUID ?
	err = graphdb.GraphDbDriver.GetExamplesFromCodeGenerationElement(elementType, elementName)
	if err != nil {
		logPanic(ctx, "error Getting examples from code generation element: %v", err)
	}

	// For method name -> 
	// 1. check caller - is application, module or methods, config
	// Method- > belongs to ->class-> is a pyaedtGroup -> <>
	string callerObjType := graphdb.GraphDbDriver.GetMethodCaller(elementName, guid)
	if callObjType == nil {
		errMsg := fmt.Sprintf("error adding code gen relationships to graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	err, string [] params := graphdb.GraphDbDriver.GetParameters(elementName, guid)
        if err != nil {
		errMsg := fmt.Sprintf("error reading parameters  graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	err, string []rets := graphdb.GraphDbDriver.GetReturns(elementName, guid)
        if err != nil {
		errMsg := fmt.Sprintf("error reading return type from  graphdb: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}

	// kapatil: Create context prompt
	// <Method> takes _,_,_, as arguments and returns _
	// For example:
	// application object like hfss calls create_circle takes _,_,_ arguments 
	// and returns ...
	// For example:
	// example-1, 2, 3



}


// GetElementMethodExamplesFromGraphDb  graph database.
//
// Tags:
//   - @displayName: Get examples from Graph DB for method element
//
// Parameters:
//   - elementName - string
// 
func GetElementMethodContextFromGraphDb(element sharedtypes.CodeGenerationElement) {
	ctx := &logging.ContextMap{}

	// kapatil : instead of element names, can we use GUID ?
	err = graphdb.GraphDbDriver.GetExamplesFromCodeGenerationElement(elementType, elementName)
	if err != nil {
		logPanic(ctx, "error Getting examples from code generation element: %v", err)
	}

	// For method name -> 
	err = graphdb.GraphDbDriver.GetExamplesFromGraphDb(elementName)
	if err != nil {
		errMsg := fmt.Sprintf("error getting examples: %v", err)
		logging.Log.Error(ctx, errMsg)
		panic(errMsg)
	}
}



