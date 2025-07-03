import { SideNav, TopNav } from './Navigation'

import { renderWithProviders } from 'utils/test'

describe('Navigation', () => {
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
})
