package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	bag "github.com/lherman-cs/go-rosbag"
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
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

func pointArrayToCoordinates(points []any) [][]float64 {
	coords := make([][]float64, 0)
	for _, p := range points {
		if ptMap, ok := p.(map[string]any); ok {
			x, xok := ptMap["x"].(float64)
			y, yok := ptMap["y"].(float64)
			if xok && yok {
				coords = append(coords, []float64{x, y})
			}
		}
	}
	if len(coords) > 0 {
		coords = append(coords, coords[0]) // close polygon
	}
	return coords
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: extract_from_bag <input.bag> <output.geojson>")
		os.Exit(1)
	}

	inPath := os.Args[1]
	outPath := os.Args[2]

	f, err := os.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := bag.NewDecoder(f)
	var features []GeoFeature

	for {
		rec, err := d.Read()
		if err != nil {
			break
		}

		msg, ok := rec.(*bag.RecordMessageData)
		if !ok {
			continue
		}

		if msg.Conn.Topic == "/map/metadata" || msg.Conn.Topic == "/map" {
			continue // skip raw map data
		}

		data := make(map[string]any)
		err = msg.ViewAs(data)
		if err != nil {
			continue
		}

		for key, val := range data {
			if pts, ok := val.([]any); ok && len(pts) > 0 {
				coords := pointArrayToCoordinates(pts)
				if len(coords) == 0 {
					continue
				}
				features = append(features, GeoFeature{
					Type: "Feature",
					Geometry: GeoGeometry{
						Type:        "Polygon",
						Coordinates: [][][2]float64{coords},
					},
					Properties: map[string]any{
						"source":  msg.Conn.Topic,
						"field":   key,
						"stamp":   time.Unix(msg.Header.Timestamp.Sec, 0).Format(time.RFC3339),
					},
				})
			}
		}
		rec.Close()
	}

	geo := GeoJSON{
		Type:     "FeatureCollection",
		Features: features,
	}

	out, err := os.Create(outPath)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = json.NewEncoder(out).Encode(geo)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Converted", inPath, "to", outPath)
}
