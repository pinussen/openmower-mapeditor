package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	bag "github.com/akamensky/rosbag"
)

type GeoJSON struct {
	Type     string        `json:"type"`
	Features []GeoFeature  `json:"features"`
}

type GeoFeature struct {
	Type       string         `json:"type"`
	Geometry   GeoGeometry    `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

type GeoGeometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

func convertBagToGeoJSON(bagPath, outputPath string) error {
	file, err := os.Open(bagPath)
	if err != nil {
		return fmt.Errorf("could not open bag: %w", err)
	}
	defer file.Close()

	reader, err := bag.NewReader(file)
	if err != nil {
		return fmt.Errorf("could not read bag: %w", err)
	}
	defer reader.Close()

	var features []GeoFeature

	for reader.HasNext() {
		conn, msg, err := reader.ReadMessage()
		if err != nil {
			log.Printf("warn: failed to read msg: %v", err)
			continue
		}

		if conn.Topic == "/xbot_monitoring/map" {
			// TODO: parse actual ROS msg binary blob properly
			// Dummy geometry for now
			features = append(features, GeoFeature{
				Type: "Feature",
				Geometry: GeoGeometry{
					Type:        "Polygon",
					Coordinates: [][]float64{{18.06, 59.33}, {18.07, 59.33}, {18.07, 59.34}, {18.06, 59.34}, {18.06, 59.33}},
				},
				Properties: map[string]any{
					"source": conn.Topic,
					"stamp":  time.Now().Format(time.RFC3339),
				},
			})
		}
	}

	geo := GeoJSON{
		Type:     "FeatureCollection",
		Features: features,
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("could not create output: %w", err)
	}
	defer out.Close()

	return json.NewEncoder(out).Encode(geo)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
		os.Exit(1)
	}

	inPath := os.Args[1]
	outPath := os.Args[2]

	err := convertBagToGeoJSON(inPath, outPath)
	if err != nil {
		log.Fatal("conversion failed:", err)
	}

	fmt.Println("✅ Converted", inPath, "to", outPath)
}
