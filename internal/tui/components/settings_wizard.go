package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kartoza/kartoza-geoserver-client/internal/models"
	"github.com/kartoza/kartoza-geoserver-client/internal/tui/styles"
)

// SettingsWizardResult contains the result of the settings wizard
type SettingsWizardResult struct {
	Confirmed    bool
	Contact      *models.GeoServerContact
	ConnectionID string
}

// SettingsWizardKeyMap defines the key bindings
type SettingsWizardKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Tab     key.Binding
	Confirm key.Binding
	Cancel  key.Binding
}

// DefaultSettingsWizardKeyMap returns the default key bindings
func DefaultSettingsWizardKeyMap() SettingsWizardKeyMap {
	return SettingsWizardKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("↑", "previous field"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("↓", "next field"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// Field indices
const (
	fieldContactPerson = iota
	fieldContactPosition
	fieldContactOrganization
	fieldContactEmail
	fieldContactPhone
	fieldContactFax
	fieldAddress
	fieldAddressCity
	fieldAddressState
	fieldAddressPostCode
	fieldAddressCountry
	fieldOnlineResource
	fieldWelcome
	fieldCount
)

// SettingsWizard is a dialog for editing GeoServer contact settings
type SettingsWizard struct {
	keyMap       SettingsWizardKeyMap
	visible      bool
	animating    bool
	animProgress float64

	// Connection info
	connectionID   string
	connectionName string

	// Form fields
	inputs       []textinput.Model
	currentField int

	// Dimensions
	width  int
	height int

	// Current tab (0=Contact, 1=Address, 2=Service)
	currentTab int

	// Callbacks
	onConfirm func(SettingsWizardResult)
	onCancel  func()
}

// NewSettingsWizard creates a new settings wizard
func NewSettingsWizard(connectionID, connectionName string, contact *models.GeoServerContact) *SettingsWizard {
	inputs := make([]textinput.Model, fieldCount)

	// Initialize all inputs
	for i := 0; i < fieldCount; i++ {
		inputs[i] = textinput.New()
		inputs[i].CharLimit = 256
	}

	// Set placeholders and labels
	inputs[fieldContactPerson].Placeholder = "John Smith"
	inputs[fieldContactPosition].Placeholder = "GIS Administrator"
	inputs[fieldContactOrganization].Placeholder = "Kartoza (Pty) Ltd"
	inputs[fieldContactEmail].Placeholder = "info@example.com"
	inputs[fieldContactPhone].Placeholder = "+27 21 123 4567"
	inputs[fieldContactFax].Placeholder = "+27 21 123 4568"
	inputs[fieldAddress].Placeholder = "123 Main Street"
	inputs[fieldAddressCity].Placeholder = "Cape Town"
	inputs[fieldAddressState].Placeholder = "Western Cape"
	inputs[fieldAddressPostCode].Placeholder = "8001"
	inputs[fieldAddressCountry].Placeholder = "South Africa"
	inputs[fieldOnlineResource].Placeholder = "https://www.example.com"
	inputs[fieldWelcome].Placeholder = "Welcome to our GeoServer..."

	// Set widths
	for i := 0; i < fieldCount; i++ {
		inputs[i].Width = 40
	}

	// Populate with existing contact data
	if contact != nil {
		inputs[fieldContactPerson].SetValue(contact.ContactPerson)
		inputs[fieldContactPosition].SetValue(contact.ContactPosition)
		inputs[fieldContactOrganization].SetValue(contact.ContactOrganization)
		inputs[fieldContactEmail].SetValue(contact.ContactEmail)
		inputs[fieldContactPhone].SetValue(contact.ContactVoice)
		inputs[fieldContactFax].SetValue(contact.ContactFax)
		inputs[fieldAddress].SetValue(contact.Address)
		inputs[fieldAddressCity].SetValue(contact.AddressCity)
		inputs[fieldAddressState].SetValue(contact.AddressState)
		inputs[fieldAddressPostCode].SetValue(contact.AddressPostCode)
		inputs[fieldAddressCountry].SetValue(contact.AddressCountry)
		inputs[fieldOnlineResource].SetValue(contact.OnlineResource)
		inputs[fieldWelcome].SetValue(contact.Welcome)
	}

	// Focus first field
	inputs[0].Focus()

	return &SettingsWizard{
		keyMap:         DefaultSettingsWizardKeyMap(),
		visible:        true,
		animating:      true,
		animProgress:   0,
		connectionID:   connectionID,
		connectionName: connectionName,
		inputs:         inputs,
		currentField:   0,
		currentTab:     0,
	}
}

// SettingsWizardAnimationMsg is sent to animate the wizard
type SettingsWizardAnimationMsg struct{}

// SetCallbacks sets the confirmation and cancellation callbacks
func (w *SettingsWizard) SetCallbacks(onConfirm func(SettingsWizardResult), onCancel func()) {
	w.onConfirm = onConfirm
	w.onCancel = onCancel
}

// SetSize sets the wizard dimensions
func (w *SettingsWizard) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// IsVisible returns whether the wizard is visible
func (w *SettingsWizard) IsVisible() bool {
	return w.visible
}

// Init initializes the wizard
func (w *SettingsWizard) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		func() tea.Msg {
			return SettingsWizardAnimationMsg{}
		},
	)
}

// Update handles messages
func (w *SettingsWizard) Update(msg tea.Msg) (*SettingsWizard, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case SettingsWizardAnimationMsg:
		if w.animating {
			w.animProgress += 0.15
			if w.animProgress >= 1.0 {
				w.animProgress = 1.0
				w.animating = false
			} else {
				cmds = append(cmds, func() tea.Msg {
					return SettingsWizardAnimationMsg{}
				})
			}
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, w.keyMap.Cancel):
			w.close()
			if w.onCancel != nil {
				w.onCancel()
			}

		case key.Matches(msg, w.keyMap.Confirm):
			result := SettingsWizardResult{
				Confirmed:    true,
				ConnectionID: w.connectionID,
				Contact:      w.buildContact(),
			}
			w.close()
			if w.onConfirm != nil {
				w.onConfirm(result)
			}

		case key.Matches(msg, w.keyMap.Down), key.Matches(msg, w.keyMap.Tab):
			w.nextField()

		case key.Matches(msg, w.keyMap.Up):
			w.prevField()

		default:
			// Update the current input
			var cmd tea.Cmd
			w.inputs[w.currentField], cmd = w.inputs[w.currentField].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return w, tea.Batch(cmds...)
}

func (w *SettingsWizard) nextField() {
	w.inputs[w.currentField].Blur()
	w.currentField = (w.currentField + 1) % fieldCount
	w.inputs[w.currentField].Focus()
	w.updateTab()
}

func (w *SettingsWizard) prevField() {
	w.inputs[w.currentField].Blur()
	w.currentField = (w.currentField - 1 + fieldCount) % fieldCount
	w.inputs[w.currentField].Focus()
	w.updateTab()
}

func (w *SettingsWizard) updateTab() {
	switch {
	case w.currentField <= fieldContactFax:
		w.currentTab = 0
	case w.currentField <= fieldAddressCountry:
		w.currentTab = 1
	default:
		w.currentTab = 2
	}
}

func (w *SettingsWizard) buildContact() *models.GeoServerContact {
	return &models.GeoServerContact{
		ContactPerson:       w.inputs[fieldContactPerson].Value(),
		ContactPosition:     w.inputs[fieldContactPosition].Value(),
		ContactOrganization: w.inputs[fieldContactOrganization].Value(),
		ContactEmail:        w.inputs[fieldContactEmail].Value(),
		ContactVoice:        w.inputs[fieldContactPhone].Value(),
		ContactFax:          w.inputs[fieldContactFax].Value(),
		Address:             w.inputs[fieldAddress].Value(),
		AddressCity:         w.inputs[fieldAddressCity].Value(),
		AddressState:        w.inputs[fieldAddressState].Value(),
		AddressPostCode:     w.inputs[fieldAddressPostCode].Value(),
		AddressCountry:      w.inputs[fieldAddressCountry].Value(),
		OnlineResource:      w.inputs[fieldOnlineResource].Value(),
		Welcome:             w.inputs[fieldWelcome].Value(),
	}
}

func (w *SettingsWizard) close() {
	w.visible = false
}

// View renders the wizard
func (w *SettingsWizard) View() string {
	if !w.visible {
		return ""
	}

	var b strings.Builder

	// Title
	title := styles.DialogTitleStyle.Render("Service Metadata")
	b.WriteString(title)
	b.WriteString("\n")

	// Connection name
	b.WriteString(styles.HelpTextStyle.Render("Connection: "))
	b.WriteString(styles.PanelHeaderStyle.Render(w.connectionName))
	b.WriteString("\n\n")

	// Tabs
	tabs := []string{"Contact", "Address", "Service"}
	var tabLine strings.Builder
	for i, tab := range tabs {
		if i == w.currentTab {
			tabLine.WriteString(styles.ActiveItemStyle.Render(" " + tab + " "))
		} else {
			tabLine.WriteString(styles.ItemStyle.Render(" " + tab + " "))
		}
		tabLine.WriteString(" ")
	}
	b.WriteString(tabLine.String())
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 50))
	b.WriteString("\n\n")

	// Render fields based on current tab
	switch w.currentTab {
	case 0: // Contact
		w.renderField(&b, "Name:", fieldContactPerson)
		w.renderField(&b, "Position:", fieldContactPosition)
		w.renderField(&b, "Organization:", fieldContactOrganization)
		w.renderField(&b, "Email:", fieldContactEmail)
		w.renderField(&b, "Phone:", fieldContactPhone)
		w.renderField(&b, "Fax:", fieldContactFax)
	case 1: // Address
		w.renderField(&b, "Street:", fieldAddress)
		w.renderField(&b, "City:", fieldAddressCity)
		w.renderField(&b, "State:", fieldAddressState)
		w.renderField(&b, "Postal Code:", fieldAddressPostCode)
		w.renderField(&b, "Country:", fieldAddressCountry)
	case 2: // Service
		w.renderField(&b, "Website:", fieldOnlineResource)
		w.renderField(&b, "Welcome:", fieldWelcome)
	}

	b.WriteString("\n")
	b.WriteString(styles.HelpTextStyle.Render("Tab/↓ next  Shift+Tab/↑ prev  Ctrl+S save  Esc cancel"))

	// Create dialog box
	dialogWidth := 60
	dialog := styles.DialogStyle.Width(dialogWidth).Render(b.String())

	// Apply animation
	if w.animating {
		lines := strings.Split(dialog, "\n")
		visibleLines := int(float64(len(lines)) * w.animProgress)
		if visibleLines < 1 {
			visibleLines = 1
		}
		if visibleLines < len(lines) {
			dialog = strings.Join(lines[:visibleLines], "\n")
		}
	}

	return styles.Center(w.width, w.height, dialog)
}

func (w *SettingsWizard) renderField(b *strings.Builder, label string, fieldIndex int) {
	prefix := "  "
	if fieldIndex == w.currentField {
		prefix = "▸ "
	}

	labelStyle := styles.HelpTextStyle.Width(14)
	b.WriteString(prefix)
	b.WriteString(labelStyle.Render(label))
	b.WriteString(w.inputs[fieldIndex].View())
	b.WriteString("\n")
}
