#!/bin/bash
set -e

# Source ROS environment
source /opt/ros/noetic/setup.bash

# Check if the bag file exists
if [ -f "/data/ros/map.bag" ]; then
    echo "📍 Found map.bag file"
    rosbag info /data/ros/map.bag
    
    # Convert to GeoJSON if the tool is available
    if command -v rosbag2geojson &> /dev/null; then
        echo "🔄 Converting ROS bag to GeoJSON..."
        rosbag2geojson -in /data/ros/map.bag -out /data/map.geojson || echo "⚠️  Conversion failed"
    fi
else
    echo "ℹ️  No map.bag file found in /data/ros/"
fi

# Start the Flask application
echo "🚀 Starting Flask application..."
exec python3 backend.py
