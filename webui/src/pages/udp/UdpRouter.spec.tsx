import { UdpRouterRender } from './UdpRouter'

import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { renderWithProviders } from 'utils/test'

describe('<UdpRouterPage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <UdpRouterRender name="mock-router" data={undefined} error={new Error('Test error')} />,
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <UdpRouterRender name="mock-router" data={undefined} error={undefined} />,
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <UdpRouterRender name="mock-router" data={{} as ResourceDetailDataType} error={undefined} />,
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
      <UdpRouterRender name="mock-router" data={mockData as any} error={undefined} />,
    )

    const routerStructure = getByTestId('router-structure')
    expect(routerStructure.innerHTML).toContain(':443')
    expect(routerStructure.innerHTML).toContain(':8000')
    expect(routerStructure.innerHTML).toContain('udp-all@docker')
    expect(routerStructure.innerHTML).toContain('udp-all</span>')
    expect(routerStructure.innerHTML).toContain('UDP Router')
    expect(routerStructure.innerHTML).not.toContain('HTTP Router')

    const routerDetailsSection = getByTestId('router-details')
    const routerDetailsPanel = routerDetailsSection.querySelector(':scope > div:nth-child(1)')

    expect(routerDetailsPanel?.innerHTML).toContain('Status')
    expect(routerDetailsPanel?.innerHTML).toContain('Success')
    expect(routerDetailsPanel?.innerHTML).toContain('Provider')
    expect(routerDetailsPanel?.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(routerDetailsPanel?.innerHTML).toContain('Name')
    expect(routerDetailsPanel?.innerHTML).toContain('udp-all@docker')
    expect(routerDetailsPanel?.innerHTML).toContain('Entrypoints')
    expect(routerDetailsPanel?.innerHTML).toContain('web</')
    expect(routerDetailsPanel?.innerHTML).toContain('web-secured')
    expect(routerDetailsPanel?.innerHTML).toContain('udp-all</')

    expect(getByTestId('/udp/services/udp-all@docker')).toBeInTheDocument()
  })
})
