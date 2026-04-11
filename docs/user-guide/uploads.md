# File Uploads

CloudBench supports uploading geospatial data files directly to GeoServer.

## Supported File Types

| Type | Extensions | Target |
|------|------------|--------|
| Shapefile | `.zip` | Data Store |
| GeoPackage | `.gpkg` | Data Store |
| GeoTIFF | `.tif`, `.tiff` | Coverage Store |
| GeoJSON | `.geojson`, `.json` | Data Store |
| SLD Style | `.sld` | Styles |
| CSS Style | `.css` | Styles |

## Uploading Files

### Via Web UI

1. Select a workspace in the tree
2. Click the **Upload** button (arrow icon)
3. Select your file
4. Monitor upload progress
5. Layer is automatically published

### Chunked Uploads

Large files are automatically uploaded in chunks:

- Default chunk size: 5 MB
- Progress tracking per chunk
- Resume capability on failure
- Files up to 10 GB supported

## Upload Process

1. **Initialize**: Create upload session
2. **Upload chunks**: Send file in pieces
3. **Assemble**: Combine chunks on server
4. **Publish**: Create store and layer in GeoServer

## Shapefile Requirements

When uploading shapefiles:

- Package as a ZIP file
- Include all required files:
  - `.shp` - Shape geometry
  - `.shx` - Shape index
  - `.dbf` - Attributes
  - `.prj` - Projection (recommended)

## GeoPackage

GeoPackage files can contain:

- Multiple vector layers
- Raster tiles
- Feature attributes

All layers in the GeoPackage are published.

## GeoTIFF

GeoTIFF requirements:

- Embedded or external georeferencing
- Supported pixel types
- Optional internal tiling for performance

## Troubleshooting

### Upload Fails

1. Check file size limits
2. Verify file format is supported
3. Check GeoServer logs for errors

### Layer Not Visible

1. Verify layer is enabled
2. Check coordinate system
3. Validate bounding box
