# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-02-08

### Added

#### Web Interface
- **Modern React/Chakra UI Web Interface** - A beautiful, responsive web interface with Kartoza branding
  - Dashboard views for connections, workspaces, data stores, coverage stores, and layers
  - Hero-style welcome panel with feature cards
  - Gradient headers and modern card-based layouts
  - Dark/light mode support

#### Layer Preview
- **MapLibre GL Layer Preview** - Interactive map preview for layers and layer groups
  - WMS layer rendering with automatic bounds detection
  - WMTS/GeoWebCache tile preview support
  - Layer metadata panel showing title, SRS, bounding box, and status
  - Auto-update preview when selecting different layers
  - Zoom/pan controls and layer info display

#### GeoWebCache Integration
- **Tile Cache Management** - Full GeoWebCache integration for tile caching
  - Seed, Reseed, and Truncate operations
  - Grid set and tile format selection
  - Zoom level range configuration
  - Thread count adjustment for seeding
  - Real-time task progress monitoring
  - Task cancellation support
  - Cache preview using WMTS tiles

#### Service Metadata
- **GeoServer Settings Editor** - Edit server contact and service metadata
  - Contact information (person, organization, position)
  - Address details (street, city, state, postal code, country)
  - Communication details (phone, fax, email)
  - Online resource and welcome message
  - Tabbed interface for organized editing

#### Layer Metadata
- **Comprehensive Layer Metadata Editing** - Full layer metadata management
  - Basic info: title, abstract, keywords
  - Technical details: native CRS, SRS, bounding boxes
  - Settings: enabled, advertised, queryable toggles
  - Attribution: title, URL, logo
  - Metadata links management
  - Support for both feature types and coverages

#### Layer Groups
- **Layer Group Management** - Create and configure layer groups
  - Add/remove layers from groups
  - Reorder layers with drag-and-drop style controls
  - Group mode selection (SINGLE, OPAQUE_CONTAINER, NAMED, CONTAINER, EO)
  - Title and abstract editing

#### Upload Functionality
- **Data Upload** - Upload geospatial data to GeoServer
  - Shapefile (ZIP) upload with automatic store/layer creation
  - GeoPackage upload with layer publishing
  - GeoTIFF upload to coverage stores
  - Progress indication during upload
  - Auto-focus on newly uploaded resources

#### Multi-Connection Support
- **Multiple GeoServer Connections** - Manage multiple servers simultaneously
  - Add, edit, and delete connections
  - Connection status indicators
  - Quick connection switching

#### TUI Enhancements
- **Settings Wizard** - TUI wizard for editing GeoServer contact information
  - Multi-field form with tab navigation
  - Input validation
  - Keyboard-driven interface

- **Cache Wizard** - TUI wizard for tile cache operations
  - Operation type selection (Seed/Reseed/Truncate)
  - Grid set and format selection
  - Zoom range configuration
  - Thread count adjustment

#### Infrastructure
- **Dual Package Releases** - Both TUI and Web UI are now packaged separately
  - `kartoza-geoserver-client` - Terminal UI package
  - `kartoza-geoserver-web` - Web interface package
  - Separate .deb and .rpm packages for each
  - Platform-specific archives (Linux, macOS, Windows)

- **Nix Flake Support** - Full Nix integration
  - Flake with TUI and Web packages
  - Overlay for easy integration
  - Development shell with all dependencies
  - Direct running without installation: `nix run github:kartoza/kartoza-geoserver-client`

### Changed
- Improved store type detection and display
- Enhanced error handling throughout the application
- Better keyboard navigation in TUI components
- More responsive UI layouts in web interface
- Streamlined dialog workflows

### Fixed
- Layer preview now fully refreshes when switching between layers
- Store name extraction from GeoServer API hrefs
- Proper handling of coverage store types (GeoTIFF, WorldImage, ImageMosaic, etc.)
- Dialog state management for nested dialogs

## [0.1.0] - Initial Development

### Added
- Initial TUI implementation with Midnight Commander-style dual-panel interface
- GeoServer API client with authentication
- Workspace CRUD operations
- Data store and coverage store management
- Layer listing and basic information display
- Vim-style keyboard navigation (j/k/h/l)
- Connection manager with persistent storage
- Basic file browser for local filesystem
