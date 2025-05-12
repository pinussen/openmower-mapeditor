# 1) Basimage med ROS Noetic + rosbag CLI
FROM docker.io/osrf/ros:noetic-ros-base

# 2) Installera pip3 & Flask
RUN apt-get update \
 && apt-get install -y python3-pip curl \
 && pip3 install flask \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 3) Kopiera in din backend och frontend
COPY backend.py           app.py
COPY static/              static/

# 4) Kopiera konverterings‐binarien och extraktionsskriptet
COPY tools/rosbag2geojson/rosbag2geojson    /usr/local/bin/rosbag2geojson
COPY tools/extract_from_bag.py              extract_from_bag.py

# 5) Lägg in entrypoint som gör förkonvertering vid start
COPY entrypoint.sh       entrypoint.sh
RUN chmod +x entrypoint.sh

# 6) Förbered data‐volym och port
RUN mkdir /data
VOLUME ["/data"]
EXPOSE 8088

ENTRYPOINT ["./entrypoint.sh"]
CMD ["python3", "app.py"]
