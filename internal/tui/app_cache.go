// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kartoza/kartoza-cloudbench/internal/models"
	"github.com/kartoza/kartoza-cloudbench/internal/tui/components"
)

// Cache operation result message
type cacheOperationCompleteMsg struct {
	success   bool
	err       error
	operation string
	layerName string
}

// showCacheWizard displays the cache management wizard for a layer
func (a *App) showCacheWizard(node *models.TreeNode) tea.Cmd {
	client := a.getClientForNode(node)
	if client == nil {
		a.errorMsg = "No client available for this node"
		return nil
	}

	// Try to get layer cache info to populate grid sets and formats
	layerName := node.Name
	if node.Workspace != "" {
		layerName = node.Workspace + ":" + node.Name
	}

	// Default grid sets and formats
	gridSets := []string{"EPSG:4326", "EPSG:900913", "EPSG:3857"}
	formats := []string{"image/png", "image/jpeg", "image/png8"}

	// Try to fetch actual layer cache info
	if layer, err := client.GetGWCLayer(layerName); err == nil && layer != nil {
		if len(layer.GridSubsets) > 0 {
			gridSets = layer.GridSubsets
		}
		if len(layer.MimeFormats) > 0 {
			formats = layer.MimeFormats
		}
	}

	// Create the cache wizard
	a.cacheWizard = components.NewCacheWizard(node, gridSets, formats)
	a.cacheWizard.SetSize(a.width, a.height)

	// Set callbacks
	a.cacheWizard.SetCallbacks(
		func(result components.CacheWizardResult) {
			// Execute the cache operation
			a.pendingCRUDCmd = a.executeCacheOperation(result)
		},
		func() {
			// Cancelled - nothing to do
		},
	)

	return a.cacheWizard.Init()
}

// executeCacheOperation executes the cache operation based on wizard result
func (a *App) executeCacheOperation(result components.CacheWizardResult) tea.Cmd {
	return func() tea.Msg {
		// Find the client for this connection
		client := a.clients[result.ConnectionID]
		if client == nil {
			return cacheOperationCompleteMsg{
				success:   false,
				err:       fmt.Errorf("connection not found"),
				operation: "cache",
				layerName: result.LayerName,
			}
		}

		var operationName string
		var err error

		switch result.Operation {
		case components.CacheOperationSeed:
			operationName = "Seed"
			err = client.SeedLayer(result.LayerName, models.GWCSeedRequest{
				GridSetID:   result.GridSet,
				ZoomStart:   result.ZoomStart,
				ZoomStop:    result.ZoomStop,
				Format:      result.Format,
				Type:        "seed",
				ThreadCount: result.ThreadCount,
			})

		case components.CacheOperationReseed:
			operationName = "Reseed"
			err = client.SeedLayer(result.LayerName, models.GWCSeedRequest{
				GridSetID:   result.GridSet,
				ZoomStart:   result.ZoomStart,
				ZoomStop:    result.ZoomStop,
				Format:      result.Format,
				Type:        "reseed",
				ThreadCount: result.ThreadCount,
			})

		case components.CacheOperationTruncate:
			operationName = "Truncate"
			err = client.TruncateLayer(result.LayerName, result.GridSet, result.Format, result.ZoomStart, result.ZoomStop)
		}

		if err != nil {
			return cacheOperationCompleteMsg{
				success:   false,
				err:       err,
				operation: operationName,
				layerName: result.LayerName,
			}
		}

		return cacheOperationCompleteMsg{
			success:   true,
			operation: operationName,
			layerName: result.LayerName,
		}
	}
}
