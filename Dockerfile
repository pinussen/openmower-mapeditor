FROM docker.io/ros:noetic-ros-base

# 1. Installera rosbag-verktygen
RUN apt-get update && apt-get install -y --no-install-recommends \
      ros-noetic-rosbag \
      ros-noetic-rosbag-storage \
    && rm -rf /var/lib/apt/lists/*

# 2. Se till att ROS-miljön är sourcad i alla lager och vid runtime
SHELL ["/bin/bash", "-lc"]
RUN echo "source /opt/ros/noetic/setup.bash" >> /etc/bash.bashrc

# ... övriga steg: kopiera din kod, installera Python-beroenden, exponera portar, osv.
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor
RUN pip install -r requirements.txt

# Kör din app under en bash som sourcar ROS-setup
CMD ["bash", "-lc", "rosbag info /data/yourbag.bag && python app.py"]
