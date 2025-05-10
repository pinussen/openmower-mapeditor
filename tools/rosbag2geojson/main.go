
package main

import (
    "bufio"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "math"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

type Feature struct {
    Type       string                 `json:"type"`
    Properties map[string]interface{} `json:"properties"`
    Geometry   Geometry               `json:"geometry"`
}

type Geometry struct {
    Type        string        `json:"type"`
    Coordinates [][][]float64 `json:"coordinates"`
}

type FeatureCollection struct {
    Type     string    `json:"type"`
    Features []Feature `json:"features"`
}

func parseDatumFromConfig(path string) (float64, float64, error) {
    file, err := os.Open(path)
    if err != nil {
        return 0, 0, err
    }
    defer file.Close()

    var lat, lon float64
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "export OM_DATUM_LAT=") {
            val := strings.Trim(strings.Split(line, "=")[1], "\"")
            lat, _ = strconv.ParseFloat(val, 64)
        }
        if strings.HasPrefix(line, "export OM_DATUM_LONG=") {
            val := strings.Trim(strings.Split(line, "=")[1], "\"")
            lon, _ = strconv.ParseFloat(val, 64)
        }
    }
    if lat == 0 || lon == 0 {
        return 0, 0, errors.New("datum values not found")
    }
    return lat, lon, nil
}

func transformXYToLatLon(x, y, datumLat, datumLon float64) (float64, float64) {
    lat := datumLat + (y / 111320.0)
    lon := datumLon + (x / (40075000.0 * math.Cos(datumLat*math.Pi/180) / 360.0))
    return lon, lat
}

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: rosbag2geojson <input.bag> <output.geojson>")
        os.Exit(1)
    }

 //   bagPath := os.Args[1]
    outPath := os.Args[2]
    datumLat, datumLon, err := parseDatumFromConfig("/boot/openmower/mower_config.txt")
    if err != nil {
        log.Fatal("failed to read datum:", err)
    }

    fmt.Println("Using datum:")
    fmt.Println("  LAT:", datumLat)
    fmt.Println("  LON:", datumLon)

    // Dummy polygon for debug purposes
    polygon := [][]float64{}
    for _, pt := range [][2]float64{
        {0, 0},
        {1, 0},
        {1, 1},
        {0, 1},
        {0, 0},
    } {
        fmt.Printf("  Original XY: (%.3f, %.3f)\n", pt[0], pt[1])
        lon, lat := transformXYToLatLon(pt[0], pt[1], datumLat, datumLon)
        fmt.Printf("  Transformed to lat/lon: (%.6f, %.6f)\n", lat, lon)
        polygon = append(polygon, []float64{lon, lat})
    }

    geo := FeatureCollection{
        Type: "FeatureCollection",
        Features: []Feature{
            {
                Type: "Feature",
                Properties: map[string]interface{}{
                    "id": "working_area",
                },
                Geometry: Geometry{
                    Type:        "Polygon",
                    Coordinates: [][][]float64{polygon},
                },
            },
        },
    }

    if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
        log.Fatal(err)
    }
    f, err := os.Create(outPath)
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    enc.Encode(geo)
    fmt.Println("✅ Wrote GeoJSON with transformed coords to", outPath)
}
