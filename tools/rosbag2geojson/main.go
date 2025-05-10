package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

func createDummyFeature() Feature {
	return Feature{
		Type: "Feature",
		Properties: map[string]interface{}{
			"id": "working_area",
		},
		Geometry: Geometry{
			Type: "Polygon",
			Coordinates: [][][]float64{
				{
					{20.0, 59.0},
					{20.001, 59.0},
					{20.001, 59.001},
					{20.0, 59.001},
					{20.0, 59.0},
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

	outputPath := os.Args[2]

	geo := FeatureCollection{
		Type:     "FeatureCollection",
		Features: []Feature{createDummyFeature()},
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

	fmt.Println("✅ Dummy GeoJSON written to", outputPath)
}
