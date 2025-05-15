package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	rb "github.com/pinussen/rosbag2geojson"
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

	var fc rb.FeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		log.Fatal("Failed to parse GeoJSON:", err)
	}

	// Get datum from config if not provided
	datumLat, datumLon := *datumLatFlag, *datumLonFlag
	if datumLat == 0 || datumLon == 0 {
		datumLat, datumLon = rb.ReadDatum("/boot/openmower/mower_config.txt")
	}

	// Create temporary directory for message files
	tmpDir, err := os.MkdirTemp("", "geojson2rosbag-*")
	if err != nil {
		log.Fatal("Failed to create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)

	// Process features and create message files
	dockingPointFile := ""
	mowingAreaFile := ""

	for _, feature := range fc.Features {
		switch feature.Properties["type"] {
		case "docking_point":
			// Parse Point coordinates
			var coords []float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				log.Fatal("Failed to parse docking point coordinates:", err)
			}
			
			// Convert to local coordinates
			x, y := rb.WGSToLocal(coords[0], coords[1], datumLat, datumLon)
			
			// Create docking point file
			dockingPointFile = fmt.Sprintf("%s/docking_point.yaml", tmpDir)
			f, err := os.Create(dockingPointFile)
			if err != nil {
				log.Fatal("Failed to create docking point file:", err)
			}
			
			// Write docking point YAML
			fmt.Fprintf(f, `header:
  seq: 0
  stamp: {secs: %d, nsecs: 0}
  frame_id: "map"
position:
  x: %f
  y: %f
  z: 0.0
orientation:
  x: 0.0
  y: 0.0
  z: 0.0
  w: 1.0
`, time.Now().Unix(), x, y)
			f.Close()

		case "working_area":
			// Parse Polygon coordinates
			var coords [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				log.Fatal("Failed to parse polygon coordinates:", err)
			}
			
			// Create mowing area file
			mowingAreaFile = fmt.Sprintf("%s/mowing_area.yaml", tmpDir)
			f, err := os.Create(mowingAreaFile)
			if err != nil {
				log.Fatal("Failed to create mowing area file:", err)
			}
			
			// Write mowing area YAML
			fmt.Fprintf(f, `header:
  seq: 0
  stamp: {secs: %d, nsecs: 0}
  frame_id: "map"
points:
`, time.Now().Unix())
			
			for _, point := range coords[0] { // First ring only
				x, y := rb.WGSToLocal(point[0], point[1], datumLat, datumLon)
				fmt.Fprintf(f, "- {x: %f, y: %f}\n", x, y)
			}
			f.Close()
		}
	}

	// Start rosbag record with signal handling
	recordCmd := exec.Command("rosbag", "record", "-O", *outFile, "/docking_point", "/mowing_areas")
	
	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the record process
	if err := recordCmd.Start(); err != nil {
		log.Fatal("Failed to start rosbag record:", err)
	}

	// Ensure cleanup on exit
	defer func() {
		if recordCmd.Process != nil {
			recordCmd.Process.Signal(syscall.SIGINT)
			recordCmd.Wait()
		}
	}()

	// Wait for rosbag to initialize
	time.Sleep(2 * time.Second)

	// Publish messages
	if dockingPointFile != "" {
		cmd := exec.Command("rostopic", "pub", "-f", dockingPointFile, "/docking_point", "geometry_msgs/Pose", "-1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("Failed to publish docking point:", err)
		}
	}

	if mowingAreaFile != "" {
		cmd := exec.Command("rostopic", "pub", "-f", mowingAreaFile, "/mowing_areas", "openmower_msgs/MowingAreaList", "-1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("Failed to publish mowing area:", err)
		}
	}

	// Wait a moment for messages to be recorded
	time.Sleep(1 * time.Second)

	// Clean shutdown
	if recordCmd.Process != nil {
		recordCmd.Process.Signal(syscall.SIGINT)
		recordCmd.Wait()
	}

	log.Printf("✅ Created bag file: %s", *outFile)
} 