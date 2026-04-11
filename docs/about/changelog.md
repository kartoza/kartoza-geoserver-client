# Changelog

All notable changes to CloudBench are documented here.

## [Unreleased]

### Added
- Layer preview system with session management
- CARTO basemap for map previews
- Django backend replacing Go implementation
- Chunked file upload with CSRF support
- PostgreSQL service integration via pg_service.conf
- S3 storage browser
- QGIS project management

### Changed
- Switched from Go to Python/Django backend
- Updated basemap from OSM to CARTO
- Improved layer styles API to return string names

### Fixed
- CSRF token handling in file uploads
- Layer styles returning objects instead of strings
- Upload parameter naming (fileSize, connectionId)

## [0.1.0] - 2024-02-15

### Added
- Initial release
- GeoServer connection management
- Web UI with React/Chakra UI
- TUI with Textual
- Layer preview with MapLibre GL
- File upload support

---

Made with 💗 by [Kartoza](https://kartoza.com)
