import { RouterDetail } from 'components/routers/RouterDetail'
import { renderWithProviders } from 'utils/test'

describe('<TcpRouterPage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <RouterDetail name="mock-router" data={undefined} error={new Error('Test error')} protocol="tcp" />,
      { route: '/tcp/routers/mock-router', withPage: true },
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <RouterDetail name="mock-router" data={undefined} error={undefined} protocol="tcp" />,
      { route: '/tcp/routers/mock-router', withPage: true },
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <RouterDetail name="mock-router" data={{} as Resource.DetailsData} error={undefined} protocol="tcp" />,
      { route: '/tcp/routers/mock-router', withPage: true },
    )
    expect(getByTestId('Not found page')).toBeInTheDocument()
  })

  it('should render the router details', async () => {
    const mockData = {
      entryPoints: ['web-tcp'],
      service: 'tcp-all',
      rule: 'HostSNI(`*`)',
      status: 'enabled',
      using: ['web-secured', 'web'],
      name: 'tcp-all@docker',
      provider: 'docker',
      middlewares: [
        {
          status: 'enabled',
          usedBy: ['foo@docker', 'bar@file'],
          name: 'middleware00@docker',
          provider: 'docker',
          type: 'middleware00',
        },
        {
          status: 'enabled',
          usedBy: ['foo@docker', 'bar@file'],
          name: 'middleware01@docker',
          provider: 'docker',
          type: 'middleware01',
        },
      ],
      hasValidMiddlewares: true,
      entryPointsData: [
        {
          address: ':8000',
          name: 'web',
        },
        {
          address: ':443',
          name: 'web-secured',
        },
      ],
    }

    const { getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <RouterDetail name="mock-router" data={mockData as any} error={undefined} protocol="tcp" />,
      { route: '/tcp/routers/tcp-all@docker', withPage: true },
    )

    const routerStructure = getByTestId('router-structure')
    expect(routerStructure.innerHTML).toContain(':443')
    expect(routerStructure.innerHTML).toContain(':8000')
    expect(routerStructure.innerHTML).toContain('TCP Router')
    expect(routerStructure.innerHTML).not.toContain('HTTP Router')

    const routerDetailsSection = getByTestId('router-details')

    expect(routerDetailsSection?.innerHTML).toContain('Status')
    expect(routerDetailsSection?.innerHTML).toContain('Success')
    expect(routerDetailsSection?.innerHTML).toContain('Provider')
    expect(routerDetailsSection?.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(routerStructure.innerHTML).toContain('middleware00')
    expect(routerStructure.innerHTML).toContain('middleware01')

    expect(getByTestId('/tcp/services/tcp-all@docker')).toBeInTheDocument()
  })
})
