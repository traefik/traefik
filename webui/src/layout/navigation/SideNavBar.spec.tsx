import { screen } from '@testing-library/react'

import { SideNav } from './SideNavBar'

import { VersionContext } from 'contexts/version'
import { renderWithProviders } from 'utils/test'

const renderSideNav = (dashboardName: string) =>
  renderWithProviders(
    <VersionContext.Provider value={{ showHubButton: false, version: '1.0.0', dashboardName }}>
      <SideNav isExpanded={true} onSidePanelToggle={() => {}} />
    </VersionContext.Provider>,
  )

describe('<SideNav /> InstanceBadge placement', () => {
  it('renders the badge below the logo when dashboardName is set', () => {
    renderSideNav('int')
    expect(screen.getByTestId('instance-badge')).toBeInTheDocument()
    expect(screen.getByTestId('instance-badge-below')).toBeInTheDocument()
  })

  it('renders nothing when dashboardName is empty', () => {
    renderSideNav('')
    expect(screen.queryByTestId('instance-badge')).toBeNull()
    expect(screen.queryByTestId('instance-badge-below')).toBeNull()
  })
})
