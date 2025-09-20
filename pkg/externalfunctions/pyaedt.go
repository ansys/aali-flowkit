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
	//"encoding/json"
	//"encoding/xml"
	//"os"
	//"path/filepath"
	//"strings"
	//"sync"

	//"github.com/ansys/aali-flowkit/pkg/privatefunctions/codegeneration"
	"github.com/ansys/aali-flowkit/pkg/privatefunctions/graphdb"
	//"github.com/qdrant/go-client/qdrant"

	//qdrant_utils "github.com/ansys/aali-flowkit/pkg/privatefunctions/qdrant"
	//"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
	//"github.com/google/uuid"
	//"github.com/pandodao/tokenizer-go"
	//"github.com/tmc/langchaingo/documentloaders"
	//"github.com/tmc/langchaingo/schema"
	//"github.com/tmc/langchaingo/textsplitter"
)

// GetElementContextFromGraphDb  graph database.
//
// Tags:
//   - @displayName: Get context information from Graph DB for method or class element
//
// Parameters:
//   - elementName - string
//   - elementType - string
func GetElementContextFromGraphDb(dbResponses []sharedtypes.ApiDbResponse) {
	ctx := &logging.ContextMap{}
	var exampleName []string
	var err error
	//graphdb.Initialize()
	// kapatil : instead of element names, can we use GUID ?
	// Assuming this is a single entry point
	if len(dbResponses) > 0 {
		elementType := dbResponses[0].Type
		elementName := dbResponses[0].Name
		exampleName, err = graphdb.GraphDbDriver.GetExamplesFromCodeGenerationElement(elementType, elementName)
		if err != nil {
			logPanic(ctx, "error Getting examples from code generation element: %v", err)
		}
		for ex, _ := range exampleName {
			logging.Log.Debugf(ctx, "Reading examples %v", ex)
		}
	} else {
		logging.Log.Debugf(ctx, "Graph DB no entry point found!!!")
	}
	// For method name ->
	// 1. check caller - is application, module or methods, config
	// Method- > belongs to ->class-> is a pyaedtGroup -> <>
	// string
	//callerObjType = graphdb.GraphDbDriver.GetMethodCaller(elementName, guid)
	//if callObjType == nil {
	//	errMsg := fmt.Sprintf("error adding code gen relationships to graphdb: %v", err)
	//			logging.Log.Error(ctx, errMsg)
	//			panic(errMsg)
	//		}
	//
	//	//string[]
	//	err, params = graphdb.GraphDbDriver.GetParameters(elementName, guid)
	//      if err != nil {
	//		errMsg := fmt.Sprintf("error reading parameters  graphdb: %v", err)
	//		logging.Log.Error(ctx, errMsg)
	//		panic(errMsg)
	//	}

	// rets []string
	//	err, rets = graphdb.GraphDbDriver.GetReturns(elementName, guid)
	//      if err != nil {
	//		errMsg := fmt.Sprintf("error reading return type from  graphdb: %v", err)
	//		logging.Log.Error(ctx, errMsg)
	//		panic(errMsg)
	//	}

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
//func GetElementMethodContextFromGraphDb(element sharedtypes.CodeGenerationElement) {
//	ctx := &logging.ContextMap{}

// kapatil : instead of element names, can we use GUID ?
//	err = graphdb.GraphDbDriver.GetExamplesFromCodeGenerationElement(elementType, elementName)
//	if err != nil {
//		logPanic(ctx, "error Getting examples from code generation element: %v", err)
//	}

// For method name ->
//	err = graphdb.GraphDbDriver.GetExamplesFromGraphDb(elementName)
//	if err != nil {
//		errMsg := fmt.Sprintf("error getting examples: %v", err)
//		logging.Log.Error(ctx, errMsg)
//		panic(errMsg)
//	}
//}
