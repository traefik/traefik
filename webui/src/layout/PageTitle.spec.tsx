import { waitFor } from '@testing-library/react'

import PageTitle from './PageTitle'

import { VersionContext } from 'contexts/version'
import { renderWithProviders } from 'utils/test'

describe('<PageTitle />', () => {
  it('should render default title without page title or dashboard name', async () => {
    renderWithProviders(<PageTitle />)

    await waitFor(() => {
      expect(document.title).toBe('Traefik Proxy')
    })
  })

  it('should render with page title', async () => {
    renderWithProviders(<PageTitle title="Dashboard" />)

    await waitFor(() => {
      expect(document.title).toBe('Dashboard - Traefik Proxy')
    })
  })

  it('should render with dashboard name', async () => {
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: 'MyDashboard' }}>
        <PageTitle />
      </VersionContext.Provider>,
    )

    await waitFor(() => {
      expect(document.title).toBe('Traefik Proxy [MyDashboard]')
    })
  })

  it('should render with page title and dashboard name', async () => {
    renderWithProviders(
      <VersionContext.Provider value={{ showHubButton: false, version: '', dashboardName: 'MyDashboard' }}>
        <PageTitle title="Dashboard" />
      </VersionContext.Provider>,
    )

    await waitFor(() => {
      expect(document.title).toBe('Dashboard - Traefik Proxy [MyDashboard]')
    })
  })
})
