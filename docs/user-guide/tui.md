# TUI Guide

CloudBench includes a Terminal User Interface (TUI) built with Python Textual.

## Starting the TUI

```bash
python -m tui
```

Or using the Nix shortcut:
```bash
runtui
```

## Navigation

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` / `l` | Select / Expand |
| `Backspace` / `h` | Go back / Collapse |
| `Tab` | Switch panels |
| `q` / `Ctrl+C` | Quit |

### Connection Manager

| Key | Action |
|-----|--------|
| `a` | Add connection |
| `e` | Edit connection |
| `d` | Delete connection |
| `t` | Test connection |
| `Enter` | Connect |

## Features

### GeoServer Browser
- Browse workspaces, stores, and layers
- View resource metadata
- Expand/collapse tree nodes

### Connection Management
- Store multiple GeoServer connections
- Credentials securely stored
- Quick connect/disconnect

## Authentication

The TUI supports token-based authentication:

1. On first run, enter your credentials
2. A token is saved to `~/.config/kartoza-cloudbench/auth.json`
3. Subsequent runs use the saved token

To logout:
```bash
rm ~/.config/kartoza-cloudbench/auth.json
```
