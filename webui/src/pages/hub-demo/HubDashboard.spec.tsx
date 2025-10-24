import { waitFor } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

import HubDashboard from './HubDashboard'
import * as scriptVerification from './workers/scriptVerification'

import { renderWithProviders } from 'utils/test'

vi.mock('./workers/scriptVerification', () => ({
  verifyScriptSignature: vi.fn(),
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
  const mockVerifyScriptSignature = vi.mocked(scriptVerification.verifyScriptSignature)

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should render loading state during script verification', async () => {
    mockVerifyScriptSignature.mockImplementation(() => new Promise((resolve) => setTimeout(() => resolve(true), 100)))

    const { getByTestId } = renderWithProviders(<HubDashboard path="dashboard" />, {
      route: '/hub-dashboard',
    })

    expect(getByTestId('loading')).toBeInTheDocument()

    await waitFor(() => {
      expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
    })
  })

  it('should render the custom web component when signature is verified', async () => {
    mockVerifyScriptSignature.mockResolvedValue(true)

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
    mockVerifyScriptSignature.mockResolvedValue(false)

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
    mockVerifyScriptSignature.mockResolvedValue(true)

    renderWithProviders(<HubDashboard path="dashboard" />)

    await waitFor(() => {
      expect(mockVerifyScriptSignature).toHaveBeenCalledWith(
        'MCowBQYDK2VwAyEAWMBZ0pMBaL/s8gNXxpAPCIQ8bxjnuz6bQFwGYvjXDfg=',
        'https://assets.traefik.io/hub-ui-demo.js',
        'https://assets.traefik.io/hub-ui-demo.js.sig',
      )
    })
  })

  it('should set theme attribute based on dark mode', async () => {
    mockVerifyScriptSignature.mockResolvedValue(true)

    const { container } = renderWithProviders(<HubDashboard path="dashboard" />)

    await waitFor(() => {
      expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
    })

    const hubComponent = container.querySelector('hub-ui-demo-app')

    expect(hubComponent?.getAttribute('theme')).toMatch(/light|dark/)
  })

  it('should handle path with :id parameter correctly', async () => {
    mockVerifyScriptSignature.mockResolvedValue(true)

    const { container } = renderWithProviders(<HubDashboard path="gateways:id" />)

    await waitFor(() => {
      expect(mockVerifyScriptSignature).toHaveBeenCalledTimes(1)
    })

    const hubComponent = container.querySelector('hub-ui-demo-app')
    expect(hubComponent).toBeInTheDocument()
    expect(hubComponent?.getAttribute('path')).toBe('gateways/test-id')
  })
})
