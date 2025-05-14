# Use Ubuntu 20.04 ARM64 base
FROM ubuntu:20.04

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

# Install basic dependencies
RUN apt-get update && apt-get install -y \
    curl \
    gnupg2 \
    lsb-release \
    python3-pip \
    python3-flask \
    git \
    build-essential \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# Add ROS repository and keys
RUN curl -sSL https://raw.githubusercontent.com/ros/rosdistro/master/ros.key -o /usr/share/keyrings/ros-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/ros-archive-keyring.gpg] http://packages.ros.org/ros/ubuntu $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/ros1.list > /dev/null

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

# Initialize rosdep
RUN rosdep init && rosdep update

# Copy application code
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor

# Create data directory
RUN mkdir -p /data

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV ROS_DISTRO=noetic

# Source ROS environment
RUN echo "source /opt/ros/noetic/setup.bash" >> ~/.bashrc

# Make entrypoint executable
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Start application
ENTRYPOINT ["/entrypoint.sh"]