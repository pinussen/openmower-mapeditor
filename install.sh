#!/bin/bash
set -e

REPO_URL="https://github.com/pinussen/openmower-mapeditor.git"
REPO_DIR="/opt/openmower-mapeditor"
SERVICE_NAME="mapeditor.service"
CONTAINER_NAME="openmower-mapeditor"
BINARY_PATH="$REPO_DIR/tools/rosbag2geojson/rosbag2geojson"

# Stop and clean
systemctl stop "$SERVICE_NAME" || true
podman rm -f "$CONTAINER_NAME" || true
rm -rf "$REPO_DIR"

# Clone latest
git clone "$REPO_URL" "$REPO_DIR"
cd "$REPO_DIR"

# Build rosbag2geojson binary
cd tools/rosbag2geojson
go mod tidy
go build -o rosbag2geojson
cd ../..

# Build container
podman build -t "$CONTAINER_NAME" .

# Prepare data volume and dummy geojson
mkdir -p "$REPO_DIR/data"
touch "$REPO_DIR/data/map.geojson"

# Install and enable systemd service
cp "$SERVICE_NAME" /etc/systemd/system/
systemctl daemon-reexec
systemctl enable --now "$SERVICE_NAME"

echo "✅ Karteditorn kör nu på http://<din-ip>:8088"
