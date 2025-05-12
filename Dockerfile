# Dockerfile for OpenMower Map Editor
FROM docker.io/python:3.11-slim

# 1) Install system tools & ROS Noetic rosbag
RUN apt-get update \
 && apt-get install -y \
      curl \
      gnupg2 \
      lsb-release \
 && curl -sSL https://raw.githubusercontent.com/ros/rosdistro/master/ros.asc \
      | apt-key add - \
 && echo "deb http://packages.ros.org/ros/ubuntu focal main" \
      > /etc/apt/sources.list.d/ros-latest.list \
 && apt-get update \
 && apt-get install -y \
      ros-noetic-rosbag \
 && rm -rf /var/lib/apt/lists/*

# 2) Install Flask
WORKDIR /app
RUN pip install flask

# 3) Copy your backend/front-end & converter
COPY backend.py    /app/app.py
COPY static/       /app/static/
COPY tools/rosbag2geojson/rosbag2geojson  /usr/local/bin/rosbag2geojson
COPY tools/extract_from_bag.py            /opt/openmower-mapeditor/tools/extract_from_bag.py
COPY entrypoint.sh     /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

# 4) Prepare map volume and expose port
RUN mkdir /data
VOLUME ["/data"]
EXPOSE 8088

# 5) Launch
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["python3", "app.py"]