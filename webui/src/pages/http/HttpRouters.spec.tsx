import * as useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { HttpRouters as HttpRoutersPage, HttpRoutersRender, makeRowRender } from 'pages/http/HttpRouters'
import { useFetchWithPaginationMock } from 'utils/mocks'
import { renderWithProviders } from 'utils/test'

describe('<HttpRoutersPage />', () => {
  it('should render the routers list', () => {
    const pages = [
      {
        service: 'jaeger_v2-example-beta1',
        rule: 'Host(`jaeger-v2-example-beta1`)',
        status: 'enabled',
        using: ['web-secured', 'web'],
        name: 'jaeger_v2-example-beta1@docker',
        provider: 'docker',
      },
      {
        middlewares: ['middleware00@docker', 'middleware01@docker', 'middleware02@docker'],
        service: 'unexistingservice',
        rule: 'Path(`somethingreallyunexpected`)',
        error: ['the service "unexistingservice@file" does not exist'],
        status: 'disabled',
        using: ['web-secured', 'web'],
        name: 'orphan-router@file',
        provider: 'file',
      },
      {
        entryPoints: ['web-redirect'],
        middlewares: ['redirect@file'],
        service: 'api2_v2-example-beta1',
        rule: 'Host(`server`)',
        status: 'enabled',
        using: ['web-redirect'],
        name: 'server-redirect@docker',
        provider: 'docker',
      },
      {
        entryPoints: ['web-secured'],
        service: 'api2_v2-example-beta1',
        rule: 'Host(`server`)',
        tls: {},
        status: 'enabled',
        using: ['web-secured'],
        name: 'server-secured@docker',
        provider: 'docker',
      },
    ].map(makeRowRender())
    const mock = vi
      .spyOn(useFetchWithPagination, 'default')
      .mockImplementation(() => useFetchWithPaginationMock({ pages }))

    const { container, getByTestId } = renderWithProviders(<HttpRoutersPage />)

    expect(mock).toHaveBeenCalled()
    expect(getByTestId('HTTP Routers page')).toBeInTheDocument()
    const tbody = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[1]
    expect(tbody.querySelectorAll('a[role="row"]')).toHaveLength(4)

    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).not.toContain('testid="tls-on"')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('Host(`jaeger-v2-example-beta1`)')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toIncludeMultiple(['web-secured', 'web'])
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('jaeger_v2-example-beta1@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('jaeger_v2-example-beta1')
    expect(tbody.querySelectorAll('a[role="row"]')[0].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('testid="disabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).not.toContain('testid="tls-on"')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('Path(`somethingreallyunexpected`)')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toIncludeMultiple(['web-secured', 'web'])
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('orphan-router@file')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('unexistingservice')
    expect(tbody.querySelectorAll('a[role="row"]')[1].querySelector('svg[data-testid="file"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).not.toContain('testid="tls-on"')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('Host(`server`)')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toIncludeMultiple(['web-redirect'])
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('server-redirect@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('api2_v2-example-beta1')
    expect(tbody.querySelectorAll('a[role="row"]')[2].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('testid="tls-on"')
    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('Host(`server`)')
    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toIncludeMultiple(['web-secured'])
    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('server-secured@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('api2_v2-example-beta1')
    expect(tbody.querySelectorAll('a[role="row"]')[3].querySelector('svg[data-testid="docker"]')).toBeTruthy()
  })

  it('should render "No data available" when the API returns empty array', async () => {
    const { container, getByTestId } = renderWithProviders(
      <HttpRoutersRender
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
