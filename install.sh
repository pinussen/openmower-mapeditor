#!/bin/bash
set -e

REPO_DIR="/opt/openmower-mapeditor"
SERVICE_FILE="mapeditor.service"
CONTAINER_NAME="openmower-mapeditor"

# Clean up old installation
echo "🧹 Cleaning up old installation..."
systemctl stop "$SERVICE_FILE" || true
podman rm -f "$CONTAINER_NAME" || true
rm -rf "$REPO_DIR"

# Clone repository
echo "📥 Cloning repository..."
git clone https://github.com/pinussen/openmower-mapeditor.git "$REPO_DIR"
cd "$REPO_DIR"

# Build Go tool
echo "🔨 Building rosbag2geojson tool..."
cd tools/rosbag2geojson
go mod tidy
GOARCH=arm64 go build -v -o rosbag2geojson
cp rosbag2geojson /usr/local/bin/
cd ../..

# Build container
echo "🏗️ Building container..."
podman build --platform linux/arm64 -t "$CONTAINER_NAME" .

# Install service
echo "🔧 Installing service..."
cp "$SERVICE_FILE" /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now "$SERVICE_FILE"

echo "✅ Map editor is now running on http://<your-ip>:8088"
echo "📝 Check logs with: podman logs -f $CONTAINER_NAME"
