import rosbag
import json
import sys
import geometry_msgs.msg

if len(sys.argv) != 3:
    print("Usage: extract_geojson.py <input.bag> <output.geojson>")
    sys.exit(1)

bag_path = sys.argv[1]
out_path = sys.argv[2]

gj = {
    "type": "FeatureCollection",
    "features": []
}

print(f"Reading {bag_path}...")

with rosbag.Bag(bag_path) as bag:
    for topic, msg, t in bag.read_messages(topics=['/xbot_monitoring/map']):
        if not msg.polygon.points:
            continue

        coords = []
        for pt in msg.polygon.points:
            coords.append([pt.x, pt.y])
        coords.append(coords[0])  # close the polygon

        feature = {
            "type": "Feature",
            "properties": {},
            "geometry": {
                "type": "Polygon",
                "coordinates": [coords]
            }
        }
        gj["features"].append(feature)

with open(out_path, "w") as f:
    json.dump(gj, f, indent=2)

print(f"✅ Saved {len(gj['features'])} features to {out_path}")
