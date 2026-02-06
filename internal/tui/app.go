package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-geoserver-client/internal/api"
	"github.com/kartoza/kartoza-geoserver-client/internal/config"
	"github.com/kartoza/kartoza-geoserver-client/internal/models"
	"github.com/kartoza/kartoza-geoserver-client/internal/preview"
	"github.com/kartoza/kartoza-geoserver-client/internal/tui/components"
	"github.com/kartoza/kartoza-geoserver-client/internal/tui/screens"
	"github.com/kartoza/kartoza-geoserver-client/internal/tui/styles"
)

// Screen represents the current screen
type Screen int

const (
	ScreenMain Screen = iota
	ScreenConnections
	ScreenUpload
	ScreenHelp
)

// CRUDOperation represents the type of CRUD operation
type CRUDOperation int

const (
	CRUDNone CRUDOperation = iota
	CRUDCreate
	CRUDEdit
	CRUDDelete
)

// Panel represents which panel is active
type Panel int

const (
	PanelLeft Panel = iota
	PanelRight
)

// AppKeyMap defines the global key bindings
type AppKeyMap struct {
	Quit        key.Binding
	Tab         key.Binding
	Help        key.Binding
	Connections key.Binding
	Upload      key.Binding
	Refresh     key.Binding
	Escape      key.Binding
}

// DefaultAppKeyMap returns the default key bindings
func DefaultAppKeyMap() AppKeyMap {
	return AppKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		Help: key.NewBinding(
			key.WithKeys("?", "f1"),
			key.WithHelp("?", "help"),
		),
		Connections: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "connections"),
		),
		Upload: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "upload"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r", "ctrl+r"),
			key.WithHelp("r", "refresh"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
	}
}

// Messages
type (
	workspacesLoadedMsg     struct{ workspaces []models.Workspace }
	dataStoresLoadedMsg     struct{ node *models.TreeNode; stores []models.DataStore }
	coverageStoresLoadedMsg struct{ node *models.TreeNode; stores []models.CoverageStore }
	stylesLoadedMsg         struct{ node *models.TreeNode; styles []models.Style }
	layerGroupsLoadedMsg    struct{ node *models.TreeNode; groups []models.LayerGroup }
	layersLoadedMsg         struct{ node *models.TreeNode; layers []models.Layer }
	connectionTestMsg       struct{ success bool; err error; version string }
	uploadCompleteMsg       struct{ success bool; err error }
	errMsg                  struct{ err error }
	// CRUD operation messages
	crudCompleteMsg struct{ success bool; err error; operation string }
)

// App is the main TUI application
type App struct {
	config            *config.Config
	version           string
	client            *api.Client
	fileBrowser       *components.FileBrowser
	treeView          *components.TreeView
	connectionsScreen *screens.ConnectionsScreen
	screen            Screen
	activePanel       Panel
	keyMap            AppKeyMap
	width             int
	height            int
	spinner           spinner.Model
	loading           bool
	statusMsg         string
	errorMsg          string
	showHelp          bool

	// CRUD dialog state
	crudDialog    *components.Dialog
	crudOperation CRUDOperation
	crudNode      *models.TreeNode
	crudNodeType  models.NodeType

	// Store wizard state
	storeWizard *components.StoreWizard

	// Upload state
	pendingUploadFiles     []models.LocalFile
	pendingUploadWorkspace string

	// Tree state preservation
	savedTreeState components.TreeState

	// Pending CRUD command (set by dialog callbacks, executed when dialog closes)
	pendingCRUDCmd tea.Cmd

	// Info dialog state
	infoDialog *components.InfoDialog

	// Progress dialog state
	progressDialog *components.ProgressDialog

	// Track newly created items for focusing after refresh
	newlyCreatedPath string

	// Preview server for layer preview
	previewServer *preview.Server
}

// NewApp creates a new TUI application
func NewApp(cfg *config.Config, version string) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.LoadingStyle

	app := &App{
		config:            cfg,
		version:           version,
		fileBrowser:       components.NewFileBrowser(cfg.LastLocalPath),
		treeView:          components.NewTreeView(),
		connectionsScreen: screens.NewConnectionsScreen(cfg),
		screen:            ScreenMain,
		activePanel:       PanelLeft,
		keyMap:            DefaultAppKeyMap(),
		spinner:           s,
	}

	// Set the left panel as active by default
	app.fileBrowser.SetActive(true)
	app.treeView.SetActive(false)

	// Connect to active connection if available
	if conn := cfg.GetActiveConnection(); conn != nil {
		app.client = api.NewClient(conn)
		app.treeView.SetConnected(true, conn.Name)
	}

	return app
}

// Init initializes the TUI
func (a *App) Init() tea.Cmd {
	cmds := []tea.Cmd{
		a.spinner.Tick,
	}

	// Load workspaces if connected
	if a.client != nil {
		cmds = append(cmds, a.loadWorkspaces())
	}

	return tea.Batch(cmds...)
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateSizes()
		return a, nil

	case tea.KeyMsg:
		// If we have a progress dialog open, forward keys there first
		if a.progressDialog != nil && a.progressDialog.IsVisible() {
			var cmd tea.Cmd
			a.progressDialog, cmd = a.progressDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return a, tea.Batch(cmds...)
		}

		// If we have an info dialog open, forward keys there first
		if a.infoDialog != nil && a.infoDialog.IsVisible() {
			var cmd tea.Cmd
			a.infoDialog, cmd = a.infoDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if dialog was closed
			if !a.infoDialog.IsVisible() {
				a.infoDialog = nil
			}
			return a, tea.Batch(cmds...)
		}

		// If we have a store wizard open, forward keys there first
		if a.storeWizard != nil && a.storeWizard.IsVisible() {
			var cmd tea.Cmd
			a.storeWizard, cmd = a.storeWizard.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if wizard was closed
			if !a.storeWizard.IsVisible() {
				a.storeWizard = nil
				// Execute pending CRUD command if any
				if a.pendingCRUDCmd != nil {
					cmds = append(cmds, a.pendingCRUDCmd)
					a.pendingCRUDCmd = nil
				}
			}
			return a, tea.Batch(cmds...)
		}

		// If we have a CRUD dialog open, forward keys there first
		if a.crudDialog != nil && a.crudDialog.IsVisible() {
			var cmd tea.Cmd
			a.crudDialog, cmd = a.crudDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if dialog was closed
			if !a.crudDialog.IsVisible() {
				a.crudDialog = nil
				a.crudOperation = CRUDNone
				// Execute pending CRUD command if any
				if a.pendingCRUDCmd != nil {
					cmds = append(cmds, a.pendingCRUDCmd)
					a.pendingCRUDCmd = nil
				}
			}
			return a, tea.Batch(cmds...)
		}

		// If we're in the connections screen and editing a field, forward keys there first
		if a.screen == ScreenConnections && a.connectionsScreen.IsEditingField() {
			var cmd tea.Cmd
			a.connectionsScreen, cmd = a.connectionsScreen.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return a, tea.Batch(cmds...)
		}

		// Handle global keys
		switch {
		case key.Matches(msg, a.keyMap.Quit):
			// Save config before quitting
			a.config.LastLocalPath = a.fileBrowser.CurrentPath()
			_ = a.config.Save()
			return a, tea.Quit

		case key.Matches(msg, a.keyMap.Tab):
			if a.screen == ScreenMain {
				a.switchPanel()
				return a, nil
			}

		case key.Matches(msg, a.keyMap.Help):
			a.showHelp = !a.showHelp
			return a, nil

		case key.Matches(msg, a.keyMap.Escape):
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
			if a.screen != ScreenMain {
				a.screen = ScreenMain
				return a, nil
			}

		case key.Matches(msg, a.keyMap.Connections):
			if a.screen == ScreenMain {
				a.screen = ScreenConnections
				return a, nil
			}

		case key.Matches(msg, a.keyMap.Upload):
			if a.screen == ScreenMain {
				return a, a.handleUpload()
			}

		case key.Matches(msg, a.keyMap.Refresh):
			if a.screen == ScreenMain {
				if a.activePanel == PanelLeft {
					a.fileBrowser.Refresh()
				} else if a.client != nil {
					a.treeView.Clear()
					return a, a.loadWorkspaces()
				}
				return a, nil
			}
		}

		// Handle screen-specific keys
		if a.screen == ScreenConnections && !a.showHelp {
			var cmd tea.Cmd
			a.connectionsScreen, cmd = a.connectionsScreen.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if we should connect after user selects a connection
			if conn := a.connectionsScreen.GetActiveConnection(); conn != nil && a.client == nil {
				a.client = api.NewClient(conn)
				a.treeView.SetConnected(true, conn.Name)
				cmds = append(cmds, a.loadWorkspaces())
			}
		} else if a.screen == ScreenMain && !a.showHelp {
			if a.activePanel == PanelLeft {
				var cmd tea.Cmd
				a.fileBrowser, cmd = a.fileBrowser.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			} else {
				var cmd tea.Cmd
				a.treeView, cmd = a.treeView.Update(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}

				// Check if we need to load children for the selected node
				if key.Matches(msg, key.NewBinding(key.WithKeys("enter", "l", "right"))) {
					if node := a.treeView.SelectedNode(); node != nil && !node.IsLoaded && !node.IsLoading {
						cmd = a.loadNodeChildren(node)
						if cmd != nil {
							cmds = append(cmds, cmd)
						}
					}
				}
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case workspacesLoadedMsg:
		a.loading = false
		a.buildWorkspaceTree(msg.workspaces)
		a.treeView.Refresh()
		// If we have a newly created item, navigate to it
		if a.newlyCreatedPath != "" {
			a.treeView.NavigateToPath(a.newlyCreatedPath)
			a.newlyCreatedPath = ""
		} else if a.savedTreeState.CursorPath != "" || len(a.savedTreeState.ExpandedPaths) > 0 {
			// Restore tree state if we have saved state (after CRUD operations or upload)
			a.treeView.RestoreState(a.savedTreeState)
			// Clear saved state after restoration
			a.savedTreeState = components.TreeState{}
		}

	case dataStoresLoadedMsg:
		msg.node.IsLoading = false
		msg.node.IsLoaded = true
		for _, store := range msg.stores {
			child := models.NewTreeNode(store.Name, models.NodeTypeDataStore)
			child.Workspace = msg.node.Workspace
			msg.node.AddChild(child)
		}
		a.treeView.Refresh()

	case coverageStoresLoadedMsg:
		msg.node.IsLoading = false
		msg.node.IsLoaded = true
		for _, store := range msg.stores {
			child := models.NewTreeNode(store.Name, models.NodeTypeCoverageStore)
			child.Workspace = msg.node.Workspace
			msg.node.AddChild(child)
		}
		a.treeView.Refresh()

	case stylesLoadedMsg:
		msg.node.IsLoading = false
		msg.node.IsLoaded = true
		for _, style := range msg.styles {
			child := models.NewTreeNode(style.Name, models.NodeTypeStyle)
			child.Workspace = msg.node.Workspace
			msg.node.AddChild(child)
		}
		a.treeView.Refresh()

	case layerGroupsLoadedMsg:
		msg.node.IsLoading = false
		msg.node.IsLoaded = true
		for _, group := range msg.groups {
			child := models.NewTreeNode(group.Name, models.NodeTypeLayerGroup)
			child.Workspace = msg.node.Workspace
			msg.node.AddChild(child)
		}
		a.treeView.Refresh()

	case layersLoadedMsg:
		msg.node.IsLoading = false
		msg.node.IsLoaded = true
		for _, layer := range msg.layers {
			child := models.NewTreeNode(layer.Name, models.NodeTypeLayer)
			child.Workspace = msg.node.Workspace
			msg.node.AddChild(child)
		}
		a.treeView.Refresh()

	case connectionTestMsg:
		a.loading = false
		if msg.success {
			a.statusMsg = fmt.Sprintf("Connected to GeoServer %s", msg.version)
			a.errorMsg = ""
			return a, a.loadWorkspaces()
		} else {
			a.errorMsg = fmt.Sprintf("Connection failed: %v", msg.err)
			a.treeView.SetConnected(false, "")
		}

	case screens.ConnectionTestMsg:
		// Forward to connections screen and handle result
		var cmd tea.Cmd
		a.connectionsScreen, cmd = a.connectionsScreen.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if msg.Success {
			// Update connection status
			if conn := a.config.GetActiveConnection(); conn != nil {
				a.client = api.NewClient(conn)
				a.treeView.SetConnected(true, conn.Name)
				cmds = append(cmds, a.loadWorkspaces())
			}
		}

	case uploadCompleteMsg:
		a.loading = false
		if msg.success {
			a.statusMsg = "Upload completed successfully"
			a.errorMsg = ""
			// Refresh the tree
			a.treeView.Clear()
			return a, a.loadWorkspaces()
		} else {
			a.errorMsg = fmt.Sprintf("Upload failed: %v", msg.err)
		}

	case crudCompleteMsg:
		a.loading = false
		if msg.success {
			a.statusMsg = msg.operation + " completed successfully"
			a.errorMsg = ""
			// Refresh the tree
			a.treeView.Clear()
			return a, a.loadWorkspaces()
		} else {
			a.errorMsg = fmt.Sprintf("%s failed: %v", msg.operation, msg.err)
		}

	case components.TreeNewMsg:
		// Show create dialog based on node type
		return a, a.showCreateDialog(msg.Node, msg.NodeType)

	case components.TreeEditMsg:
		// Show edit dialog for the node
		return a, a.showEditDialog(msg.Node)

	case components.TreeDeleteMsg:
		// Show delete confirmation dialog
		return a, a.showDeleteDialog(msg.Node)

	case components.DialogAnimationMsg:
		// Forward to dialog if we have one
		if a.crudDialog != nil {
			wasVisible := a.crudDialog.IsVisible()
			var cmd tea.Cmd
			a.crudDialog, cmd = a.crudDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if dialog was closed during animation
			if wasVisible && !a.crudDialog.IsVisible() {
				a.crudDialog = nil
				a.crudOperation = CRUDNone
				// Execute pending CRUD command if any
				if a.pendingCRUDCmd != nil {
					cmds = append(cmds, a.pendingCRUDCmd)
					a.pendingCRUDCmd = nil
				}
			}
		}

	case components.StoreWizardAnimationMsg:
		// Forward to wizard if we have one
		if a.storeWizard != nil {
			wasVisible := a.storeWizard.IsVisible()
			var cmd tea.Cmd
			a.storeWizard, cmd = a.storeWizard.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if wizard was closed during animation
			if wasVisible && !a.storeWizard.IsVisible() {
				a.storeWizard = nil
				// Execute pending CRUD command if any
				if a.pendingCRUDCmd != nil {
					cmds = append(cmds, a.pendingCRUDCmd)
					a.pendingCRUDCmd = nil
				}
			}
		}

	case components.InfoDialogAnimationMsg:
		// Forward to info dialog if we have one
		if a.infoDialog != nil && a.infoDialog.IsVisible() {
			var cmd tea.Cmd
			a.infoDialog, cmd = a.infoDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if dialog was closed during animation
			if !a.infoDialog.IsVisible() {
				a.infoDialog = nil
			}
		}

	case components.ProgressDialogAnimationMsg:
		// Forward to progress dialog if we have one
		if a.progressDialog != nil {
			wasVisible := a.progressDialog.IsVisible()
			var cmd tea.Cmd
			a.progressDialog, cmd = a.progressDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			// Check if dialog was closed during animation
			if wasVisible && !a.progressDialog.IsVisible() {
				// Trigger tree refresh if upload completed successfully
				if a.progressDialog.IsDone() && a.progressDialog.GetError() == nil {
					a.treeView.Clear()
					cmds = append(cmds, a.loadWorkspaces())
				}
				a.progressDialog = nil
			}
		}

	case components.ProgressUpdateMsg:
		// Forward to progress dialog if we have one
		if a.progressDialog != nil {
			var cmd tea.Cmd
			a.progressDialog, cmd = a.progressDialog.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case UploadNextMsg:
		// Continue uploading the next file
		cmds = append(cmds,
			components.SendProgressUpdate("Uploading Files", msg.Index, len(msg.Files), msg.Files[msg.Index].Name, false, nil),
			a.uploadFile(msg.Files, msg.Workspace, msg.Index),
		)

	case components.FileInfoMsg:
		// Show info dialog for file
		a.infoDialog = components.NewFileInfoDialog(msg.File)
		a.infoDialog.SetSize(a.width, a.height)
		return a, a.infoDialog.Init()

	case components.TreeInfoMsg:
		// Show info dialog for tree node
		a.infoDialog = components.NewTreeNodeInfoDialog(msg.Node)
		a.infoDialog.SetSize(a.width, a.height)
		return a, a.infoDialog.Init()

	case components.TreePreviewMsg:
		// Open layer preview in browser
		return a, a.openLayerPreview(msg.Node)

	case errMsg:
		a.loading = false
		a.errorMsg = msg.err.Error()
	}

	return a, tea.Batch(cmds...)
}

// View renders the TUI
func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	var content string

	switch a.screen {
	case ScreenMain:
		content = a.renderMainScreen()
	case ScreenConnections:
		content = a.renderConnectionsScreen()
	case ScreenHelp:
		content = a.renderHelpScreen()
	default:
		content = a.renderMainScreen()
	}

	if a.showHelp {
		content = a.renderHelpOverlay(content)
	}

	// Render CRUD dialog overlay
	if a.crudDialog != nil && a.crudDialog.IsVisible() {
		a.crudDialog.SetSize(a.width, a.height)
		content = a.crudDialog.View()
	}

	// Render store wizard overlay
	if a.storeWizard != nil && a.storeWizard.IsVisible() {
		a.storeWizard.SetSize(a.width, a.height)
		content = a.storeWizard.View()
	}

	// Render info dialog overlay
	if a.infoDialog != nil && a.infoDialog.IsVisible() {
		a.infoDialog.SetSize(a.width, a.height)
		content = a.infoDialog.View()
	}

	// Render progress dialog overlay (highest priority)
	if a.progressDialog != nil && a.progressDialog.IsVisible() {
		a.progressDialog.SetSize(a.width, a.height)
		content = a.progressDialog.View()
	}

	return content
}

// renderMainScreen renders the main dual-panel screen
func (a *App) renderMainScreen() string {
	// Calculate panel widths
	panelWidth := (a.width - 1) / 2 // -1 for the separator
	panelHeight := a.height - 4     // -4 for title bar and status bar

	// Update component sizes
	a.fileBrowser.SetSize(panelWidth, panelHeight)
	a.treeView.SetSize(panelWidth, panelHeight)

	// Render panels
	leftPanel := a.fileBrowser.View()
	rightPanel := a.treeView.View()

	// Join panels horizontally
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Title bar
	title := styles.TitleStyle.Render(" 🌍 Kartoza GeoServer Client")
	titleBar := lipgloss.PlaceHorizontal(a.width, lipgloss.Left, title)

	// Status bar
	statusBar := a.renderStatusBar()

	// Help bar
	helpBar := a.renderHelpBar()

	// Combine everything
	return styles.JoinVertical(titleBar, panels, statusBar, helpBar)
}

// renderStatusBar renders the status bar
func (a *App) renderStatusBar() string {
	var status string

	if a.loading {
		status = a.spinner.View() + " Loading..."
	} else if a.errorMsg != "" {
		status = styles.ErrorStyle.Render("✗ " + a.errorMsg)
	} else if a.statusMsg != "" {
		status = styles.SuccessStyle.Render("✓ " + a.statusMsg)
	} else {
		// Show current selection info
		if a.activePanel == PanelLeft {
			if file := a.fileBrowser.SelectedFile(); file != nil {
				status = fmt.Sprintf(" %s %s", file.Type.Icon(), file.Path)
			}
		} else {
			if node := a.treeView.SelectedNode(); node != nil {
				status = fmt.Sprintf(" %s %s", node.Type.Icon(), node.Path())
			}
		}
	}

	return styles.StatusBarStyle.Width(a.width).Render(status)
}

// renderHelpBar renders the help bar at the bottom
func (a *App) renderHelpBar() string {
	var items []string

	items = append(items, styles.RenderHelpKey("Tab", "switch"))
	items = append(items, styles.RenderHelpKey("↑↓", "navigate"))
	items = append(items, styles.RenderHelpKey("Enter", "open"))
	items = append(items, styles.RenderHelpKey("i", "info"))

	// Show CRUD shortcuts when on the tree panel and connected
	if a.activePanel == PanelRight && a.treeView.IsConnected() {
		items = append(items, styles.RenderHelpKey("n", "new"))
		items = append(items, styles.RenderHelpKey("e", "edit"))
		items = append(items, styles.RenderHelpKey("d", "delete"))
		// Show preview option for layers and stores
		if node := a.treeView.SelectedNode(); node != nil {
			switch node.Type {
			case models.NodeTypeLayer, models.NodeTypeLayerGroup,
				models.NodeTypeDataStore, models.NodeTypeCoverageStore:
				items = append(items, styles.RenderHelpKey("o", "preview"))
			}
		}
	}

	items = append(items, styles.RenderHelpKey("c", "connections"))
	items = append(items, styles.RenderHelpKey("u", "upload"))
	items = append(items, styles.RenderHelpKey("r", "refresh"))
	items = append(items, styles.RenderHelpKey("?", "help"))
	items = append(items, styles.RenderHelpKey("q", "quit"))

	return styles.HelpBarStyle.Width(a.width).Render(strings.Join(items, "  "))
}

// renderConnectionsScreen renders the connections management screen
func (a *App) renderConnectionsScreen() string {
	a.connectionsScreen.SetSize(a.width, a.height)
	return a.connectionsScreen.View()
}

// renderHelpScreen renders the help screen
func (a *App) renderHelpScreen() string {
	return a.renderHelpOverlay("")
}

// renderHelpOverlay renders the help overlay
func (a *App) renderHelpOverlay(background string) string {
	var b strings.Builder

	title := styles.DialogTitleStyle.Render("Help - Keyboard Shortcuts")
	b.WriteString(title)
	b.WriteString("\n\n")

	sections := []struct {
		title string
		keys  [][2]string
	}{
		{
			title: "Navigation",
			keys: [][2]string{
				{"↑/k", "Move up"},
				{"↓/j", "Move down"},
				{"Enter/l", "Open / Expand"},
				{"Backspace/h", "Back / Collapse"},
				{"Tab", "Switch panel"},
				{"PgUp/PgDn", "Page up/down"},
				{"Home/End", "Go to start/end"},
			},
		},
		{
			title: "Actions",
			keys: [][2]string{
				{"Space", "Select file"},
				{"u", "Upload selected"},
				{"r", "Refresh"},
				{"c", "Manage connections"},
			},
		},
		{
			title: "General",
			keys: [][2]string{
				{"?/F1", "Toggle help"},
				{"Esc", "Back / Close"},
				{"q/Ctrl+C", "Quit"},
			},
		},
	}

	for _, section := range sections {
		b.WriteString(styles.PanelHeaderStyle.Render(section.title))
		b.WriteString("\n")
		for _, kv := range section.keys {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				styles.HelpKeyStyle.Width(12).Render(kv[0]),
				styles.HelpTextStyle.Render(kv[1])))
		}
		b.WriteString("\n")
	}

	b.WriteString(styles.HelpTextStyle.Render("Press ? or Esc to close"))

	dialog := styles.DialogStyle.Width(50).Render(b.String())
	return styles.Center(a.width, a.height, dialog)
}

// switchPanel switches between left and right panels
func (a *App) switchPanel() {
	if a.activePanel == PanelLeft {
		a.activePanel = PanelRight
		a.fileBrowser.SetActive(false)
		a.treeView.SetActive(true)
	} else {
		a.activePanel = PanelLeft
		a.fileBrowser.SetActive(true)
		a.treeView.SetActive(false)
	}
}

// updateSizes updates component sizes based on window size
func (a *App) updateSizes() {
	panelWidth := (a.width - 1) / 2
	panelHeight := a.height - 4

	a.fileBrowser.SetSize(panelWidth, panelHeight)
	a.treeView.SetSize(panelWidth, panelHeight)
}

// loadWorkspaces loads workspaces from the server
func (a *App) loadWorkspaces() tea.Cmd {
	a.loading = true
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{fmt.Errorf("not connected")}
		}

		workspaces, err := a.client.GetWorkspaces()
		if err != nil {
			return errMsg{err}
		}

		return workspacesLoadedMsg{workspaces}
	}
}

// buildWorkspaceTree builds the tree structure from workspaces
func (a *App) buildWorkspaceTree(workspaces []models.Workspace) {
	root := models.NewTreeNode("GeoServer", models.NodeTypeRoot)
	root.Expanded = true

	for _, ws := range workspaces {
		wsNode := models.NewTreeNode(ws.Name, models.NodeTypeWorkspace)
		wsNode.Workspace = ws.Name

		// Add category nodes
		dsNode := models.NewTreeNode("Data Stores", models.NodeTypeDataStores)
		dsNode.Workspace = ws.Name
		wsNode.AddChild(dsNode)

		csNode := models.NewTreeNode("Coverage Stores", models.NodeTypeCoverageStores)
		csNode.Workspace = ws.Name
		wsNode.AddChild(csNode)

		stylesNode := models.NewTreeNode("Styles", models.NodeTypeStyles)
		stylesNode.Workspace = ws.Name
		wsNode.AddChild(stylesNode)

		layersNode := models.NewTreeNode("Layers", models.NodeTypeLayers)
		layersNode.Workspace = ws.Name
		wsNode.AddChild(layersNode)

		lgNode := models.NewTreeNode("Layer Groups", models.NodeTypeLayerGroups)
		lgNode.Workspace = ws.Name
		wsNode.AddChild(lgNode)

		root.AddChild(wsNode)
	}

	a.treeView.SetRoot(root)
}

// loadNodeChildren loads children for a tree node
func (a *App) loadNodeChildren(node *models.TreeNode) tea.Cmd {
	if a.client == nil || node.IsLoaded || node.IsLoading {
		return nil
	}

	node.IsLoading = true

	switch node.Type {
	case models.NodeTypeDataStores:
		return func() tea.Msg {
			stores, err := a.client.GetDataStores(node.Workspace)
			if err != nil {
				node.IsLoading = false
				node.HasError = true
				node.ErrorMsg = err.Error()
				return errMsg{err}
			}
			return dataStoresLoadedMsg{node: node, stores: stores}
		}

	case models.NodeTypeCoverageStores:
		return func() tea.Msg {
			stores, err := a.client.GetCoverageStores(node.Workspace)
			if err != nil {
				node.IsLoading = false
				node.HasError = true
				node.ErrorMsg = err.Error()
				return errMsg{err}
			}
			return coverageStoresLoadedMsg{node: node, stores: stores}
		}

	case models.NodeTypeStyles:
		return func() tea.Msg {
			styles, err := a.client.GetStyles(node.Workspace)
			if err != nil {
				node.IsLoading = false
				node.HasError = true
				node.ErrorMsg = err.Error()
				return errMsg{err}
			}
			return stylesLoadedMsg{node: node, styles: styles}
		}

	case models.NodeTypeLayers:
		return func() tea.Msg {
			layers, err := a.client.GetLayers(node.Workspace)
			if err != nil {
				node.IsLoading = false
				node.HasError = true
				node.ErrorMsg = err.Error()
				return errMsg{err}
			}
			return layersLoadedMsg{node: node, layers: layers}
		}

	case models.NodeTypeLayerGroups:
		return func() tea.Msg {
			groups, err := a.client.GetLayerGroups(node.Workspace)
			if err != nil {
				node.IsLoading = false
				node.HasError = true
				node.ErrorMsg = err.Error()
				return errMsg{err}
			}
			return layerGroupsLoadedMsg{node: node, groups: groups}
		}
	}

	return nil
}

// handleUpload handles file upload - shows confirmation dialog first
func (a *App) handleUpload() tea.Cmd {
	if a.client == nil {
		a.errorMsg = "Not connected to GeoServer"
		return nil
	}

	// Get selected files
	selectedFiles := a.fileBrowser.SelectedFiles()
	if len(selectedFiles) == 0 {
		// Use current file if none selected
		if file := a.fileBrowser.SelectedFile(); file != nil && !file.IsDir {
			selectedFiles = []models.LocalFile{*file}
		}
	}

	if len(selectedFiles) == 0 {
		a.errorMsg = "No files selected for upload"
		return nil
	}

	// Get target workspace from tree selection
	targetNode := a.treeView.SelectedNode()
	var workspace string
	if targetNode != nil {
		workspace = targetNode.Workspace
	}

	if workspace == "" {
		a.errorMsg = "Select a workspace in the GeoServer tree first"
		return nil
	}

	// Store pending upload info
	a.pendingUploadFiles = selectedFiles
	a.pendingUploadWorkspace = workspace

	// Build confirmation message
	var fileList strings.Builder
	for i, file := range selectedFiles {
		if i > 0 {
			fileList.WriteString("\n")
		}
		fileList.WriteString(fmt.Sprintf("  %s %s", file.Type.Icon(), file.Name))
		if i >= 4 && len(selectedFiles) > 5 {
			fileList.WriteString(fmt.Sprintf("\n  ... and %d more files", len(selectedFiles)-5))
			break
		}
	}

	message := fmt.Sprintf("Upload %d file(s) to workspace '%s'?\n\nSource files:\n%s\n\nDestination: %s",
		len(selectedFiles), workspace, fileList.String(), workspace)

	a.crudDialog = components.NewConfirmDialog("Confirm Upload", message)
	a.crudDialog.SetSize(a.width, a.height)

	a.crudDialog.SetCallbacks(
		func(result components.DialogResult) {
			if result.Confirmed {
				a.pendingCRUDCmd = a.executeUpload()
			}
		},
		func() {
			// Cancel - clear pending upload
			a.pendingUploadFiles = nil
			a.pendingUploadWorkspace = ""
		},
	)

	return a.crudDialog.Init()
}

// executeUpload performs the actual file upload with progress dialog
func (a *App) executeUpload() tea.Cmd {
	if len(a.pendingUploadFiles) == 0 || a.pendingUploadWorkspace == "" {
		a.errorMsg = "No upload pending"
		return nil
	}

	selectedFiles := a.pendingUploadFiles
	workspace := a.pendingUploadWorkspace

	// Clear pending state
	a.pendingUploadFiles = nil
	a.pendingUploadWorkspace = ""

	// Save tree state before upload
	a.savedTreeState = a.treeView.SaveState()

	// Build list of file names for the progress dialog
	fileNames := make([]string, len(selectedFiles))
	for i, f := range selectedFiles {
		fileNames[i] = f.Name
	}

	// Create progress dialog
	a.progressDialog = components.NewProgressDialog("Uploading Files", "📤", fileNames)
	a.progressDialog.SetSize(a.width, a.height)

	// Start the upload in a goroutine and return the init command
	return tea.Batch(
		a.progressDialog.Init(),
		a.startUpload(selectedFiles, workspace),
	)
}

// UploadNextMsg signals to upload the next file
type UploadNextMsg struct {
	Files     []models.LocalFile
	Workspace string
	Index     int
}

// startUpload starts the upload process by uploading the first file
func (a *App) startUpload(files []models.LocalFile, workspace string) tea.Cmd {
	if len(files) == 0 {
		return nil
	}
	// Send progress update for the first file and start uploading
	return tea.Batch(
		components.SendProgressUpdate("Uploading Files", 0, len(files), files[0].Name, false, nil),
		a.uploadFile(files, workspace, 0),
	)
}

// uploadFile uploads a single file and returns a command to continue or finish
func (a *App) uploadFile(files []models.LocalFile, workspace string, index int) tea.Cmd {
	return func() tea.Msg {
		file := files[index]
		storeName := strings.TrimSuffix(file.Name, filepath.Ext(file.Name))

		var err error
		switch file.Type {
		case models.FileTypeShapefile:
			err = a.client.UploadShapefile(workspace, storeName, file.Path)
		case models.FileTypeGeoTIFF:
			err = a.client.UploadGeoTIFF(workspace, storeName, file.Path)
		case models.FileTypeGeoPackage:
			err = a.client.UploadGeoPackage(workspace, storeName, file.Path)
		case models.FileTypeSLD, models.FileTypeCSS:
			format := "sld"
			if file.Type == models.FileTypeCSS {
				format = "css"
			}
			err = a.client.UploadStyle(workspace, storeName, file.Path, format)
		default:
			err = fmt.Errorf("unsupported file type: %s", file.Type)
		}

		if err != nil {
			return components.ProgressUpdateMsg{
				ID:       "Uploading Files",
				Current:  index,
				Total:    len(files),
				ItemName: file.Name,
				Done:     true,
				Error:    err,
			}
		}

		// Check if there are more files
		if index+1 < len(files) {
			return UploadNextMsg{
				Files:     files,
				Workspace: workspace,
				Index:     index + 1,
			}
		}

		// All files uploaded successfully
		return components.ProgressUpdateMsg{
			ID:       "Uploading Files",
			Current:  len(files),
			Total:    len(files),
			ItemName: "",
			Done:     true,
			Error:    nil,
		}
	}
}

// openLayerPreview opens the layer preview in the browser
func (a *App) openLayerPreview(node *models.TreeNode) tea.Cmd {
	return func() tea.Msg {
		if a.client == nil {
			return errMsg{err: fmt.Errorf("not connected to GeoServer")}
		}

		var layerName string
		var layerType string
		var storeName string
		var storeType string

		switch node.Type {
		case models.NodeTypeLayer, models.NodeTypeLayerGroup:
			layerName = node.Name
			storeName = node.StoreName
			storeType = node.StoreType
			layerType = "vector"
			if node.StoreType == "coveragestore" {
				layerType = "raster"
			}
		case models.NodeTypeDataStore:
			// For data stores, use the store name as layer name (GeoServer convention)
			layerName = node.Name
			storeName = node.Name
			storeType = "datastore"
			layerType = "vector"
		case models.NodeTypeCoverageStore:
			// For coverage stores, use the store name as layer name
			layerName = node.Name
			storeName = node.Name
			storeType = "coveragestore"
			layerType = "raster"
		default:
			return errMsg{err: fmt.Errorf("can only preview layers, layer groups, and stores")}
		}

		// Create layer info
		layerInfo := &preview.LayerInfo{
			Name:         layerName,
			Workspace:    node.Workspace,
			StoreName:    storeName,
			StoreType:    storeType,
			GeoServerURL: a.client.BaseURL(),
			Type:         layerType,
		}

		// Start or update preview server
		if a.previewServer == nil {
			a.previewServer = preview.NewServer()
		}

		url, err := a.previewServer.Start(layerInfo)
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to start preview server: %w", err)}
		}

		// Open browser
		if err := preview.OpenBrowser(url); err != nil {
			return errMsg{err: fmt.Errorf("failed to open browser: %w", err)}
		}

		return nil
	}
}

// showCreateDialog shows a dialog to create a new item
func (a *App) showCreateDialog(contextNode *models.TreeNode, nodeType models.NodeType) tea.Cmd {
	if a.client == nil {
		a.errorMsg = "Not connected to GeoServer"
		return nil
	}

	// Get workspace from context
	workspace := ""
	if contextNode != nil {
		workspace = contextNode.Workspace
	}

	switch nodeType {
	case models.NodeTypeWorkspace:
		// Use simple dialog for workspace (just needs a name)
		fields := []components.DialogField{
			{Name: "name", Label: "Name", Placeholder: "workspace-name"},
		}
		a.crudDialog = components.NewInputDialog("Create Workspace", fields)
		a.crudDialog.SetSize(a.width, a.height)
		a.crudOperation = CRUDCreate
		a.crudNode = contextNode
		a.crudNodeType = nodeType

		a.crudDialog.SetCallbacks(
			func(result components.DialogResult) {
				if result.Confirmed {
					a.pendingCRUDCmd = a.executeCRUDCreate(result.Values)
				}
			},
			func() {},
		)
		return a.crudDialog.Init()

	case models.NodeTypeDataStore:
		// Use wizard for data store (needs type selection + configuration)
		if workspace == "" {
			a.errorMsg = "Select a workspace first"
			return nil
		}
		a.storeWizard = components.NewDataStoreWizard(workspace)
		a.storeWizard.SetSize(a.width, a.height)
		a.crudNode = contextNode

		a.storeWizard.SetCallbacks(
			func(result components.StoreWizardResult) {
				if result.Confirmed {
					a.pendingCRUDCmd = a.executeDataStoreCreate(workspace, result)
				}
			},
			func() {},
		)
		return a.storeWizard.Init()

	case models.NodeTypeCoverageStore:
		// Use wizard for coverage store
		if workspace == "" {
			a.errorMsg = "Select a workspace first"
			return nil
		}
		a.storeWizard = components.NewCoverageStoreWizard(workspace)
		a.storeWizard.SetSize(a.width, a.height)
		a.crudNode = contextNode

		a.storeWizard.SetCallbacks(
			func(result components.StoreWizardResult) {
				if result.Confirmed {
					a.pendingCRUDCmd = a.executeCoverageStoreCreate(workspace, result)
				}
			},
			func() {},
		)
		return a.storeWizard.Init()

	default:
		a.errorMsg = "Cannot create this type of item"
		return nil
	}
}

// showEditDialog shows a dialog to edit an item
func (a *App) showEditDialog(node *models.TreeNode) tea.Cmd {
	if a.client == nil {
		a.errorMsg = "Not connected to GeoServer"
		return nil
	}

	var title string
	var fields []components.DialogField

	switch node.Type {
	case models.NodeTypeWorkspace:
		title = "Edit Workspace"
		fields = []components.DialogField{
			{Name: "name", Label: "Name", Placeholder: "workspace-name", Value: node.Name},
		}

	case models.NodeTypeDataStore:
		title = "Edit Data Store"
		fields = []components.DialogField{
			{Name: "name", Label: "Name", Placeholder: "datastore-name", Value: node.Name},
		}

	case models.NodeTypeCoverageStore:
		title = "Edit Coverage Store"
		fields = []components.DialogField{
			{Name: "name", Label: "Name", Placeholder: "coveragestore-name", Value: node.Name},
		}

	default:
		a.errorMsg = "Cannot edit this type of item"
		return nil
	}

	a.crudDialog = components.NewInputDialog(title, fields)
	a.crudDialog.SetSize(a.width, a.height)
	a.crudOperation = CRUDEdit
	a.crudNode = node
	a.crudNodeType = node.Type

	// Set callbacks
	a.crudDialog.SetCallbacks(
		func(result components.DialogResult) {
			if result.Confirmed {
				a.pendingCRUDCmd = a.executeCRUDEdit(result.Values)
			}
		},
		func() {
			// Cancel - dialog will close automatically
		},
	)

	return a.crudDialog.Init()
}

// showDeleteDialog shows a confirmation dialog to delete an item
func (a *App) showDeleteDialog(node *models.TreeNode) tea.Cmd {
	if a.client == nil {
		a.errorMsg = "Not connected to GeoServer"
		return nil
	}

	message := fmt.Sprintf("Are you sure you want to delete %s '%s'?\nThis action cannot be undone.",
		node.Type.String(), node.Name)

	a.crudDialog = components.NewConfirmDialog("Delete "+node.Type.String(), message)
	a.crudDialog.SetSize(a.width, a.height)
	a.crudOperation = CRUDDelete
	a.crudNode = node
	a.crudNodeType = node.Type

	// Set callbacks
	a.crudDialog.SetCallbacks(
		func(result components.DialogResult) {
			if result.Confirmed {
				a.pendingCRUDCmd = a.executeCRUDDelete()
			}
		},
		func() {
			// Cancel - dialog will close automatically
		},
	)

	return a.crudDialog.Init()
}

// executeCRUDCreate executes the create operation for workspaces
func (a *App) executeCRUDCreate(values map[string]string) tea.Cmd {
	name := strings.TrimSpace(values["name"])
	if name == "" {
		a.errorMsg = "Name is required"
		return nil
	}

	// Set path to navigate to after creation (workspace name is the path)
	a.newlyCreatedPath = name

	a.loading = true
	return func() tea.Msg {
		operation := "Create workspace"
		err := a.client.CreateWorkspace(name)
		return crudCompleteMsg{success: err == nil, err: err, operation: operation}
	}
}

// executeDataStoreCreate executes the data store creation
func (a *App) executeDataStoreCreate(workspace string, result components.StoreWizardResult) tea.Cmd {
	name := strings.TrimSpace(result.Values["name"])
	if name == "" {
		a.errorMsg = "Store name is required"
		return nil
	}

	// Set path to navigate to after creation (workspace/Data Stores/storename)
	a.newlyCreatedPath = workspace + "/Data Stores/" + name

	a.loading = true
	return func() tea.Msg {
		operation := fmt.Sprintf("Create data store '%s'", name)
		err := a.client.CreateDataStore(workspace, name, result.DataStoreType, result.Values)
		return crudCompleteMsg{success: err == nil, err: err, operation: operation}
	}
}

// executeCoverageStoreCreate executes the coverage store creation
func (a *App) executeCoverageStoreCreate(workspace string, result components.StoreWizardResult) tea.Cmd {
	name := strings.TrimSpace(result.Values["name"])
	if name == "" {
		a.errorMsg = "Store name is required"
		return nil
	}

	url := result.Values["url"]
	if url == "" {
		a.errorMsg = "File path is required"
		return nil
	}

	// Set path to navigate to after creation (workspace/Coverage Stores/storename)
	a.newlyCreatedPath = workspace + "/Coverage Stores/" + name

	a.loading = true
	return func() tea.Msg {
		operation := fmt.Sprintf("Create coverage store '%s'", name)
		err := a.client.CreateCoverageStore(workspace, name, result.CoverageStoreType, url)
		return crudCompleteMsg{success: err == nil, err: err, operation: operation}
	}
}

// executeCRUDEdit executes the edit operation
func (a *App) executeCRUDEdit(values map[string]string) tea.Cmd {
	newName := strings.TrimSpace(values["name"])
	if newName == "" {
		a.errorMsg = "Name is required"
		return nil
	}

	if a.crudNode == nil {
		a.errorMsg = "No item selected"
		return nil
	}

	oldName := a.crudNode.Name
	if newName == oldName {
		return nil // No change
	}

	workspace := a.crudNode.Workspace
	nodeType := a.crudNodeType

	// Set path to navigate to after rename (the renamed item)
	switch nodeType {
	case models.NodeTypeWorkspace:
		a.newlyCreatedPath = newName
	case models.NodeTypeDataStore:
		a.newlyCreatedPath = workspace + "/Data Stores/" + newName
	case models.NodeTypeCoverageStore:
		a.newlyCreatedPath = workspace + "/Coverage Stores/" + newName
	}

	a.loading = true
	return func() tea.Msg {
		var err error
		var operation string

		switch nodeType {
		case models.NodeTypeWorkspace:
			operation = "Rename workspace"
			err = a.client.UpdateWorkspace(oldName, newName)

		case models.NodeTypeDataStore:
			operation = "Rename data store"
			err = a.client.UpdateDataStore(workspace, oldName, newName)

		case models.NodeTypeCoverageStore:
			operation = "Rename coverage store"
			err = a.client.UpdateCoverageStore(workspace, oldName, newName)
		}

		return crudCompleteMsg{success: err == nil, err: err, operation: operation}
	}
}

// executeCRUDDelete executes the delete operation
func (a *App) executeCRUDDelete() tea.Cmd {
	if a.crudNode == nil {
		a.errorMsg = "No item selected"
		return nil
	}

	// Navigate to parent after deletion
	if a.crudNode.Parent != nil && a.crudNode.Parent.Type != models.NodeTypeRoot {
		a.newlyCreatedPath = a.crudNode.Parent.Path()
	} else {
		// If deleting a workspace, just save current state
		a.savedTreeState = a.treeView.SaveState()
	}

	nodeName := a.crudNode.Name
	workspace := a.crudNode.Workspace
	nodeType := a.crudNodeType

	a.loading = true
	return func() tea.Msg {
		var err error
		var operation string

		switch nodeType {
		case models.NodeTypeWorkspace:
			operation = "Delete workspace"
			err = a.client.DeleteWorkspace(nodeName, true)

		case models.NodeTypeDataStore:
			operation = "Delete data store"
			err = a.client.DeleteDataStore(workspace, nodeName, true)

		case models.NodeTypeCoverageStore:
			operation = "Delete coverage store"
			err = a.client.DeleteCoverageStore(workspace, nodeName, true)

		case models.NodeTypeLayer:
			operation = "Delete layer"
			err = a.client.DeleteLayer(workspace, nodeName)

		case models.NodeTypeStyle:
			operation = "Delete style"
			err = a.client.DeleteStyle(workspace, nodeName, true)

		case models.NodeTypeLayerGroup:
			operation = "Delete layer group"
			err = a.client.DeleteLayerGroup(workspace, nodeName)
		}

		return crudCompleteMsg{success: err == nil, err: err, operation: operation}
	}
}
