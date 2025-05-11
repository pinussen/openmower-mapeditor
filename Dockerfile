# Dockerfile for OpenMower Map Editor
FROM docker.io/library/python:3.11-slim

WORKDIR /app

# Flask + ev. fler dependencies
RUN pip install flask

# Kopiera koden
COPY backend.py    /app/app.py
COPY static/       /app/static/
# Go-binären som byggts på hosten
COPY tools/rosbag2geojson/rosbag2geojson /usr/local/bin/rosbag2geojson

# Vårt entrypoint
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# Dir där bag + output monteras
RUN mkdir /data
VOLUME ["/data"]

EXPOSE 8088

ENTRYPOINT ["/app/entrypoint.sh"]
