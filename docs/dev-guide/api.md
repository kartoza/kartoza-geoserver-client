# API Reference

CloudBench exposes a REST API under `/api/`.

## Authentication

Currently, the API supports session authentication. API tokens are planned.

## Connections

### List Connections

```http
GET /api/connections
```

Response:
```json
[
  {
    "id": "conn_123",
    "name": "Local GeoServer",
    "url": "http://localhost:8600/geoserver",
    "is_active": true
  }
]
```

### Create Connection

```http
POST /api/connections
Content-Type: application/json

{
  "name": "My GeoServer",
  "url": "http://geoserver.example.com/geoserver",
  "username": "admin",
  "password": "geoserver"
}
```

## Workspaces

### List Workspaces

```http
GET /api/workspaces/{conn_id}
```

### Create Workspace

```http
POST /api/workspaces/{conn_id}
Content-Type: application/json

{
  "name": "my_workspace"
}
```

## Layers

### List Layers

```http
GET /api/layers/{conn_id}/{workspace}
```

### Get Layer Details

```http
GET /api/layers/{conn_id}/{workspace}/{layer}
```

### Get Layer Styles

```http
GET /api/layerstyles/{conn_id}/{workspace}/{layer}
```

Response:
```json
{
  "defaultStyle": "polygon",
  "additionalStyles": ["line", "point"]
}
```

## Preview

### Start Preview Session

```http
POST /api/preview/
Content-Type: application/json

{
  "connId": "conn_123",
  "workspace": "topp",
  "layerName": "states"
}
```

Response:
```json
{
  "url": "/api/preview/abc-123"
}
```

### Get Layer Info

```http
GET /api/preview/{session_id}/api/layer
```

### Get Layer Metadata

```http
GET /api/preview/{session_id}/api/metadata
```

## Upload

### Initialize Upload

```http
POST /api/upload/init
Content-Type: application/json

{
  "connectionId": "conn_123",
  "workspace": "topp",
  "filename": "data.gpkg",
  "fileSize": 1048576,
  "chunkSize": 5242880
}
```

### Upload Chunk

```http
POST /api/upload/chunk
Content-Type: multipart/form-data

sessionId: abc-123
chunkIndex: 0
chunk: <binary data>
```

### Complete Upload

```http
POST /api/upload/complete
Content-Type: application/json

{
  "sessionId": "abc-123"
}
```

## Error Responses

All errors return JSON:

```json
{
  "error": "Description of the error"
}
```

HTTP status codes:
- `400`: Bad request
- `401`: Unauthorized
- `404`: Not found
- `500`: Server error
