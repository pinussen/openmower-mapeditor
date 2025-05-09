# Dockerfile for OpenMower Map Editor (Podman-compatible)
FROM docker.io/library/python:3.11-slim

WORKDIR /app

# Install Flask
RUN pip install flask

# Copy backend and frontend
COPY backend.py /app/app.py
COPY static/ /app/static/
COPY tools/rosbag2geojson/rosbag2geojson /usr/local/bin/rosbag2geojson
COPY tools/extract_geojson.py /opt/openmower-mapeditor/tools/extract_geojson.py


# Create data directory
RUN mkdir /data
VOLUME ["/data"]

# Expose editor port
EXPOSE 8088

CMD ["python", "app.py"]