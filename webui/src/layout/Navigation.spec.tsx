import { waitFor } from '@testing-library/react'

import { SideNav, TopNav } from './Navigation'

import useHubUpgradeButton from 'hooks/use-hub-upgrade-button'
import { renderWithProviders } from 'utils/test'

vi.mock('hooks/use-hub-upgrade-button')

const mockUseHubUpgradeButton = vi.mocked(useHubUpgradeButton)

describe('Navigation', () => {
  beforeEach(() => {
    mockUseHubUpgradeButton.mockReturnValue({
      signatureVerified: false,
      scriptBlobUrl: null,
      isCustomElementDefined: false,
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  it('should render the side navigation bar', async () => {
    const { container } = renderWithProviders(<SideNav isExpanded={false} onSidePanelToggle={() => {}} />)

    expect(container.innerHTML).toContain('HTTP')
    expect(container.innerHTML).toContain('TCP')
    expect(container.innerHTML).toContain('UDP')
    expect(container.innerHTML).toContain('Plugins')
  })

  it('should render the top navigation bar', async () => {
    const { container } = renderWithProviders(<TopNav />)

    expect(container.innerHTML).toContain('theme-switcher')
    expect(container.innerHTML).toContain('help-menu')
  })

  describe('hub-button-app rendering', () => {
    it('should NOT render hub-button-app when signatureVerified is false', async () => {
      mockUseHubUpgradeButton.mockReturnValue({
        signatureVerified: false,
        scriptBlobUrl: null,
        isCustomElementDefined: false,
      })

      const { container } = renderWithProviders(<TopNav />)

      const hubButtonApp = container.querySelector('hub-button-app')
      expect(hubButtonApp).toBeNull()
    })

    it('should NOT render hub-button-app when scriptBlobUrl is null', async () => {
      mockUseHubUpgradeButton.mockReturnValue({
        signatureVerified: true,
        scriptBlobUrl: null,
        isCustomElementDefined: false,
      })

      const { container } = renderWithProviders(<TopNav />)

      const hubButtonApp = container.querySelector('hub-button-app')
      expect(hubButtonApp).toBeNull()
    })

    it('should render hub-button-app when signatureVerified is true and scriptBlobUrl exists', async () => {
      mockUseHubUpgradeButton.mockReturnValue({
        signatureVerified: true,
        scriptBlobUrl: 'blob:http://localhost:3000/mock-blob-url',
        isCustomElementDefined: false,
      })

      const { container } = renderWithProviders(<TopNav />)

      await waitFor(() => {
        const hubButtonApp = container.querySelector('hub-button-app')
        expect(hubButtonApp).not.toBeNull()
      })
    })

    it('should NOT render hub-button-app when noHubButton prop is true', async () => {
      mockUseHubUpgradeButton.mockReturnValue({
        signatureVerified: true,
        scriptBlobUrl: 'blob:http://localhost:3000/mock-blob-url',
        isCustomElementDefined: false,
      })

      const { container } = renderWithProviders(<TopNav noHubButton={true} />)

      const hubButtonApp = container.querySelector('hub-button-app')
      expect(hubButtonApp).toBeNull()
    })
  })
})
