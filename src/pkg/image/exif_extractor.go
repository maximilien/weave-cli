package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

// EXIFData represents extracted EXIF metadata
type EXIFData struct {
	Make         string  `json:"make,omitempty"`
	Model        string  `json:"model,omitempty"`
	DateTime     string  `json:"datetime,omitempty"`
	Orientation  int     `json:"orientation,omitempty"`
	Width        int     `json:"width,omitempty"`
	Height       int     `json:"height,omitempty"`
	FNumber      float64 `json:"f_number,omitempty"`
	ExposureTime string  `json:"exposure_time,omitempty"`
	ISO          int     `json:"iso,omitempty"`
	FocalLength  string  `json:"focal_length,omitempty"`
	GPSLatitude  float64 `json:"gps_latitude,omitempty"`
	GPSLongitude float64 `json:"gps_longitude,omitempty"`
	GPSAltitude  float64 `json:"gps_altitude,omitempty"`
}

// ExtractEXIF extracts EXIF metadata from an image file
func ExtractEXIF(filePath string) (*EXIFData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open image file: %w", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// Many images don't have EXIF data, this is not an error
		return &EXIFData{}, nil
	}

	data := &EXIFData{}

	// Extract camera info
	if make, err := x.Get(exif.Make); err == nil {
		data.Make, _ = make.StringVal()
	}
	if model, err := x.Get(exif.Model); err == nil {
		data.Model, _ = model.StringVal()
	}
	if dateTime, err := x.Get(exif.DateTime); err == nil {
		data.DateTime, _ = dateTime.StringVal()
	}

	// Extract orientation
	if orientation, err := x.Get(exif.Orientation); err == nil {
		if val, err := orientation.Int(0); err == nil {
			data.Orientation = val
		}
	}

	// Extract image dimensions
	if width, err := x.Get(exif.PixelXDimension); err == nil {
		if val, err := width.Int(0); err == nil {
			data.Width = val
		}
	}
	if height, err := x.Get(exif.PixelYDimension); err == nil {
		if val, err := height.Int(0); err == nil {
			data.Height = val
		}
	}

	// Extract camera settings
	if fNumber, err := x.Get(exif.FNumber); err == nil {
		if num, denom, err := fNumber.Rat2(0); err == nil && denom != 0 {
			data.FNumber = float64(num) / float64(denom)
		}
	}
	if exposureTime, err := x.Get(exif.ExposureTime); err == nil {
		data.ExposureTime = exposureTime.String()
	}
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if val, err := iso.Int(0); err == nil {
			data.ISO = val
		}
	}
	if focalLength, err := x.Get(exif.FocalLength); err == nil {
		data.FocalLength = focalLength.String()
	}

	// Extract GPS data
	lat, lon, err := x.LatLong()
	if err == nil {
		data.GPSLatitude = lat
		data.GPSLongitude = lon
	}

	return data, nil
}

// ExtractEXIFFromBytes extracts EXIF metadata from image bytes
func ExtractEXIFFromBytes(imageData []byte) (*EXIFData, error) {
	reader := bytes.NewReader(imageData)
	x, err := exif.Decode(reader)
	if err != nil {
		// Many images don't have EXIF data, this is not an error
		return &EXIFData{}, nil
	}

	data := &EXIFData{}

	// Extract camera info
	if make, err := x.Get(exif.Make); err == nil {
		data.Make, _ = make.StringVal()
	}
	if model, err := x.Get(exif.Model); err == nil {
		data.Model, _ = model.StringVal()
	}
	if dateTime, err := x.Get(exif.DateTime); err == nil {
		data.DateTime, _ = dateTime.StringVal()
	}

	// Extract orientation
	if orientation, err := x.Get(exif.Orientation); err == nil {
		if val, err := orientation.Int(0); err == nil {
			data.Orientation = val
		}
	}

	// Extract image dimensions
	if width, err := x.Get(exif.PixelXDimension); err == nil {
		if val, err := width.Int(0); err == nil {
			data.Width = val
		}
	}
	if height, err := x.Get(exif.PixelYDimension); err == nil {
		if val, err := height.Int(0); err == nil {
			data.Height = val
		}
	}

	// Extract camera settings
	if fNumber, err := x.Get(exif.FNumber); err == nil {
		if num, denom, err := fNumber.Rat2(0); err == nil && denom != 0 {
			data.FNumber = float64(num) / float64(denom)
		}
	}
	if exposureTime, err := x.Get(exif.ExposureTime); err == nil {
		data.ExposureTime = exposureTime.String()
	}
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if val, err := iso.Int(0); err == nil {
			data.ISO = val
		}
	}
	if focalLength, err := x.Get(exif.FocalLength); err == nil {
		data.FocalLength = focalLength.String()
	}

	// Extract GPS data
	lat, lon, err := x.LatLong()
	if err == nil {
		data.GPSLatitude = lat
		data.GPSLongitude = lon
	}

	return data, nil
}

// ToJSON converts EXIFData to JSON string
func (e *EXIFData) ToJSON() (string, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("failed to marshal EXIF data: %w", err)
	}
	return string(data), nil
}

// IsEmpty returns true if no EXIF data was extracted
func (e *EXIFData) IsEmpty() bool {
	return e.Make == "" && e.Model == "" && e.DateTime == "" &&
		e.Width == 0 && e.Height == 0 && e.GPSLatitude == 0 && e.GPSLongitude == 0
}
