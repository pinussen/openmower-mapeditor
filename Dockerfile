# Dockerfile for OpenMower Map Editor (Podman-compatible)
FROM docker.io/library/python:3.11-slim

WORKDIR /app

# Install Flask
RUN pip install flask

# Copy backend and frontend
COPY backend.py /app/app.py
COPY static/ /app/static/

# Create data directory
RUN mkdir /data
VOLUME ["/data"]

# Expose editor port
EXPOSE 8088

CMD ["python", "app.py"]