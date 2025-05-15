package main

import (
    "encoding/json"
    "log"
    "math"
    "os"
    "strconv"
    "strings"
)

type Geometry struct {
    Type        string          `json:"type"`
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