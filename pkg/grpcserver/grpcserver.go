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

package grpcserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"

	"github.com/ansys/aali-flowkit/pkg/externalfunctions"
	"github.com/ansys/aali-sharedtypes/pkg/aaliflowkitgrpc"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/typeconverters"

	"github.com/ansys/aali-flowkit/pkg/internalstates"
	"github.com/ansys/aali-sharedtypes/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// server is used to implement grpc_definition.ExternalFunctionsServer.
type server struct {
	aaliflowkitgrpc.UnimplementedExternalFunctionsServer
}

// StartServer starts the gRPC server
// The server listens on the port specified in the configuration file
// The server implements the ExternalFunctionsServer interface
func StartServer() {
	// Get webserver address
	webserverAddress, err := config.HandleLegacyPortDefinition(config.GlobalConfig.FLOWKIT_ADDRESS, config.GlobalConfig.EXTERNALFUNCTIONS_GRPC_PORT)
	if err != nil {
		logging.Log.Fatalf(&logging.ContextMap{}, "Error getting webserver address: %v", err)
	}

	// Create listener on the specified address
	lis, err := net.Listen("tcp", webserverAddress)
	if err != nil {
		logging.Log.Fatalf(&logging.ContextMap{}, "failed to listen: %v", err)
	}

	// Check if SSL is enabled and load the server's certificate and private key
	var opts []grpc.ServerOption
	if config.GlobalConfig.USE_SSL {
		creds, err := credentials.NewServerTLSFromFile(
			config.GlobalConfig.SSL_CERT_PUBLIC_KEY_FILE,
			config.GlobalConfig.SSL_CERT_PRIVATE_KEY_FILE,
		)
		if err != nil {
			logging.Log.Fatalf(&logging.ContextMap{}, "failed to load SSL certificates: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Add API key authentication interceptor if an API key is provided
	if config.GlobalConfig.FLOWKIT_API_KEY != "" {
		opts = append(opts, grpc.UnaryInterceptor(apiKeyAuthInterceptor(config.GlobalConfig.FLOWKIT_API_KEY)))
	}

	// Set gRPC message size limits
	opts = append(opts, grpc.MaxRecvMsgSize(1024*1024*1024)) // 1 GB receive limit
	opts = append(opts, grpc.MaxSendMsgSize(1024*1024*1024)) // 1 GB send limit

	// Create the gRPC server with the options
	s := grpc.NewServer(opts...)
	aaliflowkitgrpc.RegisterExternalFunctionsServer(s, &server{})
	logging.Log.Infof(&logging.ContextMap{}, "Aali FlowKit started successfully; gRPC server listening on address '%s'...\n", webserverAddress)
	if err := s.Serve(lis); err != nil {
		logging.Log.Fatalf(&logging.ContextMap{}, "failed to serve: %v", err)
	}
}

// apiKeyAuthInterceptor is a gRPC server interceptor that checks for a valid API key in the metadata of the request
// The API key is passed as a string parameter
//
// Parameters:
// - apiKey: a string containing the API key
//
// Returns:
// - grpc.UnaryServerInterceptor: a gRPC server interceptor
func apiKeyAuthInterceptor(apiKey string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract API key from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		receivedApiKeys := md["x-api-key"]
		if len(receivedApiKeys) == 0 || receivedApiKeys[0] != apiKey {
			return nil, status.Errorf(codes.Unauthenticated, "invalid API key")
		}

		// Continue handling the request
		return handler(ctx, req)
	}
}

// HealthCheck checks the health of the gRPC server
//
// Parameters:
// - ctx: the context of the request
// - req: the request to check the health of the server
//
// Returns:
// - aaliflowkitgrpc.HealthCheckResponse: a response indicating the health of the server
// - error: an error if the health check fails
func (s *server) HealthCheck(ctx context.Context, req *aaliflowkitgrpc.HealthRequest) (*aaliflowkitgrpc.HealthResponse, error) {
	// return a successful health check response
	return &aaliflowkitgrpc.HealthResponse{
		Status: "OK",
	}, nil
}

// GetVersion returns the version of the Aali FlowKit server
//
// Parameters:
// - ctx: the context of the request
// - req: the request to get the version of the server
//
// Returns:
// - aaliflowkitgrpc.VersionResponse: a response containing the version of the server
// - error: an error if the version retrieval fails
func (s *server) GetVersion(ctx context.Context, req *aaliflowkitgrpc.VersionRequest) (*aaliflowkitgrpc.VersionResponse, error) {
	// Get the version from the file
	version := getAaliFlowktiVersion(&logging.ContextMap{})

	// If the version is empty, return an error
	if version == "" {
		return nil, status.Errorf(codes.Internal, "failed to retrieve version")
	}

	// Return the version response
	return &aaliflowkitgrpc.VersionResponse{
		Version: version,
	}, nil
}

// ListFunctions lists all available function from the external functions package
//
// Parameters:
// - ctx: the context of the request
// - req: the request to list all available functions
//
// Returns:
// - aaliflowkitgrpc.ListOfFunctions: a list of all available functions
// - error: an error if the function fails
func (s *server) ListFunctions(ctx context.Context, req *aaliflowkitgrpc.ListFunctionsRequest) (*aaliflowkitgrpc.ListFunctionsResponse, error) {

	// return all available functions
	return &aaliflowkitgrpc.ListFunctionsResponse{Functions: internalstates.AvailableFunctions}, nil
}

// RunFunction runs a function from the external functions package
// The function is identified by the function id
// The function inputs are passed as a list of FunctionInput
//
// Parameters:
// - ctx: the context of the request
// - req: the request to run a function
//
// Returns:
// - aaliflowkitgrpc.FunctionOutputs: the outputs of the function
// - error: an error if the function fails
func (s *server) RunFunction(ctx context.Context, req *aaliflowkitgrpc.FunctionInputs) (output *aaliflowkitgrpc.FunctionOutputs, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("error occured in gRPC server aali-flowkit during RunFunction of '%v': %v", req.Name, r)
		}
	}()

	// get function definition from available functions
	functionDefinition, ok := internalstates.AvailableFunctions[req.Name]
	if !ok {
		return nil, fmt.Errorf("function with name %s not found", req.Name)
	}

	// create input slice
	inputs := make([]interface{}, len(functionDefinition.Input))

	// unmarshal json string values for each input into the correct type
	for i, input := range req.Inputs {
		var err error
		inputs[i], err = typeconverters.ConvertStringToGivenType(input.Value, functionDefinition.Input[i].GoType)
		if err != nil {
			return nil, fmt.Errorf("error converting input '%s' of function '%s' to type '%s': %v", input.Name, req.Name, functionDefinition.Input[i].GoType, err)
		}

		// check for option sets and convert values
		if len(functionDefinition.Input[i].Options) > 0 {
			// convert value to correct type
			inputs[i], err = convertOptionSetValues(functionDefinition.Name, input.Name, inputs[i])
			if err != nil {
				return nil, fmt.Errorf("error converting option set input '%s' of function '%s' to type '%s': %v", input.Name, req.Name, functionDefinition.Input[i].GoType, err)
			}
		}
	}

	// get externalfunctions package and the function
	function, exists := externalfunctions.ExternalFunctionsMap[functionDefinition.Name]
	if !exists {
		return nil, fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}
	funcValue := reflect.ValueOf(function)
	if !funcValue.IsValid() {
		return nil, fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}

	// Prepare arguments for the function
	args := []reflect.Value{}
	for i, input := range inputs {
		if input == nil && i < len(functionDefinition.Input) {
			// Handle missing parameters with standard datatype defaults
			expectedType := functionDefinition.Input[i].GoType
			switch expectedType {
			case "bool":
				args = append(args, reflect.ValueOf(false))
			case "int":
				args = append(args, reflect.ValueOf(0))
			case "string":
				args = append(args, reflect.ValueOf(""))
			case "float64":
				args = append(args, reflect.ValueOf(0.0))
			case "map[uint]float32":
				args = append(args, reflect.ValueOf(make(map[uint]float32)))
			default:
				// Catch-all default for other types
				args = append(args, reflect.Zero(reflect.TypeOf(input)))
			}
		} else {
			args = append(args, reflect.ValueOf(input))
		}
	}

	// Call the function
	results := funcValue.Call(args)

	// create output slice
	outputs := []*aaliflowkitgrpc.FunctionOutput{}
	for i, result := range results {
		// marshal value to json string
		value, err := typeconverters.ConvertGivenTypeToString(result.Interface(), functionDefinition.Output[i].GoType)
		if err != nil {
			return nil, fmt.Errorf("error converting output %s to string: %v", functionDefinition.Output[i].Name, err)
		}

		// append output to slice
		outputs = append(outputs, &aaliflowkitgrpc.FunctionOutput{
			Name:   functionDefinition.Output[i].Name,
			GoType: functionDefinition.Output[i].GoType,
			Value:  value,
		})
	}

	// return outputs
	return &aaliflowkitgrpc.FunctionOutputs{Name: req.Name, Outputs: outputs}, nil
}

// StreamFunction streams a function from the external functions package
// The function is identified by the function id
// The function inputs are passed as a list of FunctionInput
//
// Parameters:
// - req: the request to stream a function
// - stream: the stream to send the function outputs
//
// Returns:
// - error: an error if the function fails
func (s *server) StreamFunction(req *aaliflowkitgrpc.FunctionInputs, stream aaliflowkitgrpc.ExternalFunctions_StreamFunctionServer) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("error occured in gRPC server aali-flowkit during StreamFunction of '%v': %v", req.Name, r)
		}
	}()

	// get function definition from available functions
	functionDefinition, ok := internalstates.AvailableFunctions[req.Name]
	if !ok {
		return fmt.Errorf("function with id %s not found", req.Name)
	}

	// create input slice
	inputs := make([]interface{}, len(functionDefinition.Input))

	// unmarshal json string values for each input into the correct type
	for i, input := range req.Inputs {
		var err error
		inputs[i], err = typeconverters.ConvertStringToGivenType(input.Value, functionDefinition.Input[i].GoType)
		if err != nil {
			return fmt.Errorf("error converting input '%s' of function '%s' to type '%s': %v", input.Name, req.Name, functionDefinition.Input[i].GoType, err)
		}

		// check for option sets and convert values
		if len(functionDefinition.Input[i].Options) > 0 {
			// convert value to correct type
			inputs[i], err = convertOptionSetValues(functionDefinition.Name, input.Name, inputs[i])
			if err != nil {
				return fmt.Errorf("error converting option set input '%s' of function '%s' to type '%s': %v", input.Name, req.Name, functionDefinition.Input[i].GoType, err)
			}
		}
	}

	// get externalfunctions package and the function
	function, exists := externalfunctions.ExternalFunctionsMap[functionDefinition.Name]
	if !exists {
		return fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}
	funcValue := reflect.ValueOf(function)
	if !funcValue.IsValid() {
		return fmt.Errorf("function %s not found in externalfunctions package", functionDefinition.Name)
	}

	// Prepare arguments for the function
	args := []reflect.Value{}
	for i, input := range inputs {
		if input == nil && i < len(functionDefinition.Input) {
			// Handle missing parameters with standard datatype defaults
			expectedType := functionDefinition.Input[i].GoType
			switch expectedType {
			case "bool":
				args = append(args, reflect.ValueOf(false))
			case "int":
				args = append(args, reflect.ValueOf(0))
			case "string":
				args = append(args, reflect.ValueOf(""))
			case "float64":
				args = append(args, reflect.ValueOf(0.0))
			case "map[uint]float32":
				args = append(args, reflect.ValueOf(make(map[uint]float32)))
			default:
				// Catch-all default for other types
				args = append(args, reflect.Zero(reflect.TypeOf(input)))
			}
		} else {
			args = append(args, reflect.ValueOf(input))
		}
	}

	// Call the function
	results := funcValue.Call(args)

	// get stream channel from results
	var streamChannel *chan string
	for i, output := range functionDefinition.Output {
		if output.GoType == "*chan string" {
			streamChannel = results[i].Interface().(*chan string)
		}
	}

	// listen to channel and send to stream
	var counter int32
	var previousOutput *aaliflowkitgrpc.StreamOutput
	for message := range *streamChannel {
		// create output
		output := &aaliflowkitgrpc.StreamOutput{
			MessageCounter: counter,
			IsLast:         false,
			Value:          message,
		}

		// send output to stream
		if counter > 0 {
			err := stream.Send(previousOutput)
			if err != nil {
				return err
			}
		}

		// save output to previous output
		previousOutput = output

		// increment counter
		counter++
	}

	// send last message
	output := &aaliflowkitgrpc.StreamOutput{
		MessageCounter: counter,
		IsLast:         true,
		Value:          previousOutput.Value,
	}
	err = stream.Send(output)
	if err != nil {
		return err
	}

	return nil
}

// convertOptionSetValues converts the option set values for the given function and input
//
// Parameters:
// - functionName: a string containing the function name
// - inputName: a string containing the input name
// - inputValue: an interface containing the input value
//
// Returns:
// - interface: an interface containing the converted value
// - error: an error containing the error message
func convertOptionSetValues(functionName string, inputName string, inputValue interface{}) (interface{}, error) {
	defer func() {
		r := recover()
		if r != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Panic occured in convertOptionSetValues: %v", r)
		}
	}()

	switch functionName {

	case "AppendMessageHistory":

		switch inputName {

		case "role":
			return externalfunctions.AppendMessageHistoryRole(inputValue.(string)), nil

		default:
			return nil, fmt.Errorf("unsupported input for function %v: '%s'", functionName, inputName)
		}
	}

	return nil, fmt.Errorf("unsupported function: '%s'", functionName)
}

// getAaliFlowktiVersion reads the agent's version from a file and returns the value.
//
// Returns:
//   - string: Version
func getAaliFlowktiVersion(ctx *logging.ContextMap) string {
	// Read the version from a file; the file is expected to be at ROOT level and called VERSION
	file := "VERSION"
	versionFile, err := os.ReadFile(file)
	if err != nil {
		logging.Log.Errorf(ctx, "Error reading version file: %s\n", err)
		return ""
	}

	version := strings.TrimSpace(string(versionFile))
	return version
}
