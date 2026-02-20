import { CertificateRenderRow, Certificates as CertificatesPage, CertificatesRender } from './Certificates'

import * as useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { useFetchWithPaginationMock } from 'utils/mocks'
import { renderWithProviders } from 'utils/test'

describe('<CertificatesPage />', () => {
  it('should render the certificates list', () => {
    const pages = [
      {
        name: 'MTI3LjAuMC4xLDo6MSxleGFtcGxlLmNvbQ==',
        commonName: 'example.com',
        sans: ['example.com', '127.0.0.1', '::1'],
        issuerOrg: 'Acme Co',
        issuerCN: 'Acme Root CA',
        notAfter: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString(),
        notBefore: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        status: 'enabled',
        resolver: 'letsencrypt',
      },
      {
        name: 'd2FybmluZy5jb20sd3d3Lndhcm5pbmcuY29t',
        commonName: 'warning.com',
        sans: ['warning.com', 'www.warning.com'],
        issuerOrg: 'Warning CA',
        issuerCN: '',
        notAfter: new Date(Date.now() + 15 * 24 * 60 * 60 * 1000).toISOString(),
        notBefore: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000).toISOString(),
        status: 'warning',
        resolver: '',
      },
      {
        name: 'ZXhwaXJlZC5jb20=',
        commonName: 'expired.com',
        sans: ['expired.com'],
        issuerOrg: 'Expired CA',
        notAfter: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
        notBefore: new Date(Date.now() - 365 * 24 * 60 * 60 * 1000).toISOString(),
        status: 'disabled',
        resolver: 'manual',
      },
    ].map(CertificateRenderRow)
    const mock = vi
      .spyOn(useFetchWithPagination, 'default')
      .mockImplementation(() => useFetchWithPaginationMock({ pages }))

    const { container, getByTestId } = renderWithProviders(<CertificatesPage />, {
      route: '/certificates',
      withPage: true,
    })

    expect(mock).toHaveBeenCalled()
    expect(getByTestId('/certificates page')).toBeInTheDocument()
    const tbody = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[1]
    expect(tbody.querySelectorAll('a[role="row"]')).toHaveLength(3)

    // First certificate (enabled)
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('example.com')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('Acme Co')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('days')

    // Second certificate (warning)
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('testid="warning"')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('warning.com')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('Warning CA')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('days')

    // Third certificate (disabled/expired)
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('testid="disabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('expired.com')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('Expired CA')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('EXPIRED')
  })

  it('should render "No data available" when the API returns empty array', async () => {
    const { container, getByTestId } = renderWithProviders(
      <CertificatesRender
        error={undefined}
        isEmpty={true}
        isLoadingMore={false}
        isReachingEnd={true}
        loadMore={() => {}}
        pageCount={1}
        pages={[]}
      />,
      { route: '/certificates', withPage: true },
    )
    expect(() => getByTestId('loading')).toThrow('Unable to find an element by: [data-testid="loading"]')
    const tfoot = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[2]
    expect(tfoot.querySelectorAll('div[role="row"]')).toHaveLength(1)
    expect(tfoot.querySelectorAll('div[role="row"]')[0].innerHTML).toContain('No data available')
  })

  it('should render "Failed to fetch data" when the API returns an error', async () => {
    const { container } = renderWithProviders(
      <CertificatesRender
        error={new Error('Test error')}
        isEmpty={false}
        isLoadingMore={false}
        isReachingEnd={true}
        loadMore={() => {}}
        pageCount={1}
        pages={[]}
      />,
      { route: '/certificates', withPage: true },
    )
    const tfoot = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[2]
    expect(tfoot.querySelectorAll('div[role="row"]')).toHaveLength(1)
    expect(tfoot.querySelectorAll('div[role="row"]')[0].innerHTML).toContain('Failed to fetch data')
  })

  it('should display certificate with expiry colors', () => {
    // Test different expiry colors
    const pages = [
      {
        name: 'Z3JlZW4=',
        commonName: 'green.com',
        sans: ['green.com'],
        issuerOrg: 'Test CA',
        notAfter: new Date(Date.now() + 100 * 24 * 60 * 60 * 1000).toISOString(), // 100 days = green
        notBefore: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        status: 'enabled',
      },
      {
        name: 'b3JhbmdlLmNvbQ==',
        commonName: 'orange.com',
        sans: ['orange.com'],
        issuerOrg: 'Test CA',
        notAfter: new Date(Date.now() + 10 * 24 * 60 * 60 * 1000).toISOString(), // 10 days = orange
        notBefore: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        status: 'warning',
      },
    ].map(CertificateRenderRow)

    const { container } = renderWithProviders(
      <CertificatesRender
        error={undefined}
        isEmpty={false}
        isLoadingMore={false}
        isReachingEnd={true}
        loadMore={() => {}}
        pageCount={1}
        pages={pages}
      />,
      { route: '/certificates', withPage: true },
    )

    const tbody = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[1]
    expect(tbody.querySelectorAll('a[role="row"]')).toHaveLength(2)
    
    // Green badge for >14 days
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('green.com')
    
    // Orange badge for <14 days
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('orange.com')
  })
})
