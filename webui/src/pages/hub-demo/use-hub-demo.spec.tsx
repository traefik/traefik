import { renderHook, waitFor } from '@testing-library/react'
import { ReactNode } from 'react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

import { useHubDemo } from './use-hub-demo'

import verifySignature from 'utils/workers/scriptVerification'

vi.mock('utils/workers/scriptVerification', () => ({
  default: vi.fn(),
}))

const MOCK_ROUTES_MANIFEST = {
  routes: [
    {
      path: '/dashboard',
      label: 'Dashboard',
      icon: 'dashboard',
      contentPath: 'dashboard',
    },
    {
      path: '/gateway',
      label: 'Gateway',
      icon: 'gateway',
      contentPath: 'gateway',
      dynamicSegments: [':id'],
      activeMatches: ['/gateway/:id'],
    },
  ],
}

describe('useHubDemo', () => {
  const mockVerifySignature = vi.mocked(verifySignature)

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  const setupMockVerification = (manifest: HubDemo.Manifest) => {
    const encoder = new TextEncoder()
    const mockScriptContent = encoder.encode(JSON.stringify(manifest))

    mockVerifySignature.mockResolvedValue({
      verified: true,
      scriptContent: mockScriptContent.buffer,
    })
  }

  describe('basic functions', () => {
    const mockVerifySignature = vi.mocked(verifySignature)

    beforeEach(() => {
      vi.clearAllMocks()
    })

    afterEach(() => {
      vi.restoreAllMocks()
    })

    it('should return null when signature verification fails', async () => {
      mockVerifySignature.mockResolvedValue({
        verified: false,
      })

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(mockVerifySignature).toHaveBeenCalled()
      })

      await new Promise((resolve) => setTimeout(resolve, 10))

      expect(result.current.routes).toBeNull()
      expect(result.current.navigationItems).toBeNull()
    })

    it('should return null when scriptContent is missing', async () => {
      mockVerifySignature.mockResolvedValue({
        verified: true,
        scriptContent: undefined,
      })

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(mockVerifySignature).toHaveBeenCalled()
      })

      await new Promise((resolve) => setTimeout(resolve, 10))

      expect(result.current.routes).toBeNull()
      expect(result.current.navigationItems).toBeNull()
    })

    it('should handle errors during manifest fetch', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      mockVerifySignature.mockRejectedValue(new Error('Network error'))

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith('Failed to load hub demo manifest:', expect.any(Error))
      })

      expect(result.current.routes).toBeNull()
      expect(result.current.navigationItems).toBeNull()

      consoleErrorSpy.mockRestore()
    })

    it('should handle invalid JSON in manifest', async () => {
      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      const encoder = new TextEncoder()
      const invalidJson = encoder.encode('{ invalid json }')

      mockVerifySignature.mockResolvedValue({
        verified: true,
        scriptContent: invalidJson.buffer,
      })

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith('Failed to load hub demo manifest:', expect.any(Error))
      })

      expect(result.current.routes).toBeNull()
      expect(result.current.navigationItems).toBeNull()

      consoleErrorSpy.mockRestore()
    })
  })

  describe('routes generation', () => {
    it('should generate routes with correct base path', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(result.current.routes).not.toBeNull()
      })

      expect(result.current.routes).toHaveLength(3)
      expect(result.current.routes![0].path).toBe('/hub/dashboard')
      expect(result.current.routes![1].path).toBe('/hub/gateway')
      expect(result.current.routes![2].path).toBe('/hub/gateway/:id')
    })

    it('should generate routes for dynamic segments', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(result.current.routes).not.toBeNull()
      })

      expect(result.current.routes).toHaveLength(3)
      expect(result.current.routes![0].path).toBe('/hub/dashboard')
      expect(result.current.routes![1].path).toBe('/hub/gateway')
      expect(result.current.routes![2].path).toBe('/hub/gateway/:id')
    })

    it('should render HubDashboard with correct contentPath for dynamic segments', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(result.current.routes).not.toBeNull()
      })

      const baseRoute = result.current.routes![1]
      const dynamicRoute = result.current.routes![2]

      expect(baseRoute.element).toBeDefined()
      expect(dynamicRoute.element).toBeDefined()

      const baseElement = baseRoute.element as ReactNode & { props?: { path: string } }
      const dynamicElement = dynamicRoute.element as ReactNode & { props?: { path: string } }

      expect((baseElement as { props: { path: string } }).props.path).toBe('gateway')
      expect((dynamicElement as { props: { path: string } }).props.path).toBe('gateway:id')
    })

    it('should update routes when basePath changes', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result, rerender } = renderHook(({ basePath }) => useHubDemo(basePath), {
        initialProps: { basePath: '/hub' },
      })

      await waitFor(() => {
        expect(result.current.routes).not.toBeNull()
      })

      expect(result.current.routes![0].path).toBe('/hub/dashboard')

      rerender({ basePath: '/demo' })

      expect(result.current.routes![0].path).toBe('/demo/dashboard')
    })
  })

  describe('navigation items generation', () => {
    it('should generate navigation items with correct icons', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(result.current.navigationItems).not.toBeNull()
      })

      expect(result.current.navigationItems).toHaveLength(2)
      expect(result.current.navigationItems![0].label).toBe('Dashboard')
      expect(result.current.navigationItems![0].path).toBe('/hub/dashboard')
      expect(result.current.navigationItems![0].icon).toBeDefined()
      expect(result.current.navigationItems![1].label).toBe('Gateway')
    })

    it('should include activeMatches in navigation items', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(result.current.navigationItems).not.toBeNull()
      })

      expect(result.current.navigationItems![1].activeMatches).toEqual(['/hub/gateway/:id'])
    })

    it('should update navigation items when basePath changes', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result, rerender } = renderHook(({ basePath }) => useHubDemo(basePath), {
        initialProps: { basePath: '/hub' },
      })

      await waitFor(() => {
        expect(result.current.navigationItems).not.toBeNull()
      })

      expect(result.current.navigationItems![0].path).toBe('/hub/dashboard')

      rerender({ basePath: '/demo' })

      expect(result.current.navigationItems![0].path).toBe('/demo/dashboard')
    })

    it('should handle unknown icon types gracefully', async () => {
      const manifestWithUnknownIcon: HubDemo.Manifest = {
        routes: [
          {
            path: '/unknown',
            label: 'Unknown',
            icon: 'unknown-icon-type',
            contentPath: 'unknown',
          },
        ],
      }

      setupMockVerification(manifestWithUnknownIcon)

      const { result } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(result.current.navigationItems).not.toBeNull()
      })

      expect(result.current.navigationItems![0].icon).toBeUndefined()
    })
  })

  describe('memoization', () => {
    it('should not regenerate routes when manifest and basePath are unchanged', async () => {
      setupMockVerification(MOCK_ROUTES_MANIFEST)

      const { result, rerender } = renderHook(() => useHubDemo('/hub'))

      await waitFor(() => {
        expect(result.current.routes).not.toBeNull()
      })

      const firstRoutes = result.current.routes
      const firstNavItems = result.current.navigationItems

      rerender()

      expect(result.current.routes).toBe(firstRoutes)
      expect(result.current.navigationItems).toBe(firstNavItems)
    })
  })
})
