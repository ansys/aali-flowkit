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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ansys/aali-sharedtypes/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
)

// CheckApiKeyAuthMongoDb checks if the given API key is valid and has access to the service.
//
// Tags:
//   - @displayName: Verify API Key
//
// Parameters:
//   - apiKey: The API key to check.
//   - mongoDbUrl: The URL of the MongoDB database.
//   - mongoDatabaseName: The name of the MongoDB database.
//   - mongoDbCollectionName: The name of the MongoDB collection.
//
// Returns:
//   - isAuthenticated: A boolean indicating whether the API key is authenticated.
func CheckApiKeyAuthMongoDb(apiKey string, mongoDbUrl string, mongoDatabaseName string, mongoDbCollectionName string) (isAuthenticated bool) {

	// create mongoDb context
	mongoDbContext, err := mongoDbInitializeClient(mongoDbUrl, mongoDatabaseName, mongoDbCollectionName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error initializing mongoDb client: %v", err)
		panic(err)
	}
	defer mongoDbContext.Client.Disconnect(context.Background())

	// check if customer exists
	exists, customer, err := mongoDbGetCustomerByApiKey(mongoDbContext, apiKey)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error getting customer by API key: %v", err)
		panic(err)
	}
	if !exists {
		logging.Log.Warnf(&logging.ContextMap{}, "Authenticating failed: given API key not found in database")
		return false
	}

	// check if customer is allowed access
	if customer.AccessDenied {
		logging.Log.Warnf(&logging.ContextMap{}, "Authenticating failed: access denied for given API key")
		return false
	}

	return true
}

// CheckCreateUserIdMongoDb checks if a user ID exists in the MongoDB database and creates it if it doesn't.
//
// Tags:
//   - @displayName: Check and Create User ID
//
// Parameters:
//   - userId: The user ID to check.
//   - tokenLimitForNewUsers: The token limit for new users.
//   - mongoDbUrl: The URL of the MongoDB database.
//   - mongoDatabaseName: The name of the MongoDB database.
//   - mongoDbCollectionName: The name of the MongoDB collection.
//
// Returns:
//   - existingUser: A boolean indicating whether the user ID already exists.
func CheckCreateUserIdMongoDb(userId string, temporaryTokenLimit int, hoursUntilTokenLimitReset int, modelId []string, mongoDbUrl string, mongoDatabaseName string, mongoDbCollectionName string) (existingUser bool) {

	// create mongoDb context
	mongoDbContext, err := mongoDbInitializeClient(mongoDbUrl, mongoDatabaseName, mongoDbCollectionName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error initializing mongoDb client: %v", err)
		panic(err)
	}
	defer mongoDbContext.Client.Disconnect(context.Background())

	// check if customer for userid exists if not, create it
	existingUser, _, err = mongoDbGetCreateCustomerByUserId(mongoDbContext, userId, temporaryTokenLimit, hoursUntilTokenLimitReset, modelId)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error getting or creating customer by userId: %v", err)
		panic(err)
	}

	return existingUser
}

// UpdateTotalTokenCountForCustomerMongoDb updates the total token count for the given customer in the MongoDB database.
//
// Tags:
//   - @displayName: Update Total Token Count
//
// Parameters:
//   - apiKey: The API key of the customer.
//   - mongoDbUrl: The URL of the MongoDB database.
//   - mongoDatabaseName: The name of the MongoDB database.
//   - mongoDbCollectionName: The name of the MongoDB collection.
//   - additionalTokenCount: The number of additional tokens to add to the total token count.
//
// Returns:
//   - tokenLimitReached: A boolean indicating whether the customer has reached the token limit.
func UpdateTotalTokenCountForCustomerMongoDb(apiKey string, mongoDbUrl string, mongoDatabaseName string, mongoDbCollectionName string, additionalTokenCount int) (tokenLimitReached bool) {

	// create mongoDb context
	mongoDbContext, err := mongoDbInitializeClient(mongoDbUrl, mongoDatabaseName, mongoDbCollectionName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error initializing mongoDb client: %v", err)
		panic(err)
	}
	defer mongoDbContext.Client.Disconnect(context.Background())

	// update token count
	err = mongoDbAddToTotalTokenCount(mongoDbContext, "api_key", apiKey, additionalTokenCount)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error updating total token count for customer: %v", err)
		panic(err)
	}

	// check if customer is over the limit
	exists, customer, err := mongoDbGetCustomerByApiKey(mongoDbContext, apiKey)
	if err != nil || !exists {
		logging.Log.Errorf(&logging.ContextMap{}, "Error getting customer by API key: %v", err)
		panic(err)
	}
	if customer.TotalTokenCount >= customer.TokenLimit {
		return true
	}

	return false
}

// UpdateTotalTokenCountForUserIdMongoDb updates the total token count for the given user ID in the MongoDB database.
//
// Tags:
//   - @displayName: Update Total Token Count by User ID
//
// Parameters:
//   - userId: The user ID of the customer.
//   - mongoDbUrl: The URL of the MongoDB database.
//   - mongoDatabaseName: The name of the MongoDB database.
//   - mongoDbCollectionName: The name of the MongoDB collection.
//   - additionalTokenCount: The number of additional tokens to add to the total token count.
//
// Returns:
//   - tokenLimitReached: A boolean indicating whether the customer has reached the token limit.
func UpdateTotalTokenCountForUserIdMongoDb(userId string, mongoDbUrl string, mongoDatabaseName string, mongoDbCollectionName string, additionalInputTokenCount int, additionalOutputTokenCount int, hoursUntilTokenLimitReset int) (tokenLimitReached bool) {
	// create mongoDb context
	mongoDbContext, err := mongoDbInitializeClient(mongoDbUrl, mongoDatabaseName, mongoDbCollectionName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error initializing mongoDb client: %v", err)
		panic(err)
	}
	defer mongoDbContext.Client.Disconnect(context.Background())

	// update token count
	tokenLimitReached, err = mongoDbAddToInputOutputTokenCountAndCheckLimit(mongoDbContext, userId, additionalInputTokenCount, additionalOutputTokenCount, hoursUntilTokenLimitReset)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error updating total token count for customer: %v", err)
		panic(err)
	}

	return tokenLimitReached
}

// DenyCustomerAccessAndSendWarningMongoDb denies access to the customer and sends a warning if necessary.
//
// Tags:
//   - @displayName: Deny Customer Access
//
// Parameters:
//   - apiKey: The API key of the customer.
//   - mongoDbUrl: The URL of the MongoDB database.
//   - mongoDatabaseName: The name of the MongoDB database.
//   - mongoDbCollectionName: The name of the MongoDB collection.
//
// Returns:
//   - customerName: The name of the customer.
//   - sendWarning: A boolean indicating whether a warning should be sent to the customer.
func DenyCustomerAccessAndSendWarningMongoDb(apiKey string, mongoDbUrl string, mongoDatabaseName string, mongoDbCollectionName string) (customerName string, sendWarning bool) {
	// create mongoDb context
	mongoDbContext, err := mongoDbInitializeClient(mongoDbUrl, mongoDatabaseName, mongoDbCollectionName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error initializing mongoDb client: %v", err)
		panic(err)
	}
	defer mongoDbContext.Client.Disconnect(context.Background())

	// check if warning for customer needs to be sent
	exists, customer, err := mongoDbGetCustomerByApiKey(mongoDbContext, apiKey)
	if err != nil || !exists {
		logging.Log.Errorf(&logging.ContextMap{}, "Error getting customer by API key: %v", err)
		panic(err)
	}
	if !customer.WarningSent {
		sendWarning = true
	}

	// deny customer access and set warning sent
	err = mongoDbUpdateAccessAndWarning(mongoDbContext, "api_key", apiKey)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error updating access and warning for customer: %v", err)
		panic(err)
	}

	return customer.CustomerName, sendWarning
}

// DenyCustomerAccessAndSendWarningMongoDbUserId denies access to the customer by user ID and sends a warning if necessary.
//
// Tags:
//   - @displayName: Deny Customer Access by User ID
//
// Parameters:
//   - userId: The user ID of the customer.
//   - mongoDbUrl: The URL of the MongoDB database.
//   - mongoDatabaseName: The name of the MongoDB database.
//   - mongoDbCollectionName: The name of the MongoDB collection.
//
// Returns:
//   - sendWarning: A boolean indicating whether a warning should be sent to the customer.
func DenyCustomerAccessAndSendWarningMongoDbUserId(userId string, mongoDbUrl string, mongoDatabaseName string, mongoDbCollectionName string) (sendWarning bool) {
	// create mongoDb context
	mongoDbContext, err := mongoDbInitializeClient(mongoDbUrl, mongoDatabaseName, mongoDbCollectionName)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error initializing mongoDb client: %v", err)
		panic(err)
	}
	defer mongoDbContext.Client.Disconnect(context.Background())

	// check if warning for customer needs to be sent
	customer := &MongoDbCustomerObjectDisco{}
	filter := bson.M{"user_id": userId}
	err = mongoDbContext.Collection.FindOne(context.Background(), filter).Decode(&customer)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error finding customer by user ID: %v", err)
		panic(err)
	}
	if !customer.WarningSent {
		sendWarning = true
	}

	// deny customer access and set warning sent
	err = mongoDbUpdateAccessAndWarning(mongoDbContext, "user_id", userId)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error updating access and warning for customer: %v", err)
		panic(err)
	}

	return sendWarning
}

// SendLogicAppNotificationEmail sends a POST request to the email service.
//
// Tags:
//   - @displayName: Send Email Notification
//
// Parameters:
//   - endpoint: The email service endpoint.
//   - email: The email address.
//   - subject: The email subject.
//   - content: The email content.
func SendLogicAppNotificationEmail(logicAppEndpoint string, email string, subject string, content string) {
	// Create the request body
	requestBody := EmailRequest{
		Email:   email,
		Subject: subject,
		Content: content,
	}

	// Convert the request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error marshaling JSON: %v", err)
		panic(fmt.Errorf("error marshaling JSON: %v", err))
	}

	// Create a new HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create the POST request
	req, err := http.NewRequest("POST", logicAppEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error creating request: %v", err)
		panic(fmt.Errorf("error creating request: %v", err))
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		logging.Log.Errorf(&logging.ContextMap{}, "Error sending request: %v", err)
		panic(fmt.Errorf("error sending request: %v", err))
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logging.Log.Errorf(&logging.ContextMap{}, "Unexpected status code: %d", resp.StatusCode)
		panic(fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}
}

// CreateMessageWithVariable creates a message with a variable.
//
// Tags:
//   - @displayName: Create Message with Variable
//
// Parameters:
//   - message: The message to create.
//   - variable: The variable to insert into the message.
//
// Returns:
//   - updatedMessage: The updated message with the variable inserted.
func CreateMessageWithVariable(message string, variable string) (updatedMessage string) {
	// check for {{variable}} in message and replace with variable value
	updatedMessage = strings.ReplaceAll(message, "{{variable}}", variable)
	return updatedMessage
}
