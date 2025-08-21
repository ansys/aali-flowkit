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
	"fmt"
	"strings"
)

// validateToken checks the authentication token for security issues.
// It prevents CRLF injection attacks by checking for carriage return and line feed characters.
// This is critical for preventing header injection attacks in HTTP requests.
func validateToken(token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}
	
	// Check for CRLF characters that could be used for header injection
	if strings.ContainsAny(token, "\r\n") {
		return fmt.Errorf("token contains invalid characters (CR/LF)")
	}
	
	// Additional security checks can be added here in the future
	// For example: length limits, character restrictions, etc.
	
	return nil
}

// sanitizeError removes sensitive information from error messages.
// It replaces any occurrence of the token with a redacted placeholder to prevent
// accidental token leakage in logs, error messages, or stack traces.
func sanitizeError(err error, token string) error {
	if err == nil {
		return nil
	}
	
	// If no token provided, return error as-is
	if token == "" {
		return err
	}
	
	// Replace all occurrences of the token with a redacted placeholder
	errorMsg := err.Error()
	sanitizedMsg := strings.ReplaceAll(errorMsg, token, "***REDACTED***")
	
	// Also check for base64 encoded version (in case it appears in error messages)
	// This is especially important if the token gets encoded somewhere in the stack
	// Note: We don't encode here as we're using simple Bearer tokens, not Basic auth
	
	return fmt.Errorf("%s", sanitizedMsg)
}

// sanitizeString removes sensitive information from any string.
// This is useful for sanitizing log messages, debug output, etc.
func sanitizeString(str string, token string) string {
	if token == "" {
		return str
	}
	return strings.ReplaceAll(str, token, "***REDACTED***")
}