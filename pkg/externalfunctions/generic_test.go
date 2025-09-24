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
	"reflect"
	"testing"

	"github.com/ansys/aali-sharedtypes/pkg/sharedtypes"
)

func TestParseSlashCommands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []sharedtypes.SlashCommand
	}{
		{
			name:     "Simple global command",
			input:    "/ban",
			expected: []sharedtypes.SlashCommand{{Scope: "global", Command: "ban"}},
		},
		{
			name:     "Scoped command",
			input:    "@admin /ban",
			expected: []sharedtypes.SlashCommand{{Scope: "admin", Command: "ban"}},
		},
		{
			name:     "Command with arguments (should ignore arguments)",
			input:    "/ban user123 reason spam",
			expected: []sharedtypes.SlashCommand{{Scope: "global", Command: "ban"}},
		},
		{
			name:     "Scoped command with arguments",
			input:    "@admin /ban user123 reason spam",
			expected: []sharedtypes.SlashCommand{{Scope: "admin", Command: "ban"}},
		},
		{
			name:  "Multiple commands on different lines",
			input: "/ban\n@admin /kick\n/mute",
			expected: []sharedtypes.SlashCommand{
				{Scope: "global", Command: "ban"},
				{Scope: "admin", Command: "kick"},
				{Scope: "global", Command: "mute"},
			},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: []sharedtypes.SlashCommand{},
		},
		{
			name:     "No slash commands",
			input:    "This is just regular text without commands",
			expected: []sharedtypes.SlashCommand{},
		},
		{
			name:  "Mixed content with commands",
			input: "Please @admin /ban this user\nSome other text\n/help",
			expected: []sharedtypes.SlashCommand{
				{Scope: "admin", Command: "ban"},
				{Scope: "global", Command: "help"},
			},
		},
		{
			name:     "Command in middle of line",
			input:    "Execute @moderator /timeout now",
			expected: []sharedtypes.SlashCommand{{Scope: "moderator", Command: "timeout"}},
		},
		{
			name:  "Multiple scopes",
			input: "@admin /ban\n@moderator /kick\n@user /help",
			expected: []sharedtypes.SlashCommand{
				{Scope: "admin", Command: "ban"},
				{Scope: "moderator", Command: "kick"},
				{Scope: "user", Command: "help"},
			},
		},
		{
			name:     "Invalid format - scope without command",
			input:    "@admin",
			expected: []sharedtypes.SlashCommand{},
		},
		{
			name:     "Invalid format - scope with text but no slash command",
			input:    "@admin please do something",
			expected: []sharedtypes.SlashCommand{},
		},
		{
			name:  "Whitespace handling",
			input: "  @admin   /ban   \n  /help  ",
			expected: []sharedtypes.SlashCommand{
				{Scope: "admin", Command: "ban"},
				{Scope: "global", Command: "help"},
			},
		},
		{
			name:  "Empty lines",
			input: "/ban\n\n\n@admin /kick\n\n",
			expected: []sharedtypes.SlashCommand{
				{Scope: "global", Command: "ban"},
				{Scope: "admin", Command: "kick"},
			},
		},
		{
			name:  "Same scope multiple times",
			input: "@admin /ban\n@admin /kick",
			expected: []sharedtypes.SlashCommand{
				{Scope: "admin", Command: "ban"},
				{Scope: "admin", Command: "kick"},
			},
		},
		{
			name:  "Global scope multiple times",
			input: "/ban\n/kick\n/mute",
			expected: []sharedtypes.SlashCommand{
				{Scope: "global", Command: "ban"},
				{Scope: "global", Command: "kick"},
				{Scope: "global", Command: "mute"},
			},
		},
		{
			name:  "Complex mixed scenario",
			input: "User says: @admin /ban user123\nThen someone else: /help\nAnd finally @moderator /timeout user456",
			expected: []sharedtypes.SlashCommand{
				{Scope: "admin", Command: "ban"},
				{Scope: "global", Command: "help"},
				{Scope: "moderator", Command: "timeout"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseSlashCommands(tt.input)

			// Handle nil vs empty slice comparison
			if len(result) == 0 && len(tt.expected) == 0 {
				return // Both are empty, test passes
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseSlashCommands() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
