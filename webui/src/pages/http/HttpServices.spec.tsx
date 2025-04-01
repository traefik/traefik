import { HttpServices as HttpServicesPage, HttpServicesRender, makeRowRender } from './HttpServices'

import * as useFetchWithPagination from 'hooks/use-fetch-with-pagination'
import { useFetchWithPaginationMock } from 'utils/mocks'
import { renderWithProviders } from 'utils/test'

describe('<HttpServicesPage />', () => {
  it('should render the services list', () => {
    const pages = [
      {
        loadBalancer: { servers: [{ url: 'http://10.0.1.12:80' }], passHostHeader: true },
        status: 'enabled',
        usedBy: ['server-redirect@docker', 'server-secured@docker'],
        serverStatus: { 'http://10.0.1.12:80': 'UP' },
        name: 'api2_v2-example-beta1@docker',
        provider: 'docker',
        type: 'loadbalancer',
      },
      {
        loadBalancer: {
          servers: [{ url: 'http://10.0.1.11:80' }, { url: 'http://10.0.1.12:80' }],
          passHostHeader: true,
        },
        status: 'enabled',
        usedBy: ['web@docker'],
        serverStatus: { 'http://10.0.1.11:80': 'UP' },
        name: 'api_v2-example-beta2@docker',
        provider: 'docker',
        type: 'loadbalancer',
      },
      {
        weighted: { sticky: { cookie: { name: 'chocolat', secure: true, httpOnly: true } } },
        status: 'enabled',
        usedBy: ['foo@docker'],
        name: 'canary1@docker',
        provider: 'docker',
        type: 'weighted',
      },
      {
        weighted: { sticky: { cookie: {} } },
        status: 'enabled',
        usedBy: ['fii@docker'],
        name: 'canary2@file',
        provider: 'file',
        type: 'weighted',
      },
    ].map(makeRowRender(() => {}))
    const mock = vi
      .spyOn(useFetchWithPagination, 'default')
      .mockImplementation(() => useFetchWithPaginationMock({ pages }))

    const { container, getByTestId } = renderWithProviders(<HttpServicesPage />)

    expect(mock).toHaveBeenCalled()
    expect(getByTestId('HTTP Services page')).toBeInTheDocument()
    expect(container.querySelectorAll('tbody tr')).toHaveLength(4)

    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('api2_v2-example-beta1@docker')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('loadbalancer')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('1')
    expect(container.querySelectorAll('tbody tr')[0].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('api_v2-example-beta2@docker')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('loadbalancer')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('2')
    expect(container.querySelectorAll('tbody tr')[1].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('canary1@docker')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('weighted')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('0')
    expect(container.querySelectorAll('tbody tr')[2].innerHTML).toContain('img alt="docker"')

    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('testid="enabled"')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('canary2@file')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('weighted')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('0')
    expect(container.querySelectorAll('tbody tr')[3].innerHTML).toContain('img alt="file"')
  })

  it('should render "No data available" when the API returns empty array', async () => {
    const { container, getByTestId } = renderWithProviders(
      <HttpServicesRender
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
