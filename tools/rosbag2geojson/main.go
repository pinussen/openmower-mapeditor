package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

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

// readDatum hämtar OM_DATUM_LAT och OM_DATUM_LONG från configfilen
func readDatum(path string) (float64, float64) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("could not read mower_config.txt:", err)
	}
	var lat, lon float64
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "export OM_DATUM_LAT") {
			parts := strings.Split(line, "=")
			lat, _ = strconv.ParseFloat(strings.Trim(parts[1], `"`), 64)
		}
		if strings.HasPrefix(line, "export OM_DATUM_LONG") {
			parts := strings.Split(line, "=")
			lon, _ = strconv.ParseFloat(strings.Trim(parts[1], `"`), 64)
		}
	}
	return lat, lon
}

// localToWGS konverterar lokala x,y (meter) till lon,lat
func localToWGS(x, y, datumLat, datumLon float64) (float64, float64) {
	lat := datumLat + y/111111.0
	lon := datumLon + x/(111111.0*math.Cos(datumLat*math.Pi/180.0))
	return lon, lat
}

// parsePointsBlock läser block av ”x:…, y:…”-rader och returnerar ett ring (stängt polygon)
func parsePointsBlock(lines []string, datumLat, datumLon float64) [][][]float64 {
	var ring [][]float64
	for _, line := range lines {
		parts := strings.Fields(line)
		x, _ := strconv.ParseFloat(strings.TrimPrefix(parts[0], "x:"), 64)
		y, _ := strconv.ParseFloat(strings.TrimPrefix(parts[1], "y:"), 64)
		lon, lat := localToWGS(x, y, datumLat, datumLon)
		ring = append(ring, []float64{lon, lat})
	}
	// Stäng polygonen
	if len(ring) > 0 && (ring[0][0] != ring[len(ring)-1][0] || ring[0][1] != ring[len(ring)-1][1]) {
		ring = append(ring, ring[0])
	}
	return [][][]float64{ring}
}

// classifyZone ger rätt id-baserat på namnet
func classifyZone(name string) string {
	l := strings.ToLower(name)
	if strings.Contains(l, "exclusion") {
		return "exclusion_zone"
	} else if strings.Contains(l, "navigation") || strings.Contains(l, "transport") {
		return "transport_zone"
	}
	return "working_area"
}

// parseBagText extraherar MapArea från rosbag-info (yaml-utdata)
func parseBagText(path string, datumLat, datumLon float64) FeatureCollection {
	// Kör rosbag info i YAML-format
	cmd := exec.Command("rosbag", "info", "--yaml", path)
	outBytes, err := cmd.Output()
	if err != nil {
		log.Fatal("rosbag info failed:", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(outBytes)))

	var features []Feature
	var currentName string
	var collecting bool
	var block []string
	counter := map[string]int{}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "name:") {
			currentName = strings.Trim(strings.TrimPrefix(line, "name:"), ` "`)
		}
		if strings.Contains(line, "points:") {
			collecting = true
			block = nil
			continue
		}
		if collecting {
			lineTrim := strings.TrimSpace(line)
			if strings.HasPrefix(lineTrim, "x:") {
				block = append(block, lineTrim)
				continue
			}
			// sluta samla när vi når tom rad eller nytt objekt
			if lineTrim == "" {
				coords := parsePointsBlock(block, datumLat, datumLon)
				zoneID := classifyZone(currentName)
				counter[zoneID]++
				f := Feature{
					Type:       "Feature",
					Properties: map[string]interface{}{"id": fmt.Sprintf("%s_%d", zoneID, counter[zoneID])},
				}
				f.Geometry.Type = "Polygon"
				f.Geometry.Coordinates = coords
				features = append(features, f)
				collecting = false
			}
		}
	}

	return FeatureCollection{Type: "FeatureCollection", Features: features}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
		os.Exit(1)
	}
	bagPath := os.Args[1]
	outPath := os.Args[2]

	datumLat, datumLon := readDatum("/boot/openmower/mower_config.txt")
	log.Println("➡️  Konverterar ROS-bag till GeoJSON…")
	log.Printf("   Datum LAT: %.6f  LON: %.6f\n", datumLat, datumLon)

	geo := parseBagText(bagPath, datumLat, datumLon)

	outFile, err := os.Create(outPath)
	if err != nil {
		log.Fatal("could not create output file:", err)
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(geo); err != nil {
		log.Fatal("failed to encode GeoJSON:", err)
	}

	log.Println("✅ Klar:", outPath)
}
