// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-cloudbench/internal/api"
	"github.com/kartoza/kartoza-cloudbench/internal/config"
	"github.com/kartoza/kartoza-cloudbench/internal/integration"
	"github.com/kartoza/kartoza-cloudbench/internal/postgres"
)

// BridgeWizardStep represents the current step in the bridge wizard
type BridgeWizardStep int

const (
	BridgeStepSelectPGService BridgeWizardStep = iota
	BridgeStepSelectGeoServer
	BridgeStepSelectWorkspace
	BridgeStepEnterStoreName
	BridgeStepSelectSchema
	BridgeStepSelectTables
	BridgeStepConfirm
	BridgeStepCreating
	BridgeStepComplete
)

// BridgeWizard handles creating a PostgreSQL to GeoServer bridge
type BridgeWizard struct {
	width  int
	height int

	step           BridgeWizardStep
	storeNameInput textinput.Model

	// Selection items
	pgServices    []string
	pgServiceIdx  int
	geoServers    []string
	geoServerIdx  int
	workspaces    []string
	workspaceIdx  int
	schemas       []string
	schemaIdx     int
	tables        []string
	tableSelected map[string]bool

	selectedPGService string
	selectedGeoServer string
	selectedWorkspace string
	selectedSchema    string
	selectedTables    []string
	storeName         string
	publishLayers     bool

	config  *config.Config
	clients map[string]*api.Client

	result    *integration.LinkedStore
	err       error
	creating  bool
	cancelled bool
}

// BridgeResultMsg is sent when bridge creation completes
type BridgeResultMsg struct {
	Link *integration.LinkedStore
	Err  error
}

// NewBridgeWizard creates a new bridge wizard
func NewBridgeWizard(cfg *config.Config, clients map[string]*api.Client, width, height int) *BridgeWizard {
	w := &BridgeWizard{
		width:         width,
		height:        height,
		config:        cfg,
		clients:       clients,
		tableSelected: make(map[string]bool),
	}

	// Initialize store name input
	ti := textinput.New()
	ti.Placeholder = "my_postgis_store"
	ti.Width = 40
	w.storeNameInput = ti

	// Load PostgreSQL services
	w.loadPGServices()

	return w
}

func (w *BridgeWizard) loadPGServices() {
	w.pgServices = nil
	w.pgServiceIdx = 0

	if postgres.PGServiceFileExists() {
		services, err := postgres.ParsePGServiceFile()
		if err == nil {
			for _, svc := range services {
				w.pgServices = append(w.pgServices, svc.Name)
			}
		}
	}
}

func (w *BridgeWizard) loadGeoServerConnections() {
	w.geoServers = nil
	w.geoServerIdx = 0

	for _, conn := range w.config.Connections {
		w.geoServers = append(w.geoServers, conn.ID)
	}
}

func (w *BridgeWizard) loadWorkspaces() {
	w.workspaces = nil
	w.workspaceIdx = 0

	client := w.clients[w.selectedGeoServer]
	if client != nil {
		workspaces, err := client.GetWorkspaces()
		if err == nil {
			for _, ws := range workspaces {
				w.workspaces = append(w.workspaces, ws.Name)
			}
		}
	}
}

func (w *BridgeWizard) loadSchemas() {
	w.schemas = []string{"public"}
	w.schemaIdx = 0

	services, err := postgres.ParsePGServiceFile()
	if err == nil {
		svc, err := postgres.GetServiceByName(services, w.selectedPGService)
		if err == nil {
			db, err := svc.Connect()
			if err == nil {
				rows, err := db.Query(`
					SELECT schema_name
					FROM information_schema.schemata
					WHERE schema_name NOT LIKE 'pg_%'
					  AND schema_name != 'information_schema'
					ORDER BY schema_name
				`)
				if err == nil {
					w.schemas = nil
					for rows.Next() {
						var name string
						if rows.Scan(&name) == nil {
							w.schemas = append(w.schemas, name)
						}
					}
					rows.Close()
				}
				db.Close()
			}
		}
	}
}

func (w *BridgeWizard) loadTables() {
	w.tables = nil
	w.tableSelected = make(map[string]bool)

	tables, err := integration.GetAvailableTables(w.selectedPGService)
	if err == nil {
		w.tables = tables
	}
}

// Init implements tea.Model
func (w *BridgeWizard) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w *BridgeWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case BridgeResultMsg:
		w.result = msg.Link
		w.err = msg.Err
		w.step = BridgeStepComplete
		return w, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if w.step == BridgeStepComplete {
				return w, nil
			}
			w.cancelled = true
			return w, nil

		case "up", "k":
			w.moveSelection(-1)

		case "down", "j":
			w.moveSelection(1)

		case "enter":
			return w.handleEnter()

		case " ":
			// Toggle table selection
			if w.step == BridgeStepSelectTables && len(w.tables) > 0 {
				idx := 0
				for i, t := range w.tables {
					if w.tableSelected[t] {
						continue
					}
					idx = i
					break
				}
				if idx < len(w.tables) {
					table := w.tables[idx]
					w.tableSelected[table] = !w.tableSelected[table]
				}
			}
		}
	}

	// Update text input when on that step
	if w.step == BridgeStepEnterStoreName {
		var cmd tea.Cmd
		w.storeNameInput, cmd = w.storeNameInput.Update(msg)
		return w, cmd
	}

	return w, nil
}

func (w *BridgeWizard) moveSelection(delta int) {
	switch w.step {
	case BridgeStepSelectPGService:
		w.pgServiceIdx = clamp(w.pgServiceIdx+delta, 0, len(w.pgServices)-1)
	case BridgeStepSelectGeoServer:
		w.geoServerIdx = clamp(w.geoServerIdx+delta, 0, len(w.geoServers)-1)
	case BridgeStepSelectWorkspace:
		w.workspaceIdx = clamp(w.workspaceIdx+delta, 0, len(w.workspaces)-1)
	case BridgeStepSelectSchema:
		w.schemaIdx = clamp(w.schemaIdx+delta, 0, len(w.schemas)-1)
	}
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if max >= 0 && val > max {
		return max
	}
	return val
}

func (w *BridgeWizard) handleEnter() (tea.Model, tea.Cmd) {
	switch w.step {
	case BridgeStepSelectPGService:
		if len(w.pgServices) > 0 {
			w.selectedPGService = w.pgServices[w.pgServiceIdx]
			w.loadGeoServerConnections()
			w.step = BridgeStepSelectGeoServer
		}

	case BridgeStepSelectGeoServer:
		if len(w.geoServers) > 0 {
			w.selectedGeoServer = w.geoServers[w.geoServerIdx]
			w.loadWorkspaces()
			w.step = BridgeStepSelectWorkspace
		}

	case BridgeStepSelectWorkspace:
		if len(w.workspaces) > 0 {
			w.selectedWorkspace = w.workspaces[w.workspaceIdx]
			w.storeNameInput.SetValue(fmt.Sprintf("%s_store", w.selectedPGService))
			w.storeNameInput.Focus()
			w.step = BridgeStepEnterStoreName
		}

	case BridgeStepEnterStoreName:
		w.storeName = w.storeNameInput.Value()
		if w.storeName != "" {
			w.loadSchemas()
			w.step = BridgeStepSelectSchema
		}

	case BridgeStepSelectSchema:
		if len(w.schemas) > 0 {
			w.selectedSchema = w.schemas[w.schemaIdx]
			w.loadTables()
			w.step = BridgeStepSelectTables
		}

	case BridgeStepSelectTables:
		// Collect selected tables
		w.selectedTables = nil
		for table, selected := range w.tableSelected {
			if selected {
				w.selectedTables = append(w.selectedTables, table)
			}
		}
		w.publishLayers = len(w.selectedTables) > 0
		w.step = BridgeStepConfirm

	case BridgeStepConfirm:
		w.step = BridgeStepCreating
		return w, w.createBridge()

	case BridgeStepComplete:
		return w, nil
	}

	return w, nil
}

func (w *BridgeWizard) createBridge() tea.Cmd {
	return func() tea.Msg {
		opts := integration.BridgeOptions{
			PGServiceName:         w.selectedPGService,
			GeoServerConnectionID: w.selectedGeoServer,
			Workspace:             w.selectedWorkspace,
			StoreName:             w.storeName,
			Schema:                w.selectedSchema,
			Tables:                w.selectedTables,
			PublishLayers:         w.publishLayers,
		}

		link, err := integration.CreateBridge(w.config, w.clients, opts)
		return BridgeResultMsg{Link: link, Err: err}
	}
}

// View implements tea.Model
func (w *BridgeWizard) View() string {
	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Faint(true)

	s.WriteString(titleStyle.Render("PostgreSQL to GeoServer Bridge Wizard"))
	s.WriteString("\n\n")

	switch w.step {
	case BridgeStepSelectPGService:
		s.WriteString("Select PostgreSQL Service:\n\n")
		if len(w.pgServices) == 0 {
			s.WriteString("  (no services found - configure pg_service.conf first)\n")
		} else {
			for i, svc := range w.pgServices {
				if i == w.pgServiceIdx {
					s.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s\n", svc)))
				} else {
					s.WriteString(normalStyle.Render(fmt.Sprintf("    %s\n", svc)))
				}
			}
		}

	case BridgeStepSelectGeoServer:
		s.WriteString("Select GeoServer Connection:\n\n")
		if len(w.geoServers) == 0 {
			s.WriteString("  (no connections - add a GeoServer connection first)\n")
		} else {
			for i, gs := range w.geoServers {
				if i == w.geoServerIdx {
					s.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s\n", gs)))
				} else {
					s.WriteString(normalStyle.Render(fmt.Sprintf("    %s\n", gs)))
				}
			}
		}

	case BridgeStepSelectWorkspace:
		s.WriteString("Select Target Workspace:\n\n")
		if len(w.workspaces) == 0 {
			s.WriteString("  (no workspaces - create a workspace first)\n")
		} else {
			for i, ws := range w.workspaces {
				if i == w.workspaceIdx {
					s.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s\n", ws)))
				} else {
					s.WriteString(normalStyle.Render(fmt.Sprintf("    %s\n", ws)))
				}
			}
		}

	case BridgeStepEnterStoreName:
		s.WriteString("Enter a name for the PostGIS data store:\n\n")
		s.WriteString(w.storeNameInput.View())
		s.WriteString("\n\nPress Enter to continue")

	case BridgeStepSelectSchema:
		s.WriteString("Select PostgreSQL Schema:\n\n")
		for i, schema := range w.schemas {
			if i == w.schemaIdx {
				s.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s\n", schema)))
			} else {
				s.WriteString(normalStyle.Render(fmt.Sprintf("    %s\n", schema)))
			}
		}

	case BridgeStepSelectTables:
		s.WriteString("Select Tables to Publish (Space to toggle):\n\n")
		if len(w.tables) == 0 {
			s.WriteString("  (no spatial tables found in this schema)\n")
		} else {
			for _, table := range w.tables {
				checkbox := "[ ]"
				if w.tableSelected[table] {
					checkbox = "[x]"
				}
				s.WriteString(fmt.Sprintf("  %s %s\n", checkbox, table))
			}
		}
		s.WriteString("\nPress Enter to continue (skip table selection to create store only)")

	case BridgeStepConfirm:
		s.WriteString("Summary:\n\n")
		s.WriteString(fmt.Sprintf("  PostgreSQL Service: %s\n", w.selectedPGService))
		s.WriteString(fmt.Sprintf("  GeoServer:          %s\n", w.selectedGeoServer))
		s.WriteString(fmt.Sprintf("  Workspace:          %s\n", w.selectedWorkspace))
		s.WriteString(fmt.Sprintf("  Store Name:         %s\n", w.storeName))
		s.WriteString(fmt.Sprintf("  Schema:             %s\n", w.selectedSchema))
		if len(w.selectedTables) > 0 {
			s.WriteString(fmt.Sprintf("  Tables to Publish:  %s\n", strings.Join(w.selectedTables, ", ")))
		} else {
			s.WriteString("  Tables to Publish:  (none)\n")
		}
		s.WriteString("\nPress Enter to create the bridge, Esc to cancel")

	case BridgeStepCreating:
		s.WriteString("Creating PostGIS data store...")

	case BridgeStepComplete:
		if w.err != nil {
			s.WriteString(fmt.Sprintf("Error: %s\n\nPress Esc to close", w.err.Error()))
		} else {
			s.WriteString("Bridge created successfully!\n\n")
			s.WriteString(fmt.Sprintf("Store '%s' is now available in workspace '%s'\n", w.storeName, w.selectedWorkspace))
			s.WriteString("\nPress Esc to close")
		}
	}

	s.WriteString("\n\n")
	s.WriteString(lipgloss.NewStyle().Faint(true).Render("↑/↓: Navigate  Enter: Select  Esc: Cancel"))

	return s.String()
}

// IsCancelled returns true if the wizard was cancelled
func (w *BridgeWizard) IsCancelled() bool {
	return w.cancelled
}

// IsComplete returns true if the wizard is complete
func (w *BridgeWizard) IsComplete() bool {
	return w.step == BridgeStepComplete
}

// GetResult returns the result of the bridge creation
func (w *BridgeWizard) GetResult() (*integration.LinkedStore, error) {
	return w.result, w.err
}
