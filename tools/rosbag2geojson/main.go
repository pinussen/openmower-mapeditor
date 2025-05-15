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

// readDatum reads OM_DATUM_LAT and OM_DATUM_LONG from config file
func readDatum(path string) (float64, float64) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: could not read mower_config.txt: %v", err)
		// Return default values for testing
		return 59.3293, 18.0686
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

// localToWGS converts local x,y (meters) to lon,lat
func localToWGS(x, y, datumLat, datumLon float64) (float64, float64) {
	lat := datumLat + y/111111.0
	lon := datumLon + x/(111111.0*math.Cos(datumLat*math.Pi/180.0))
	return lon, lat
}

// parsePointsBlock reads block of "x:..., y:..." lines and returns a ring (closed polygon)
func parsePointsBlock(lines []string, datumLat, datumLon float64) [][][]float64 {
	var ring [][]float64
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		x, _ := strconv.ParseFloat(strings.TrimPrefix(parts[0], "x:"), 64)
		y, _ := strconv.ParseFloat(strings.TrimPrefix(parts[1], "y:"), 64)
		lon, lat := localToWGS(x, y, datumLat, datumLon)
		ring = append(ring, []float64{lon, lat})
	}
	// Close the polygon
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

// parseBagText extracts MapArea and Pose from rosbag-info (yaml output)
func parseBagText(path string, datumLat, datumLon float64) FeatureCollection {
	// Run rosbag info in YAML format
	cmd := exec.Command("rosbag", "info", "--yaml", path)
	outBytes, err := cmd.Output()
	if err != nil {
		log.Fatal("rosbag info failed:", err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(outBytes)))

	var features []Feature
	var currentTopic string
	var collecting bool
	var block []string
	counter := map[string]int{}

	for scanner.Scan() {
		line := scanner.Text()
		
		// Track which topic we're processing
		if strings.Contains(line, "topics:") {
			collecting = false
		} else if strings.Contains(line, "docking_point") {
			currentTopic = "docking_point"
			collecting = true
			block = nil
		} else if strings.Contains(line, "mowing_areas") {
			currentTopic = "mowing_areas"
			collecting = true
			block = nil
		}

		if collecting {
			lineTrim := strings.TrimSpace(line)
			if strings.HasPrefix(lineTrim, "x:") || strings.HasPrefix(lineTrim, "y:") {
				block = append(block, lineTrim)
			} else if len(block) > 0 && lineTrim == "" {
				// Process collected points based on topic
				if currentTopic == "docking_point" {
					// For docking point, create a single point feature
					if len(block) >= 2 {
						x, _ := strconv.ParseFloat(strings.TrimPrefix(block[0], "x:"), 64)
						y, _ := strconv.ParseFloat(strings.TrimPrefix(block[1], "y:"), 64)
						lon, lat := localToWGS(x, y, datumLat, datumLon)
						features = append(features, Feature{
							Type: "Feature",
							Properties: map[string]interface{}{
								"id":   "docking_point",
								"type": "docking_point",
							},
							Geometry: struct {
								Type        string        `json:"type"`
								Coordinates [][][]float64 `json:"coordinates"`
							}{
								Type:        "Point",
								Coordinates: [][][]float64{{[]float64{lon, lat}}},
							},
						})
					}
				} else if currentTopic == "mowing_areas" {
					// For mowing areas, create a polygon feature
					coords := parsePointsBlock(block, datumLat, datumLon)
					if len(coords[0]) > 2 { // Need at least 3 points for a valid polygon
						zoneID := classifyZone(currentTopic)
						counter[zoneID]++
						features = append(features, Feature{
							Type: "Feature",
							Properties: map[string]interface{}{
								"id":   fmt.Sprintf("%s_%d", zoneID, counter[zoneID]),
								"type": zoneID,
							},
							Geometry: struct {
								Type        string        `json:"type"`
								Coordinates [][][]float64 `json:"coordinates"`
							}{
								Type:        "Polygon",
								Coordinates: coords,
							},
						})
					}
				}
				block = nil
			}
		}
	}

	return FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
		os.Exit(1)
	}
	bagPath := os.Args[1]
	outPath := os.Args[2]

	datumLat, datumLon := readDatum("/boot/openmower/mower_config.txt")
	log.Printf("➡️  Converting ROS bag to GeoJSON...")
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

	log.Printf("✅ Done: %s", outPath)
}
