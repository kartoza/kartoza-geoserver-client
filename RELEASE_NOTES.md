# Release Notes - v0.1.0

## Kartoza GeoServer Client

**The first public release of Kartoza GeoServer Client!**

A powerful, Midnight Commander-style Terminal User Interface (TUI) for managing GeoServer instances. Built with Go and the Bubble Tea framework, this application brings the power of GeoServer administration to your terminal.

---

## Highlights

### Dual-Panel Interface
Navigate your local filesystem on the left while exploring GeoServer resources on the right. Seamlessly upload geospatial data to any workspace with a few keystrokes.

### Multi-Connection Support
Manage multiple GeoServer instances simultaneously. Each connection appears as a root node in the tree, allowing you to work across different servers in a single session.

### Beautiful TUI Experience
- Smooth spring-physics animations using Harmonica
- Vim-style keyboard navigation (j/k/h/l)
- Rich visual feedback with icons and colors
- Responsive layout that adapts to terminal size

---

## Features

### File Browser
- **Supported formats**: Shapefile (.shp, .zip), GeoPackage (.gpkg), GeoTIFF (.tif, .tiff), GeoJSON, SLD styles, CSS styles
- **Multi-select**: Select multiple files for batch upload
- **Visual icons**: Quick identification of file types

### GeoServer Explorer
- **Full hierarchy**: Browse workspaces, data stores, coverage stores, layers, styles, and layer groups
- **Lazy loading**: Resources load on-demand for performance
- **Real-time refresh**: Update the tree without losing expansion state

### CRUD Operations
- **Workspaces**: Create, edit, delete with OGC service configuration
- **Data Stores**: PostGIS, Shapefile Directory, GeoPackage, WFS
- **Coverage Stores**: GeoTIFF, World Image, Image Mosaic, Image Pyramid, ArcGrid
- **Layers**: Enable/disable, set advertised and queryable status

### File Upload
- **Progress tracking**: Visual progress dialog for multi-file uploads
- **Auto-publish**: Uploaded files are automatically published as layers
- **Verification**: WFS-based verification confirms successful uploads

### Layer Preview
- **Built-in preview server**: Opens in your default browser
- **MapLibre GL**: Hardware-accelerated WebGL rendering
- **Feature query**: Click to view attributes
- **Auto-zoom**: Automatically fits layer extent

### Connection Manager
- **Store credentials**: Save connections with username/password
- **Test connections**: Verify connectivity before use
- **Server info**: View GeoServer version and build details

---

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Tab` | Switch panels |
| `j/k` or `↑/↓` | Navigate |
| `Enter` or `l` | Open/Expand |
| `Backspace` or `h` | Back/Collapse |
| `Space` | Select file |
| `u` | Upload |
| `c` | Connections |
| `i` | Info |
| `n` | New |
| `e` | Edit |
| `d` | Delete |
| `o` | Preview |
| `r` | Refresh |
| `?` | Help |
| `q` | Quit |

---

## Installation

### Download Binary

Download the appropriate package for your system from the assets below.

### Using Nix

```bash
nix run github:kartoza/kartoza-geoserver-client
```

### Linux (Debian/Ubuntu)

```bash
sudo dpkg -i kartoza-geoserver-client_0.1.0_amd64.deb
```

### Linux (Fedora/RHEL)

```bash
sudo rpm -i kartoza-geoserver-client_0.1.0_x86_64.rpm
```

### macOS

```bash
tar xzf kartoza-geoserver-client_0.1.0_darwin_amd64.tar.gz
sudo mv kartoza-geoserver-client /usr/local/bin/
```

### Windows

1. Download the `.zip` file
2. Extract to a folder
3. Add to PATH or run directly

---

## Requirements

- GeoServer 2.20+ with REST API enabled
- Terminal with Unicode support
- 256-color terminal (recommended)

---

## Configuration

Configuration is stored in `~/.config/kartoza-geoserver-client/config.json`

---

## Known Limitations

- GeoJSON upload not yet supported (coming in v0.2.0)
- Password stored in plaintext (keyring integration planned)
- GeoTIFF verification requires manual check

---

## Getting Started

1. Launch the application: `kartoza-geoserver-client`
2. Press `c` to open the connection manager
3. Press `a` to add a new GeoServer connection
4. Enter your GeoServer URL, username, and password
5. Press `Enter` to connect
6. Navigate and explore!

---

## Contributing

We welcome contributions! Please see our [GitHub repository](https://github.com/kartoza/kartoza-geoserver-client) for:
- Bug reports and feature requests
- Pull request guidelines
- Development setup instructions

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Harmonica](https://github.com/charmbracelet/harmonica) - Animations
- [MapLibre GL](https://maplibre.org/) - Map rendering
- [GeoServer](https://geoserver.org/) - The powerful open source server

---

**Made with by [Kartoza](https://kartoza.com)**
