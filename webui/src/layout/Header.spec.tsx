import { renderWithProviders } from 'utils/test'

import Header from './Header'

describe('<Header />', () => {
  it('should render the NavBar', async () => {
    const { container } = renderWithProviders(<Header />)

    expect(container.innerHTML).toContain('HTTP')
    expect(container.innerHTML).toContain('TCP')
    expect(container.innerHTML).toContain('UDP')
    expect(container.innerHTML).toContain('Plugins')
    expect(container.innerHTML).toContain('theme-switcher')
    expect(container.innerHTML).toContain('help-menu')
  })
})
