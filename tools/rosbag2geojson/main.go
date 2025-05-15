package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Geometry struct {
	Type        string      `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   Geometry               `json:"geometry"`
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

// localToWGS converts local x,y (meters) to lon,lat using Haversine formula
func localToWGS(x, y, datumLat, datumLon float64) (float64, float64) {
	// Earth's radius in meters
	const R = 6378137.0

	// Convert datum to radians
	lat1 := datumLat * math.Pi / 180.0
	lon1 := datumLon * math.Pi / 180.0

	// Calculate new latitude
	// Using Haversine formula rearranged for latitude
	lat2 := math.Asin(math.Sin(lat1)*math.Cos(y/R) + 
			math.Cos(lat1)*math.Sin(y/R)*math.Cos(0))

	// Calculate new longitude
	// Using Haversine formula rearranged for longitude
	lon2 := lon1 + math.Atan2(math.Sin(x/R)*math.Cos(lat1),
			math.Cos(x/R) - math.Sin(lat1)*math.Sin(lat2))

	// Convert back to degrees
	return lon2 * 180.0 / math.Pi, lat2 * 180.0 / math.Pi
}

// readPoseFromBag reads the docking point pose from the bag file
func readPoseFromBag(bagPath string, datumLat, datumLon float64) *Feature {
	cmd := exec.Command("rostopic", "echo", "-b", bagPath, "-n", "1", "/docking_point")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to read docking point: %v", err)
		return nil
	}

	log.Printf("Docking point output:\n%s", string(output))
	
	var x, y float64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "x:") {
			x, _ = strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "x:")), 64)
		}
		if strings.HasPrefix(line, "y:") {
			y, _ = strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "y:")), 64)
		}
	}

	lon, lat := localToWGS(x, y, datumLat, datumLon)
	log.Printf("Creating docking point at lon: %f, lat: %f", lon, lat)
	
	coords, _ := json.Marshal([]float64{lon, lat})
	return &Feature{
		Type: "Feature",
		Properties: map[string]interface{}{
			"id":   "docking_point",
			"type": "docking_point",
		},
		Geometry: Geometry{
			Type:        "Point",
			Coordinates: coords,
		},
	}
}

// readMapAreaFromBag reads the mowing area from the bag file
func readMapAreaFromBag(bagPath string, datumLat, datumLon float64) *Feature {
	cmd := exec.Command("rostopic", "echo", "-b", bagPath, "-n", "1", "/mowing_areas")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to read mowing areas: %v", err)
		return nil
	}

	log.Printf("Mowing areas output:\n%s", string(output))

	var points [][]float64
	var currentPoint []float64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "x:") {
			x, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "x:")), 64)
			currentPoint = []float64{x}
		}
		if strings.HasPrefix(line, "y:") {
			y, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "y:")), 64)
			if len(currentPoint) == 1 {
				currentPoint = append(currentPoint, y)
				points = append(points, currentPoint)
			}
		}
	}

	if len(points) < 3 {
		log.Printf("Warning: Not enough points for a polygon: %d", len(points))
		return nil
	}

	// Convert all points to WGS84
	var coordinates [][]float64
	for _, p := range points {
		lon, lat := localToWGS(p[0], p[1], datumLat, datumLon)
		coordinates = append(coordinates, []float64{lon, lat})
	}

	// Close the polygon
	if len(coordinates) > 0 {
		coordinates = append(coordinates, coordinates[0])
	}

	log.Printf("Created polygon with %d coordinates", len(coordinates))

	coords, _ := json.Marshal([][][]float64{coordinates})
	return &Feature{
		Type: "Feature",
		Properties: map[string]interface{}{
			"id":   "working_area_1",
			"type": "working_area",
		},
		Geometry: Geometry{
			Type:        "Polygon",
			Coordinates: coords,
		},
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

	var features []Feature

	// Read docking point
	if dockingPoint := readPoseFromBag(bagPath, datumLat, datumLon); dockingPoint != nil {
		features = append(features, *dockingPoint)
	}

	// Read mowing area
	if mowingArea := readMapAreaFromBag(bagPath, datumLat, datumLon); mowingArea != nil {
		features = append(features, *mowingArea)
	}

	geo := FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}

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
