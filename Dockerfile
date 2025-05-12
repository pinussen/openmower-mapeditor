# 1) Hämta officiell ROS Noetic-base (includes rosbag CLI, finns för arm64)
FROM ros:noetic-ros-base

# 2) Installera Python3, pip och Flask
RUN apt-get update \
 && apt-get install -y python3-pip \
 && pip3 install --no-cache-dir flask \
 && rm -rf /var/lib/apt/lists/*

# 3) Skapa app-mapp och kopiera in koden
WORKDIR /app
# Kopiera backend (flask-app), entrypoint och frontend
COPY backend.py  app.py
COPY entrypoint.sh  entrypoint.sh
COPY static/    static/
RUN chmod +x entrypoint.sh

# 4) Kopiera Go-binaryn för rosbag2geojson
COPY tools/rosbag2geojson/rosbag2geojson /usr/local/bin/rosbag2geojson

# 5) Förbered data-volym
RUN mkdir /data
VOLUME ["/data"]

# 6) Exponera port och startkommando
EXPOSE 8088
ENTRYPOINT ["./entrypoint.sh"]
CMD ["python3", "app.py"]
