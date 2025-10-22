import { TcpServiceRender } from './TcpService'

import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { renderWithProviders } from 'utils/test'

describe('<TcpServicePage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <TcpServiceRender name="mock-service" data={undefined} error={new Error('Test error')} />,
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <TcpServiceRender name="mock-service" data={undefined} error={undefined} />,
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <TcpServiceRender name="mock-service" data={{} as ResourceDetailDataType} error={undefined} />,
    )
    expect(getByTestId('Not found page')).toBeInTheDocument()
  })

  it('should render the service', async () => {
    const mockData = {
      loadBalancer: {
        servers: [
          {
            address: 'http://10.0.1.12:80',
          },
        ],
        terminationDelay: 10,
        healthCheck: {
          interval: '30s',
          timeout: '10s',
          port: 8080,
          unhealthyInterval: '1m',
          send: 'PING',
          expect: 'PONG',
        },
      },
      serverStatus: {
        'http://10.0.1.12:80': 'UP',
      },
      status: 'enabled',
      usedBy: ['router-test1@docker'],
      name: 'service-test1',
      provider: 'docker',
      type: 'loadbalancer',
      routers: [
        {
          entryPoints: ['web-redirect'],
          middlewares: ['redirect@file'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-redirect'],
          name: 'router-test1@docker',
          provider: 'docker',
        },
      ],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <TcpServiceRender name="mock-service" data={mockData as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'service-test1')
    expect(titleTags.length).toBe(1)

    const serviceDetails = getByTestId('tcp-service-details')
    expect(serviceDetails.innerHTML).toContain('Type')
    expect(serviceDetails.innerHTML).toContain('loadbalancer')
    expect(serviceDetails.innerHTML).toContain('Provider')
    expect(serviceDetails.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(serviceDetails.innerHTML).toContain('Status')
    expect(serviceDetails.innerHTML).toContain('Success')
    expect(serviceDetails.innerHTML).toContain('Termination Delay')
    expect(serviceDetails.innerHTML).toContain('10 ms')

    const healthCheck = getByTestId('tcp-health-check')
    expect(healthCheck.innerHTML).toContain('Interval')
    expect(healthCheck.innerHTML).toContain('30s')
    expect(healthCheck.innerHTML).toContain('Timeout')
    expect(healthCheck.innerHTML).toContain('10s')
    expect(healthCheck.innerHTML).toContain('Port')
    expect(healthCheck.innerHTML).toContain('8080')
    expect(healthCheck.innerHTML).toContain('Unhealthy Interval')
    expect(healthCheck.innerHTML).toContain('1m')
    expect(healthCheck.innerHTML).toContain('Send')
    expect(healthCheck.innerHTML).toContain('PING')
    expect(healthCheck.innerHTML).toContain('Expect')
    expect(healthCheck.innerHTML).toContain('PONG')

    const serversList = getByTestId('tcp-servers-list')
    expect(serversList.childNodes.length).toBe(1)
    expect(serversList.innerHTML).toContain('http://10.0.1.12:80')

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(tableBody?.innerHTML).toContain('router-test1@docker')
  })

  it('should render the service servers from the serverStatus property', async () => {
    const mockData = {
      loadBalancer: {
        terminationDelay: 10,
      },
      status: 'enabled',
      usedBy: ['router-test1@docker', 'router-test2@docker'],
      serverStatus: {
        'http://10.0.1.12:81': 'UP',
      },
      name: 'service-test2',
      provider: 'docker',
      type: 'loadbalancer',
      routers: [
        {
          entryPoints: ['web-redirect'],
          middlewares: ['redirect@file'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-redirect'],
          name: 'router-test1@docker',
          provider: 'docker',
        },
        {
          entryPoints: ['web-secured'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-secured'],
          name: 'router-test2@docker',
          provider: 'docker',
        },
      ],
    }

    const { getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <TcpServiceRender name="mock-service" data={mockData as any} error={undefined} />,
    )

    const serversList = getByTestId('tcp-servers-list')
    expect(serversList.childNodes.length).toBe(1)
    expect(serversList.innerHTML).toContain('http://10.0.1.12:81')

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(2)
    expect(tableBody?.innerHTML).toContain('router-test1@docker')
    expect(tableBody?.innerHTML).toContain('router-test2@docker')
  })

  it('should not render used by routers table if the usedBy property is empty', async () => {
    const mockData = {
      status: 'enabled',
      usedBy: [],
      name: 'service-test3',
      provider: 'docker',
      type: 'loadbalancer',
      routers: [],
    }

    const { getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <TcpServiceRender name="mock-service" data={mockData as any} error={undefined} />,
    )

    expect(() => {
      getByTestId('routers-table')
    }).toThrow('Unable to find an element by: [data-testid="routers-table"]')
  })

  it('should render weighted services', async () => {
    const mockData = {
      weighted: {
        services: [
          {
            name: 'service1@docker',
            weight: 80,
          },
          {
            name: 'service2@kubernetes',
            weight: 20,
          },
        ],
      },
      status: 'enabled',
      usedBy: ['router-test1@docker'],
      name: 'weighted-service-test',
      provider: 'docker',
      type: 'weighted',
      routers: [
        {
          entryPoints: ['tcp'],
          service: 'weighted-service-test',
          rule: 'HostSNI(`*`)',
          status: 'enabled',
          using: ['tcp'],
          name: 'router-test1@docker',
          provider: 'docker',
        },
      ],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <TcpServiceRender name="mock-service" data={mockData as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'weighted-service-test')
    expect(titleTags.length).toBe(1)

    const serviceDetails = getByTestId('tcp-service-details')
    expect(serviceDetails.innerHTML).toContain('Type')
    expect(serviceDetails.innerHTML).toContain('weighted')
    expect(serviceDetails.innerHTML).toContain('Provider')
    expect(serviceDetails.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(serviceDetails.innerHTML).toContain('Status')
    expect(serviceDetails.innerHTML).toContain('Success')

    const weightedServices = getByTestId('tcp-weighted-services')
    expect(weightedServices.childNodes.length).toBe(2)
    expect(weightedServices.innerHTML).toContain('service1@docker')
    expect(weightedServices.innerHTML).toContain('80')
    expect(weightedServices.innerHTML).toContain('service2@kubernetes')
    expect(weightedServices.innerHTML).toContain('20')
    expect(weightedServices.querySelector('svg[data-testid="docker"]')).toBeTruthy()
  })
})
