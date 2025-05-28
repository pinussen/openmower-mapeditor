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

	rb "github.com/pinussen/rosbag2geojson/pkg"
)

func main() {
	inFile := flag.String("in", "", "Input GeoJSON file")
	outFile := flag.String("out", "map_new.bag", "Output bag file")
	datumLatFlag := flag.Float64("lat", 0, "Datum latitude (optional)")
	datumLonFlag := flag.Float64("lon", 0, "Datum longitude (optional)")
	flag.Parse()

    if *inFile == "" {
        fmt.Println("Usage: geojson2rosbag -in map.geojson [-out map.bag] [-lat datum_lat -lon datum_lon]")
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
	log.Printf("Created temp dir: %s", tmpDir)
	defer func() {
		log.Printf("Cleaning up temp dir: %s", tmpDir)
		os.RemoveAll(tmpDir)
	}()

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
			log.Printf("Docking point coordinates: %v", coords)
			
			// Convert to local coordinates
			x, y := rb.WGSToLocal(coords[0], coords[1], datumLat, datumLon)
			log.Printf("Converted to local coordinates: x=%f, y=%f", x, y)
			
			// Create docking point file
			dockingPointFile = fmt.Sprintf("%s/docking_point.yaml", tmpDir)
			f, err := os.Create(dockingPointFile)
			if err != nil {
				log.Fatal("Failed to create docking point file:", err)
			}
			
			// Write docking point YAML
			yamlContent := fmt.Sprintf(`header:
  seq: 0
  stamp: {secs: %d, nsecs: 0}
  frame_id: "map"
pose:
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
			
			if _, err := f.WriteString(yamlContent); err != nil {
				log.Fatal("Failed to write YAML:", err)
			}
			log.Printf("Wrote docking point YAML to %s:\n%s", dockingPointFile, yamlContent)
			f.Close()

		case "working_area":
			// Parse Polygon coordinates
			var coords [][][]float64
			if err := json.Unmarshal(feature.Geometry.Coordinates, &coords); err != nil {
				log.Fatal("Failed to parse polygon coordinates:", err)
			}
			log.Printf("Working area coordinates: %v", coords)
			
			// Create mowing area file
			mowingAreaFile = fmt.Sprintf("%s/mowing_area.yaml", tmpDir)
			f, err := os.Create(mowingAreaFile)
			if err != nil {
				log.Fatal("Failed to create mowing area file:", err)
			}
			
			// Write mowing area YAML
			yamlContent := fmt.Sprintf(`header:
  seq: 0
  stamp: {secs: %d, nsecs: 0}
  frame_id: "map"
points:
`, time.Now().Unix())
			
			if _, err := f.WriteString(yamlContent); err != nil {
				log.Fatal("Failed to write YAML header:", err)
			}

			for _, point := range coords[0] { // First ring only
				x, y := rb.WGSToLocal(point[0], point[1], datumLat, datumLon)
				pointYaml := fmt.Sprintf("- {x: %f, y: %f}\n", x, y)
				if _, err := f.WriteString(pointYaml); err != nil {
					log.Fatal("Failed to write point:", err)
				}
			}
			log.Printf("Wrote mowing area YAML to %s", mowingAreaFile)
			f.Close()
		}
	}

	// Start rosbag record with signal handling
	recordCmd := exec.Command("rosbag", "record", "-O", *outFile, "/docking_point", "/mowing_areas")
	recordCmd.Stdout = os.Stdout
	recordCmd.Stderr = os.Stderr
	
	// Create a channel to receive OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the record process
	log.Printf("Starting rosbag record: %v", recordCmd.Args)
	if err := recordCmd.Start(); err != nil {
		log.Fatal("Failed to start rosbag record:", err)
	}

	// Ensure cleanup on exit
	defer func() {
		if recordCmd.Process != nil {
			log.Printf("Stopping rosbag record")
			recordCmd.Process.Signal(syscall.SIGINT)
			recordCmd.Wait()
		}
	}()

	// Wait for rosbag to initialize
	time.Sleep(2 * time.Second)

	// Publish messages
	if dockingPointFile != "" {
		cmd := exec.Command("rostopic", "pub", "-f", dockingPointFile, "/docking_point", "geometry_msgs/PoseStamped", "-1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Printf("Publishing docking point with command: %v", cmd.Args)
		log.Printf("YAML file contents:\n%s", yamlContent)
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				log.Printf("rostopic pub stderr: %s", string(exitErr.Stderr))
			}
			log.Fatal("Failed to create docking point message:", err)
		}
	}

	if mowingAreaFile != "" {
		cmd := exec.Command("rostopic", "pub", "-f", mowingAreaFile, "/mowing_areas", "openmower_msgs/MowingAreaList", "-1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		log.Printf("Publishing mowing area with command: %v", cmd.Args)
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				log.Printf("rostopic pub stderr: %s", string(exitErr.Stderr))
			}
			log.Fatal("Failed to publish mowing area:", err)
		}
	}

	// Wait a moment for messages to be recorded
	time.Sleep(1 * time.Second)

	// Clean shutdown
	if recordCmd.Process != nil {
		log.Printf("Stopping rosbag record")
		recordCmd.Process.Signal(syscall.SIGINT)
		recordCmd.Wait()
	}

	log.Printf("✅ Created bag file: %s", *outFile)
} 