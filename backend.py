# backend.py

from flask import Flask, request, jsonify, send_from_directory
import os, json, subprocess

app = Flask(__name__)

GEOJSON_PATH = "/data/map.geojson"

@app.route("/load", methods=["GET"])
def load_geojson():
    if os.path.exists(GEOJSON_PATH):
        with open(GEOJSON_PATH, "r") as f:
            data = json.load(f)
        return jsonify(data)
    return jsonify({"type": "FeatureCollection", "features": []})

@app.route("/save", methods=["POST"])
def save_geojson():
    data = request.get_json()
    with open(GEOJSON_PATH, "w") as f:
        json.dump(data, f, indent=2)
    return "Saved", 200

@app.route("/extract", methods=["POST"])
def extract():
    try:
        subprocess.run([
            "/opt/openmower-mapeditor/tools/rosbag2geojson",
            "/bag/map.bag",
            GEOJSON_PATH
        ], check=True)
        return "Extracted", 200
    except subprocess.CalledProcessError as e:
        return f"Extraction failed: {e}", 500

@app.route("/")
def index():
    return send_from_directory("static", "index.html")

@app.route("/<path:path>")
def static_proxy(path):
    return send_from_directory("static", path)

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8088)