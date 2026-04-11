/**
 * Tests for providers store.
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { useProvidersStore } from './providersStore'

// Create a mock fetch function
const mockFetch = vi.fn()

describe('providersStore', () => {
  beforeEach(() => {
    // Stub global fetch with our mock
    vi.stubGlobal('fetch', mockFetch)

    // Reset store state
    useProvidersStore.setState({
      providers: [],
      isLoading: false,
      error: null,
    })
    mockFetch.mockReset()
  })

  describe('fetchProviders', () => {
    it('should fetch providers successfully', async () => {
      const mockProviders = [
        {
          id: 'geoserver',
          name: 'GeoServer',
          description: 'Test',
          enabled: true,
          experimental: false,
        },
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ providers: mockProviders }),
      })

      await useProvidersStore.getState().fetchProviders()

      expect(useProvidersStore.getState().providers).toEqual(mockProviders)
      expect(useProvidersStore.getState().isLoading).toBe(false)
      expect(useProvidersStore.getState().error).toBeNull()
    })

    it('should handle fetch error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
      })

      await useProvidersStore.getState().fetchProviders()

      expect(useProvidersStore.getState().error).toBe('Failed to fetch providers')
      expect(useProvidersStore.getState().isLoading).toBe(false)
    })

    it('should set loading state while fetching', async () => {
      let resolvePromise: (value: unknown) => void
      const promise = new Promise((resolve) => {
        resolvePromise = resolve
      })

      mockFetch.mockReturnValueOnce(promise)

      const fetchPromise = useProvidersStore.getState().fetchProviders()

      expect(useProvidersStore.getState().isLoading).toBe(true)

      resolvePromise!({
        ok: true,
        json: () => Promise.resolve({ providers: [] }),
      })

      await fetchPromise

      expect(useProvidersStore.getState().isLoading).toBe(false)
    })
  })

  describe('isProviderEnabled', () => {
    it('should return true for enabled provider', () => {
      useProvidersStore.setState({
        providers: [
          {
            id: 'geoserver',
            name: 'GeoServer',
            description: 'Test',
            enabled: true,
            experimental: false,
          },
        ],
      })

      expect(useProvidersStore.getState().isProviderEnabled('geoserver')).toBe(true)
    })

    it('should return false for disabled provider', () => {
      useProvidersStore.setState({
        providers: [
          {
            id: 's3',
            name: 'S3',
            description: 'Test',
            enabled: false,
            experimental: true,
          },
        ],
      })

      expect(useProvidersStore.getState().isProviderEnabled('s3')).toBe(false)
    })

    it('should return false for non-existent provider', () => {
      expect(useProvidersStore.getState().isProviderEnabled('nonexistent')).toBe(false)
    })
  })

  describe('updateProviders', () => {
    it('should update providers successfully', async () => {
      const mockProviders = [
        {
          id: 's3',
          name: 'S3',
          description: 'Test',
          enabled: true,
          experimental: true,
        },
      ]

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ providers: mockProviders }),
      })

      await useProvidersStore.getState().updateProviders([
        { id: 's3', enabled: true },
      ])

      expect(useProvidersStore.getState().providers).toEqual(mockProviders)
    })

    it('should handle update error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
      })

      await useProvidersStore.getState().updateProviders([
        { id: 's3', enabled: true },
      ])

      expect(useProvidersStore.getState().error).toBe('Failed to update providers')
    })
  })
})
