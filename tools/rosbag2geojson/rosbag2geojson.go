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

// WGSToLocal converts lon,lat to local x,y (meters) using Haversine formula
func WGSToLocal(lon, lat, datumLat, datumLon float64) (float64, float64) {
	// Earth's radius in meters
	const R = 6378137.0

	// Convert to radians
	lat1 := datumLat * math.Pi / 180.0
	lon1 := datumLon * math.Pi / 180.0
	lat2 := lat * math.Pi / 180.0
	lon2 := lon * math.Pi / 180.0

	// Calculate x distance
	deltaLon := lon2 - lon1
	x := R * deltaLon * math.Cos(lat1)

	// Calculate y distance
	y := R * (lat2 - lat1)

	return x, y
}

// testRoundTrip tests coordinate conversion accuracy
func testRoundTrip(x, y float64, datumLat, datumLon float64) {
	log.Printf("Original local coordinates: x=%.3f, y=%.3f", x, y)
	
	// Convert to WGS84
	lon, lat := localToWGS(x, y, datumLat, datumLon)
	log.Printf("Converted to WGS84: lon=%.6f, lat=%.6f", lon, lat)
	
	// Convert back to local
	x2, y2 := WGSToLocal(lon, lat, datumLat, datumLon)
	log.Printf("Converted back to local: x=%.3f, y=%.3f", x2, y2)
	
	// Calculate error
	dx := x2 - x
	dy := y2 - y
	error := math.Sqrt(dx*dx + dy*dy)
	log.Printf("Error in meters: %.3f", error)
}

// readPoseFromBag reads the docking point pose from the bag file
func readPoseFromBag(bagPath string, datumLat, datumLon float64) *Feature {
	// First try to get message type
	cmd := exec.Command("rosbag", "info", bagPath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to get bag info: %v", err)
		return nil
	}

	log.Printf("Bag info output:\n%s", string(output))

	// Check if it's PoseStamped or Pose
	isPoseStamped := strings.Contains(string(output), "geometry_msgs/PoseStamped")
	log.Printf("Message type detection - isPoseStamped: %v", isPoseStamped)
	
	// Read the message
	cmd = exec.Command("rostopic", "echo", "-b", bagPath, "-n", "1", "/docking_point")
	output, err = cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to read docking point: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("stderr: %s", string(exitErr.Stderr))
		}
		return nil
	}

	log.Printf("Docking point output:\n%s", string(output))
	
	var x, y float64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "x:") {
			// For PoseStamped, we need to check if we're in the pose.position section
			if isPoseStamped && !strings.Contains(string(output), "position:") {
				continue
			}
			x, _ = strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "x:")), 64)
		}
		if strings.HasPrefix(line, "y:") {
			// For PoseStamped, we need to check if we're in the pose.position section
			if isPoseStamped && !strings.Contains(string(output), "position:") {
				continue
			}
			y, _ = strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "y:")), 64)
		}
	}

	// Add round-trip test
	log.Printf("\n=== Testing docking point coordinate conversion ===")
	testRoundTrip(x, y, datumLat, datumLon)
	log.Printf("===============================================\n")

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
	// First try to get message type
	cmd := exec.Command("rosbag", "info", bagPath)
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to get bag info: %v", err)
		return nil
	}

	log.Printf("Bag info output:\n%s", string(output))

	// Check message type
	isMowingAreaList := strings.Contains(string(output), "openmower_msgs/MowingAreaList")
	log.Printf("Message type detection - isMowingAreaList: %v", isMowingAreaList)
	
	cmd = exec.Command("rostopic", "echo", "-b", bagPath, "-n", "1", "/mowing_areas")
	output, err = cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to read mowing areas: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("stderr: %s", string(exitErr.Stderr))
		}
		return nil
	}

	log.Printf("Mowing areas output:\n%s", string(output))

	var points [][]float64
	var currentPoint []float64
	lines := strings.Split(string(output), "\n")
	inPoints := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// For MowingAreaList type, we need to check if we're in the points section
		if isMowingAreaList && strings.Contains(line, "points:") {
			inPoints = true
			continue
		}
		
		if strings.HasPrefix(line, "x:") {
			if isMowingAreaList && !inPoints {
				continue
			}
			x, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "x:")), 64)
			currentPoint = []float64{x}
		}
		if strings.HasPrefix(line, "y:") {
			if isMowingAreaList && !inPoints {
				continue
			}
			y, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "y:")), 64)
			if len(currentPoint) == 1 {
				currentPoint = append(currentPoint, y)
				points = append(points, currentPoint)
				
				// Add round-trip test for each point
				log.Printf("\n=== Testing mowing area point %d coordinate conversion ===", len(points))
				testRoundTrip(currentPoint[0], currentPoint[1], datumLat, datumLon)
				log.Printf("===============================================\n")
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
