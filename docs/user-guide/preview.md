# Layer Preview

CloudBench includes an interactive map viewer for previewing layers.

## Starting a Preview

1. Click on a layer in the tree browser
2. Click the **Preview** button (eye icon)
3. The map preview opens in the main content area

## Map Controls

### Navigation

- **Pan**: Click and drag
- **Zoom**: Scroll wheel or +/- buttons
- **Rotate** (3D mode): Ctrl + drag

### View Modes

| Mode | Description |
|------|-------------|
| 2D | Standard flat map view |
| 3D | Tilted perspective view |
| Globe | Full 3D Cesium globe |

## Style Selection

The preview includes a style picker:

1. Click the style dropdown in the header
2. Select from available styles
3. Map updates immediately
4. Legend icons show style preview

## Layer Information

Click the **Info** button to see:

- Layer title and abstract
- Coordinate reference system
- Bounding box
- Enabled/advertised status

## Basemap

The preview uses CARTO basemaps:

- Light theme for clear visibility
- Attribution included

## WMS Parameters

The preview uses WMS for layer display:

- Format: PNG with transparency
- SRS: EPSG:3857 (Web Mercator)
- Tiled: 256x256 tiles

## 3D Globe Preview

For 3D visualization:

1. Click the **Globe** button
2. Opens Cesium-based viewer
3. Full 3D terrain and imagery
4. Layer draped on terrain

## Troubleshooting

### Blank Preview

1. Check GeoServer is accessible
2. Verify layer is enabled
3. Check browser console for errors

### Wrong Location

1. Verify layer bounding box
2. Check coordinate system settings
3. Ensure data is georeferenced
