# syntax=docker/dockerfile:1
FROM ghcr.io/cedbossneo/openmower-gui:master

# Install ROS Noetic
RUN apt-get update && apt-get install -y \
    ros-noetic-desktop-full \
    && rm -rf /var/lib/apt/lists/*

# Lägg till ROS-miljön direkt i varje kommando
RUN echo "source /opt/ros/noetic/setup.bash" >> /etc/bash.bashrc

# Kopiera in din mapeditor-kod
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor

# Starta applikationen
CMD bash -c "source /opt/ros/noetic/setup.bash && rosbag info /data/yourbag.bag && python app.py"