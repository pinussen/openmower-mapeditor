# 1) Base on ROS Noetic (Ubuntu Focal under the hood, includes rosbag CLI)
FROM docker.io/ros:noetic-ros-base

# 2) Install pip3 & Flask
RUN apt-get update \
 && apt-get install -y python3-pip \
 && pip3 install flask \
 && rm -rf /var/lib/apt/lists/*

# 3) Copy your app + converter
WORKDIR /app
COPY backend.py        app.py
COPY stati:contentReference[oaicite:5]{index=5}sbag2geo:contentReference[oaicite:6]{index=6}pt/openm:contentReference[oaicite:7]{index=7}xtract_from_bag.py

# 4) Add our entrypoint that does the pre-conversion
COPY entrypoint.sh     entrypoint.sh
RUN chmod +x entrypoint.sh

# 5) Prepare the data volume & port
RUN mkdir /data
VOLUME ["/data"]
EXPOSE 8088

ENTRYPOINT ["./entrypoint.sh"]
CMD ["python3", "app.py"]
