import { makeRowRender, UdpRouters as UdpRoutersPage, UdpRoutersRender } from './UdpRouters'

import * as useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { useFetchWithPaginationMock } from 'utils/mocks'
import { renderWithProviders } from 'utils/test'

describe('<UdpRoutersPage />', () => {
  it('should render the routers list', () => {
    const pages = [
      {
        entryPoints: ['web-udp'],
        service: 'udp-all',
        rule: 'HostSNI(`*`)',
        status: 'enabled',
        using: ['web-secured', 'web'],
        name: 'udp-all@docker00',
        provider: 'docker',
      },
      {
        entryPoints: ['web-udp'],
        service: 'udp-all',
        rule: 'HostSNI(`*`)',
        status: 'disabled',
        using: ['web-secured', 'web'],
        name: 'udp-all@docker01',
        provider: 'docker',
      },
      {
        entryPoints: ['web-udp'],
        service: 'udp-all',
        rule: 'HostSNI(`*`)',
        status: 'enabled',
        using: ['web-secured', 'web'],
        name: 'udp-all@docker02',
        provider: 'docker',
      },
    ].map(makeRowRender())
    const mock = vi
      .spyOn(useFetchWithPagination, 'default')
      .mockImplementation(() => useFetchWithPaginationMock({ pages }))

    const { container, getByTestId } = renderWithProviders(<UdpRoutersPage />)

    expect(mock).toHaveBeenCalled()
    expect(getByTestId('UDP Routers page')).toBeInTheDocument()
    const tbody = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[1]
    expect(tbody.querySelectorAll('a[role="row"]')).toHaveLength(3)

    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toIncludeMultiple(['web-udp'])
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('udp-all@docker00')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('udp-all')
    expect(tbody.querySelectorAll('a[role="row"]')[0].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('testid="disabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toIncludeMultiple(['web-udp'])
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('udp-all@docker01')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('udp-all')
    expect(tbody.querySelectorAll('a[role="row"]')[1].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toIncludeMultiple(['web-udp'])
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('udp-all@docker02')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('udp-all')
    expect(tbody.querySelectorAll('a[role="row"]')[2].querySelector('svg[data-testid="docker"]')).toBeTruthy()
  })

  it('should render "No data available" when the API returns empty array', async () => {
    const { container, getByTestId } = renderWithProviders(
      <UdpRoutersRender
        error={undefined}
        isEmpty={true}
        isLoadingMore={false}
        isReachingEnd={true}
        loadMore={() => {}}
        pageCount={1}
        pages={[]}
      />,
    )
    expect(() => getByTestId('loading')).toThrow('Unable to find an element by: [data-testid="loading"]')
    const tfoot = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[2]
    expect(tfoot.querySelectorAll('div[role="row"]')).toHaveLength(1)
    expect(tfoot.querySelectorAll('div[role="row"]')[0].innerHTML).toContain('No data available')
  })
})
