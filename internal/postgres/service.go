package postgres

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

// ServiceEntry represents a PostgreSQL service configuration from pg_service.conf
type ServiceEntry struct {
	Name     string
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
	SSLMode  string
	Options  map[string]string
	Hidden   bool // True if the service is commented out in pg_service.conf
}

// ParsePGServiceFile parses the pg_service.conf file from standard locations
func ParsePGServiceFile() ([]ServiceEntry, error) {
	paths := getPGServicePaths()

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return parsePGServiceFileAt(path)
		}
	}

	return nil, fmt.Errorf("no pg_service.conf found in standard locations")
}

// getPGServicePaths returns possible pg_service.conf locations
func getPGServicePaths() []string {
	paths := []string{}

	// Check PGSERVICEFILE env var first
	if envPath := os.Getenv("PGSERVICEFILE"); envPath != "" {
		paths = append(paths, envPath)
	}

	// User's home directory
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".pg_service.conf"))
	}

	// System-wide locations
	paths = append(paths, "/etc/pg_service.conf")
	paths = append(paths, "/etc/postgresql-common/pg_service.conf")

	return paths
}

// parsePGServiceFileAt parses a pg_service.conf file at the given path
// It also detects hidden (commented-out) services prefixed with #[servicename]
func parsePGServiceFileAt(path string) ([]ServiceEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var services []ServiceEntry
	var current *ServiceEntry
	isHidden := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for hidden service header #[servicename]
		if strings.HasPrefix(line, "#[") && strings.HasSuffix(line, "]") {
			if current != nil {
				services = append(services, *current)
			}
			serviceName := strings.TrimSuffix(strings.TrimPrefix(line, "#["), "]")
			current = &ServiceEntry{
				Name:    serviceName,
				Options: make(map[string]string),
				Hidden:  true,
			}
			isHidden = true
			continue
		}

		// Check for normal service header [servicename]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if current != nil {
				services = append(services, *current)
			}
			serviceName := strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			current = &ServiceEntry{
				Name:    serviceName,
				Options: make(map[string]string),
				Hidden:  false,
			}
			isHidden = false
			continue
		}

		// Skip regular comments (not part of hidden service)
		if strings.HasPrefix(line, "#") && !isHidden {
			continue
		}

		// For hidden services, strip the # prefix from key=value lines
		if isHidden && strings.HasPrefix(line, "#") {
			line = strings.TrimPrefix(line, "#")
			line = strings.TrimSpace(line)
		}

		// Parse key=value pairs
		if current != nil && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				switch key {
				case "host":
					current.Host = value
				case "port":
					current.Port = value
				case "dbname":
					current.DBName = value
				case "user":
					current.User = value
				case "password":
					current.Password = value
				case "sslmode":
					current.SSLMode = value
				default:
					current.Options[key] = value
				}
			}
		}
	}

	// Don't forget the last service
	if current != nil {
		services = append(services, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

// ConnectionString returns a PostgreSQL connection string for the service
func (s *ServiceEntry) ConnectionString() string {
	parts := []string{}

	if s.Host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", s.Host))
	}
	if s.Port != "" {
		parts = append(parts, fmt.Sprintf("port=%s", s.Port))
	}
	if s.DBName != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", s.DBName))
	}
	if s.User != "" {
		parts = append(parts, fmt.Sprintf("user=%s", s.User))
	}
	if s.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", s.Password))
	}
	if s.SSLMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", s.SSLMode))
	} else {
		parts = append(parts, "sslmode=prefer")
	}

	for k, v := range s.Options {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(parts, " ")
}

// Connect creates a database connection to this service
func (s *ServiceEntry) Connect() (*sql.DB, error) {
	return sql.Open("postgres", s.ConnectionString())
}

// TestConnection tests if a connection can be established
func (s *ServiceEntry) TestConnection() error {
	db, err := s.Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Ping()
}

// GetServiceByName finds a service by name from the list
func GetServiceByName(services []ServiceEntry, name string) (*ServiceEntry, error) {
	for _, s := range services {
		if s.Name == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("service '%s' not found", name)
}

// GetPGServiceFilePath returns the path to the user's pg_service.conf file
func GetPGServiceFilePath() string {
	// Check PGSERVICEFILE env var first
	if envPath := os.Getenv("PGSERVICEFILE"); envPath != "" {
		return envPath
	}
	// Default to user's home directory
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".pg_service.conf")
	}
	return ""
}

// PGServiceFileExists checks if the pg_service.conf file exists
func PGServiceFileExists() bool {
	paths := getPGServicePaths()
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// SaveServiceEntry saves or updates a service entry in pg_service.conf
func SaveServiceEntry(entry ServiceEntry) error {
	path := GetPGServiceFilePath()
	if path == "" {
		return fmt.Errorf("could not determine pg_service.conf path")
	}

	// Read existing services
	var services []ServiceEntry
	if _, err := os.Stat(path); err == nil {
		services, _ = parsePGServiceFileAt(path)
	}

	// Update or add the entry
	found := false
	for i, s := range services {
		if s.Name == entry.Name {
			services[i] = entry
			found = true
			break
		}
	}
	if !found {
		services = append(services, entry)
	}

	// Write back to file
	return writePGServiceFile(path, services)
}

// DeleteServiceEntry removes a service entry from pg_service.conf
func DeleteServiceEntry(name string) error {
	path := GetPGServiceFilePath()
	if path == "" {
		return fmt.Errorf("could not determine pg_service.conf path")
	}

	services, err := parsePGServiceFileAt(path)
	if err != nil {
		return err
	}

	// Filter out the entry to delete
	var filtered []ServiceEntry
	for _, s := range services {
		if s.Name != name {
			filtered = append(filtered, s)
		}
	}

	return writePGServiceFile(path, filtered)
}

// writePGServiceFile writes service entries to a pg_service.conf file
// Hidden services are written with # prefix on all lines
func writePGServiceFile(path string, services []ServiceEntry) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var content strings.Builder
	content.WriteString("# PostgreSQL service configuration\n")
	content.WriteString("# Generated by Kartoza CloudBench\n\n")

	for _, s := range services {
		prefix := ""
		if s.Hidden {
			prefix = "#"
		}
		content.WriteString(fmt.Sprintf("%s[%s]\n", prefix, s.Name))
		if s.Host != "" {
			content.WriteString(fmt.Sprintf("%shost=%s\n", prefix, s.Host))
		}
		if s.Port != "" {
			content.WriteString(fmt.Sprintf("%sport=%s\n", prefix, s.Port))
		}
		if s.DBName != "" {
			content.WriteString(fmt.Sprintf("%sdbname=%s\n", prefix, s.DBName))
		}
		if s.User != "" {
			content.WriteString(fmt.Sprintf("%suser=%s\n", prefix, s.User))
		}
		if s.Password != "" {
			content.WriteString(fmt.Sprintf("%spassword=%s\n", prefix, s.Password))
		}
		if s.SSLMode != "" {
			content.WriteString(fmt.Sprintf("%ssslmode=%s\n", prefix, s.SSLMode))
		}
		for k, v := range s.Options {
			content.WriteString(fmt.Sprintf("%s%s=%s\n", prefix, k, v))
		}
		content.WriteString("\n")
	}

	return os.WriteFile(path, []byte(content.String()), 0600)
}

// SetServiceHidden sets the hidden state for a service
func SetServiceHidden(name string, hidden bool) error {
	path := GetPGServiceFilePath()
	if path == "" {
		return fmt.Errorf("could not determine pg_service.conf path")
	}

	services, err := parsePGServiceFileAt(path)
	if err != nil {
		return err
	}

	found := false
	for i, s := range services {
		if s.Name == name {
			services[i].Hidden = hidden
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("service '%s' not found", name)
	}

	return writePGServiceFile(path, services)
}

// ServerStats represents statistics about a PostgreSQL server
type ServerStats struct {
	// Server info
	Version         string `json:"version"`
	ServerStartTime string `json:"server_start_time"`
	Uptime          string `json:"uptime"`
	Host            string `json:"host"`
	Port            string `json:"port"`

	// Database info
	DatabaseName string `json:"database_name"`
	DatabaseSize string `json:"database_size"`
	DatabaseOID  int64  `json:"database_oid"`

	// Connection stats
	MaxConnections    int `json:"max_connections"`
	CurrentConns      int `json:"current_connections"`
	ActiveConns       int `json:"active_connections"`
	IdleConns         int `json:"idle_connections"`
	IdleInTxConns     int `json:"idle_in_transaction_connections"`
	WaitingConns      int `json:"waiting_connections"`
	ConnectionPercent int `json:"connection_percent"`

	// Database stats
	NumBackends      int    `json:"num_backends"`
	XactCommit       int64  `json:"xact_commit"`
	XactRollback     int64  `json:"xact_rollback"`
	BlksRead         int64  `json:"blks_read"`
	BlksHit          int64  `json:"blks_hit"`
	TupReturned      int64  `json:"tup_returned"`
	TupFetched       int64  `json:"tup_fetched"`
	TupInserted      int64  `json:"tup_inserted"`
	TupUpdated       int64  `json:"tup_updated"`
	TupDeleted       int64  `json:"tup_deleted"`
	CacheHitRatio    string `json:"cache_hit_ratio"`
	DeadTuples       int64  `json:"dead_tuples"`
	LiveTuples       int64  `json:"live_tuples"`
	TableCount       int    `json:"table_count"`
	IndexCount       int    `json:"index_count"`
	ViewCount        int    `json:"view_count"`
	FunctionCount    int    `json:"function_count"`
	SchemaCount      int    `json:"schema_count"`

	// Replication
	IsInRecovery bool   `json:"is_in_recovery"`
	ReplayLag    string `json:"replay_lag,omitempty"`

	// Extensions
	InstalledExtensions []string `json:"installed_extensions"`

	// PostGIS specific
	HasPostGIS      bool   `json:"has_postgis"`
	PostGISVersion  string `json:"postgis_version,omitempty"`
	GeometryColumns int    `json:"geometry_columns,omitempty"`
	RasterColumns   int    `json:"raster_columns,omitempty"`
}

// GetServerStats retrieves comprehensive statistics about the PostgreSQL server
func (s *ServiceEntry) GetServerStats() (*ServerStats, error) {
	db, err := s.Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	stats := &ServerStats{
		Host:         s.Host,
		Port:         s.Port,
		DatabaseName: s.DBName,
	}

	// Get version
	var version string
	if err := db.QueryRow("SELECT version()").Scan(&version); err == nil {
		stats.Version = version
	}

	// Get server start time and uptime
	var startTime, uptime string
	if err := db.QueryRow(`
		SELECT
			pg_postmaster_start_time()::text,
			age(now(), pg_postmaster_start_time())::text
	`).Scan(&startTime, &uptime); err == nil {
		stats.ServerStartTime = startTime
		stats.Uptime = uptime
	}

	// Get max connections
	var maxConns int
	if err := db.QueryRow("SHOW max_connections").Scan(&maxConns); err == nil {
		stats.MaxConnections = maxConns
	}

	// Get current connection stats
	if err := db.QueryRow(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE state = 'active') as active,
			COUNT(*) FILTER (WHERE state = 'idle') as idle,
			COUNT(*) FILTER (WHERE state = 'idle in transaction') as idle_in_tx,
			COUNT(*) FILTER (WHERE wait_event_type IS NOT NULL) as waiting
		FROM pg_stat_activity
		WHERE backend_type = 'client backend'
	`).Scan(&stats.CurrentConns, &stats.ActiveConns, &stats.IdleConns, &stats.IdleInTxConns, &stats.WaitingConns); err == nil {
		if stats.MaxConnections > 0 {
			stats.ConnectionPercent = (stats.CurrentConns * 100) / stats.MaxConnections
		}
	}

	// Get database stats
	if err := db.QueryRow(`
		SELECT
			datid,
			numbackends,
			xact_commit,
			xact_rollback,
			blks_read,
			blks_hit,
			tup_returned,
			tup_fetched,
			tup_inserted,
			tup_updated,
			tup_deleted
		FROM pg_stat_database
		WHERE datname = current_database()
	`).Scan(
		&stats.DatabaseOID, &stats.NumBackends,
		&stats.XactCommit, &stats.XactRollback,
		&stats.BlksRead, &stats.BlksHit,
		&stats.TupReturned, &stats.TupFetched,
		&stats.TupInserted, &stats.TupUpdated, &stats.TupDeleted,
	); err == nil {
		// Calculate cache hit ratio
		totalBlks := stats.BlksRead + stats.BlksHit
		if totalBlks > 0 {
			ratio := float64(stats.BlksHit) * 100 / float64(totalBlks)
			stats.CacheHitRatio = fmt.Sprintf("%.2f%%", ratio)
		} else {
			stats.CacheHitRatio = "N/A"
		}
	}

	// Get database size
	var dbSize string
	if err := db.QueryRow("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize); err == nil {
		stats.DatabaseSize = dbSize
	}

	// Get tuple stats
	if err := db.QueryRow(`
		SELECT
			COALESCE(SUM(n_live_tup), 0),
			COALESCE(SUM(n_dead_tup), 0)
		FROM pg_stat_user_tables
	`).Scan(&stats.LiveTuples, &stats.DeadTuples); err != nil {
		// Ignore error
	}

	// Get object counts
	db.QueryRow(`SELECT COUNT(*) FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema')`).Scan(&stats.TableCount)
	db.QueryRow(`SELECT COUNT(*) FROM pg_indexes WHERE schemaname NOT IN ('pg_catalog', 'information_schema')`).Scan(&stats.IndexCount)
	db.QueryRow(`SELECT COUNT(*) FROM pg_views WHERE schemaname NOT IN ('pg_catalog', 'information_schema')`).Scan(&stats.ViewCount)
	db.QueryRow(`SELECT COUNT(*) FROM pg_proc WHERE prokind = 'f' AND pronamespace IN (SELECT oid FROM pg_namespace WHERE nspname NOT IN ('pg_catalog', 'information_schema'))`).Scan(&stats.FunctionCount)
	db.QueryRow(`SELECT COUNT(*) FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname != 'information_schema'`).Scan(&stats.SchemaCount)

	// Check if in recovery mode (standby)
	db.QueryRow("SELECT pg_is_in_recovery()").Scan(&stats.IsInRecovery)

	// Get replay lag if in recovery
	if stats.IsInRecovery {
		var replayLag sql.NullString
		db.QueryRow("SELECT COALESCE(pg_last_wal_receive_lsn() - pg_last_wal_replay_lsn(), 0)::text").Scan(&replayLag)
		if replayLag.Valid {
			stats.ReplayLag = replayLag.String
		}
	}

	// Get installed extensions
	rows, err := db.Query("SELECT extname FROM pg_extension ORDER BY extname")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var extName string
			if rows.Scan(&extName) == nil {
				stats.InstalledExtensions = append(stats.InstalledExtensions, extName)
				if extName == "postgis" {
					stats.HasPostGIS = true
				}
			}
		}
	}

	// If PostGIS is installed, get version and geometry/raster column counts
	if stats.HasPostGIS {
		db.QueryRow("SELECT PostGIS_Full_Version()").Scan(&stats.PostGISVersion)
		db.QueryRow("SELECT COUNT(*) FROM geometry_columns").Scan(&stats.GeometryColumns)
		db.QueryRow("SELECT COUNT(*) FROM raster_columns").Scan(&stats.RasterColumns)
	}

	return stats, nil
}
