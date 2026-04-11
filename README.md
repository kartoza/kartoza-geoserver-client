# Kartoza CloudBench

A comprehensive cloud-native geospatial data management platform with TUI (Terminal User Interface) and Web UI for managing GeoServer, PostgreSQL/PostGIS, S3 storage, and more.

![Python Version](https://img.shields.io/badge/Python-3.12+-3776ab?style=flat&logo=python)
![Django](https://img.shields.io/badge/Django-5.1+-092E20?style=flat&logo=django)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## Features

### TUI (Terminal User Interface)
- **Textual-based interface** - Modern Python TUI with rich widgets
- **Connection manager** - Store and manage multiple GeoServer connections with credentials
- **GeoServer hierarchy browser** - Navigate workspaces, data stores, coverage stores, layers, styles, and layer groups
- **CRUD operations** - Create, edit, and delete workspaces, data stores, and coverage stores
- **Vim-style navigation** - Use familiar j/k keys for navigation

### Web UI
- **Modern React interface** - Beautiful Chakra UI-based web application
- **Tree browser** - Hierarchical view of all GeoServer resources
- **Layer metadata editing** - Comprehensive metadata management including title, abstract, keywords, and attribution
- **GeoWebCache management** - Seed, reseed, and truncate cached tiles with real-time progress
- **Map preview** - Interactive map preview with WMS/WMTS support using MapLibre GL
- **Server synchronization** - Replicate resources between GeoServer instances
- **Chunked file upload** - Upload large geospatial files with progress tracking
- **S3 Storage integration** - Browse and manage cloud-native geospatial data
- **PostgreSQL/PostGIS** - Direct database connections via pg_service.conf
- **QGIS Projects** - Manage and preview QGIS Server projects
- **Terria/Cesium 3D** - 3D globe visualization support

## Installation

### Using Nix (recommended)

```bash
# Enter development shell
nix develop

# Run the web server
python manage.py runserver 8080

# Run the TUI
python -m tui
```

### From source

```bash
git clone https://github.com/kartoza/kartoza-cloudbench.git
cd kartoza-cloudbench

# Create virtual environment
python -m venv venv
source venv/bin/activate

# Install dependencies
pip install -e .

# Run database migrations
python manage.py migrate

# Build frontend
cd web && npm install && npm run build && cd ..

# Run the web server
python manage.py runserver 8080
```

### Development

```bash
nix develop  # Enter development shell with all dependencies

# Run Django development server
python manage.py runserver 8080

# Run TUI
python -m tui

# Run tests
pytest

# Lint code
ruff check .
```

## Quick Start

1. **Start the server**:
   ```bash
   python manage.py runserver 8080
   ```

2. **Open the web UI**: Visit http://localhost:8080

3. **Add a GeoServer connection**: Click the + button next to "GeoServer Connections"

4. **Browse and manage**: Navigate workspaces, upload data, preview layers

## Architecture

```
kartoza-cloudbench/
├── apps/                    # Django applications
│   ├── geoserver/          # GeoServer REST API client
│   ├── postgres/           # PostgreSQL/PostGIS integration
│   ├── s3/                 # S3/MinIO storage
│   ├── upload/             # Chunked file uploads
│   ├── preview/            # Layer preview sessions
│   ├── gwc/                # GeoWebCache management
│   ├── sync/               # Server synchronization
│   └── ...
├── cloudbench/             # Django project settings
├── tui/                    # Textual TUI application
├── web/                    # React frontend
│   └── src/
│       ├── api/            # API client modules
│       ├── components/     # React components
│       └── stores/         # Zustand state stores
└── static/                 # Compiled frontend assets
```

## Configuration

### Environment Variables

```bash
# Django settings
SECRET_KEY=your-secret-key
DEBUG=true
ALLOWED_HOSTS=localhost,127.0.0.1

# Database (optional - defaults to SQLite)
DATABASE_URL=postgres://user:pass@localhost/cloudbench
```

### PostgreSQL Services

PostgreSQL connections are configured via `~/.pg_service.conf`:

```ini
[mydb]
host=localhost
port=5432
dbname=geodata
user=postgres
password=secret
```

## Supported File Types

| Type | Extensions | Upload Target |
|------|------------|---------------|
| Shapefile | `.shp`, `.zip` | Data Store |
| GeoPackage | `.gpkg` | Data Store |
| GeoTIFF | `.tif`, `.tiff` | Coverage Store |
| GeoJSON | `.geojson`, `.json` | Data Store |
| SLD Style | `.sld` | Styles |
| CSS Style | `.css` | Styles |

## API Endpoints

The REST API is available under `/api/`:

- `/api/connections` - GeoServer connections
- `/api/workspaces/<conn_id>` - Workspaces
- `/api/datastores/<conn_id>/<workspace>` - Data stores
- `/api/layers/<conn_id>/<workspace>` - Layers
- `/api/styles/<conn_id>/<workspace>` - Styles
- `/api/upload/init` - Initialize chunked upload
- `/api/preview/` - Start layer preview session
- `/api/pg/services` - PostgreSQL services

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Made with 💗 by [Kartoza](https://kartoza.com) | [Donate](https://github.com/sponsors/kartoza) | [GitHub](https://github.com/kartoza/kartoza-cloudbench)
