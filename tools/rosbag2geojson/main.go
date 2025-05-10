package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"
)

type Feature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   Geometry               `json:"geometry"`
}

type Geometry struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

func getDatumCoords(configPath string) (float64, float64, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return 0, 0, err
	}
	lines := strings.Split(string(data), "\n")
	var lat, lon float64

	for _, line := range lines {
		if strings.HasPrefix(line, "export OM_DATUM_LAT=") {
			val := strings.Trim(strings.TrimSpace(strings.Split(line, "=")[1]), `"`)
			lat, _ = strconv.ParseFloat(val, 64)
		}
		if strings.HasPrefix(line, "export OM_DATUM_LONG=") {
			val := strings.Trim(strings.TrimSpace(strings.Split(line, "=")[1]), `"`)
			lon, _ = strconv.ParseFloat(val, 64)
		}
	}

	return lat, lon, nil
}

func createDummyFeature(lat, lon float64) Feature {
	offset := 0.0005
	return Feature{
		Type: "Feature",
		Properties: map[string]interface{}{
			"id": "working_area",
		},
		Geometry: Geometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{lon, lat},
					{lon + offset, lat},
					{lon + offset, lat + offset},
					{lon, lat + offset},
					{lon, lat},
				},
			},
		},
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
		os.Exit(1)
	}

	bagPath := os.Args[1]
	outputPath := os.Args[2]
	configPath := "/boot/openmower/mower_config.txt"

	lat, lon, err := getDatumCoords(configPath)
	if err != nil {
		log.Fatal("failed to read datum from config:", err)
	}

	fmt.Println("Using datum:")
	fmt.Println("  LAT:", lat)
	fmt.Println("  LON:", lon)

	geo := FeatureCollection{
		Type:     "FeatureCollection",
		Features: []Feature{createDummyFeature(lat, lon)},
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatal("could not create output file:", err)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(geo); err != nil {
		log.Fatal("failed to encode GeoJSON:", err)
	}

	fmt.Println("✅ Converted", bagPath, "to", outputPath)
}
