package sync

import (
	"fmt"
	"strings"

	"github.com/kartoza/kartoza-geoserver-client/internal/api"
	"github.com/kartoza/kartoza-geoserver-client/internal/config"
)

// Executor handles the actual sync operations
type Executor struct {
	task         *Task
	sourceClient *api.Client
	destClient   *api.Client
	options      config.SyncOptions
	stopChan     chan struct{}
}

// Execute runs the sync operation
func (e *Executor) Execute() {
	e.task.AddLog("Analyzing source server...")

	if e.options.Workspaces {
		workspaces, err := e.sourceClient.GetWorkspaces()
		if err != nil {
			e.task.SetError(fmt.Sprintf("Failed to get workspaces: %v", err))
			return
		}

		for _, ws := range workspaces {
			if e.isStopped() {
				return
			}

			// Check workspace filter
			if len(e.options.WorkspaceFilter) > 0 {
				if !e.matchesFilter(ws.Name) {
					continue
				}
			}

			e.syncWorkspace(ws.Name)
		}
	}

	// Sync global styles
	if e.options.Styles {
		e.syncGlobalStyles()
	}

	e.task.AddLog("Sync completed!")
}

func (e *Executor) isStopped() bool {
	select {
	case <-e.stopChan:
		return true
	default:
		return false
	}
}

func (e *Executor) matchesFilter(name string) bool {
	for _, f := range e.options.WorkspaceFilter {
		if f == name {
			return true
		}
	}
	return false
}

func (e *Executor) syncWorkspace(name string) {
	e.task.IncrementTotal()
	e.task.SetCurrentItem(fmt.Sprintf("Workspace: %s", name))
	e.task.AddLog(fmt.Sprintf("Syncing workspace: %s", name))

	// Try to create workspace on destination
	err := e.destClient.CreateWorkspace(name)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "409") {
			e.task.IncrementSkipped()
			e.task.AddLog(fmt.Sprintf("Workspace %s already exists, skipping", name))
		} else {
			e.task.IncrementFailed()
			e.task.AddLog(fmt.Sprintf("Failed to create workspace %s: %v", name, err))
		}
	} else {
		e.task.IncrementDone()
		e.task.AddLog(fmt.Sprintf("Created workspace: %s", name))
	}
	e.task.UpdateProgress()

	// Sync styles for this workspace
	if e.options.Styles {
		e.syncWorkspaceStyles(name)
	}

	// Sync data stores
	if e.options.DataStores {
		e.syncDataStores(name)
	}

	// Sync coverage stores
	if e.options.CoverageStores {
		e.syncCoverageStores(name)
	}

	// Sync layer groups
	if e.options.LayerGroups {
		e.syncLayerGroups(name)
	}
}

func (e *Executor) syncWorkspaceStyles(workspace string) {
	styles, err := e.sourceClient.GetStyles(workspace)
	if err != nil {
		e.task.AddLog(fmt.Sprintf("Failed to get styles for %s: %v", workspace, err))
		return
	}

	for _, style := range styles {
		if e.isStopped() {
			return
		}

		e.task.IncrementTotal()
		e.task.SetCurrentItem(fmt.Sprintf("Style: %s:%s", workspace, style.Name))

		// Get style content from source
		sld, err := e.sourceClient.GetStyleSLD(workspace, style.Name)
		if err != nil {
			e.task.IncrementFailed()
			e.task.AddLog(fmt.Sprintf("Failed to get style %s: %v", style.Name, err))
			continue
		}

		// Create on destination
		err = e.destClient.CreateOrUpdateStyle(workspace, style.Name, sld)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				e.task.IncrementSkipped()
			} else {
				e.task.IncrementFailed()
				e.task.AddLog(fmt.Sprintf("Failed to create style %s: %v", style.Name, err))
			}
		} else {
			e.task.IncrementDone()
		}
		e.task.UpdateProgress()
	}
}

func (e *Executor) syncGlobalStyles() {
	styles, err := e.sourceClient.GetStyles("")
	if err != nil {
		e.task.AddLog(fmt.Sprintf("Failed to get global styles: %v", err))
		return
	}

	for _, style := range styles {
		if e.isStopped() {
			return
		}

		e.task.IncrementTotal()
		e.task.SetCurrentItem(fmt.Sprintf("Global Style: %s", style.Name))

		sld, err := e.sourceClient.GetStyleSLD("", style.Name)
		if err != nil {
			e.task.IncrementFailed()
			continue
		}

		err = e.destClient.CreateOrUpdateStyle("", style.Name, sld)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				e.task.IncrementSkipped()
			} else {
				e.task.IncrementFailed()
			}
		} else {
			e.task.IncrementDone()
		}
		e.task.UpdateProgress()
	}
}

func (e *Executor) syncDataStores(workspace string) {
	stores, err := e.sourceClient.GetDataStores(workspace)
	if err != nil {
		e.task.AddLog(fmt.Sprintf("Failed to get data stores for %s: %v", workspace, err))
		return
	}

	for _, store := range stores {
		if e.isStopped() {
			return
		}

		e.task.IncrementTotal()
		e.task.SetCurrentItem(fmt.Sprintf("DataStore: %s:%s", workspace, store.Name))
		e.task.AddLog(fmt.Sprintf("Syncing data store: %s", store.Name))

		// Note: Data stores often have connection-specific settings
		e.task.AddLog(fmt.Sprintf("Note: Data store %s may require manual configuration on destination", store.Name))
		e.task.IncrementSkipped()
		e.task.UpdateProgress()

		// Sync layers from this store
		if e.options.Layers {
			e.syncLayers(workspace, store.Name, "datastore")
		}
	}
}

func (e *Executor) syncCoverageStores(workspace string) {
	stores, err := e.sourceClient.GetCoverageStores(workspace)
	if err != nil {
		e.task.AddLog(fmt.Sprintf("Failed to get coverage stores for %s: %v", workspace, err))
		return
	}

	for _, store := range stores {
		if e.isStopped() {
			return
		}

		e.task.IncrementTotal()
		e.task.SetCurrentItem(fmt.Sprintf("CoverageStore: %s:%s", workspace, store.Name))
		e.task.AddLog(fmt.Sprintf("Syncing coverage store: %s", store.Name))

		e.task.AddLog(fmt.Sprintf("Note: Coverage store %s may require manual configuration on destination", store.Name))
		e.task.IncrementSkipped()
		e.task.UpdateProgress()

		// Sync coverages from this store
		if e.options.Layers {
			e.syncCoverages(workspace, store.Name)
		}
	}
}

func (e *Executor) syncLayers(workspace, store, storeType string) {
	layers, err := e.sourceClient.GetLayers(workspace)
	if err != nil {
		e.task.AddLog(fmt.Sprintf("Failed to get layers for %s: %v", workspace, err))
		return
	}

	for _, layer := range layers {
		if e.isStopped() {
			return
		}

		e.task.IncrementTotal()
		e.task.SetCurrentItem(fmt.Sprintf("Layer: %s:%s", workspace, layer.Name))

		// Get layer metadata
		metadata, err := e.sourceClient.GetLayerMetadata(workspace, layer.Name)
		if err != nil {
			e.task.IncrementFailed()
			e.task.AddLog(fmt.Sprintf("Failed to get layer metadata for %s: %v", layer.Name, err))
			continue
		}

		e.task.AddLog(fmt.Sprintf("Layer %s has title: %s", layer.Name, metadata.Title))
		e.task.IncrementSkipped() // Layers require stores to exist first
		e.task.UpdateProgress()
	}
}

func (e *Executor) syncCoverages(workspace, store string) {
	coverages, err := e.sourceClient.GetCoverages(workspace, store)
	if err != nil {
		e.task.AddLog(fmt.Sprintf("Failed to get coverages for %s:%s: %v", workspace, store, err))
		return
	}

	for _, cov := range coverages {
		if e.isStopped() {
			return
		}

		e.task.IncrementTotal()
		e.task.SetCurrentItem(fmt.Sprintf("Coverage: %s:%s", workspace, cov.Name))
		e.task.IncrementSkipped() // Coverages require stores to exist first
		e.task.UpdateProgress()
	}
}

func (e *Executor) syncLayerGroups(workspace string) {
	groups, err := e.sourceClient.GetLayerGroups(workspace)
	if err != nil {
		e.task.AddLog(fmt.Sprintf("Failed to get layer groups for %s: %v", workspace, err))
		return
	}

	for _, group := range groups {
		if e.isStopped() {
			return
		}

		e.task.IncrementTotal()
		e.task.SetCurrentItem(fmt.Sprintf("LayerGroup: %s:%s", workspace, group.Name))
		e.task.AddLog(fmt.Sprintf("Syncing layer group: %s", group.Name))

		// Layer groups depend on layers existing
		e.task.IncrementSkipped()
		e.task.UpdateProgress()
	}
}
