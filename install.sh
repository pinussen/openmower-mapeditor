#!/bin/bash
set -e

REPO_DIR="/opt/openmower-mapeditor"
SERVICE_NAME="mapeditor.service"
CONTAINER_NAME="openmower-mapeditor"

# 1) Stoppa service, ta bort eventuell container & gammal kod
systemctl stop "$SERVICE_NAME"            || true
podman rm -f "$CONTAINER_NAME"            || true
rm -rf "$REPO_DIR"

# 2) Klona repot på nytt
git clone https://github.com/pinussen/openmower-mapeditor.git "$REPO_DIR"
cd "$REPO_DIR"

# 3) Bygg Go-konvertern och installera på hosten (för pre-extract)
cd tools/rosbag2geojson
rm -f rosbag2geojson
go mod tidy
go build -v -o rosbag2geojson
cp rosbag2geojson /usr/local/bin/
cd ../..

# 4) Förbered data-mapp
mkdir -p data

# 5) Kör en förkonvertering om .bag finns på hosten
BAG_FILE="/root/ros_home/.ros/map.bag"
if [ -f "$BAG_FILE" ]; then
  echo "➡️  Konverterar ROS-bag till GeoJSON…"
  rosbag2geojson "$BAG_FILE" data/map.geojson || true
fi

# 6) Bygg docker/podman-imagen
podman build -t "$CONTAINER_NAME" .

# 7) Säkerställ att det alltid finns en fil att serve:a
mkdir -p data
touch data/map.geojson

# 8) Installera & starta systemd-servicen
cp "$SERVICE_NAME" /etc/systemd/system/
systemctl daemon-reexec
systemctl enable --now "$SERVICE_NAME"

echo "✅ Karteditorn kör nu på http://<din-ip>:8088"
