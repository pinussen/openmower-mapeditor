#!/bin/bash
set -e

REPO_DIR="/opt/openmower-mapeditor"
SERVICE_NAME="mapeditor.service"
CONTAINER_NAME="openmower-mapeditor"

# Stop and clean
systemctl stop "$SERVICE_NAME" || true
podman rm -f "$CONTAINER_NAME" || true
rm -rf "$REPO_DIR"

# Re-clone fresh
git clone https://github.com/pinussen/openmower-mapeditor.git "$REPO_DIR"
cd "$REPO_DIR"

# Build Go converter
cd tools/rosbag2geojson

go mod tidy
go build -o rosbag2geojson
cp rosbag2geojson /usr/local/bin/
cd ../..

# Extract map if .bag exists
if [ -f /root/ros_home/.ros/map.bag ]; then
  /usr/local/bin/rosbag2geojson /root/ros_home/.ros/map.bag "$REPO_DIR/data/map.geojson" || true
fi

# Build container
podman build -t "$CONTAINER_NAME" .

# Ensure data dir exists
mkdir -p "$REPO_DIR/data"
touch "$REPO_DIR/data/map.geojson"

# Install service
cp "$SERVICE_NAME" /etc/systemd/system/
systemctl daemon-reexec
systemctl enable --now "$SERVICE_NAME"

echo "✅ Karteditorn kör nu på http://<din-ip>:8088"
