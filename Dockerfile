# Use Ubuntu 20.04 ARM64 base
FROM ubuntu:20.04

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install basic dependencies first
RUN apt-get update && apt-get install -y \
    curl \
    gnupg2 \
    lsb-release \
    python3-pip \
    python3-flask \
    git \
    build-essential \
    software-properties-common \
    ca-certificates \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Install Go 1.24
RUN wget -q https://go.dev/dl/go1.24.3.linux-arm64.tar.gz && \
    tar -C /usr/local -xzf go1.24.3.linux-arm64.tar.gz && \
    rm go1.24.3.linux-arm64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

# Add ROS repository
RUN wget -qO - https://raw.githubusercontent.com/ros/rosdistro/master/ros.key | apt-key add - && \
    echo "deb http://packages.ros.org/ros/ubuntu $(lsb_release -cs) main" > /etc/apt/sources.list.d/ros1.list

# Install ROS Noetic (minimal installation first)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ros-noetic-ros-core \
    && rm -rf /var/lib/apt/lists/*

# Install additional ROS packages
RUN apt-get update && \
    apt-get install -y \
    python3-rosdep \
    python3-rosinstall \
    python3-rosinstall-generator \
    python3-wstool \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy application code
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor

# Build and install rosbag2geojson
RUN cd tools/rosbag2geojson && \
    go mod tidy && \
    GOARCH=arm64 go build -v -o rosbag2geojson && \
    cp rosbag2geojson /usr/local/bin/ && \
    chmod +x /usr/local/bin/rosbag2geojson

# Create data directory
RUN mkdir -p /data

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV ROS_DISTRO=noetic
ENV FLASK_APP=backend.py
ENV FLASK_ENV=development

# Source ROS environment
RUN echo "source /opt/ros/noetic/setup.bash" >> ~/.bashrc

# Make entrypoint executable
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Start application
ENTRYPOINT ["/entrypoint.sh"]