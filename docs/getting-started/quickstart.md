# Quick Start

This guide will help you get CloudBench running and connected to your first GeoServer instance.

## 1. Start the Server

```bash
# Using Nix
nix develop
python manage.py runserver 8080

# Or using pip
python manage.py runserver 8080
```

## 2. Open the Web UI

Navigate to [http://localhost:8080](http://localhost:8080) in your browser.

## 3. Add a GeoServer Connection

1. In the left sidebar, find "GeoServer Connections"
2. Click the **+** button to add a new connection
3. Enter your GeoServer details:
   - **Name**: A friendly name for this connection
   - **URL**: Your GeoServer URL (e.g., `http://localhost:8600/geoserver`)
   - **Username**: GeoServer admin username
   - **Password**: GeoServer admin password
4. Click **Test Connection** to verify
5. Click **Save** to add the connection

## 4. Browse Your Data

Once connected, you can:

- **Expand workspaces** to see data stores, coverage stores, and layers
- **Click on a layer** to see its details in the right panel
- **Preview layers** on an interactive map
- **Upload data** to create new stores and layers

## 5. Upload Geospatial Data

1. Select a workspace in the tree
2. Click the **Upload** button (arrow icon)
3. Select your file (Shapefile ZIP, GeoPackage, GeoTIFF, etc.)
4. The file will be uploaded and published automatically

## 6. Preview Layers

1. Click on any layer in the tree
2. Click the **Preview** button (eye icon)
3. An interactive map will appear showing your layer

## Using the TUI

CloudBench also includes a terminal user interface:

```bash
python -m tui
```

Navigate using:
- **j/k** or **↑/↓**: Move up/down
- **Enter**: Select/expand
- **Tab**: Switch panels
- **q**: Quit

## Next Steps

- [Web UI Guide](../user-guide/web-ui.md) - Learn all Web UI features
- [GeoServer Management](../user-guide/geoserver.md) - Manage workspaces, stores, and layers
- [Configuration](configuration.md) - Customize CloudBench settings
