// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

// Package icons provides Nerd Font icons for the TUI.
// These are monochrome, tasteful icons that work in any Nerd Font-enabled terminal.
// See: https://www.nerdfonts.com/cheat-sheet
package icons

// Nerd Font icons - these require a Nerd Font compatible terminal/font
// Using primarily Codicon (cod), Material Design (md), and Font Awesome (fa) icons
const (
	// Application
	Globe      = "\uf0ac" // fa-globe - main app icon
	Server     = "\uf233" // fa-server - connection/server
	Database   = "\uf1c0" // fa-database - data store
	Image      = "\uf03e" // fa-image - coverage/raster
	Map        = "\uf279" // fa-map - map/wms
	Grid       = "\uf00a" // fa-th - grid/wmts
	Folder     = "\uf07b" // fa-folder - workspace/directory
	FolderOpen = "\uf07c" // fa-folder-open - expanded folder
	File       = "\uf15b" // fa-file - generic file
	FileAlt    = "\uf15c" // fa-file-alt - document/layer
	Files      = "\uf0c5" // fa-copy - multiple files/layers
	Palette    = "\uf53f" // fa-palette - styles
	Brush      = "\uf1fc" // fa-paint-brush - single style
	Book       = "\uf02d" // fa-book - layer group
	Books      = "\uf5db" // fa-books - layer groups collection
	Archive    = "\uf187" // fa-archive - geopackage
	Photo      = "\uf03e" // fa-image - geotiff
	Code       = "\uf121" // fa-code - geojson/css

	// Status indicators
	Check        = "\uf00c" // fa-check - success/enabled
	Cross        = "\uf00d" // fa-times - error/disabled
	CircleFill   = "\uf111" // fa-circle - filled bullet
	CircleEmpty  = "\uf10c" // fa-circle-o - empty bullet
	ChevronDown  = "\uf078" // fa-chevron-down - expanded
	ChevronRight = "\uf054" // fa-chevron-right - collapsed
	CaretRight   = "\uf0da" // fa-caret-right - active/editing
	Spinner      = "\uf110" // fa-spinner - loading
	Warning      = "\uf071" // fa-exclamation-triangle - warning
	Info         = "\uf05a" // fa-info-circle - info
	Question     = "\uf128" // fa-question - unknown

	// Actions
	Search   = "\uf002" // fa-search - search
	Plus     = "\uf067" // fa-plus - add/new
	Minus    = "\uf068" // fa-minus - remove
	Edit     = "\uf044" // fa-edit - edit
	Trash    = "\uf1f8" // fa-trash - delete
	Download = "\uf019" // fa-download - download
	Upload   = "\uf093" // fa-upload - upload
	Sync     = "\uf021" // fa-refresh - sync/refresh
	Eye      = "\uf06e" // fa-eye - preview
	Cog      = "\uf013" // fa-cog - settings
	Link     = "\uf0c1" // fa-link - connection

	// Navigation
	ArrowUp    = "\uf062" // fa-arrow-up
	ArrowDown  = "\uf063" // fa-arrow-down
	ArrowLeft  = "\uf060" // fa-arrow-left
	ArrowRight = "\uf061" // fa-arrow-right
	Home       = "\uf015" // fa-home

	// Misc
	Clock  = "\uf017" // fa-clock-o - time
	Cube   = "\uf1b2" // fa-cube - store
	Cubes  = "\uf1b3" // fa-cubes - stores
	Layer  = "\uf5fd" // fa-layer-group - layer
	Layers = "\uf5fd" // fa-layer-group - layers
	Stop   = "\uf04d" // fa-stop - stop
	Play   = "\uf04b" // fa-play - play/start
)

// Convenience aliases for common use cases
const (
	// Tree node icons
	Root           = Globe
	Connection     = Server
	Workspace      = Folder
	DataStores     = Cubes
	CoverageStores = Image
	WMSStores      = Map
	WMTSStores     = Grid
	DataStore      = Database
	CoverageStore  = Photo
	WMSStore       = Map
	WMTSStore      = Grid
	LayersIcon     = Layers
	LayerIcon      = Layer
	StylesIcon     = Palette
	StyleIcon      = Brush
	LayerGroups    = Books
	LayerGroup     = Book
	Unknown        = Question

	// File icons
	Directory   = Folder
	Shapefile   = Map
	GeoPackage  = Archive
	GeoTIFF     = Photo
	GeoJSON     = Code
	SLDFile     = Palette
	CSSFile     = Palette
	GenericFile = File

	// Status
	Success    = Check
	Error      = Cross
	Enabled    = Check
	Disabled   = Cross
	Online     = CircleFill
	Offline    = CircleEmpty
	Circle     = CircleFill // alias for selection marker
	Selected   = CircleFill
	Unselected = CircleEmpty
	Expanded   = ChevronDown
	Collapsed  = ChevronRight
	Active     = CaretRight
	Loading    = Spinner
)
