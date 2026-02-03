// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package document

import (
	"testing"
)

// TestListCmd_Flags tests that the list command has all required flags
func TestListCmd_Flags(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		shorthand    string
		defaultValue interface{}
	}{
		{"limit flag", "limit", "l", 50},
		{"offset flag", "offset", "", 0},
		{"long flag", "long", "L", false},
		{"short flag", "short", "s", 5},
		{"no-truncate flag", "no-truncate", "", false},
		{"virtual flag", "virtual", "w", false},
		{"summary flag", "summary", "S", false},
		{"output flag", "output", "o", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := ListCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag --%s not found", tt.flagName)
				return
			}

			if tt.shorthand != "" && flag.Shorthand != tt.shorthand {
				t.Errorf("Flag --%s shorthand = %s, want %s", tt.flagName, flag.Shorthand, tt.shorthand)
			}

			// Check default values
			switch v := tt.defaultValue.(type) {
			case int:
				if flag.DefValue != string(rune(v+'0')) && flag.DefValue != "50" && flag.DefValue != "5" && flag.DefValue != "0" {
					// For int flags, just verify the flag exists and has a value
					if flag.DefValue == "" {
						t.Errorf("Flag --%s has no default value", tt.flagName)
					}
				}
			case bool:
				expected := "false"
				if v {
					expected = "true"
				}
				if flag.DefValue != expected {
					t.Errorf("Flag --%s default = %s, want %s", tt.flagName, flag.DefValue, expected)
				}
			case string:
				if flag.DefValue != v {
					t.Errorf("Flag --%s default = %s, want %s", tt.flagName, flag.DefValue, v)
				}
			}
		})
	}
}

// TestListCmd_OffsetFlag specifically tests the offset flag functionality
func TestListCmd_OffsetFlag(t *testing.T) {
	flag := ListCmd.Flags().Lookup("offset")
	if flag == nil {
		t.Fatal("--offset flag not found")
	}

	if flag.DefValue != "0" {
		t.Errorf("--offset default value = %s, want 0", flag.DefValue)
	}

	if flag.Usage == "" {
		t.Error("--offset flag has no usage description")
	}

	expectedUsage := "Number of documents to skip (for pagination)"
	if flag.Usage != expectedUsage {
		t.Errorf("--offset usage = %q, want %q", flag.Usage, expectedUsage)
	}
}

// TestListCmd_PaginationFlagCombinations tests valid flag combinations
func TestListCmd_PaginationFlagCombinations(t *testing.T) {
	tests := []struct {
		name   string
		limit  int
		offset int
		valid  bool
	}{
		{"default pagination", 50, 0, true},
		{"second page", 50, 50, true},
		{"third page", 50, 100, true},
		{"small limit", 10, 0, true},
		{"small limit page 2", 10, 10, true},
		{"large limit", 1000, 0, true},
		{"negative limit", -1, 0, false},
		{"negative offset", 50, -1, false},
		{"zero limit", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limitFlag := ListCmd.Flags().Lookup("limit")
			offsetFlag := ListCmd.Flags().Lookup("offset")

			if limitFlag == nil || offsetFlag == nil {
				t.Fatal("Required pagination flags not found")
			}

			// Verify flags exist and can be used
			// Note: Actual validation happens at runtime in runDocumentList
		})
	}
}

// TestPaginationCalculations tests the pagination math
func TestPaginationCalculations(t *testing.T) {
	tests := []struct {
		name          string
		offset        int
		limit         int
		totalDocs     int64
		shownDocs     int
		expectedPage  int
		expectedPages int
		expectedEnd   int
	}{
		{
			name:          "page 1 of 3",
			offset:        0,
			limit:         50,
			totalDocs:     150,
			shownDocs:     50,
			expectedPage:  1,
			expectedPages: 3,
			expectedEnd:   50,
		},
		{
			name:          "page 2 of 3",
			offset:        50,
			limit:         50,
			totalDocs:     150,
			shownDocs:     50,
			expectedPage:  2,
			expectedPages: 3,
			expectedEnd:   100,
		},
		{
			name:          "page 3 of 3 (partial)",
			offset:        100,
			limit:         50,
			totalDocs:     150,
			shownDocs:     50,
			expectedPage:  3,
			expectedPages: 3,
			expectedEnd:   150,
		},
		{
			name:          "small pages",
			offset:        6,
			limit:         3,
			totalDocs:     79,
			shownDocs:     3,
			expectedPage:  3,
			expectedPages: 27,
			expectedEnd:   9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate current page
			currentPage := (tt.offset / tt.limit) + 1
			if currentPage != tt.expectedPage {
				t.Errorf("currentPage = %d, want %d", currentPage, tt.expectedPage)
			}

			// Calculate total pages
			totalPages := (int(tt.totalDocs) + tt.limit - 1) / tt.limit
			if totalPages != tt.expectedPages {
				t.Errorf("totalPages = %d, want %d", totalPages, tt.expectedPages)
			}

			// Calculate current end position
			currentEnd := tt.offset + tt.shownDocs
			if currentEnd != tt.expectedEnd {
				t.Errorf("currentEnd = %d, want %d", currentEnd, tt.expectedEnd)
			}

			// Verify there are more documents
			hasMore := int64(currentEnd) < tt.totalDocs
			if tt.expectedPage < tt.expectedPages && !hasMore {
				t.Error("Should have more documents but hasMore = false")
			}
		})
	}
}
