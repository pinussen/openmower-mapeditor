# syntax=docker/dockerfile:1
FROM ghcr.io/cedbossneo/openmower-gui:master

# Lägg till ROS-miljön direkt i varje kommando
RUN echo "source /opt/ros/noetic/setup.bash" >> /etc/bash.bashrc

# Kopiera in din mapeditor-kod
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor

# Installera Python-beroenden (om du har requirements.txt)
#RUN pip install -r requirements.txt

# Starta applikationen utan att använda SHELL
CMD bash -c "source /opt/ros/noetic/setup.bash && rosbag info /data/yourbag.bag && python app.py"