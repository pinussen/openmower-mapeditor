#!/bin/sh
set -e

# 1) Om det ligger en bag-fil monterad, konvertera direkt
if [ -f /data/map.bag ]; then
  echo "➡️  Konverterar ROS-bag till GeoJSON…"
  rosbag2geojson /data/map.bag /data/map.geojson || true
fi

# 2) Starta Flask-appen
exec python3 /app/app.py
