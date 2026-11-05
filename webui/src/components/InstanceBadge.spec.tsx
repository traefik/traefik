import { screen } from '@testing-library/react'

import InstanceBadge from './InstanceBadge'

import { VersionContext } from 'contexts/version'
import { renderWithProviders } from 'utils/test'

describe('<InstanceBadge />', () => {
  it('renders nothing when dashboardName is empty', () => {
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: '' }}>
        <InstanceBadge />
      </VersionContext.Provider>,
    )

    expect(screen.queryByTestId('instance-badge')).toBeNull()
  })

  it('renders the badge with the dashboardName text when set', () => {
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: 'int' }}>
        <InstanceBadge />
      </VersionContext.Provider>,
    )

    expect(screen.getByTestId('instance-badge')).toHaveTextContent('int')
  })

  it('keeps the full dashboardName in the DOM for long names', () => {
    const longName = 'production-eu-west-1-cluster-a-ingress-external'
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: longName }}>
        <InstanceBadge />
      </VersionContext.Provider>,
    )

    expect(screen.getByTestId('instance-badge')).toHaveTextContent(longName)
  })
})
