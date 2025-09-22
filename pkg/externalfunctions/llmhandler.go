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
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

// PerformVectorEmbeddingRequest performs a vector embedding request to LLM
//
// Tags:N
//   - @displayName: Embeddings
//
// Parameters:
//   - input: the input string
//   - includeSparse: flag to include sparse vectors (false for dense-only, true for hybrid search)
//
// Returns:
//   - embeddedVector: the embedded vector in float32 format
//   - sparseVector: the sparse embedded vector as term_id->weight map (only when includeSparse=true)
func PerformVectorEmbeddingRequest(input string, includeSparse bool) (embeddedVector []float32, sparseVector map[uint]float32) {
	// Use the provided parameter directly
	shouldIncludeSparse := includeSparse

	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Use hybrid embeddings if requested, otherwise use existing dense-only logic
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, shouldIncludeSparse, nil)
	defer close(responseChannel)

	var denseEmbedding []float32
	var sparseEmbedding map[uint]float32
	var err error

	for response := range responseChannel {
		if response.Type == "error" {
			panic(response.Error)
		}

		fmt.Printf("Received embeddings response.")

		// Process dense embedding (same as before)
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
		denseEmbedding, err = convertToFloat32Slice(interfaceArray)
		if err != nil {
			errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Process sparse embedding if available (added functionality)
		if shouldIncludeSparse && response.LexicalWeights != nil {
			fmt.Print("Processing sparse embedding...\n")
			if sparseInterface, ok := response.LexicalWeights.(map[string]interface{}); ok {
				sparseEmbedding = make(map[uint]float32)
				for key, value := range sparseInterface {
					if keyUint, err := strconv.ParseUint(key, 10, 32); err == nil {
						if valueFloat, ok := value.(float64); ok {
							sparseEmbedding[uint(keyUint)] = float32(valueFloat)
						}
					}
				}
			}
		}

		// Exit after processing first response
		break
	}

	return denseEmbedding, sparseEmbedding
}

// PerformVectorEmbeddingRequestWithTokenLimitCatch performs a vector embedding request to LLM
// and catches the token limit error message
//
// Tags:
//   - @displayName: Embeddings with Token Limit Catch
//
// Parameters:
//   - input: the input string
//
// Returns:
//   - embeddedVector: the embedded vector in float32 format
func PerformVectorEmbeddingRequestWithTokenLimitCatch(input string, tokenLimitMessage string) (embeddedVector []float32, tokenLimitReached bool, responseMessage string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, false, nil)
	defer close(responseChannel)

	// Process the first response and close the channel
	var embedding32 []float32
	var err error
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			if strings.Contains(response.Error.Message, "tokens") {
				return nil, true, tokenLimitMessage
			} else {
				panic(response.Error)
			}
		}

		// Log LLM response
		logging.Log.Debugf(&logging.ContextMap{}, "Received embeddings response.")

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}
		embedding32, err = convertToFloat32Slice(interfaceArray)
		if err != nil {
			errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Mark that the first response has been received
		firstResponseReceived := true

		// Exit the loop after processing the first response
		if firstResponseReceived {
			break
		}
	}

	return embedding32, false, ""
}

// PerformBatchEmbeddingRequest performs a batch vector embedding request to LLM
//
// Tags:
//   - @displayName: Batch Embeddings
//
// Parameters:
//   - input: the input strings
//
// Returns:
//   - embeddedVectors: the embedded vectors in float32 format
func PerformBatchEmbeddingRequest(input []string) (embeddedVectors [][]float32) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send embeddings request
	responseChannel := sendEmbeddingsRequest(input, llmHandlerEndpoint, false, nil)
	defer close(responseChannel)

	// Process the first response and close the channel
	embedding32Array := make([][]float32, len(input))
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Log LLM response
		logging.Log.Debugf(&logging.ContextMap{}, "Received batch embeddings response.")

		// Get embedded vector array
		interfaceArray, ok := response.EmbeddedData.([]interface{})
		if !ok {
			errMessage := "error converting embedded data to interface array"
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		for i, interfaceArrayElement := range interfaceArray {
			lowerInterfaceArray, ok := interfaceArrayElement.([]interface{})
			if !ok {
				errMessage := "error converting embedded data to interface array"
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
			embedding32, err := convertToFloat32Slice(lowerInterfaceArray)
			if err != nil {
				errMessage := fmt.Sprintf("error converting embedded data to float32 slice: %v", err)
				logging.Log.Error(&logging.ContextMap{}, errMessage)
				panic(errMessage)
			}
			embedding32Array[i] = embedding32
		}

		// Mark that the first response has been received
		firstResponseReceived := true

		// Exit the loop after processing the first response
		if firstResponseReceived {
			break
		}
	}

	return embedding32Array
}

// PerformBatchHybridEmbeddingRequest performs a batch hybrid embedding request to LLM
// returning the sparse and dense embeddings
//
// Tags:
//   - @displayName: Batch Hybrid Embeddings
//
// Parameters:
//   - input: the input strings
//
// Returns:
//   - denseEmbeddings: the dense embeddings in float32 format
//   - sparseEmbeddings: the sparse embeddings in map format
func PerformBatchHybridEmbeddingRequest(input []string, maxBatchSize int) (denseEmbeddings [][]float32, sparseEmbeddings []map[uint]float32) {
	processedEmbeddings := 0

	// Process data in batches
	for i := 0; i < len(input); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(input) {
			end = len(input)
		}

		// Create a batch of data to send to LLM handler
		batchTextToEmbed := input[i:end]

		// Send http request
		batchDenseEmbeddings, batchLexicalWeights, err := llmHandlerPerformVectorEmbeddingRequest(batchTextToEmbed, true)
		if err != nil {
			errMessage := fmt.Sprintf("Error performing batch embedding request: %v", err)
			logging.Log.Error(&logging.ContextMap{}, errMessage)
			panic(errMessage)
		}

		// Add the embeddings to the list
		denseEmbeddings = append(denseEmbeddings, batchDenseEmbeddings...)
		sparseEmbeddings = append(sparseEmbeddings, batchLexicalWeights...)

		processedEmbeddings += len(batchTextToEmbed)
		logging.Log.Debugf(&logging.ContextMap{}, "Processed %d embeddings", processedEmbeddings)
	}

	return denseEmbeddings, sparseEmbeddings
}

// PerformKeywordExtractionRequest performs a keywords extraction request to LLM
//
// Tags:
//   - @displayName: Keyword Extraction
//
// Parameters:
//   - input: the input string
//   - maxKeywordsSearch: the maximum number of keywords to search for
//
// Returns:
//   - keywords: the keywords extracted from the input string as a slice of strings
func PerformKeywordExtractionRequest(input string, maxKeywordsSearch uint32) (keywords []string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequestNoHistory(input, "keywords", maxKeywordsSearch, llmHandlerEndpoint, nil, nil)
	defer close(responseChannel)

	// Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Received keywords response.")

	// Unmarshal JSON data into the result variable
	err := json.Unmarshal([]byte(responseAsStr), &keywords)
	if err != nil {
		errMessage := fmt.Sprintf("Error unmarshalling keywords response from aali-llm: %v", err)
		logging.Log.Error(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	for words := range keywords {
		logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Keyword %v ", words)
	}

	// Return the response
	return keywords
}

// PerformSummaryRequest performs a summary request to LLM
//
// Tags:
//   - @displayName: Summary
//
// Parameters:
//   - input: the input string
//
// Returns:
//   - summary: the summary extracted from the input string
func PerformSummaryRequest(input string) (summary string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequestNoHistory(input, "summary", 1, llmHandlerEndpoint, nil, nil)
	defer close(responseChannel)

	// Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	logging.Log.Debugf(&logging.ContextMap{}, "Received summary response.")

	// Return the response
	return responseAsStr
}

// PerformGeneralRequest performs a general chat completion request to LLM
//
// Tags:
//   - @displayName: General LLM Request
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - isStream: the stream flag
//   - systemPrompt: the system prompt
//
// Returns:
//   - message: the generated message
//   - stream: the stream channel
func PerformGeneralRequest(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil, nil, nil)
	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", "", false, "")

		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Return the response
	return responseAsStr, nil
}

// PerformGeneralRequestWithImages performs a general request to LLM with images
//
// Tags:
//   - @displayName: General LLM Request (with Images)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - images: the images
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralRequestWithImages(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, images []string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil, nil, images)
	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", "", false, "")

		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Return the response
	return responseAsStr, nil
}

// PerformGeneralModelSpecificationRequest performs a specified request to LLM with a configured model and Systemprompt.
//
// Tags:
//   - @displayName: General LLM Request (Specified System Prompt)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - modelId: the model ID
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralModelSpecificationRequest(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt map[string]string, modelIds []string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	fmt.Printf("[%s] type of alpsRequest inside modelspecification %T\n", time.Now().Format("2006-01-02 15:04:05.000"), systemPrompt)
	logging.Log.Infof(&logging.ContextMap{}, "[%s] type of alpsRequest inside modelspecification %T\n", time.Now().Format("2006-01-02 15:04:05.000"), systemPrompt)

	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT
	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", "", false, "")

		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Return the response
	return responseAsStr, nil
}

// PerformGeneralRequestSpecificModel performs a general request to LLM with a specific model
//
// Tags:
//   - @displayName: General LLM Request (Specific Models)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - modelId: the model ID
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralRequestSpecificModel(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, modelIds []string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", "", false, "")

		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Return the response
	return responseAsStr, nil
}

// PerformGeneralRequestSpecificModel performs a general request to LLM with a specific model
//
// Tags:
//   - @displayName: General LLM Request (Specific Models & Model Options)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - modelId: the model ID
//   - modelOptions: the model options
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralRequestSpecificModelAndModelOptions(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, modelIds []string, modelOptions sharedtypes.ModelOptions) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, &modelOptions, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", "", false, "")

		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Return the response
	return responseAsStr, nil
}

// PerformGeneralRequestSpecificModelModelOptionsAndImages performs a general request to LLM with a specific model including model options and images
//
// Tags:
//   - @displayName: General LLM Request (Specific Models, Model Options & Images)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - isStream: the flag to indicate whether the response should be streamed
//   - systemPrompt: the system prompt
//   - modelId: the model ID
//   - modelOptions: the model options
//   - images: the images to include in the request
//
// Returns:
//   - message: the response message
//   - stream: the stream channel
func PerformGeneralRequestSpecificModelModelOptionsAndImages(input string, history []sharedtypes.HistoricMessage, isStream bool, systemPrompt string, modelIds []string, modelOptions sharedtypes.ModelOptions, images []string, modelCategory []string) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, modelCategory, &modelOptions, images)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		// Create a stream channel
		streamChannel := make(chan string, 400)

		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, false, false, "", 0, 0, "", "", "", false, "")

		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// Return the response
	return responseAsStr, nil
}

// PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput performs a general request to LLM with a specific model
// and returns the token count using OpenAI token count model. Does not stream the response.
//
// Tags:
//   - @displayName: General LLM Request (Specific Models, No Stream, OpenAI Token Output)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - systemPrompt: the system prompt
//   - modelIds: the model IDs of the AI models to use
//   - tokenCountModelName: the model name to use for token count
//
// Returns:
//   - message: the response message
//   - tokenCount: the token count
func PerformGeneralRequestSpecificModelNoStreamWithOpenAiTokenOutput(input string, history []sharedtypes.HistoricMessage, systemPrompt string, modelIds []string, tokenCountModelName string) (message string, tokenCount int) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, nil, nil)
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// get input token count
	totalTokenCount, err := openAiTokenCount(tokenCountModelName, input+systemPrompt)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting input token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
		panic(errorMessage)
	}

	// get history token count
	for _, message := range history {
		historyTokenCount, err := openAiTokenCount(tokenCountModelName, message.Content)
		if err != nil {
			errorMessage := fmt.Sprintf("Error getting history token count: %v", err)
			logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
			panic(errorMessage)
		}
		totalTokenCount += historyTokenCount
	}

	// get the output token count
	outputTokenCount, err := openAiTokenCount(tokenCountModelName, responseAsStr)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting output token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
		panic(errorMessage)
	}
	totalTokenCount += outputTokenCount

	// log token count
	logging.Log.Debugf(&logging.ContextMap{}, "Total token count: %d", totalTokenCount)

	// Return the response
	return responseAsStr, totalTokenCount
}

// PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiTokenOutput performs a general request to LLM with a specific model
// and model options, and returns the token count using OpenAI token count model. Does not stream the response.
//
// Tags:
//   - @displayName: General LLM Request (Specific Models, Model Options, No Stream, OpenAI Token Output)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - systemPrompt: the system prompt
//   - modelIds: the model IDs of the AI models to use
//   - modelOptions: the model options
//   - tokenCountModelName: the model name to use for token count
//
// Returns:
//   - message: the response message
//   - tokenCount: the token count
func PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiTokenOutput(input string, history []sharedtypes.HistoricMessage, systemPrompt string, modelIds []string, modelOptions sharedtypes.ModelOptions, tokenCountModelName string) (message string, tokenCount int) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, &modelOptions, nil)
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// get input token count
	totalTokenCount, err := openAiTokenCount(tokenCountModelName, input+systemPrompt)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting input token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
		panic(errorMessage)
	}

	// get history token count
	for _, message := range history {
		historyTokenCount, err := openAiTokenCount(tokenCountModelName, message.Content)
		if err != nil {
			errorMessage := fmt.Sprintf("Error getting history token count: %v", err)
			logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
			panic(errorMessage)
		}
		totalTokenCount += historyTokenCount
	}

	// get the output token count
	outputTokenCount, err := openAiTokenCount(tokenCountModelName, responseAsStr)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting output token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
		panic(errorMessage)
	}
	totalTokenCount += outputTokenCount

	// log token count
	logging.Log.Debugf(&logging.ContextMap{}, "Total token count: %d", totalTokenCount)

	// Return the response
	return responseAsStr, totalTokenCount
}

// PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiInputOutputTokenOutput performs a general request to LLM with a specific model
// and model options, and returns the token count using OpenAI token count model. Does not stream the response.
//
// Tags:
//   - @displayName: General LLM Request (Specific Models, Model Options, No Stream, OpenAI Input & Output Token Output)
//
// Parameters:
//   - input: the user input
//   - history: the conversation history
//   - systemPrompt: the system prompt
//   - modelIds: the model IDs of the AI models to use
//   - modelOptions: the model options
//   - tokenCountModelName: the model name to use for token count
//
// Returns:
//   - message: the response message
//   - inputTokenCount: the input token count
//   - outputTokenCount: the output token count
func PerformGeneralRequestSpecificModelAndModelOptionsNoStreamWithOpenAiInputOutputTokenOutput(input string, history []sharedtypes.HistoricMessage, systemPrompt string, modelIds []string, modelOptions sharedtypes.ModelOptions, tokenCountModelName string) (message string, inputTokenCount int, outputTokenCount int) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, modelIds, nil, &modelOptions, nil)
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	// get input token count
	var err error
	inputTokenCount, err = openAiTokenCount(tokenCountModelName, input+systemPrompt)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting input token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
		panic(errorMessage)
	}
	for _, message := range history {
		historyTokenCount, err := openAiTokenCount(tokenCountModelName, message.Content)
		if err != nil {
			errorMessage := fmt.Sprintf("Error getting history token count: %v", err)
			logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
			panic(errorMessage)
		}
		inputTokenCount += historyTokenCount
	}

	// get the output token count
	outputTokenCount, err = openAiTokenCount(tokenCountModelName, responseAsStr)
	if err != nil {
		errorMessage := fmt.Sprintf("Error getting output token count: %v", err)
		logging.Log.Errorf(&logging.ContextMap{}, "%v", errorMessage)
		panic(errorMessage)
	}

	// log token count
	logging.Log.Debugf(&logging.ContextMap{}, "Input token count: %d; Output token count: %d", inputTokenCount, outputTokenCount)

	// Return the response
	return responseAsStr, inputTokenCount, outputTokenCount
}

func parseAPINames(input string) (listApis []string) {
	// expected input
	// `python \n api1,api2,api3\n`
	libContext := "ansys.aedt.core"
	aedtApps := []string{"desktop", "hfss", "maxwell", "circuit", "icepak", "hfss3dlayout", "mechanical", "rmxprt", "emit", "maxwellcircuit"}
	clearList := strings.ReplaceAll(input, "\n", "")
	clearList = strings.ReplaceAll(clearList, "```", "")
	clearList = strings.ReplaceAll(clearList, "plaintext", "")
	clearList = strings.ReplaceAll(clearList, "python", "")
	clearList = strings.ReplaceAll(clearList, "```", "")
	listRawApis := strings.Split(clearList, ",") // ignore python
	// add suffix ansys.aedt.core ?
	// we get imports and function names here
	// map it as class to function names
	for _, api := range listRawApis {
		api = strings.TrimSpace(api)
		api = strings.TrimPrefix(api, "\"")
		api = strings.TrimSuffix(api, "\"")
		api = strings.Trim(api, "`")
		//logging.Log.Debugf(&logging.ContextMap{}, "processing %s", api)

		if strings.Contains(api, libContext) {
			// get class names
			// create absolute name
			//pass
		} else {
			// get function name
			temp := strings.Split(api, ".")
			if len(temp) > 0 {
				funcName := temp[len(temp)-1]
				//check is aaedt app
				for _, app := range aedtApps {
					if strings.ToLower(funcName) == app {
						funcName = libContext + "." + strings.ToLower(funcName) + "." + funcName + ".__init__"
					}
				}
				listApis = append(listApis, funcName)
			}
		}

	}
	return listApis
}

// PyaedtCodeValidationLoop performs a code validation request to LLM
//
// Tags:
//   - @displayName: Pyaedt Code Validation loop
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - isStream: the stream flag
//
// Returns:
//   - message: the generated code
//   - stream: the stream channel
func PyaedtCodeValidationLoop(input string, history []sharedtypes.HistoricMessage, isStream bool, validateCode bool) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "code", history, 0, "", llmHandlerEndpoint, nil, nil, nil, nil)
	var responseAsStr string
	validateCode = true
	validationCount := 2
	var pythonCodeTemp string
	var latestAPISignatures []string
	for cnt := range validationCount {
		for response := range responseChannel {
			// Check if the response is an error
			if response.Type == "error" {
				panic(response.Error)
			}

			// Accumulate the responses
			responseAsStr += *(response.ChatData)

			// If we are at the last message, break the loop
			if *(response.IsLast) {
				break
			}
		}

		// Extract the code from the response

		pythonCode, err := extractPythonCode(responseAsStr)
		pythonCodeTemp = pythonCode
		// kapatil:
		// Get latest API signatures for all
		//var latestAPISignatures []string
		listAPIPrompt := "For following code, list only apis as comma separated values and do  not explain anyting"
		//listAPIPrompt += "Python Code:\n"
		listAPIPrompt += responseAsStr
		logging.Log.Debugf(&logging.ContextMap{}, "**Query APIs for %s", listAPIPrompt)
		// API list LLM and send chat request
		responseApiList := sendChatRequestNoStreaming(listAPIPrompt, "code", nil, 0, "", llmHandlerEndpoint, nil, nil, nil, nil)
		apisUsed := parseAPINames(responseApiList)
		logging.Log.Debugf(&logging.ContextMap{}, "Apis read: %v", apisUsed)
		latestAPISignatures = GetLatestApiSignaturesForApis(apisUsed)

		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error extracting Python code: %v, Couldn't validate code", err)
			break
		} else {
			// Validate the Python code
			valid, _, err := validatePythonCode(pythonCode)
			logging.Log.Debugf(&logging.ContextMap{}, "Python code is valid %v", valid)
			if valid {
				break
			}
			if err != nil {
				// parse errors
				// kapatil: redo code generation
				// Prompt: Following errors are found in code, fix code w.r.t pyaedt code library
				errPrompt := GetValidationPrompt(err.Error(), latestAPISignatures)
				time.Sleep(3 * time.Second)
				if errPrompt != "" {
					errPrompt += "Pyaedt script:\n " + pythonCode
					// Set up WebSocket connection with LLM and send chat request
					responseChannel = sendChatRequest(errPrompt, "code", history, 0, "", llmHandlerEndpoint, nil, nil, nil, nil)
					logging.Log.Debugf(&logging.ContextMap{}, "Request review : %v", errPrompt)
				}
			} else {
				break
			}
		}
		logging.Log.Debugf(&logging.ContextMap{}, "***Validation Loop %d************************", cnt)
	} //validationloop
	logging.Log.Debugf(&logging.ContextMap{}, "Validation done!")
	tempPrompt := "return this python code no explaination\n" + pythonCodeTemp
	responseChannel = sendChatRequest(tempPrompt, "code", history, 0, "", llmHandlerEndpoint, nil, nil, nil, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {
		logging.Log.Debugf(&logging.ContextMap{}, "Streaming ..")
		validateCode = false
		// Create a stream channel
		streamChannel := make(chan string, 400)
		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, validateCode, false, "", 0, 0, "", "", "", false, "")
		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// Return the response
	return responseAsStr, nil
}

// PerformCodeLLMRequest performs a code generation request to LLM
//
// Tags:
//   - @displayName: Code LLM Request
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - isStream: the stream flag
//
// Returns:
//   - message: the generated code
//   - stream: the stream channel
func PerformCodeLLMRequest(input string, history []sharedtypes.HistoricMessage, isStream bool, validateCode bool) (message string, stream *chan string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseChannel := sendChatRequest(input, "code", history, 0, "", llmHandlerEndpoint, nil, nil, nil, nil)

	// If isStream is true, create a stream channel and return asap
	if isStream {

		// Create a stream channel
		streamChannel := make(chan string, 400)
		// Start a goroutine to transfer the data from the response channel to the stream channel
		go transferDatafromResponseToStreamChannel(&responseChannel, &streamChannel, validateCode, false, "", 0, 0, "", "", "", false, "")
		// Return the stream channel
		return "", &streamChannel
	}

	// Close the response channel
	defer close(responseChannel)

	// else Process all responses
	var responseAsStr string
	for response := range responseChannel {
		// Check if the response is an error
		if response.Type == "error" {
			panic(response.Error)
		}

		// Accumulate the responses
		responseAsStr += *(response.ChatData)

		// If we are at the last message, break the loop
		if *(response.IsLast) {
			break
		}
	}

	validateCode = true
	// Code validation
	if validateCode {

		// Extract the code from the response
		pythonCode, err := extractPythonCode(responseAsStr)
		if err != nil {
			logging.Log.Errorf(&logging.ContextMap{}, "Error extracting Python code: %v", err)
		} else {

			// Validate the Python code
			valid, warnings, err := validatePythonCode(pythonCode)
			if err != nil {
				logging.Log.Debugf(&logging.ContextMap{}, "Error validating code")
			} else {
				if valid {
					if warnings {
						responseAsStr += "\nCode has warnings."
					} else {
						responseAsStr += "\nCode is valid."
					}
				} else {

					responseAsStr += "\nCode is invalid."
				}
			}
		}
	}
	// Return the response
	return responseAsStr, nil
}

// PerformGeneralRequestNoStreaming performs a general chat completion request to LLM without streaming
//
// Tags:
//   - @displayName: General LLM Request (no streaming)
//
// Parameters:
//   - input: the input string
//   - history: the conversation history
//   - systemPrompt: the system prompt
//
// Returns:
//   - message: the generated message
func PerformGeneralRequestNoStreaming(input string, history []sharedtypes.HistoricMessage, systemPrompt string) (message string) {
	// get the LLM handler endpoint
	llmHandlerEndpoint := config.GlobalConfig.LLM_HANDLER_ENDPOINT

	// Set up WebSocket connection with LLM and send chat request
	responseString := sendChatRequestNoStreaming(input, "general", history, 0, systemPrompt, llmHandlerEndpoint, nil, nil, nil, nil)

	// Return the response
	return responseString
}

// BuildLibraryContext builds the context string for the query
//
// Tags:
//   - @displayName: Library Context
//
// Parameters:
//   - message: the message string
//   - libraryContext: the library context string
//
// Returns:
//   - messageWithContext: the message with context
func BuildLibraryContext(message string, libraryContext string) (messageWithContext string) {
	// Check if "pyansys" is in the library context
	message = libraryContext + " " + message

	return message
}

// BuildFinalQueryForGeneralLLMRequest builds the final query for a general
// request to LLM. The final query is a markdown string that contains the
// original request and the examples from the KnowledgeDB.
//
// Tags:
//   - @displayName: Final Query (General LLM Request)
//
// Parameters:
//   - request: the original request
//   - knowledgedbResponse: the KnowledgeDB response
//
// Returns:
//   - finalQuery: the final query
func BuildFinalQueryForGeneralLLMRequest(request string, knowledgedbResponse []sharedtypes.DbResponse) (finalQuery string) {

	// If there is no response from the KnowledgeDB, return the original request
	if len(knowledgedbResponse) == 0 {
		return request
	}

	// Build the final query using the KnowledgeDB response and the original request
	finalQuery = "Based on the following examples:\n\n--- INFO START ---\n"
	for _, example := range knowledgedbResponse {
		finalQuery += example.Text + "\n"
	}
	finalQuery += "--- INFO END ---\n\n" + request + "\n"

	// Return the final query
	return finalQuery
}

// PyaedtBuildFinalQueryForCodeLLMRequest builds the final query for a code generation
// request to LLM. The final query is a markdown string that contains the
// original request and the code examples from the KnowledgeDB.
//
// Tags:
//   - @displayName: Pyaedt Final Query (Code LLM Request)
//
// Parameters:
//   - request: the original request
//   - knowledgedbResponse: the KnowledgeDB response
//   - userGuideSearch: include user guide citations
//   - citations: citations string
//   - elementContext: String context prompt
//   - design context: context from the active design
//
// Returns:
//   - finalQuery: the final query
func PyaedtBuildFinalQueryForCodeLLMRequest(request string, knowledgedbResponse []sharedtypes.ExampleDbResponse, userGuideSearch bool, citations []string, elementContexts []string, designContext string) (finalQuery string) {
	// Build the final query using the KnowledgeDB response and the original request
	// We have to use the text from the DB response and the original request.
	//
	// Design context is a string that we get from the AEDT session. It contains
	// information about the current design, project, application, and PyAEDT version.
	// It is in the following format:
	// "{'designContext': {'design': 'MyDesign', 'project': 'MyProject',
	// 'selections': [], 'application': 'MyApplication',
	// 'pyaedtVersion': '0.xx.x', 'type': 'Generic', 'units': 'xx'}}"
	//
	// The code will parse this string and extract the information to the final prompt.
	// {'designContext':
	// 		{
	// 			'design': 'MyDesign',
	// 			'project': 'MyProject',
	// 			'selections': [],
	// 			'application': 'MyApplication',
	// 			'pyaedtVersion': '0.xx.x',
	// 			'type': 'Generic',
	// 			'units': 'xxx'
	// 		}
	// }
	//
	// The prompt should be in the following format:
	//
	// ******************************************************************************
	/// <Element Context here>
	//  Based on the following reference links and examples:
	// --- REFERENCE LINKS START ---
	// {citation_url_1},
	// {citation_url_2},
	// {citation_url_3}
	// --- END REFERENCE LINKS --
	//
	// --- START EXAMPLE {response_n}---
	// >>> Summary:
	// {knowledge_db_response_n_summary}
	//
	// >>> Code snippet:
	// ```python
	// {knowledge_db_response_n_text}
	// ```
	// --- END EXAMPLE {response_n}---
	//
	// --- START EXAMPLE {response_n}---
	// ...
	// --- END EXAMPLE {response_n}---
	//

	// Generate the Python code for the following request: {original request}
	//
	// Hard requirements (do not violate):
	// - Include **all imports** actually used. Follow the template for {Pyaedt version}}: {import template}
	// - Provide an **Initialization** section that **explicitly** declares the known information as follows:
	//   - Use PyAEDT version: {Pyaedt version}
	//   - AEDT version: {AEDT version}
	//   - Design name: {Design name}
	//   - Application: {Application name}
	//   - Project name: {Project name}
	//   - Selections: {selections, if any}
	//
	// The following statements are examples of how to initialize different applications, refer to these examples and initialization accordingly:
	//  - {application name}: {initialization template}
	//	- ... (other applications as applicable)
	// ******************************************************************************

	// Construct final query prompt.
	finalQuery = "You are a Python expert with experience in writing complete, functional PyAEDT scripts. These scripts typically include python code for tasks such as geometry creation, boundary setup, and analysis setups - especially for HFSS (or other AnsysEM tools as applicable). Your task is to write valid Python code using PyAEDT APIs.\n"
	if len(elementContexts) > 0 {
		// assuming we get the first element context only
		finalQuery += elementContexts[0]

	}

	// Get the citations from the user guide search
	if userGuideSearch {
		finalQuery += "\nBased on the following PyAEDT documentation links: \n\n"

		for i, citation := range citations {
			finalQuery += "--- REFERENCE LINKS START " + fmt.Sprint(i+1) + " ---\n"
			finalQuery += citation + "\n"
			finalQuery += "--- END REFERENCE LINKS " + fmt.Sprint(i+1) + " ---\n\n"
		}
		finalQuery += "And following examples:\n\n"
	} else {
		finalQuery += "\nBased on the following examples:\n\n"

	}

	// Get the examples from the knowledge DB
	if len(knowledgedbResponse) > 0 {
		for i, element := range knowledgedbResponse {
			// Add the example number
			// logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Reading knowledge DB response")
			finalQuery += "--- START EXAMPLE " + fmt.Sprint(i+1) + "---\n"
			finalQuery += "* Summary:\n" + element.Summary + "\n\n"
			finalQuery += "* Code snippet:\n```python\n" + element.Text + "\n```\n"
			finalQuery += "--- END EXAMPLE " + fmt.Sprint(i+1) + "---\n\n"
			// logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Initial Query %s", finalQuery)

		}
	} else {
		logging.Log.Debugf(&logging.ContextMap{}, "No relevant examples found in DB.")
	}

	// Kaumudi: Rephrase
	//newRequest := RephraseRequest_kapatil(request)
	newRequest := request

	// Pass in the original request without blank in the front and end
	finalQuery += "Generate the Python code for the following request: **" + strings.TrimSpace(newRequest) + "** \n"

	// Convert designContext to a JSON format: map[string]any
	convertDesignContext := func(designContext string, format string) (any, error) {
		// Replace single quotes with double quotes for valid JSON
		designContext = strings.ReplaceAll(designContext, "'", "\"")

		// Fix newline characters in string literals by escaping them
		designContext = strings.ReplaceAll(designContext, "\n", "\\n")

		// Parse the JSON string into a map
		var contextData map[string]interface{}
		err := json.Unmarshal([]byte(designContext), &contextData)
		if err != nil {
			return "", fmt.Errorf("failed to parse designContext: %v", err)
		}

		if format == "JSON" {
			// Convert back to JSON with indent 2.
			jsonBytes, err := json.MarshalIndent(contextData, "", "  ")
			if err != nil {
				return "", fmt.Errorf("failed to marshal to JSON: %v", err)
			}
			return string(jsonBytes), nil
		} else if format == "Map" {
			// Convert back to map[string]any format
			result := make(map[string]any)
			for key, value := range contextData {
				result[key] = value
			}
			return result, nil
		} else {
			return "", fmt.Errorf("unknown format: %s", format)
		}

	}

	var generationType, design, project, application, pyaedtVersion string
	var selections []string
	if designContext == "" {
		logging.Log.Info(&logging.ContextMap{}, "No design context provided. Using default strings for design, project, application, and pyaedtVersion.")
		design = "MyDesign"
		project = "MyProject"
		application = "MyApplication"
		pyaedtVersion = "0.19.0" // Default version: the latest one by Sep 2025.
		selections = []string{}
	} else {
		// Cutoff designContext and only process generic context.
		pattern := `'type'\s*:\s*'[^']*'`

		// Use regex to find the pattern
		re := regexp.MustCompile(pattern)
		match := re.FindStringIndex(designContext)

		if match == nil {
			// If pattern not found, try with double quotes format
			pattern = `"type"\s*:\s*"[^"]*"`
			re = regexp.MustCompile(pattern)
			match = re.FindStringIndex(designContext)

			if match == nil {
				logging.Log.Warnf(&logging.ContextMap{}, "Cutoff pattern 'type' field not found in designContext")
				return designContext
			}
		}

		// Get the end position of the match (after the 'type' field and its value)
		endPos := match[1]

		// Extract substring up to the end of the 'type' field
		designContextGeneric := designContext[:endPos]

		// Add proper closing braces
		designContextGeneric += "}}"

		// Convert designContextGeneric to map[string]any
		designContextMap, err := convertDesignContext(designContextGeneric, "Map")
		if err != nil {
			logging.Log.Warn(&logging.ContextMap{}, "Failed to convert designContext to map: %v", err)
			designContextMap = make(map[string]any)
		} else {
			// Successfully converted designContext to map
			logging.Log.Debugf(&logging.ContextMap{}, "Successfully converted designContext to map: %v", designContextMap)
		}

		if nestedContext, ok := designContextMap.(map[string]any)["designContext"].(map[string]any); ok {
			// Extract basic context information.
			if val, ok := nestedContext["type"]; ok {
				if strVal, ok := val.(string); ok {
					generationType = strVal

					logging.Log.Info(&logging.ContextMap{}, "Design context generation type: %s", generationType)
				}
			} else {
				logging.Log.Error(&logging.ContextMap{}, "Missing generation type in design context.")
			}

			// Extract design name
			if val, ok := nestedContext["design"]; ok {
				if strVal, ok := val.(string); ok {
					design = strVal
				}
			} else {
				logging.Log.Debugf(&logging.ContextMap{}, "No design name found in design context. Using default.")
				design = "MyDesign"
			}

			// Extract project name.
			if val, ok := nestedContext["project"]; ok {
				if strVal, ok := val.(string); ok {
					project = strVal
				}
			} else {
				logging.Log.Debugf(&logging.ContextMap{}, "No project name found in design context. Using default.")
				project = "MyProject"
			}

			// Extract application name.
			if val, ok := nestedContext["application"]; ok {
				if strVal, ok := val.(string); ok {
					application = strVal
				}
			} else {
				logging.Log.Debugf(&logging.ContextMap{}, "No application name found in design context. Using default.")
				application = "MyApplication"
			}

			// Extract PyAEDT version.
			if val, ok := nestedContext["pyaedtVersion"]; ok {
				if strVal, ok := val.(string); ok {
					pyaedtVersion = strVal
				}
			} else {
				logging.Log.Debugf(&logging.ContextMap{}, "No PyAEDT version found in design context. Using default.")
				pyaedtVersion = "0.19.0"
			}

			// Extract selections.
			if val, ok := nestedContext["selections"]; ok {
				if interfaceSlice, ok := val.([]interface{}); ok {
					selections = make([]string, 0, len(interfaceSlice))
					for _, item := range interfaceSlice {
						if strItem, ok := item.(string); ok {
							selections = append(selections, strItem)
						} else {
							logging.Log.Warnf(&logging.ContextMap{}, "Selection item is not a string: %v (type: %T)", item, item)
						}
					}
				} else if sliceVal, ok := val.([]string); ok {
					selections = sliceVal
				} else {
					logging.Log.Warnf(&logging.ContextMap{}, "Selections field is not a slice, found type: %T, value: %v", val, val)
					selections = []string{}
				}
			} else {
				logging.Log.Debugf(&logging.ContextMap{}, "No selections found in design context. Using default.")
				selections = []string{}
			}
		} else {
			logging.Log.Error(&logging.ContextMap{}, "Missing generation type in design context.")
		}

		// Store designContext to a JSON file.
		dumpJSONToFile := func(jsonData, filename string) error {
			// Create the file
			file, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer file.Close()

			// Write JSON data to file
			_, err = file.WriteString(jsonData)
			if err != nil {
				return fmt.Errorf("failed to write to file: %v", err)
			}

			return nil
		}

		// Store designContext to a JSON file.
		// TODO: accumulate design contexts and store them to a single file with timestamp? Or overwrite the previous one?
		// For now, overwrite the previous one.
		designContextJSONResult, err := convertDesignContext(designContextGeneric, "JSON")
		if err != nil {
			logging.Log.Warn(&logging.ContextMap{}, "Failed to convert designContext to JSON: %v", err)
			// Use default empty JSON
			err = dumpJSONToFile("{}", "design_context.json")
			if err != nil {
				logging.Log.Warn(&logging.ContextMap{}, "Failed to dump default JSON to file: %v", err)
			}
		} else {
			// Type assert to string
			if designContextJSON, ok := designContextJSONResult.(string); ok {
				logging.Log.Debugf(&logging.ContextMap{}, "Design context as JSON:\n%s", designContextJSON)

				// Dump to file
				fileName := "design_context.json"
				err = dumpJSONToFile(designContextJSON, fileName)
				if err != nil {
					logging.Log.Warn(&logging.ContextMap{}, "Failed to dump JSON to file: %v", err)
				} else {
					logging.Log.Debugf(&logging.ContextMap{}, "Successfully dumped design context JSON to file: %s", fileName)
				}
			} else {
				logging.Log.Warn(&logging.ContextMap{}, "Failed to assert designContext result to string")
				// Fallback to default
				err = dumpJSONToFile("{}", "design_context.json")
				if err != nil {
					logging.Log.Warn(&logging.ContextMap{}, "Failed to dump fallback JSON to file: %v", err)
				}
			}
		}
	}

	// ==============================
	// Imports and initilization templates for different PyAEDT versions
	version_mapper := map[string]string{
		"0.19.0": "2025.1",
	}
	import_templates := map[string]string{
		"0.19.0": "```python\nimport ansys.aedt.core as pyaedt```",
	}
	init_templates := map[string]map[string]string{
		"0.19.0": {
			"Desktop":        "```\nDesktop(version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None)\n```",
			"Hfss":           "```\nHfss(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Q3d":            "```\nQ3d(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Q2d":            "```\nQ2d(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Maxwell2d":      "```\nMaxwell2d(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Maxwell3d":      "```\nMaxwell3d(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Icepak":         "```\nIcepak(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Hfss3dLayout":   "```\nHfss3dLayout(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, ic_mode:bool|None, remove_lock:bool|None)\n```",
			"Mechanical":     "```\nMechanical(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Rmxprt":         "```\nRmxprt(project:str|None, design:str|None, solution_type:str|None, model_units:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Circuit":        "```\nCircuit(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"MaxwellCircuit": "```\nMaxwellCircuit(project:str|None, design:str|None, solution_type:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"Emit":           "```\nEmit(project:str|None, design:str|None, solution_type:str|None, version:str|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
			"TwinBuilder":    "```\nTwinBuilder(project:str|None, design:str|None, solution_type:str|None, setup:str|None, version:str|int|float|None, non_graphical:bool|None, new_desktop:bool|None, close_on_exit:bool|None, student_version:bool|None, machine:str|None, port:int|None, aedt_process_id:int|None, remove_lock:bool|None)\n```",
		},
	}
	// ==============================

	// Include initialization template to prompt.
	finalQuery += "\nHard requirements (do not violate):\n- Include **all imports** actually used. Follow the template for PyAEDT version " + pyaedtVersion + ": " + import_templates[pyaedtVersion] + "\n"
	finalQuery += "- Provide an **Initialization** section that **explicitly** declares the known information as follows:\n"

	if _, ok := version_mapper[pyaedtVersion]; !ok {
		logging.Log.Warnf(&logging.ContextMap{}, "Unknown PyAEDT version: %s. Defaulting to 0.19.0", pyaedtVersion)
		pyaedtVersion = "0.19.0"
	}
	finalQuery += "  - Use PyAEDT version: " + pyaedtVersion + "\n"
	finalQuery += "  - AEDT version: " + version_mapper[pyaedtVersion] + "\n"
	finalQuery += "  - Design name: " + design + "\n"
	finalQuery += "  - Application: " + application + "\n"

	logging.Log.Debugf(&logging.ContextMap{}, "!!!!Selections: %v", selections)
	// if selections is empty, skip it.
	if selections != nil && len(selections) > 0 {
		finalQuery += "  - Selections: " + strings.Join(selections, ", ") + "\n"
	}
	finalQuery += "  - Project name: " + project + "\n\n"
	finalQuery += "The following statements are examples of how to initialize different applications, refer to these examples and initialization accordingly: \n"

	for appName, init_template := range init_templates[pyaedtVersion] {
		finalQuery += "\n- " + appName + ":\n" + init_template + "\n"
	}

	finalQuery += "\n\n"

	logging.Log.Debugf(&logging.ContextMap{}, "=================== Final Query %v ===================", finalQuery)

	// Return the final query
	return finalQuery
}

func RephraseRequest_kapatil(request string) (result string) {
	input := strings.ToLower(request)
	bef, after, found := strings.Cut(input, "launch aedt")
	if found {
		result = bef + "create desktop instance" + after
	} else {
		result = input
	}

	return result

}

// BuildFinalQueryForCodeLLMRequest builds the final query for a code generation
// request to LLM. The final query is a markdown string that contains the
// original request and the code examples from the KnowledgeDB.
//
// Tags:
//   - @displayName: Final Query (Code LLM Request)
//
// Parameters:
//   - request: the original request
//   - knowledgedbResponse: the KnowledgeDB response
//
// Returns:
//   - finalQuery: the final query
func BuildFinalQueryForCodeLLMRequest(request string, knowledgedbResponse []sharedtypes.DbResponse) (finalQuery string) {
	// Build the final query using the KnowledgeDB response and the original request
	// We have to use the text from the DB response and the original request.
	//
	// The prompt should be in the following format:
	//
	// ******************************************************************************
	// Based on the following examples:
	//
	// --- START EXAMPLE {response_n}---
	// >>> Summary:
	// {knowledge_db_response_n_summary}
	//
	// >>> Code snippet:
	// ```python
	// {knowledge_db_response_n_text}
	// ```
	// --- END EXAMPLE {response_n}---
	//
	// --- START EXAMPLE {response_n}---
	// ...
	// --- END EXAMPLE {response_n}---
	//
	// Generate the Python code for the following request:
	//
	// >>> Request:
	// {original_request}
	// ******************************************************************************

	// If there is no response from the KnowledgeDB, return the original request
	if len(knowledgedbResponse) > 0 {
		// Initial request

		finalQuery = "Based on the following examples:\n\n"

		for i, element := range knowledgedbResponse {
			// Add the example number
			logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Reading knowledge DB response")
			finalQuery += "--- START EXAMPLE " + fmt.Sprint(i+1) + "---\n"
			finalQuery += ">>> Summary:\n" + element.Summary + "\n\n"
			finalQuery += ">>> Code snippet:\n```python\n" + element.Text + "\n```\n"
			finalQuery += "--- END EXAMPLE " + fmt.Sprint(i+1) + "---\n\n"
			// logging.Log.Debugf(&logging.ContextMap{}, "kapatil: Initial Query %s", finalQuery)
		}

	} else {
		logging.Log.Debugf(&logging.ContextMap{}, "Zero knowledge DB reponse found")
	}

	// Kaumudi: Rephrase
	new_request := RephraseRequest_kapatil(request)

	// Pass in the original request
	finalQuery += "Generate the Python code for the following request:\n>>> Request:\n" + new_request + "\n"

	// Return the final query
	return finalQuery
}

type AppendMessageHistoryRole string

const (
	user      AppendMessageHistoryRole = "user"
	assistant AppendMessageHistoryRole = "assistant"
	system    AppendMessageHistoryRole = "system"
)

// AppendMessageHistory appends a new message to the conversation history
//
// Tags:
//   - @displayName: Append Message History
//
// Parameters:
//   - newMessage: the new message
//   - role: the role of the message
//   - history: the conversation history
//
// Returns:
//   - updatedHistory: the updated conversation history
func AppendMessageHistory(newMessage string, role AppendMessageHistoryRole, history []sharedtypes.HistoricMessage) (updatedHistory []sharedtypes.HistoricMessage) {
	switch role {
	case user:
	case assistant:
	case system:
	default:
		errMessage := fmt.Sprintf("Invalid role used for 'AppendMessageHistory': %v", role)
		logging.Log.Warn(&logging.ContextMap{}, errMessage)
		panic(errMessage)
	}

	// skip for empty messages
	if newMessage == "" {
		return history
	}

	// Create a new HistoricMessage
	newMessageHistory := sharedtypes.HistoricMessage{
		Role:    string(role),
		Content: newMessage,
	}

	// Append the new message to the history
	history = append(history, newMessageHistory)

	return history
}

// ShortenMessageHistory shortens the conversation history to a maximum length.
// It will retain only the most recent messages and older messages will be
// removed.
//
// Tags:
//   - @displayName: Shorten History
//
// Parameters:
//   - history: the conversation history
//   - maxLength: the maximum length of the conversation history
//
// Returns:
//   - updatedHistory: the updated conversation history
func ShortenMessageHistory(history []sharedtypes.HistoricMessage, maxLength int) (updatedHistory []sharedtypes.HistoricMessage) {
	if len(history) <= maxLength {
		return history
	}

	return history[len(history)-maxLength:]
}

// CheckTokenLimitReached checks if the query exceeds the token limit for the specified model
//
// Tags:
//   - @displayName: Check Token Limit Reached
//
// Parameters:
//   - query: the query string
//   - tokenLimit: the token limit
//   - modelName: the name of the model to check against
//
// Returns:
//   - tokenLimitReached: true if the token limit is reached, false otherwise
func CheckTokenLimitReached(query string, tokenLimit int, modelName string, tokenLimitMessage string) (tokenLimitReached bool, responseMessage string) {
	// Check if the query exceeds the token limit
	tokenCount, err := openAiTokenCount(modelName, query)
	if err != nil {
		panic(fmt.Sprintf("Error counting tokens: %v", err))
	}
	if tokenCount > tokenLimit {
		logging.Log.Warnf(&logging.ContextMap{}, "Query exceeds token limit: %d tokens, limit is %d tokens", tokenCount, tokenLimit)
		return true, tokenLimitMessage
	}

	return false, ""
}
