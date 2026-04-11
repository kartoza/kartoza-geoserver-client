/**
 * MSW handlers for mocking API responses in tests.
 */

import { http, HttpResponse } from 'msw'

export const handlers = [
  // Health check
  http.get('/health/', () => {
    return new HttpResponse('OK', { status: 200 })
  }),

  // Settings API
  http.get('/api/settings/', () => {
    return HttpResponse.json({
      theme: 'default',
      pingIntervalSecs: 60,
      lastLocalPath: '/home/user',
    })
  }),

  http.put('/api/settings/', async ({ request }) => {
    const body = await request.json() as Record<string, unknown>
    return HttpResponse.json({
      theme: body.theme ?? 'default',
      pingIntervalSecs: body.pingIntervalSecs ?? 60,
      lastLocalPath: body.lastLocalPath ?? '/home/user',
    })
  }),

  // Providers API
  http.get('/api/providers/', () => {
    return HttpResponse.json({
      providers: [
        {
          id: 'geoserver',
          name: 'GeoServer',
          description: 'OGC-compliant geospatial server',
          enabled: true,
          experimental: false,
        },
        {
          id: 'postgres',
          name: 'PostgreSQL',
          description: 'PostgreSQL database connections',
          enabled: true,
          experimental: false,
        },
        {
          id: 'geonode',
          name: 'GeoNode',
          description: 'Open source geospatial CMS',
          enabled: true,
          experimental: false,
        },
        {
          id: 's3',
          name: 'S3 Storage',
          description: 'S3-compatible storage',
          enabled: false,
          experimental: true,
        },
        {
          id: 'iceberg',
          name: 'Apache Iceberg',
          description: 'Iceberg tables',
          enabled: false,
          experimental: true,
        },
        {
          id: 'qgis',
          name: 'QGIS Projects',
          description: 'Local QGIS projects',
          enabled: false,
          experimental: true,
        },
        {
          id: 'qfieldcloud',
          name: 'QFieldCloud',
          description: 'QFieldCloud integration',
          enabled: false,
          experimental: true,
        },
        {
          id: 'mergin',
          name: 'Mergin Maps',
          description: 'Mergin Maps integration',
          enabled: false,
          experimental: true,
        },
      ],
    })
  }),

  // Connections API
  http.get('/api/connections/', () => {
    return HttpResponse.json([])
  }),

  http.post('/api/connections/', async ({ request }) => {
    const body = await request.json() as Record<string, unknown>
    return HttpResponse.json(
      {
        id: 'mock-conn-id',
        name: body.name,
        url: body.url,
        username: body.username,
        is_active: false,
      },
      { status: 201 }
    )
  }),

  // S3 Connections API
  http.get('/api/s3/connections/', () => {
    return HttpResponse.json([])
  }),

  // PostgreSQL Services API
  http.get('/api/postgres/services/', () => {
    return HttpResponse.json([])
  }),

  // GeoNode Connections API
  http.get('/api/geonode/connections/', () => {
    return HttpResponse.json([])
  }),

  // Iceberg Connections API
  http.get('/api/iceberg/connections/', () => {
    return HttpResponse.json([])
  }),

  // QFieldCloud Connections API
  http.get('/api/qfieldcloud/connections/', () => {
    return HttpResponse.json([])
  }),

  // Mergin Maps Connections API
  http.get('/api/mergin/connections/', () => {
    return HttpResponse.json([])
  }),

  // QGIS Projects API
  http.get('/api/qgis/projects/', () => {
    return HttpResponse.json([])
  }),
]
