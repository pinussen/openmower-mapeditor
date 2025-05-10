package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func parseDatum(filePath string) (float64, float64) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal("Failed to read mower config:", err)
	}

	var lat, lon float64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "export OM_DATUM_LAT=") {
			val := strings.Trim(strings.Split(line, "=")[1], "\"")
			lat, _ = strconv.ParseFloat(val, 64)
		}
		if strings.HasPrefix(line, "export OM_DATUM_LONG=") {
			val := strings.Trim(strings.Split(line, "=")[1], "\"")
			lon, _ = strconv.ParseFloat(val, 64)
		}
	}

	return lat, lon
}

func createDummyFeature(lat, lon float64) Feature {
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
					{lon + 0.001, lat},
					{lon + 0.001, lat + 0.001},
					{lon, lat + 0.001},
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

//	bagPath := os.Args[1]
	outputPath := os.Args[2]

	lat, lon := parseDatum("/boot/openmower/mower_config.txt")

	fmt.Println("➡️  Konverterar ROS-bag till GeoJSON...")
	fmt.Println("Using datum:")
	fmt.Println("  LAT:", lat)
	fmt.Println("  LON:", lon)

	geo := FeatureCollection{
		Type:     "FeatureCollection",
		Features: []Feature{createDummyFeature(lat, lon)},
	}

	// Create output dir if missing
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		log.Fatal("failed to create output directory:", err)
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

	fmt.Println("✅ GeoJSON written to", outputPath)
}
