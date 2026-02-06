package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-geoserver-client/internal/models"
	"github.com/kartoza/kartoza-geoserver-client/internal/tui/styles"
)

// InfoDialogAnimationMsg is sent to update animation state
type InfoDialogAnimationMsg struct {
	ID string
}

// InfoDialog displays detailed information about a resource
type InfoDialog struct {
	id          string
	title       string
	icon        string
	details     []InfoItem
	width       int
	height      int
	visible     bool

	// Harmonica physics for smooth animations
	spring        harmonica.Spring
	animScale     float64
	animVelocity  float64
	animOpacity   float64
	targetScale   float64
	targetOpacity float64
	animating     bool
	closing       bool
}

// InfoItem represents a single piece of information to display
type InfoItem struct {
	Label string
	Value string
	Style lipgloss.Style
}

// NewInfoDialog creates a new info dialog
func NewInfoDialog(title, icon string, details []InfoItem) *InfoDialog {
	return &InfoDialog{
		id:            title,
		title:         title,
		icon:          icon,
		details:       details,
		visible:       true,
		spring:        harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.5),
		animScale:     0.0,
		animOpacity:   0.0,
		targetScale:   1.0,
		targetOpacity: 1.0,
		animating:     true,
	}
}

// NewFileInfoDialog creates an info dialog for a local file
func NewFileInfoDialog(file *models.LocalFile) *InfoDialog {
	details := []InfoItem{
		{Label: "Name", Value: file.Name},
		{Label: "Path", Value: file.Path},
		{Label: "Type", Value: file.Type.String()},
		{Label: "Size", Value: formatFileSize(file.Size)},
	}

	if file.IsDir {
		details[2].Value = "Directory"
	}

	if file.Type.CanUpload() {
		details = append(details, InfoItem{
			Label: "Uploadable",
			Value: "Yes",
			Style: styles.SuccessStyle,
		})
	}

	return NewInfoDialog("File Information", file.Type.Icon(), details)
}

// NewTreeNodeInfoDialog creates an info dialog for a tree node
func NewTreeNodeInfoDialog(node *models.TreeNode) *InfoDialog {
	details := []InfoItem{
		{Label: "Name", Value: node.Name},
		{Label: "Type", Value: node.Type.String()},
		{Label: "Path", Value: node.Path()},
	}

	if node.Workspace != "" {
		details = append(details, InfoItem{Label: "Workspace", Value: node.Workspace})
	}

	if node.StoreName != "" {
		details = append(details, InfoItem{Label: "Store", Value: node.StoreName})
	}

	if node.HasError {
		details = append(details, InfoItem{
			Label: "Error",
			Value: node.ErrorMsg,
			Style: styles.ErrorStyle,
		})
	}

	status := "Ready"
	statusStyle := styles.SuccessStyle
	if node.IsLoading {
		status = "Loading..."
		statusStyle = styles.LoadingStyle
	} else if !node.IsLoaded && len(node.Children) == 0 {
		status = "Not loaded"
		statusStyle = styles.MutedStyle
	}
	details = append(details, InfoItem{
		Label: "Status",
		Value: status,
		Style: statusStyle,
	})

	if len(node.Children) > 0 {
		details = append(details, InfoItem{
			Label: "Children",
			Value: fmt.Sprintf("%d items", len(node.Children)),
		})
	}

	return NewInfoDialog("Resource Information", node.Type.Icon(), details)
}

// formatFileSize formats a file size in bytes to a human-readable string
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}

// SetSize sets the dialog size
func (d *InfoDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

// IsVisible returns whether the dialog is visible
func (d *InfoDialog) IsVisible() bool {
	return d.visible
}

// StartCloseAnimation starts the closing animation
func (d *InfoDialog) StartCloseAnimation() tea.Cmd {
	d.closing = true
	d.targetScale = 0.0
	d.targetOpacity = 0.0
	d.animating = true
	return d.animateCmd()
}

// animateCmd returns a command to continue the animation
func (d *InfoDialog) animateCmd() tea.Cmd {
	return tea.Tick(time.Second/60, func(t time.Time) tea.Msg {
		return InfoDialogAnimationMsg{ID: d.id}
	})
}

// Init initializes the dialog and starts the opening animation
func (d *InfoDialog) Init() tea.Cmd {
	return d.animateCmd()
}

// Update handles messages
func (d *InfoDialog) Update(msg tea.Msg) (*InfoDialog, tea.Cmd) {
	switch msg := msg.(type) {
	case InfoDialogAnimationMsg:
		if msg.ID != d.id {
			return d, nil
		}
		return d.updateAnimation()

	case tea.KeyMsg:
		if !d.visible || d.animating {
			return d, nil
		}

		// Any key closes the dialog
		switch msg.String() {
		case "esc", "enter", "i", "q", " ":
			return d, d.StartCloseAnimation()
		}
	}

	return d, nil
}

// updateAnimation updates the harmonica physics animation
func (d *InfoDialog) updateAnimation() (*InfoDialog, tea.Cmd) {
	if !d.animating {
		return d, nil
	}

	// Update scale using spring physics
	d.animScale, d.animVelocity = d.spring.Update(d.animScale, d.animVelocity, d.targetScale)

	// Update opacity
	opacityStep := 0.1
	if d.closing {
		d.animOpacity -= opacityStep
		if d.animOpacity < 0 {
			d.animOpacity = 0
		}
	} else {
		d.animOpacity += opacityStep
		if d.animOpacity > 1 {
			d.animOpacity = 1
		}
	}

	// Check if animation is complete
	scaleClose := abs(d.animScale-d.targetScale) < 0.01 && abs(d.animVelocity) < 0.01
	opacityClose := d.closing && d.animOpacity <= 0.01 || !d.closing && d.animOpacity >= 0.99

	if scaleClose && opacityClose {
		d.animating = false
		d.animScale = d.targetScale
		d.animOpacity = d.targetOpacity

		if d.closing {
			d.visible = false
			return d, nil
		}
	}

	return d, d.animateCmd()
}

// View renders the dialog with animation
func (d *InfoDialog) View() string {
	if !d.visible {
		return ""
	}

	var b strings.Builder

	// Title with icon
	titleText := fmt.Sprintf("%s  %s", d.icon, d.title)
	title := styles.DialogTitleStyle.Render(titleText)
	b.WriteString(title)
	b.WriteString("\n\n")

	// Separator
	separator := styles.TreeBranchStyle.Render(strings.Repeat("─", 40))
	b.WriteString(separator)
	b.WriteString("\n\n")

	// Details
	maxLabelLen := 0
	for _, item := range d.details {
		if len(item.Label) > maxLabelLen {
			maxLabelLen = len(item.Label)
		}
	}

	for _, item := range d.details {
		label := styles.HelpKeyStyle.Width(maxLabelLen + 2).Render(item.Label + ":")

		valueStyle := item.Style
		if item.Style.Value() == "" {
			valueStyle = styles.ItemStyle
		}

		// Wrap long values
		maxValueWidth := 40
		value := item.Value
		if len(value) > maxValueWidth {
			value = wrapText(value, maxValueWidth)
		}

		renderedValue := valueStyle.Render(value)
		b.WriteString(fmt.Sprintf("  %s %s\n", label, renderedValue))
	}

	b.WriteString("\n")

	// Help text
	b.WriteString(styles.HelpTextStyle.Render("Press any key to close"))

	dialogWidth := d.width/2 + 10
	if dialogWidth < 50 {
		dialogWidth = 50
	}
	if dialogWidth > 80 {
		dialogWidth = 80
	}

	// Apply animation scale
	scaledWidth := int(float64(dialogWidth) * d.animScale)
	if scaledWidth < 10 {
		scaledWidth = 10
	}

	dialogStyle := styles.DialogStyle.Width(scaledWidth)

	// Add bounce effect
	marginOffset := int((1.0 - d.animScale) * 5)
	dialogStyle = dialogStyle.MarginTop(marginOffset).MarginBottom(marginOffset)

	dialog := dialogStyle.Render(b.String())

	// Apply opacity effect
	if d.animOpacity < 1.0 && d.animOpacity > 0.5 {
		dialog = lipgloss.NewStyle().Render(dialog)
	} else if d.animOpacity <= 0.5 {
		fadedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		dialog = fadedStyle.Render(dialog)
	}

	return styles.Center(d.width, d.height, dialog)
}

// wrapText wraps text to a maximum width
func wrapText(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxWidth {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				result.WriteString(currentLine + "\n")
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		result.WriteString(currentLine)
	}

	return result.String()
}
