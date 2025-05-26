import { HttpMiddlewares as HttpMiddlewaresPage, HttpMiddlewaresRender, makeRowRender } from './HttpMiddlewares'

import * as useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { useFetchWithPaginationMock } from 'utils/mocks'
import { renderWithProviders } from 'utils/test'

describe('<HttpMiddlewaresPage />', () => {
  it('should render the middleware list', () => {
    const pages = [
      {
        addPrefix: { prefix: '/foo' },
        status: 'enabled',
        usedBy: ['web@docker'],
        name: 'add-foo@docker',
        provider: 'docker',
        type: 'addprefix',
      },
      {
        addPrefix: { prefix: '/path' },
        error: ['message 1', 'message 2'],
        status: 'disabled',
        usedBy: ['foo@docker', 'bar@file'],
        name: 'middleware00@docker',
        provider: 'docker',
        type: 'addprefix',
      },
      {
        basicAuth: {
          users: ['test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/', 'test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0'],
          usersFile: '/etc/foo/my/file/path/.htpasswd',
          realm: 'Hello you are here',
          removeHeader: true,
          headerField: 'X-WebAuth-User',
        },
        error: ['message 1', 'message 2'],
        status: 'enabled',
        usedBy: ['foo@docker', 'bar@file'],
        name: 'middleware01@docker',
        provider: 'docker',
        type: 'basicauth',
      },
      {
        buffering: {
          maxRequestBodyBytes: 42,
          memRequestBodyBytes: 42,
          maxResponseBodyBytes: 42,
          memResponseBodyBytes: 42,
          retryExpression: 'IsNetworkError() \u0026\u0026 Attempts() \u003c 2',
        },
        error: ['message 1', 'message 2'],
        status: 'enabled',
        usedBy: ['foo@docker', 'bar@file'],
        name: 'middleware02@docker',
        provider: 'docker',
        type: 'buffering',
      },
      {
        chain: {
          middlewares: [
            'middleware01@docker',
            'middleware021@docker',
            'middleware03@docker',
            'middleware06@docker',
            'middleware10@docker',
          ],
        },
        error: ['message 1', 'message 2'],
        status: 'enabled',
        usedBy: ['foo@docker', 'bar@file'],
        name: 'middleware03@docker',
        provider: 'docker',
        type: 'chain',
      },
    ].map(makeRowRender())
    const mock = vi
      .spyOn(useFetchWithPagination, 'default')
      .mockImplementation(() => useFetchWithPaginationMock({ pages }))

    const { container, getByTestId } = renderWithProviders(<HttpMiddlewaresPage />)

    expect(mock).toHaveBeenCalled()
    expect(getByTestId('HTTP Middlewares page')).toBeInTheDocument()
    const tbody = container.querySelectorAll('div[role="table"] > div[role="rowgroup"]')[1]
    expect(tbody.querySelectorAll('a[role="row"]')).toHaveLength(5)

    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('add-foo@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[0].innerHTML).toContain('addprefix')
    expect(tbody.querySelectorAll('a[role="row"]')[0].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('testid="disabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('middleware00@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[1].innerHTML).toContain('addprefix')
    expect(tbody.querySelectorAll('a[role="row"]')[1].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('middleware01@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[2].innerHTML).toContain('basicauth')
    expect(tbody.querySelectorAll('a[role="row"]')[2].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('middleware02@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[3].innerHTML).toContain('buffering')
    expect(tbody.querySelectorAll('a[role="row"]')[3].querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(tbody.querySelectorAll('a[role="row"]')[4].innerHTML).toContain('testid="enabled"')
    expect(tbody.querySelectorAll('a[role="row"]')[4].innerHTML).toContain('middleware03@docker')
    expect(tbody.querySelectorAll('a[role="row"]')[4].innerHTML).toContain('chain')
    expect(tbody.querySelectorAll('a[role="row"]')[4].querySelector('svg[data-testid="docker"]')).toBeTruthy()
  })

  it('should render "No data available" when the API returns empty array', async () => {
    const { container, getByTestId } = renderWithProviders(
      <HttpMiddlewaresRender
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
