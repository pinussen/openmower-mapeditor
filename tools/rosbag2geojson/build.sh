#!/bin/bash
set -e

echo "🔨 Building rosbag2geojson tools..."

# Build rosbag2geojson
echo "Building rosbag2geojson..."
cd cmd/rosbag2geojson
go build -v -o ../../rosbag2geojson
cd ../..

# Build geojson2rosbag
echo "Building geojson2rosbag..."
cd cmd/geojson2rosbag
go build -v -o ../../geojson2rosbag
cd ../..

echo "✅ Build complete!" 