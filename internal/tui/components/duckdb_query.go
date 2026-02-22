package components

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kartoza/kartoza-cloudbench/internal/cloudnative"
)

// DuckDBQueryMode represents the current view mode
type DuckDBQueryMode int

const (
	DuckDBQueryModeSQL DuckDBQueryMode = iota
	DuckDBQueryModeResults
	DuckDBQueryModeSchema
	DuckDBQueryModeExecuting
)

// DuckDBQuery is a component for querying Parquet files with DuckDB
type DuckDBQuery struct {
	width  int
	height int

	mode       DuckDBQueryMode
	focusIndex int

	// File info
	filePath    string
	fileName    string
	tableInfo   *cloudnative.DuckDBTableInfo

	// SQL input
	sqlInput    textarea.Model

	// Results display
	resultView  viewport.Model
	queryResult *cloudnative.DuckDBQueryResult

	// Query options
	limit  int
	offset int

	// State
	error     string
	cancelled bool
	showHelp  bool
}

// NewDuckDBQuery creates a new DuckDB query component
func NewDuckDBQuery(filePath, fileName string) *DuckDBQuery {
	sqlInput := textarea.New()
	sqlInput.Placeholder = "Enter SQL query... Use 'data' as the table name\nExample: SELECT * FROM data LIMIT 10"
	sqlInput.SetHeight(5)
	sqlInput.Focus()

	resultView := viewport.New(80, 20)

	return &DuckDBQuery{
		filePath:   filePath,
		fileName:   fileName,
		mode:       DuckDBQueryModeSQL,
		focusIndex: 0,
		sqlInput:   sqlInput,
		resultView: resultView,
		limit:      100,
		offset:     0,
	}
}

// Init implements tea.Model
func (d *DuckDBQuery) Init() tea.Cmd {
	return d.loadTableInfo()
}

// DuckDBTableInfoLoaded is a message sent when table info is loaded
type DuckDBTableInfoLoaded struct {
	Info  *cloudnative.DuckDBTableInfo
	Error error
}

// DuckDBQueryResultMsg is a message sent when a query completes
type DuckDBQueryResultMsg struct {
	Result *cloudnative.DuckDBQueryResult
	Error  error
}

func (d *DuckDBQuery) loadTableInfo() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		info, err := cloudnative.GetParquetTableInfo(ctx, d.filePath)
		return DuckDBTableInfoLoaded{Info: info, Error: err}
	}
}

func (d *DuckDBQuery) executeQuery() tea.Cmd {
	sql := d.sqlInput.Value()
	if strings.TrimSpace(sql) == "" {
		return nil
	}

	// Validate SQL
	if err := cloudnative.ValidateSQL(sql); err != nil {
		d.error = err.Error()
		return nil
	}

	// Replace 'data' with actual file path
	sql = strings.ReplaceAll(sql, "'data'", "'"+d.filePath+"'")
	sql = strings.ReplaceAll(sql, "\"data\"", "'"+d.filePath+"'")
	sql = strings.ReplaceAll(sql, " data ", " '"+d.filePath+"' ")
	sql = strings.ReplaceAll(sql, " data;", " '"+d.filePath+"';")
	sql = strings.ReplaceAll(sql, "(data)", "('"+d.filePath+"')")
	sql = strings.ReplaceAll(sql, "FROM data", "FROM '"+d.filePath+"'")
	sql = strings.ReplaceAll(sql, "from data", "from '"+d.filePath+"'")

	d.mode = DuckDBQueryModeExecuting
	d.error = ""

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		opts := cloudnative.DuckDBQueryOptions{
			SQL:    sql,
			Limit:  d.limit,
			Offset: d.offset,
		}

		result, err := cloudnative.QueryParquetFile(ctx, d.filePath, opts)
		return DuckDBQueryResultMsg{Result: result, Error: err}
	}
}

// Update implements tea.Model
func (d *DuckDBQuery) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.sqlInput.SetWidth(msg.Width - 4)
		d.resultView.Width = msg.Width - 4
		d.resultView.Height = msg.Height - 15

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if d.mode == DuckDBQueryModeResults {
				d.mode = DuckDBQueryModeSQL
				d.sqlInput.Focus()
			} else {
				d.cancelled = true
			}
			return d, nil

		case "ctrl+e", "f5":
			// Execute query
			if d.mode == DuckDBQueryModeSQL {
				return d, d.executeQuery()
			}

		case "tab":
			// Switch between modes
			switch d.mode {
			case DuckDBQueryModeSQL:
				d.mode = DuckDBQueryModeResults
				d.sqlInput.Blur()
			case DuckDBQueryModeResults:
				d.mode = DuckDBQueryModeSchema
			case DuckDBQueryModeSchema:
				d.mode = DuckDBQueryModeSQL
				d.sqlInput.Focus()
			}
			return d, nil

		case "?":
			d.showHelp = !d.showHelp
			return d, nil
		}

	case DuckDBTableInfoLoaded:
		if msg.Error != nil {
			d.error = msg.Error.Error()
		} else {
			d.tableInfo = msg.Info
			// Set default SQL
			if len(msg.Info.Columns) > 0 {
				d.sqlInput.SetValue("SELECT * FROM data LIMIT 10")
			}
		}

	case DuckDBQueryResultMsg:
		d.mode = DuckDBQueryModeResults
		if msg.Error != nil {
			d.error = msg.Error.Error()
		} else {
			d.queryResult = msg.Result
			d.error = ""
			// Update result view content
			d.updateResultView()
		}
	}

	// Update sub-components
	if d.mode == DuckDBQueryModeSQL {
		var cmd tea.Cmd
		d.sqlInput, cmd = d.sqlInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if d.mode == DuckDBQueryModeResults {
		var cmd tea.Cmd
		d.resultView, cmd = d.resultView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return d, tea.Batch(cmds...)
}

func (d *DuckDBQuery) updateResultView() {
	if d.queryResult == nil {
		d.resultView.SetContent("No results")
		return
	}

	var sb strings.Builder

	// Header
	if len(d.queryResult.Columns) > 0 {
		for i, col := range d.queryResult.Columns {
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(fmt.Sprintf("%-20s", truncate(col, 20)))
		}
		sb.WriteString("\n")
		sb.WriteString(strings.Repeat("-", len(d.queryResult.Columns)*23))
		sb.WriteString("\n")
	}

	// Rows
	for _, row := range d.queryResult.Rows {
		for i, col := range d.queryResult.Columns {
			if i > 0 {
				sb.WriteString(" | ")
			}
			val := fmt.Sprintf("%v", row[col])
			sb.WriteString(fmt.Sprintf("%-20s", truncate(val, 20)))
		}
		sb.WriteString("\n")
	}

	// Footer
	sb.WriteString(fmt.Sprintf("\n%d rows returned", d.queryResult.RowCount))
	if d.queryResult.HasMore {
		sb.WriteString(" (more available)")
	}

	d.resultView.SetContent(sb.String())
}

// View implements tea.Model
func (d *DuckDBQuery) View() string {
	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	b.WriteString(headerStyle.Render(fmt.Sprintf("DuckDB Query: %s", d.fileName)))
	b.WriteString("\n\n")

	// Table info
	if d.tableInfo != nil {
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		b.WriteString(infoStyle.Render(fmt.Sprintf("Rows: %d | Columns: %d", d.tableInfo.RowCount, len(d.tableInfo.Columns))))
		if d.tableInfo.GeometryColumn != "" {
			b.WriteString(infoStyle.Render(fmt.Sprintf(" | Geometry: %s", d.tableInfo.GeometryColumn)))
		}
		b.WriteString("\n\n")
	}

	// Mode tabs
	tabs := []string{"SQL", "Results", "Schema"}
	tabStyle := lipgloss.NewStyle().Padding(0, 2)
	activeTabStyle := tabStyle.Copy().
		Bold(true).
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))

	for i, tab := range tabs {
		if i == int(d.mode) || (d.mode == DuckDBQueryModeExecuting && i == 1) {
			b.WriteString(activeTabStyle.Render(tab))
		} else {
			b.WriteString(tabStyle.Render(tab))
		}
	}
	b.WriteString("\n\n")

	// Content based on mode
	switch d.mode {
	case DuckDBQueryModeSQL:
		b.WriteString(d.sqlInput.View())
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Press Ctrl+E or F5 to execute | Tab to switch views | Esc to close"))

	case DuckDBQueryModeResults, DuckDBQueryModeExecuting:
		if d.mode == DuckDBQueryModeExecuting {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("Executing query..."))
		} else {
			b.WriteString(d.resultView.View())
		}

	case DuckDBQueryModeSchema:
		if d.tableInfo != nil {
			schemaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
			for _, col := range d.tableInfo.Columns {
				geomIndicator := ""
				if col.Name == d.tableInfo.GeometryColumn {
					geomIndicator = " [geometry]"
				}
				b.WriteString(fmt.Sprintf("  %s %s%s\n", schemaStyle.Render(col.Name), col.Type, geomIndicator))
			}
		}
	}

	// Error display
	if d.error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("Error: " + d.error))
	}

	// Help
	if d.showHelp {
		b.WriteString("\n\n")
		helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		b.WriteString(helpStyle.Render(`
Help:
  Ctrl+E / F5  Execute query
  Tab          Switch between SQL/Results/Schema
  Esc          Close dialog (from SQL mode)
  ?            Toggle help

SQL Tips:
  - Use 'data' as the table name
  - Example: SELECT * FROM data WHERE column = 'value'
  - DuckDB spatial functions: ST_X(), ST_Y(), ST_AsText(), etc.
`))
	}

	return b.String()
}

// IsCancelled returns true if the user cancelled
func (d *DuckDBQuery) IsCancelled() bool {
	return d.cancelled
}

// GetResult returns the current query result
func (d *DuckDBQuery) GetResult() *cloudnative.DuckDBQueryResult {
	return d.queryResult
}

// truncate truncates a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
