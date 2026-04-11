# GeoServer Management

CloudBench provides comprehensive management of GeoServer instances through the REST API.

## Workspaces

### Creating a Workspace

1. Right-click on "Workspaces" in a connection
2. Select "New Workspace"
3. Enter workspace name
4. Configure options:
   - Default workspace
   - Isolated workspace

### Workspace Settings

- **Services**: Enable/disable OGC services (WMS, WFS, WCS, WMTS)
- **Security**: Workspace-level access control

## Data Stores

### Supported Types

| Type | Description |
|------|-------------|
| PostGIS | PostgreSQL/PostGIS database |
| Shapefile | Single shapefile or directory |
| GeoPackage | OGC GeoPackage file |
| WFS | External WFS service |

### Creating a Data Store

1. Select a workspace
2. Click "Add Data Store"
3. Choose store type
4. Enter connection parameters
5. Click "Save"

### PostGIS Connection

```
Host: localhost
Port: 5432
Database: geodata
Schema: public
User: postgres
Password: ********
```

## Coverage Stores

### Supported Types

| Type | Description |
|------|-------------|
| GeoTIFF | Single GeoTIFF file |
| Image Mosaic | Directory of images |
| Image Pyramid | Multi-resolution pyramid |
| WorldImage | Image with world file |

## Layers

### Layer Properties

- **Enabled**: Layer serves requests
- **Advertised**: Layer in GetCapabilities
- **Queryable**: Supports GetFeatureInfo

### Bounding Boxes

- Native bounding box (original CRS)
- Lat/Lon bounding box (EPSG:4326)
- Auto-compute from data

### Styles

- Default style
- Additional styles
- Style preview in layer viewer

## Layer Groups

Combine multiple layers into a single requestable group:

1. Select workspace
2. Click "New Layer Group"
3. Add layers to the group
4. Configure bounds and styles

## Styles

### Supported Formats

- **SLD**: Styled Layer Descriptor (XML)
- **CSS**: GeoServer CSS extension

### Uploading Styles

1. Go to Styles section
2. Click "Upload Style"
3. Select SLD or CSS file
4. Style is available for layer assignment

## Best Practices

1. **Organize with workspaces**: Group related layers
2. **Use meaningful names**: Clear, descriptive resource names
3. **Set bounding boxes**: Improves performance
4. **Enable caching**: Use GeoWebCache for production
