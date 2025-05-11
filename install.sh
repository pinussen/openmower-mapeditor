#!/bin/bash
set -e

REPO_DIR="/opt/openmower-mapeditor"
SERVICE_FILE="mapeditor.service"
CONTAINER_NAME="openmower-mapeditor"

# Rensa gamla
systemctl stop "$SERVICE_FILE"   || true
podman rm -f "$CONTAINER_NAME"   || true
rm -rf "$REPO_DIR"

# Klona
git clone https://github.com/pinussen/openmower-mapeditor.git "$REPO_DIR"
cd "$REPO_DIR"

# Bygg Go-verktyget
cd tools/rosbag2geojson
go mod tidy
go build -v -o rosbag2geojson
cp rosbag2geojson /usr/local/bin/
cd ../..

# Bygg container
podman build -t "$CONTAINER_NAME" .

# Installera service
cp "$SERVICE_FILE" /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now "$SERVICE_FILE"

echo "✅ Karteditorn kör nu på http://<din-ip>:8088"
