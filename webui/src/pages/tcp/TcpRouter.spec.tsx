import { TcpRouterRender } from './TcpRouter'

import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { renderWithProviders } from 'utils/test'

describe('<TcpRouterPage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <TcpRouterRender name="mock-router" data={undefined} error={new Error('Test error')} />,
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <TcpRouterRender name="mock-router" data={undefined} error={undefined} />,
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <TcpRouterRender name="mock-router" data={{} as ResourceDetailDataType} error={undefined} />,
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
      <TcpRouterRender name="mock-router" data={mockData as any} error={undefined} />,
    )

    const routerStructure = getByTestId('router-structure')
    expect(routerStructure.innerHTML).toContain(':443')
    expect(routerStructure.innerHTML).toContain(':8000')
    expect(routerStructure.innerHTML).toContain('tcp-all@docker')
    expect(routerStructure.innerHTML).toContain('tcp-all</span>')
    expect(routerStructure.innerHTML).toContain('TCP Router')
    expect(routerStructure.innerHTML).not.toContain('HTTP Router')

    const routerDetailsSection = getByTestId('router-details')
    const routerDetailsPanel = routerDetailsSection.querySelector(':scope > div:nth-child(1)')

    expect(routerDetailsPanel?.innerHTML).toContain('Status')
    expect(routerDetailsPanel?.innerHTML).toContain('Success')
    expect(routerDetailsPanel?.innerHTML).toContain('Provider')
    expect(routerDetailsPanel?.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(routerDetailsPanel?.innerHTML).toContain('Name')
    expect(routerDetailsPanel?.innerHTML).toContain('tcp-all@docker')
    expect(routerDetailsPanel?.innerHTML).toContain('Entrypoints')
    expect(routerDetailsPanel?.innerHTML).toContain('web</')
    expect(routerDetailsPanel?.innerHTML).toContain('web-secured')
    expect(routerDetailsPanel?.innerHTML).toContain('tcp-all</')

    const middlewaresPanel = routerDetailsSection.querySelector(':scope > div:nth-child(3)')
    const providers = Array.from(middlewaresPanel?.querySelectorAll('svg[data-testid="docker"]') || [])
    expect(middlewaresPanel?.innerHTML).toContain('middleware00')
    expect(middlewaresPanel?.innerHTML).toContain('middleware01')
    expect(middlewaresPanel?.innerHTML).toContain('Success')
    expect(providers.length).toBe(2)

    expect(getByTestId('/tcp/services/tcp-all@docker')).toBeInTheDocument()
  })
})
