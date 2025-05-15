#!/usr/bin/env python3
import os
import json
import subprocess
import logging
from flask import Flask, abort, jsonify, send_from_directory, request

# Set up logging
logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

app = Flask(__name__, static_folder='static')

# === Justera dessa om du mountar till någon annan sökväg ===
BAG_PATH     = "/data/ros/map.bag"
GEOJSON_PATH = "/opt/openmower-mapeditor/data/map.geojson"

@app.route("/")
def root():
    logger.debug("Serving root page")
    return app.send_static_file('index.html')

@app.route("/load", methods=["GET"])
def load():
    """Returnerar befintlig GeoJSON (om den finns)."""
    logger.debug("Loading GeoJSON")
    if not os.path.isfile(GEOJSON_PATH):
        return jsonify({"type": "FeatureCollection", "features": []}), 200
    with open(GEOJSON_PATH) as f:
        return f.read(), 200, {"Content-Type": "application/json"}

@app.route("/save", methods=["POST"])
def save():
    """Save the GeoJSON data."""
    logger.debug("Saving GeoJSON")
    if not request.is_json:
        abort(400, "Expected JSON data")
    try:
        with open(GEOJSON_PATH, 'w') as f:
            json.dump(request.json, f)
        return jsonify({"status": "saved"}), 200
    except Exception as e:
        app.logger.error("Save failed: %s", str(e))
        abort(500, f"Save failed: {str(e)}")

@app.route("/extract", methods=["POST"])
def extract():
    """Kör rosbag2geojson på map.bag → map.geojson."""
    logger.debug("Extracting from rosbag")
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
    # Enable debug mode for better error messages
    app.debug = True
    logger.info("Starting Flask server on 0.0.0.0:8088")
    # Try to bind to all interfaces explicitly
    app.run(host="0.0.0.0", port=8088, threaded=True)
