# Dockerfile for OpenMower Map Editor
# Byter till en officiell ROS-Noetic-basbild som redan har rosbag, msg-paket mm
FROM osrf/ros:noetic-ros-base

# Se till att vi använder Python3
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# 1) Installera apt-beroenden för att kunna köra rosbag och dess Python-API
RUN apt-get update && apt-get install -y --no-install-recommends \
      python3-rosbag \
      python3-rospkg \
      python3-geometry-msgs \
      python3-yaml \
      python3-pip \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 2) Installera Flask via pip
RUN pip3 install flask

# 3) Kopiera in vår backend + frontend
#    – backend.py innehåller Flask-servern med /extract-endpoint
#    – static/ är vår React/vanilla-JS/HTML-klient
COPY backend.py   ./app.py
COPY static/      ./static/

# 4) Kopiera in vårt Go-binära verktyg (rosbag2geojson)
#    (byggt via ditt install-skript eller via `go build` på hosten)
COPY tools/rosbag2geojson/rosbag2geojson /usr/local/bin/

# 5) Kopiera in Python-fallback (om du vill)
COPY tools/extract_from_bag.py /app/tools/extract_from_bag.py

# 6) Mappa ut och exponera data-mapp där .bag och .geojson ligger
RUN mkdir /data
VOLUME ["/data"]

# 7) Exponera HTTP-port
EXPOSE 8088

# 8) Starta Flask-appen
CMD ["python3","app.py"]
