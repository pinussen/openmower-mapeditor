#!/bin/bash
set -e

REPO_DIR="/opt/openmower-mapeditor"
SERVICE_NAME="mapeditor.service"

# Klona repo om det inte finns
if [ ! -d "$REPO_DIR" ]; then
  git clone https://github.com/placeholder/openmower-mapeditor.git "$REPO_DIR"
fi
cd "$REPO_DIR"

# Bygg containern (med full bildreferens)
podman build -t openmower-mapeditor --build-arg BASE_IMAGE=docker.io/library/python:3.11-slim .

# Kopiera systemd-tjänstfil
cp "$SERVICE_NAME" /etc/systemd/system/

# Aktivera och starta
systemctl daemon-reexec
systemctl enable --now "$SERVICE_NAME"

echo "✅ Karteditor installerad och körs nu på http://<din-ip>:8088"
