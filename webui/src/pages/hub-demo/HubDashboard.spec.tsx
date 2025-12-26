import { waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

import { PUBLIC_KEY } from './constants'
import HubDashboard, { resetCache } from './HubDashboard'

import { renderWithProviders } from 'utils/test'
import verifySignature from 'utils/workers/scriptVerification'

vi.mock('utils/workers/scriptVerification', () => ({
  default: vi.fn(),
}))

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useParams: vi.fn(() => ({ id: 'test-id' })),
  }
})

vi.mock('hooks/use-theme', () => ({
  useIsDarkMode: vi.fn(() => false),
  useTheme: vi.fn(() => ({
    selectedTheme: 'light',
    appliedTheme: 'light',
    setTheme: vi.fn(),
  })),
}))

describe('HubDashboard demo', () => {
  const mockVerifyScriptSignature = vi.mocked(verifySignature)
  let mockCreateObjectURL: ReturnType<typeof vi.fn>

  beforeEach(() => {
    vi.clearAllMocks()

    mockCreateObjectURL = vi.fn(() => 'blob:mock-url')
    globalThis.URL.createObjectURL = mockCreateObjectURL
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('without cache', () => {
    beforeEach(() => {
      resetCache()
    })

    it('should render loading state during script verification', async () => {
      const mockScriptContent = new ArrayBuffer(100)
      mockVerifyScriptSignature.mockImplementation(
        () =>
          new Promise((resolve) =>
            setTimeout(() => resolve({ verified: true, scriptContent: mockScriptContent }), 100),
          ),
      )

      const { getByTestId } = renderWithProviders(<HubDashboard path="dashboard" />, {
        route: '/hub-dashboard',
      })

      expect(getByTestId('loading')).toBeInTheDocument()

      await waitFor(() => {
        expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
      })
    })

    it('should render the custom web component when signature is verified', async () => {
      const mockScriptContent = new ArrayBuffer(100)
      mockVerifyScriptSignature.mockResolvedValue({ verified: true, scriptContent: mockScriptContent })

      const { container } = renderWithProviders(<HubDashboard path="dashboard" />, {
        route: '/hub-dashboard',
      })

      await waitFor(() => {
        expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
      })

      const hubComponent = container.querySelector('hub-ui-demo-app')
      expect(hubComponent).toBeInTheDocument()
      expect(hubComponent?.getAttribute('path')).toBe('dashboard')
      expect(hubComponent?.getAttribute('baseurl')).toBe('#/hub-dashboard')
      expect(hubComponent?.getAttribute('theme')).toBe('light')
    })

    it('should render error state when signature verification fails', async () => {
      mockVerifyScriptSignature.mockResolvedValue({ verified: false })

      const { container } = renderWithProviders(<HubDashboard path="dashboard" />)

      await waitFor(() => {
        expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
      })

      expect(container.textContent).toContain("Oops! We couldn't load the demo content")

      const errorImage = container.querySelector('img[src="/img/gopher-something-went-wrong.png"]')
      expect(errorImage).toBeInTheDocument()

      const links = container.querySelectorAll('a')
      const websiteLink = Array.from(links).find((link) => link.href.includes('traefik.io/traefik-hub'))
      const docLink = Array.from(links).find((link) => link.href.includes('doc.traefik.io/traefik-hub'))

      expect(websiteLink).toBeInTheDocument()
      expect(docLink).toBeInTheDocument()
    })

    it('should render error state when verification throws an error', async () => {
      mockVerifyScriptSignature.mockRejectedValue(new Error('Network error'))

      const { container } = renderWithProviders(<HubDashboard path="dashboard" />)

      await waitFor(() => {
        expect(container.textContent).toContain("Oops! We couldn't load the demo content")
      })
    })

    it('should call verifyScriptSignature with correct parameters', async () => {
      const mockScriptContent = new ArrayBuffer(100)
      mockVerifyScriptSignature.mockResolvedValue({ verified: true, scriptContent: mockScriptContent })

      renderWithProviders(<HubDashboard path="dashboard" />)

      await waitFor(() => {
        expect(mockVerifyScriptSignature).toHaveBeenCalledWith(
          'https://assets.traefik.io/hub-ui-demo.js',
          'https://assets.traefik.io/hub-ui-demo.js.sig',
          PUBLIC_KEY,
        )
      })
    })

    it('should set theme attribute based on dark mode', async () => {
      const mockScriptContent = new ArrayBuffer(100)
      mockVerifyScriptSignature.mockResolvedValue({ verified: true, scriptContent: mockScriptContent })

      const { container } = renderWithProviders(<HubDashboard path="dashboard" />)

      await waitFor(() => {
        expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
      })

      const hubComponent = container.querySelector('hub-ui-demo-app')

      expect(hubComponent?.getAttribute('theme')).toMatch(/light|dark/)
    })

    it('should handle path with :id parameter correctly', async () => {
      const mockScriptContent = new ArrayBuffer(100)
      mockVerifyScriptSignature.mockResolvedValue({ verified: true, scriptContent: mockScriptContent })

      const { container } = renderWithProviders(<HubDashboard path="gateways:id" />)

      await waitFor(() => {
        expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
      })

      const hubComponent = container.querySelector('hub-ui-demo-app')
      expect(hubComponent).toBeInTheDocument()
      expect(hubComponent?.getAttribute('path')).toBe('gateways/test-id')
    })
  })

  describe('with cache', () => {
    beforeEach(() => {
      resetCache()
    })

    it('should use cached blob URL without calling verifySignature again', async () => {
      const mockScriptContent = new ArrayBuffer(100)
      mockVerifyScriptSignature.mockResolvedValue({ verified: true, scriptContent: mockScriptContent })

      // First render
      const { container: firstContainer, unmount: firstUnmount } = renderWithProviders(
        <HubDashboard path="dashboard" />,
      )

      await waitFor(() => {
        expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
      })

      const firstHubComponent = firstContainer.querySelector('hub-ui-demo-app')
      expect(firstHubComponent).toBeInTheDocument()

      firstUnmount()

      mockVerifyScriptSignature.mockClear()

      // Second render - should use cache
      const { container: secondContainer } = renderWithProviders(<HubDashboard path="dashboard" />)

      await waitFor(() => {
        const secondHubComponent = secondContainer.querySelector('hub-ui-demo-app')
        expect(secondHubComponent).toBeInTheDocument()
      })

      expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(0)
    })
  })
})
