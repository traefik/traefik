import { makeRowRender, TcpMiddlewares as TcpMiddlewaresPage, TcpMiddlewaresRender } from './TcpMiddlewares'

import * as useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { useFetchWithPaginationMock } from 'utils/mocks'
import { renderWithProviders } from 'utils/test'

describe('<TcpMiddlewaresPage />', () => {
  it('should render the middlewares list', () => {
    const pages = [
      {
        inFlightConn: { amount: 10 },
        status: 'enabled',
        usedBy: ['web@docker'],
        name: 'inFlightConn-foo@docker',
        provider: 'docker',
        type: 'inFlightConn',
      },
      {
        ipWhiteList: { sourceRange: ['125.0.0.1', '125.0.0.4'] },
        error: ['message 1', 'message 2'],
        status: 'disabled',
        usedBy: ['foo@docker', 'bar@file'],
        name: 'ipWhiteList@docker',
        provider: 'docker',
        type: 'ipWhiteList',
      },
    ].map(makeRowRender(() => {}))
    const mock = vi
      .spyOn(useFetchWithPagination, 'default')
      .mockImplementation(() => useFetchWithPaginationMock({ pages }))

    const { container, getByTestId } = renderWithProviders(<TcpMiddlewaresPage />)

    expect(mock).toHaveBeenCalled()
    expect(getByTestId('TCP Middlewares page')).toBeInTheDocument()
    expect(container.querySelectorAll('tbody tr')).toHaveLength(2)

    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('inFlightConn-foo@docker')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('inFlightConn')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('testid="disabled"')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('ipWhiteList@docker')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('ipWhiteList')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('img alt="docker"')
  })

  it('should render "No data available" when the API returns empty array', async () => {
    const { container, getByTestId } = renderWithProviders(
      <TcpMiddlewaresRender
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
