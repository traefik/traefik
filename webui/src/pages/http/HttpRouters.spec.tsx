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
    expect(container.querySelectorAll('tbody tr')).toHaveLength(4)

    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).not.toContain('testid="tls-on"')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('Host(`jaeger-v2-example-beta1`)')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toIncludeMultiple(['web-secured', 'web'])
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('jaeger_v2-example-beta1@docker')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('jaeger_v2-example-beta1')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('testid="disabled"')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).not.toContain('testid="tls-on"')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('Path(`somethingreallyunexpected`)')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toIncludeMultiple(['web-secured', 'web'])
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('orphan-router@file')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('unexistingservice')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('img alt="file"')

    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).not.toContain('testid="tls-on"')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('Host(`server`)')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toIncludeMultiple(['web-redirect'])
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('server-redirect@docker')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('api2_v2-example-beta1')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('testid="tls-on"')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('Host(`server`)')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toIncludeMultiple(['web-secured'])
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('server-secured@docker')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('api2_v2-example-beta1')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('img alt="docker"')
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
    expect(container.querySelectorAll('tfoot tr')).toHaveLength(1)
    expect(container.querySelectorAll('tfoot tr')[0].innerHTML).toContain('No data available')
  })
})
