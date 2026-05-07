import { screen } from '@testing-library/react'

import InstanceBadge from './InstanceBadge'

import { VersionContext } from 'contexts/version'
import { renderWithProviders } from 'utils/test'

describe('<InstanceBadge />', () => {
  it('renders nothing when dashboardName is empty', () => {
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: '', dashboardNamePosition: 'side' }}>
        <InstanceBadge />
      </VersionContext.Provider>,
    )

    expect(screen.queryByTestId('instance-badge')).toBeNull()
  })

  it('renders the badge with the dashboardName text when set', () => {
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: 'int', dashboardNamePosition: 'side' }}>
        <InstanceBadge />
      </VersionContext.Provider>,
    )

    const badge = screen.getByTestId('instance-badge')
    expect(badge).toBeInTheDocument()
    expect(badge).toHaveAttribute('aria-label', 'Instance: int')
    expect(badge).toHaveTextContent('int')
  })

  it('truncates with ellipsis when dashboardName exceeds 32 grapheme clusters', () => {
    const longName = 'a'.repeat(40)
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: longName, dashboardNamePosition: 'side' }}>
        <InstanceBadge />
      </VersionContext.Provider>,
    )

    const badge = screen.getByTestId('instance-badge')
    expect(badge).toHaveTextContent('a'.repeat(32) + '…')
    expect(badge).toHaveAttribute('aria-label', `Instance: ${longName}`)
  })

  it('handles multi-byte unicode safely (no surrogate pair split)', () => {
    const emojiName = '🌍🌎🌏'
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: emojiName, dashboardNamePosition: 'side' }}>
        <InstanceBadge />
      </VersionContext.Provider>,
    )

    const badge = screen.getByTestId('instance-badge')
    expect(badge).toHaveTextContent(emojiName)
  })
})
