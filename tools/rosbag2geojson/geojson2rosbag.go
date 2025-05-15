package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	inFile := flag.String("in", "", "Input GeoJSON file")
	outFile := flag.String("out", "", "Output bag file")
	datumLatFlag := flag.Float64("lat", 0, "Datum latitude (optional)")
	datumLonFlag := flag.Float64("lon", 0, "Datum longitude (optional)")
	flag.Parse()

	if *inFile == "" || *outFile == "" {
		fmt.Println("Usage: geojson2rosbag -in map.geojson -out map.bag [-lat datum_lat -lon datum_lon]")
		os.Exit(1)
	}

	// Read GeoJSON
	data, err := os.ReadFile(*inFile)
	if err != nil {
		log.Fatal("Failed to read GeoJSON:", err)
	}

	var fc FeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		log.Fatal("Failed to parse GeoJSON:", err)
	}

	// Get datum from config if not provided
	datumLat, datumLon := *datumLatFlag, *datumLonFlag
	if datumLat == 0 || datumLon == 0 {
		datumLat, datumLon = readDatum("/boot/openmower/mower_config.txt")
	}

	// Create temporary files for the messages
	dockingPointFile, err := os.CreateTemp("", "docking_point_*.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dockingPointFile.Name())

	mowingAreaFile, err := os.CreateTemp("", "mowing_area_*.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(mowingAreaFile.Name())

	// Process features
	for _, feature := range fc.Features {
		switch feature.Properties["type"] {
		case "docking_point":
			// Parse Point coordinates
			var coords []float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				log.Fatal("Failed to parse docking point coordinates:", err)
			}
			// Convert to local coordinates
			x, y := WGSToLocal(coords[0], coords[1], datumLat, datumLon)
			
			// Write docking point YAML
			fmt.Fprintf(dockingPointFile, `position:
  x: %f
  y: %f
  z: 0.0
orientation:
  x: 0.0
  y: 0.0
  z: 0.0
  w: 1.0
`, x, y)

		case "working_area":
			// Parse Polygon coordinates
			var coords [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				log.Fatal("Failed to parse polygon coordinates:", err)
			}
			
			// Write mowing area YAML
			fmt.Fprintf(mowingAreaFile, "points:\n")
			for _, point := range coords[0] { // First ring only
				x, y := WGSToLocal(point[0], point[1], datumLat, datumLon)
				fmt.Fprintf(mowingAreaFile, "- x: %f\n  y: %f\n", x, y)
			}
		}
	}

	dockingPointFile.Close()
	mowingAreaFile.Close()

	// Create bag file using rostopic
	log.Println("Creating bag file...")

	// Create docking point message
	cmd := exec.Command("rostopic", "pub", "-f", dockingPointFile.Name(), "-r", "1", "-p", "/docking_point", "geometry_msgs/Pose", "--once")
	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to create docking point message:", err)
	}

	// Create mowing area message
	cmd = exec.Command("rostopic", "pub", "-f", mowingAreaFile.Name(), "-r", "1", "-p", "/mowing_areas", "mower_map/MapArea", "--once")
	if err := cmd.Run(); err != nil {
		log.Fatal("Failed to create mowing area message:", err)
	}

	log.Printf("✅ Created bag file: %s", *outFile)
} 