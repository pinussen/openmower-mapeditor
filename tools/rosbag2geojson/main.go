package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func convertBagToGeoJSON(bagPath, outputPath string) error {
	cmd := exec.Command("python3", "/opt/openmower-mapeditor/tools/extract_geojson.py", bagPath, outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
		os.Exit(1)
	}

	inPath := os.Args[1]
	outPath := os.Args[2]

	err := convertBagToGeoJSON(inPath, outPath)
	if err != nil {
		log.Fatal("conversion failed:", err)
	}

	fmt.Println("✅ Converted", inPath, "to", outPath)
}
