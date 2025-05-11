package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func readDatum(path string) (float64, float64) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("could not read mower_config.txt:", err)
	}

	var lat, lon float64
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "OM_DATUM_LAT") {
			parts := strings.Split(line, "=")
			lat, _ = strconv.ParseFloat(strings.Trim(parts[1], "\""), 64)
		}
		if strings.Contains(line, "OM_DATUM_LONG") {
			parts := strings.Split(line, "=")
			lon, _ = strconv.ParseFloat(strings.Trim(parts[1], "\""), 64)
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
					{lon + 0.0005, lat},
					{lon + 0.0005, lat + 0.0005},
					{lon, lat + 0.0005},
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

	datumPath := "/boot/openmower/mower_config.txt"
	lat, lon := readDatum(datumPath)

	geo := FeatureCollection{
		Type:     "FeatureCollection",
		Features: []Feature{createDummyFeature(lat, lon)},
	}

	outputPath := os.Args[2]
	out, err := os.Create(outputPath)
	if err != nil {
		log.Fatal("could not create output file:", err)
	}
	defer out.Close()

	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(geo); err != nil {
		log.Fatal("failed to encode GeoJSON:", err)
	}

	fmt.Println("✅ Dummy GeoJSON written to", outputPath)
}
