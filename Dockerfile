# syntax=docker/dockerfile:1
FROM osrf/ros:noetic-desktop-full

# Install additional dependencies
RUN apt-get update && apt-get install -y \
    python3-pip \
    python3-flask \
    && rm -rf /var/lib/apt/lists/*

# Copy application code
COPY . /opt/openmower-mapeditor
WORKDIR /opt/openmower-mapeditor

# Create data directory
RUN mkdir -p /data

# Set environment variables
ENV PYTHONUNBUFFERED=1

# Source ROS environment
RUN echo "source /opt/ros/noetic/setup.bash" >> ~/.bashrc

# Make entrypoint executable
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Start application
ENTRYPOINT ["/entrypoint.sh"]