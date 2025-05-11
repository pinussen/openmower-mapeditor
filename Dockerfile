# 1) Börja från officiella ROS Noetic-ros-base (innehåller rosbag CLI)
FROM docker.io/library/ros:noetic-ros-base

WORKDIR /app

# 2) Installera Flask
RUN apt-get update && \
    apt-get install -y python3-pip && \
    pip3 install flask && \
    rm -rf /var/lib/apt/lists/*

# 3) Kopiera din backend/frontend
COPY backend.py    /app/app.py
COPY static/       /app/static/

# 4) Kopiera din Go-converter (bygger den via install.sh och lägger den i /usr/local/bin)
COPY tools/rosbag2geojson/rosbag2geojson /usr/local/bin/rosbag2geojson

# 5) Se till att HTTP-serven sparar i /data
RUN mkdir /data
VOLUME ["/data"]

EXPOSE 8088

CMD ["python3", "app.py"]
