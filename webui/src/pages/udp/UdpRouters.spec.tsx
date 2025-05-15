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
    expect(container.querySelectorAll('tbody tr')).toHaveLength(3)

    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toIncludeMultiple(['web-udp'])
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('udp-all@docker00')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('udp-all')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('testid="disabled"')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toIncludeMultiple(['web-udp'])
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('udp-all@docker01')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('udp-all')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toIncludeMultiple(['web-udp'])
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('udp-all@docker02')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('udp-all')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('img alt="docker"')
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
    expect(container.querySelectorAll('tfoot tr')).toHaveLength(1)
    expect(container.querySelectorAll('tfoot tr')[0].innerHTML).toContain('No data available')
  })
})
