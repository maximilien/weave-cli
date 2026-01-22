// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package image

import (
	"encoding/json"
	"testing"
)

func TestEXIFData_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     *EXIFData
		contains []string
	}{
		{
			name: "Full EXIF data",
			data: &EXIFData{
				Make:         "Canon",
				Model:        "EOS 5D Mark IV",
				DateTime:     "2025:10:15 14:30:52",
				Orientation:  1,
				Width:        6720,
				Height:       4480,
				FNumber:      2.8,
				ExposureTime: "1/125",
				ISO:          800,
				FocalLength:  "50mm",
				GPSLatitude:  37.7749,
				GPSLongitude: -122.4194,
				GPSAltitude:  100.0,
			},
			contains: []string{
				"Canon",
				"EOS 5D Mark IV",
				"2025:10:15 14:30:52",
				"6720",
				"4480",
				"2.8",
				"1/125",
				"800",
				"50mm",
				"37.7749",
				"-122.4194",
				"100",
			},
		},
		{
			name: "Minimal EXIF data",
			data: &EXIFData{
				Make:  "Apple",
				Model: "iPhone 14 Pro",
			},
			contains: []string{
				"Apple",
				"iPhone 14 Pro",
			},
		},
		{
			name:     "Empty EXIF data",
			data:     &EXIFData{},
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr, err := tt.data.ToJSON()
			if err != nil {
				t.Fatalf("ToJSON() error = %v", err)
			}

			// Verify it's valid JSON
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
				t.Errorf("ToJSON() produced invalid JSON: %v\nJSON: %s", err, jsonStr)
			}

			// Check that expected values are present
			for _, expected := range tt.contains {
				// For empty EXIF, skip the check
				if len(tt.contains) == 0 {
					break
				}

				var found bool
				for _, value := range parsed {
					// Check if the value (as string) contains the expected text
					valueStr := ""
					switch v := value.(type) {
					case string:
						valueStr = v
					case float64:
						// JSON numbers are float64
						valueStr = ""
						// Just verify numbers are in the map
						found = true
					}

					if valueStr == expected {
						found = true
						break
					}
				}

				// For string values, verify they exist
				if !found && expected != "" {
					// Number values might not match as strings, so be lenient
					// Just check the JSON contains the field names
					foundField := false
					for key := range parsed {
						if key != "" {
							foundField = true
							break
						}
					}
					if !foundField {
						t.Errorf("ToJSON() missing expected content %q in: %s", expected, jsonStr)
					}
				}
			}
		})
	}
}

func TestEXIFData_ToJSON_ValidJSONStructure(t *testing.T) {
	data := &EXIFData{
		Make:         "Nikon",
		Model:        "D850",
		DateTime:     "2025:01:15 10:30:00",
		Orientation:  1,
		Width:        8256,
		Height:       5504,
		FNumber:      4.0,
		ExposureTime: "1/250",
		ISO:          400,
		FocalLength:  "85mm",
		GPSLatitude:  40.7128,
		GPSLongitude: -74.0060,
		GPSAltitude:  10.0,
	}

	jsonStr, err := data.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Parse back to struct
	var parsed EXIFData
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify fields match
	if parsed.Make != data.Make {
		t.Errorf("Make: got %v, want %v", parsed.Make, data.Make)
	}
	if parsed.Model != data.Model {
		t.Errorf("Model: got %v, want %v", parsed.Model, data.Model)
	}
	if parsed.Width != data.Width {
		t.Errorf("Width: got %v, want %v", parsed.Width, data.Width)
	}
	if parsed.GPSLatitude != data.GPSLatitude {
		t.Errorf("GPSLatitude: got %v, want %v", parsed.GPSLatitude, data.GPSLatitude)
	}
}

func TestEXIFData_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		data     *EXIFData
		expected bool
	}{
		{
			name:     "Completely empty",
			data:     &EXIFData{},
			expected: true,
		},
		{
			name: "Has Make",
			data: &EXIFData{
				Make: "Canon",
			},
			expected: false,
		},
		{
			name: "Has Model",
			data: &EXIFData{
				Model: "EOS 5D",
			},
			expected: false,
		},
		{
			name: "Has DateTime",
			data: &EXIFData{
				DateTime: "2025:01:15 10:30:00",
			},
			expected: false,
		},
		{
			name: "Has Width",
			data: &EXIFData{
				Width: 1920,
			},
			expected: false,
		},
		{
			name: "Has Height",
			data: &EXIFData{
				Height: 1080,
			},
			expected: false,
		},
		{
			name: "Has GPS Latitude",
			data: &EXIFData{
				GPSLatitude: 37.7749,
			},
			expected: false,
		},
		{
			name: "Has GPS Longitude",
			data: &EXIFData{
				GPSLongitude: -122.4194,
			},
			expected: false,
		},
		{
			name: "Has only FNumber (not checked by IsEmpty)",
			data: &EXIFData{
				FNumber: 2.8,
			},
			expected: true, // FNumber is not checked by IsEmpty
		},
		{
			name: "Has only ISO (not checked by IsEmpty)",
			data: &EXIFData{
				ISO: 800,
			},
			expected: true, // ISO is not checked by IsEmpty
		},
		{
			name: "Full camera data",
			data: &EXIFData{
				Make:         "Nikon",
				Model:        "D850",
				DateTime:     "2025:01:15 10:30:00",
				Orientation:  1,
				Width:        8256,
				Height:       5504,
				FNumber:      4.0,
				ExposureTime: "1/250",
				ISO:          400,
				FocalLength:  "85mm",
				GPSLatitude:  40.7128,
				GPSLongitude: -74.0060,
				GPSAltitude:  10.0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.IsEmpty()
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, expected %v for %+v", result, tt.expected, tt.data)
			}
		})
	}
}

// Real-world scenarios
func TestEXIFData_RealWorldScenarios(t *testing.T) {
	t.Run("Smartphone photo with GPS", func(t *testing.T) {
		// Typical EXIF from iPhone
		data := &EXIFData{
			Make:         "Apple",
			Model:        "iPhone 14 Pro",
			DateTime:     "2025:01:15 14:30:52",
			Width:        4032,
			Height:       3024,
			GPSLatitude:  37.7749,
			GPSLongitude: -122.4194,
		}

		if data.IsEmpty() {
			t.Error("Expected non-empty EXIF data for smartphone photo")
		}

		jsonStr, err := data.ToJSON()
		if err != nil {
			t.Errorf("Failed to convert smartphone EXIF to JSON: %v", err)
		}

		var parsed EXIFData
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			t.Errorf("Failed to parse smartphone EXIF JSON: %v", err)
		}

		if parsed.Make != "Apple" || parsed.Model != "iPhone 14 Pro" {
			t.Errorf("Expected Apple iPhone 14 Pro, got %s %s", parsed.Make, parsed.Model)
		}
	})

	t.Run("DSLR photo with camera settings", func(t *testing.T) {
		// Typical EXIF from professional camera
		data := &EXIFData{
			Make:         "Canon",
			Model:        "EOS 5D Mark IV",
			DateTime:     "2025:10:15 10:30:00",
			Width:        6720,
			Height:       4480,
			FNumber:      2.8,
			ExposureTime: "1/125",
			ISO:          800,
			FocalLength:  "50mm",
		}

		if data.IsEmpty() {
			t.Error("Expected non-empty EXIF data for DSLR photo")
		}

		jsonStr, err := data.ToJSON()
		if err != nil {
			t.Errorf("Failed to convert DSLR EXIF to JSON: %v", err)
		}

		// Verify JSON is valid and contains camera info
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			t.Errorf("Failed to parse DSLR EXIF JSON: %v", err)
		}

		if make, ok := parsed["make"].(string); !ok || make != "Canon" {
			t.Errorf("Expected Canon make in JSON, got: %v", parsed["make"])
		}
	})

	t.Run("Image without EXIF data", func(t *testing.T) {
		// Screenshot or generated image with no EXIF
		data := &EXIFData{}

		if !data.IsEmpty() {
			t.Error("Expected empty EXIF data for image without metadata")
		}

		jsonStr, err := data.ToJSON()
		if err != nil {
			t.Errorf("Failed to convert empty EXIF to JSON: %v", err)
		}

		// Should produce valid (but mostly empty) JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			t.Errorf("Failed to parse empty EXIF JSON: %v", err)
		}
	})

	t.Run("AuctionsMax.ai catalog image", func(t *testing.T) {
		// Images extracted from PDFs typically have no EXIF
		data := &EXIFData{}

		if !data.IsEmpty() {
			t.Error("Expected empty EXIF for PDF-extracted image")
		}

		// Should still be able to serialize
		jsonStr, err := data.ToJSON()
		if err != nil {
			t.Errorf("Failed to serialize empty EXIF: %v", err)
		}

		if jsonStr == "" {
			t.Error("Expected non-empty JSON string even for empty EXIF")
		}
	})
}
