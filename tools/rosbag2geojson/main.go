package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	} `json:"geometry"`
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

func localToWGS(x, y, datumLat, datumLon float64) (float64, float64) {
	lat := datumLat + y/111111.0
	lon := datumLon + x/(111111.0 * math.Cos(datumLat*math.Pi/180.0))
	return lon, lat
}

func parsePointsBlock(lines []string, datumLat, datumLon float64) [][][]float64 {
	var coords [][][]float64
	var ring [][]float64

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "x:") {
			parts := strings.Fields(line)
			x, _ := strconv.ParseFloat(strings.TrimPrefix(parts[0], "x:"), 64)
			y, _ := strconv.ParseFloat(strings.TrimPrefix(parts[1], "y:"), 64)
			lon, lat := localToWGS(x, y, datumLat, datumLon)
			ring = append(ring, []float64{lon, lat})
		}
	}
	if len(ring) > 0 {
		// Ensure polygon is closed
		if ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1] {
			ring = append(ring, ring[0])
		}
		coords = append(coords, ring)
	}

	return coords
}

func parseBagText(path string, datumLat, datumLon float64) FeatureCollection {
	cmd := exec.Command("rosbag", "play", "--pause", path)
	cmdOut := exec.Command("rosbag", "info", "--yaml", path)
	out, err := cmdOut.Output()
	if err != nil {
		log.Fatal("rosbag info failed:", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var currentType string
	var polygonLines []string
	var features []Feature
	var counter = map[string]int{}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "name:") {
			currentType = strings.Trim(strings.Split(line, ":")[1], " \"")
		}
		if strings.Contains(line, "points:") {
			polygonLines = nil
		}
		if strings.HasPrefix(line, "  - x:") {
			polygonLines = append(polygonLines, strings.TrimSpace(line))
		}
		if len(polygonLines) > 0 && strings.TrimSpace(line) == "" {
			coords := parsePointsBlock(polygonLines, datumLat, datumLon)
			id := classifyZone(currentType)
			counter[id]++
			feature := Feature{
				Type:       "Feature",
				Properties: map[string]interface{}{"id": fmt.Sprintf("%s_%d", id, counter[id])},
			}
			feature.Geometry.Type = "Polygon"
			feature.Geometry.Coordinates = coords
			features = append(features, feature)
		}
	}

	return FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}

func classifyZone(name string) string {
	name = strings.ToLower(name)
	if strings.Contains(name, "exclusion") {
		return "exclusion_zone"
	} else if strings.Contains(name, "navigation") || strings.Contains(name, "transport") {
		return "transport_zone"
	}
	return "working_area"
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
		os.Exit(1)
	}

	bagPath := os.Args[1]
	outPath := os.Args[2]
	datumLat, datumLon := readDatum("/boot/openmower/mower_config.txt")

	log.Println("➡️  Konverterar ROS-bag till GeoJSON...")
	log.Println("  LAT:", datumLat, "  LON:", datumLon)

	geojson := parseBagText(bagPath, datumLat, datumLon)

	out, err := os.Create(outPath)
	if err != nil {
		log.Fatal("could not create output file:", err)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(geojson); err != nil {
		log.Fatal("failed to encode GeoJSON:", err)
	}

	log.Println("✅ Klar:", outPath)
}
