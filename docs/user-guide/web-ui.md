# Web UI Guide

The CloudBench Web UI provides a modern, intuitive interface for managing your geospatial infrastructure.

## Interface Overview

The interface is divided into three main areas:

1. **Left Sidebar** - Tree browser for all resources
2. **Main Content** - Details panel, map preview, or data views
3. **Header** - Search, settings, and user info

## Tree Browser

The tree browser shows all your connected resources:

### GeoServer Connections
- Workspaces
  - Data Stores (vector data)
  - Coverage Stores (raster data)
  - Layers
  - Styles
  - Layer Groups

### PostgreSQL Services
- Schemas
- Tables (with geometry info)

### S3 Storage
- Buckets
- Objects (folders and files)

### QGIS Projects
- Registered QGIS Server projects

## Working with GeoServer

### Adding a Connection

1. Click **+** next to "GeoServer Connections"
2. Fill in the connection details
3. Click **Test Connection** to verify
4. Click **Save**

### Managing Workspaces

- **Create**: Right-click workspace list → New Workspace
- **Edit**: Click workspace → Edit button
- **Delete**: Click workspace → Delete button

### Uploading Data

1. Select a workspace
2. Click the **Upload** button
3. Select your file:
   - Shapefile (`.zip`)
   - GeoPackage (`.gpkg`)
   - GeoTIFF (`.tif`, `.tiff`)
   - GeoJSON (`.geojson`)
4. Large files are uploaded in chunks with progress tracking

### Layer Preview

1. Click on a layer in the tree
2. Click the **Preview** button (eye icon)
3. The map preview shows:
   - Interactive map with CARTO basemap
   - Your layer via WMS
   - Layer metadata panel
   - Style selector

### Managing Styles

- View available styles for each layer
- Switch between styles in the preview
- Upload new SLD or CSS styles

## Map Preview Features

### View Modes
- **2D** - Standard flat map view
- **3D** - Tilted perspective view
- **Globe** - Full 3D Cesium globe

### Controls
- **Pan**: Click and drag
- **Zoom**: Scroll wheel
- **Rotate** (3D): Ctrl + drag

### Style Picker
- Click the style dropdown to change layer styling
- Preview shows legend icons for each style

## GeoWebCache

Manage tile caching for your layers:

1. Select a layer
2. Open the GWC panel
3. Choose operation:
   - **Seed**: Generate new tiles
   - **Reseed**: Regenerate existing tiles
   - **Truncate**: Clear cached tiles
4. Monitor progress in real-time

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `/` | Focus search |
| `Esc` | Close panel/dialog |
| `?` | Show help |
