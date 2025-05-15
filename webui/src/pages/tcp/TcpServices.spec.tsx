import { makeRowRender, TcpServices as TcpServicesPage, TcpServicesRender } from './TcpServices'

import * as useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { useFetchWithPaginationMock } from 'utils/mocks'
import { renderWithProviders } from 'utils/test'

describe('<TcpServicesPage />', () => {
  it('should render the services list', () => {
    const pages = [
      {
        loadBalancer: { terminationDelay: 10, servers: [{ address: '10.0.1.14:8080' }] },
        status: 'enabled',
        usedBy: ['tcp-all@docker'],
        name: 'tcp-all@docker00',
        provider: 'docker',
        type: 'loadbalancer',
      },
      {
        loadBalancer: { terminationDelay: 10, servers: [{ address: '10.0.1.14:8080' }] },
        status: 'disabled',
        usedBy: ['tcp-all@docker'],
        name: 'tcp-all@docker01',
        provider: 'docker',
        type: 'loadbalancer',
      },
      {
        loadBalancer: { terminationDelay: 10, servers: [{ address: '10.0.1.14:8080' }] },
        status: 'enabled',
        usedBy: ['tcp-all@docker'],
        name: 'tcp-all@docker02',
        provider: 'docker',
        type: 'loadbalancer',
      },
    ].map(makeRowRender())
    const mock = vi
      .spyOn(useFetchWithPagination, 'default')
      .mockImplementation(() => useFetchWithPaginationMock({ pages }))

    const { container, getByTestId } = renderWithProviders(<TcpServicesPage />)

    expect(mock).toHaveBeenCalled()
    expect(getByTestId('TCP Services page')).toBeInTheDocument()
    expect(container.querySelectorAll('tbody tr')).toHaveLength(3)

    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('tcp-all@docker00')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('loadbalancer')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('1')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('testid="disabled"')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('tcp-all@docker01')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('loadbalancer')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('1')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('tcp-all@docker02')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('loadbalancer')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('1')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('img alt="docker"')
  })

  it('should render "No data available" when the API returns empty array', async () => {
    const { container, getByTestId } = renderWithProviders(
      <TcpServicesRender
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
