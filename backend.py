#!/usr/bin/env python3
import os
import subprocess
from flask import Flask, abort, jsonify

app = Flask(__name__)

# === Justera dessa om du mountar till någon annan sökväg ===
BAG_PATH     = "/data/map.bag"
GEOJSON_PATH = "/data/map.geojson"

@app.route("/load", methods=["GET"])
def load():
    """Returnerar befintlig GeoJSON (om den finns)."""
    if not os.path.isfile(GEOJSON_PATH):
        abort(404, "Ingen GeoJSON hittades")
    with open(GEOJSON_PATH) as f:
        return f.read(), 200, {"Content-Type": "application/json"}

@app.route("/extract", methods=["POST"])
def extract():
    """Kör rosbag2geojson på map.bag → map.geojson."""
    if not os.path.isfile(BAG_PATH):
        abort(404, f"map.bag hittades inte på {BAG_PATH}")
    try:
        proc = subprocess.run(
            ["rosbag2geojson", BAG_PATH, GEOJSON_PATH],
            check=True,
            capture_output=True,
            text=True
        )
        app.logger.info("rosbag2geojson stdout: %s", proc.stdout)
        app.logger.info("rosbag2geojson stderr: %s", proc.stderr)
        return jsonify({"status": "extracted"}), 200
    except subprocess.CalledProcessError as e:
        app.logger.error("Extraction failed: %s", e.stderr)
        abort(500, f"Extraction failed: {e.stderr}")

if __name__ == "__main__":
    # Kör Flask på 0.0.0.0:8088 så att hosten kommer åt den
    app.run(host="0.0.0.0", port=8088)
