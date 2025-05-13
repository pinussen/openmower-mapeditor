# syntax=docker/dockerfile:1
FROM osrf/ros:noetic-desktop-full

# Kopiera in din mapeditor-kod
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor

# Lägg till ROS-miljön direkt i varje kommando
RUN echo "source /opt/ros/noetic/setup.bash" >> /etc/bash.bashrc

# Starta applikationen
CMD bash -c "source /opt/ros/noetic/setup.bash && rosbag info /data/yourbag.bag && python app.py"