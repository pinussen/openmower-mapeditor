package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	rb "github.com/pinussen/rosbag2geojson/pkg"
)

// readPoseFromBag reads the docking point pose from the bag file
func readPoseFromBag(bagPath string, datumLat, datumLon float64) *rb.Feature {
	// Try to read the message directly
	cmd := exec.Command("rostopic", "echo", "-b", bagPath, "-n", "1", "/docking_point")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to read docking point: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("stderr: %s", string(exitErr.Stderr))
		}
		return nil
	}

	log.Printf("Docking point output:\n%s", string(output))
	
	var x, y float64
	inPosition := false
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "position:") {
			inPosition = true
			continue
		}
		if strings.HasPrefix(line, "orientation:") {
			inPosition = false
			continue
		}
		if inPosition {
			if strings.HasPrefix(line, "x:") {
				x, _ = strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "x:")), 64)
			}
			if strings.HasPrefix(line, "y:") {
				y, _ = strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "y:")), 64)
			}
		}
	}

	// Add round-trip test
	log.Printf("\n=== Testing docking point coordinate conversion ===")
	rb.TestRoundTrip(x, y, datumLat, datumLon)
	log.Printf("===============================================\n")

	lon, lat := rb.LocalToWGS(x, y, datumLat, datumLon)
	log.Printf("Creating docking point at lon: %f, lat: %f", lon, lat)
	
	coords, _ := json.Marshal([]float64{lon, lat})
	return &rb.Feature{
		Type: "Feature",
		Properties: map[string]interface{}{
			"id":   "docking_point",
			"type": "docking_point",
		},
		Geometry: rb.Geometry{
			Type:        "Point",
			Coordinates: coords,
		},
	}
}

// readMapAreaFromBag reads the mowing area from the bag file
func readMapAreaFromBag(bagPath string, datumLat, datumLon float64) *rb.Feature {
	// Try to read the message directly
	cmd := exec.Command("rostopic", "echo", "-b", bagPath, "-n", "1", "/mowing_areas")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Warning: Failed to read mowing areas: %v", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("stderr: %s", string(exitErr.Stderr))
		}
		return nil
	}

	log.Printf("Area output:\n%s", string(output))

	var points [][]float64
	var currentPoint []float64
	inPoints := false
	expectingPoint := false
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		switch {
		case strings.HasPrefix(line, "area:"):
			continue
		case strings.HasPrefix(line, "points:"):
			inPoints = true
			continue
		case inPoints && line == "-":
			expectingPoint = true
			currentPoint = nil
			continue
		case expectingPoint:
			if strings.HasPrefix(line, "x:") {
				x, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "x:")), 64)
				currentPoint = []float64{x}
			} else if strings.HasPrefix(line, "y:") && len(currentPoint) == 1 {
				y, _ := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(line, "y:")), 64)
				currentPoint = append(currentPoint, y)
				points = append(points, currentPoint)
				
				// Add round-trip test for each point
				log.Printf("\n=== Testing area point %d coordinate conversion ===", len(points))
				rb.TestRoundTrip(currentPoint[0], currentPoint[1], datumLat, datumLon)
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
		lon, lat := rb.LocalToWGS(p[0], p[1], datumLat, datumLon)
		coordinates = append(coordinates, []float64{lon, lat})
	}

	// Close the polygon
	if len(coordinates) > 0 {
		coordinates = append(coordinates, coordinates[0])
	}

	log.Printf("Created polygon with %d coordinates", len(coordinates))

	coords, _ := json.Marshal([][][]float64{coordinates})
	return &rb.Feature{
		Type: "Feature",
		Properties: map[string]interface{}{
			"id":   "working_area_1",
			"type": "working_area",
		},
		Geometry: rb.Geometry{
			Type:        "Polygon",
			Coordinates: coords,
		},
	}
}

func main() {
	if len(os.Args) != 5 {
		fmt.Println("Usage: rosbag2geojson -in <input.bag> -out <output.geojson>")
		os.Exit(1)
	}

	var bagPath, outPath string
	for i := 1; i < len(os.Args); i += 2 {
		switch os.Args[i] {
		case "-in":
			bagPath = os.Args[i+1]
		case "-out":
			outPath = os.Args[i+1]
		}
	}

	if bagPath == "" || outPath == "" {
		fmt.Println("Usage: rosbag2geojson -in <input.bag> -out <output.geojson>")
		os.Exit(1)
	}

	datumLat, datumLon := rb.ReadDatum("/boot/openmower/mower_config.txt")
	log.Printf("➡️  Konverterar ROS-bag till GeoJSON...")
	log.Printf("   Datum LAT: %.6f  LON: %.6f\n", datumLat, datumLon)

	var features []rb.Feature

	// Read docking point
	if dockingPoint := readPoseFromBag(bagPath, datumLat, datumLon); dockingPoint != nil {
		features = append(features, *dockingPoint)
	}

	// Read mowing area
	if mowingArea := readMapAreaFromBag(bagPath, datumLat, datumLon); mowingArea != nil {
		features = append(features, *mowingArea)
	}

	geo := rb.FeatureCollection{
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