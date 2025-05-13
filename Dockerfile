# syntax=docker/dockerfile:1
FROM ghcr.io/cedbossneo/openmower-gui:master

# Se till att ROS-miljön alltid är källad
SHELL ["/bin/bash", "-lc"]
RUN echo "source /opt/ros/noetic/setup.bash" >> /etc/bash.bashrc

# Kopiera in din mapeditor-kod
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor

# Installera Python-beroenden (om du har requirements.txt)
#RUN pip install -r requirements.txt

# Till sist: starta ditt kommando i en ROS-login-shell
CMD ["bash", "-lc", "rosbag info /data/yourbag.bag && python app.py"]
