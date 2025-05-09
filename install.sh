#!/bin/bash
set -e

REPO_DIR="/opt/openmower-mapeditor"
SERVICE_NAME="mapeditor.service"
CONTAINER_NAME="openmower-mapeditor"

# Stoppa och ta bort gammal container om den finns
systemctl stop "$SERVICE_NAME" || true
podman rm -f "$CONTAINER_NAME" || true

# Ta bort gammalt repo om det finns
rm -rf "$REPO_DIR"

# Klona nytt
git clone https://github.com/placeholder/openmower-mapeditor.git "$REPO_DIR"
cd "$REPO_DIR"

# Bygg containern
podman build -t "$CONTAINER_NAME" .

# Kopiera systemd-tjänst
cp "$SERVICE_NAME" /etc/systemd/system/

# Aktivera och starta
systemctl daemon-reexec
systemctl enable --now "$SERVICE_NAME"

echo "✅ Karteditor återinstallerad och körs nu på http://<din-ip>:8088"
