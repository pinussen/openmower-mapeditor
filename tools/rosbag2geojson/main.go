package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"io/ioutil"

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

func readDatum(path string) (float64, float64, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	var lat, lon float64
	for _, line := range strings.Split(string(content), "\n") {
		if strings.Contains(line, "OM_DATUM_LAT") {
			lat, _ = strconv.ParseFloat(strings.Split(strings.Split(line, "\"")[1], "\"")[0], 64)
		}
		if strings.Contains(line, "OM_DATUM_LONG") {
			lon, _ = strconv.ParseFloat(strings.Split(strings.Split(line, "\"")[1], "\"")[0], 64)
		}
	}
	return lat, lon, nil
}

func convertUTMToLatLon(x, y, datumLat, datumLon float64) (float64, float64) {
	// En förenklad översättning – i verkligheten används en UTM-projektion
	// Här antar vi 1:1 meter till decimalgrader (ungefärlig för små områden)
	lat := datumLat + (y / 111000.0)
	lon := datumLon + (x / (111000.0 * 0.6))
	return lon, lat
}

func extractWorkingAreaFromBag(bagPath string) ([][]float64, error) {
	cmd := exec.Command("rosbag", "play", bagPath, "--topic", "/map/working_area", "-r", "1", "--pause", "--quiet")
	out, err := cmd.Output()
	defer out.Close()
	_, err = out.Write(jsonData) // eller vad du nu skriver

	if err != nil {
		return nil, err
	}
	// 👇 OBS: Här behöver vi ROS/rosbag parser på riktigt – detta är placeholder
	return [][]float64{
		{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0},
	}, nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
		os.Exit(1)
	}
	bagPath := os.Args[1]
	outPath := os.Args[2]

	fmt.Println("➡️  Konverterar ROS-bag till GeoJSON...")

	datumLat, datumLon, err := readDatum("/boot/openmower/mower_config.txt")
	if err != nil {
		log.Fatal("⚠️  Kunde inte läsa datum från mower_config.txt:", err)
	}

	fmt.Println("Using datum:")
	fmt.Println("  LAT:", datumLat)
	fmt.Println("  LON:", datumLon)

	pointsUTM, err := extractWorkingAreaFromBag(bagPath)
	if err != nil {
		log.Fatal("⚠️  Kunde inte extrahera working_area:", err)
	}

	var coordinates [][]float64
	for _, p := range pointsUTM {
		lon, lat := convertUTMToLatLon(p[0], p[1], datumLat, datumLon)
		coordinates = append(coordinates, []float64{lon, lat})
	}

	geo := FeatureCollection{
		Type: "FeatureCollection",
		Features: []Feature{
			{
				Type: "Feature",
				Properties: map[string]interface{}{
					"id": "working_area",
				},
				Geometry: Geometry{
					Type:        "Polygon",
					Coordinates: [][][]float64{coordinates},
				},
			},
		},
	}

	err = os.MkdirAll(filepath.Dir(outPath), 0755)
	if err != nil {
		log.Fatal("Kunde inte skapa katalog:", err)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal("could not create output file:", err)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(geo); err != nil {
		log.Fatal("failed to encode GeoJSON:", err)
	}

	fmt.Println("✅ GeoJSON saved to:", outPath)
}
