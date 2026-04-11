# Kartoza CloudBench - Package Architecture

This document provides an annotated list of all packages in the software architecture.

## Python Packages (v0.3.0+)

### Django Project (`cloudbench/`)

| Module | Purpose |
|--------|---------|
| `cloudbench/__init__.py` | Package initialization, version info |
| `cloudbench/settings/base.py` | Base Django settings (common to all environments) |
| `cloudbench/settings/development.py` | Development-specific settings |
| `cloudbench/settings/production.py` | Production-specific settings |
| `cloudbench/urls.py` | Root URL configuration |
| `cloudbench/asgi.py` | ASGI application entry point |
| `cloudbench/wsgi.py` | WSGI application entry point |

### Django Apps (`apps/`)

| App | Purpose | Key Files |
|-----|---------|-----------|
| `apps.core` | Configuration, middleware, utilities | `config.py`, `managers.py`, `middleware.py`, `exceptions.py` |
| `apps.connections` | GeoServer connection CRUD | `views.py`, `serializers.py` |
| `apps.geoserver` | GeoServer REST API operations | `client.py`, `views.py` |
| `apps.gwc` | GeoWebCache tile management | `client.py`, `views.py` |
| `apps.postgres` | PostgreSQL/PostGIS integration | `service.py`, `schema.py`, `views.py` |
| `apps.upload` | Chunked file uploads | `views.py` (session manager) |
| `apps.s3` | S3-compatible storage | `client.py`, `duckdb.py`, `views.py` |
| `apps.ai` | AI query generation (Ollama) | `engine.py`, `views.py` |
| `apps.query` | Visual query builder | `builder.py`, `views.py` |
| `apps.bridge` | PostgreSQL to GeoServer bridge | `views.py` |
| `apps.sqlview` | SQL View layer publishing | `views.py` |
| `apps.sync` | Server synchronization | `services.py`, `views.py` |
| `apps.dashboard` | Monitoring dashboard | `views.py` |
| `apps.search` | Universal search | `services.py`, `views.py` |
| `apps.terria` | 3D viewer integration | `catalog.py`, `views.py` |
| `apps.qfieldcloud` | QFieldCloud integration | `client.py`, `views.py` |
| `apps.mergin` | Mergin Maps integration | `client.py`, `views.py` |
| `apps.geonode` | GeoNode integration | `client.py`, `views.py` |
| `apps.iceberg` | Apache Iceberg integration | `client.py`, `views.py` |
| `apps.qgis` | QGIS project management | `views.py` |

### TUI Application (`tui/`)

| Module | Purpose |
|--------|---------|
| `tui/__main__.py` | CLI entry point (Click) |
| `tui/app.py` | Main Textual App class |
| `tui/screens/home.py` | Dashboard/home screen |
| `tui/screens/connections.py` | Connection management screen |
| `tui/screens/geoserver.py` | GeoServer resource browser |
| `tui/screens/postgres.py` | PostgreSQL service browser |
| `tui/screens/s3.py` | S3 storage browser |
| `tui/screens/settings.py` | Application settings |
| `tui/widgets/tree.py` | Resource tree widget |
| `tui/widgets/progress.py` | Progress indicator widgets |
| `tui/styles/app.tcss` | Textual CSS styling |

## Frontend Packages (`web/`)

| Module | Purpose |
|--------|---------|
| `web/src/api/` | TypeScript API client modules |
| `web/src/components/` | React components (129 .tsx files) |
| `web/src/stores/` | Zustand state management |
| `web/src/hooks/` | Custom React hooks |
| `web/src/types/` | TypeScript type definitions |
| `web/src/utils/` | Utility functions and animations |
| `web/src/theme.ts` | Chakra UI theme configuration |

## Legacy Go Packages (`internal/`)

These packages are from the Go implementation (v0.2.x) and are maintained for backward compatibility:

| Package | Purpose |
|---------|---------|
| `internal/api/` | GeoServer REST API client |
| `internal/config/` | Configuration management |
| `internal/gwc/` | GeoWebCache integration |
| `internal/integration/` | Cross-system operations |
| `internal/llm/` | LLM/Ollama integration |
| `internal/models/` | Data models |
| `internal/ogr2ogr/` | Data import via ogr2ogr |
| `internal/postgres/` | PostgreSQL integration |
| `internal/preview/` | Layer preview server |
| `internal/query/` | Visual query builder |
| `internal/s3client/` | S3-compatible storage client |
| `internal/cloudnative/` | Cloud-native format conversion |
| `internal/storage/` | File storage management |
| `internal/sync/` | Server synchronization |
| `internal/terria/` | Terria catalog export |
| `internal/tui/` | Bubble Tea TUI components |
| `internal/verify/` | Upload verification |
| `internal/webserver/` | HTTP handlers (Go) |

## Nix Packages

The `flake.nix` provides the following packages:

| Package | Description |
|---------|-------------|
| `default` / `web` | Django web server (Python) |
| `tui` | Textual TUI application (Python) |
| `frontend` | Built React frontend assets |
| `go-tui` | Legacy Go TUI binary |
| `go-web` | Legacy Go web server |

## Key Dependencies

### Python Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| django | ^5.0 | Web framework |
| djangorestframework | ^3.15 | REST API |
| django-cors-headers | ^4.3 | CORS handling |
| whitenoise | ^6.6 | Static file serving |
| httpx | ^0.27 | Async HTTP client |
| psycopg | ^3.1 | PostgreSQL driver |
| boto3 | ^1.34 | S3 client |
| duckdb | ^1.0 | Parquet queries |
| lxml | ^5.2 | XML parsing |
| pydantic | ^2.7 | Data validation |
| textual | ^0.79 | TUI framework |
| rich | ^13.7 | Rich text formatting |
| click | ^8.1 | CLI parsing |
| uvicorn | ^0.29 | ASGI server |
| gunicorn | ^22.0 | WSGI server |

### Frontend Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| react | ^18.2.0 | UI framework |
| @chakra-ui/react | ^2.8.2 | UI components |
| zustand | ^4.5.0 | State management |
| maplibre-gl | ^4.0.0 | 2D map viewer |
| cesium | ^1.138.0 | 3D globe viewer |
| @tanstack/react-query | ^5.17.19 | Data fetching |
| framer-motion | ^11.0.0 | Animations |
| @codemirror/lang-sql | ^6.0.0 | SQL editor |

---

Made with love by [Kartoza](https://kartoza.com) | [Donate](https://github.com/sponsors/kartoza) | [GitHub](https://github.com/kartoza/kartoza-cloudbench)
