# OpenMower Map Editor

The OpenMower Map Editor is a tool for creating and editing maps for OpenMower. This project combines a web-based interface with backend services to extract and manage GeoJSON data from ROS bag files.

## Features

- **Web Interface**: Interactive map powered by Leaflet for viewing and editing zones.
- **GeoJSON Management**: Load and save map data in GeoJSON format.
- **ROS Bag Extraction**: Convert ROS bag files to GeoJSON using `rosbag2geojson`.

## Installation

### Prerequisites

- **Go** (to build `rosbag2geojson`)
- **Podman** or **Docker** (to run the container)
- **Python 3** and **Flask** (for the backend service)

### Installation Steps

1. Run the installation script:

   ```bash
   sudo ./install.sh

2. After installation, the service will be available at http://<your-ip>:8088.

## Usage

### Web Interface
Open your browser and navigate to http://<your-ip>:8088.
Use the buttons to:
Save: Save edited zones to GeoJSON.
Extract from .bag: Extract map data from a ROS bag file.

### Backend API
GET /load: Load existing GeoJSON data.
POST /extract: Extract GeoJSON from a ROS bag file.

## Development

### Build rosbag2geojson
cd tools/rosbag2geojson
go build -o rosbag2geojson

### Run Backend Locally
python3 [backend.py](http://_vscodecontentref_/0)

## Contributing
We welcome contributions! Please create a pull request or open an issue if you have suggestions or find bugs.

## License
This project is licensed under the MIT License. See LICENSE for more details.

