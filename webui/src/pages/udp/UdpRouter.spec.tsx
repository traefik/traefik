import { RouterDetail } from 'components/routers/RouterDetail'
import { renderWithProviders } from 'utils/test'

describe('<UdpRouterPage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <RouterDetail name="mock-router" data={undefined} error={new Error('Test error')} protocol="udp" />,
      { route: '/udp/routers/mock-router', withPage: true },
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <RouterDetail name="mock-router" data={undefined} error={undefined} protocol="udp" />,
      { route: '/udp/routers/mock-router', withPage: true },
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <RouterDetail name="mock-router" data={{} as Resource.DetailsData} error={undefined} protocol="udp" />,
      { route: '/udp/routers/mock-router', withPage: true },
    )
    expect(getByTestId('Not found page')).toBeInTheDocument()
  })

  it('should render the router details', async () => {
    const mockData = {
      entryPoints: ['web-udp'],
      service: 'udp-all',
      rule: 'HostSNI(`*`)',
      status: 'enabled',
      using: ['web-secured', 'web'],
      name: 'udp-all@docker',
      provider: 'docker',
      middlewares: undefined,
      hasValidMiddlewares: undefined,
      entryPointsData: [
        {
          address: ':443',
          name: 'web-secured',
        },
        {
          address: ':8000',
          name: 'web',
        },
      ],
    }

    const { getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <RouterDetail name="mock-router" data={mockData as any} error={undefined} protocol="udp" />,
      { route: '/udp/routers/udp-all@docker', withPage: true },
    )

    const routerStructure = getByTestId('router-structure')
    expect(routerStructure.innerHTML).toContain(':443')
    expect(routerStructure.innerHTML).toContain(':8000')
    expect(routerStructure.innerHTML).toContain('UDP Router')
    expect(routerStructure.innerHTML).not.toContain('HTTP Router')

    const routerDetailsSection = getByTestId('router-details')

    expect(routerDetailsSection?.innerHTML).toContain('Status')
    expect(routerDetailsSection?.innerHTML).toContain('Success')
    expect(routerDetailsSection?.innerHTML).toContain('Provider')
    expect(routerDetailsSection?.querySelector('svg[data-testid="docker"]')).toBeTruthy()

    expect(getByTestId('/udp/services/udp-all@docker')).toBeInTheDocument()
  })
})
