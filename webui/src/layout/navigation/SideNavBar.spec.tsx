import { screen } from '@testing-library/react'

import { SideNav } from './SideNavBar'

import { VersionContext } from 'contexts/version'
import { renderWithProviders } from 'utils/test'

const renderSideNav = (dashboardName: string, dashboardNamePosition: 'side' | 'below') =>
  renderWithProviders(
    <VersionContext.Provider
      value={{ showHubButton: false, version: '1.0.0', dashboardName, dashboardNamePosition }}
    >
      <SideNav isExpanded={true} onSidePanelToggle={() => {}} />
    </VersionContext.Provider>,
  )

describe('<SideNav /> InstanceBadge placement', () => {
  it('renders the badge inline (no below container) when position=side', () => {
    renderSideNav('int', 'side')
    expect(screen.getByTestId('instance-badge')).toBeInTheDocument()
    expect(screen.queryByTestId('instance-badge-below')).toBeNull()
  })

  it('renders the badge in a below container when position=below', () => {
    renderSideNav('int', 'below')
    expect(screen.getByTestId('instance-badge')).toBeInTheDocument()
    expect(screen.getByTestId('instance-badge-below')).toBeInTheDocument()
  })

  it('renders nothing when dashboardName is empty regardless of position', () => {
    renderSideNav('', 'below')
    expect(screen.queryByTestId('instance-badge')).toBeNull()
    expect(screen.queryByTestId('instance-badge-below')).toBeNull()
  })
})
