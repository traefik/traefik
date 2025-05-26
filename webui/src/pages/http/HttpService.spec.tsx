import { HttpServiceRender } from './HttpService'

import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { renderWithProviders } from 'utils/test'

describe('<HttpServicePage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <HttpServiceRender name="mock-service" data={undefined} error={new Error('Test error')} />,
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <HttpServiceRender name="mock-service" data={undefined} error={undefined} />,
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <HttpServiceRender name="mock-service" data={{} as ResourceDetailDataType} error={undefined} />,
    )
    expect(getByTestId('Not found page')).toBeInTheDocument()
  })

  it('should render a service with no health check or mirrors', async () => {
    const mockData = {
      loadBalancer: {
        servers: [
          {
            url: 'http://10.0.1.12:80',
          },
        ],
        passHostHeader: true,
      },
      status: 'enabled',
      usedBy: ['router-test1@docker', 'router-test2@docker'],
      serverStatus: {
        'http://10.0.1.12:80': 'UP',
      },
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

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpServiceRender name="mock-service" data={mockData as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'service-test1')
    expect(titleTags.length).toBe(1)

    const serviceDetails = getByTestId('service-details')
    expect(serviceDetails.innerHTML).toContain('Type')
    expect(serviceDetails.innerHTML).toContain('loadbalancer')
    expect(serviceDetails.innerHTML).toContain('Provider')
    expect(serviceDetails.innerHTML).toContain('docker')
    expect(serviceDetails.innerHTML).toContain('Status')
    expect(serviceDetails.innerHTML).toContain('Success')
    expect(serviceDetails.innerHTML).toContain('Pass Host Header')
    expect(serviceDetails.innerHTML).toContain('True')

    const serversList = getByTestId('servers-list')
    expect(serversList.childNodes.length).toBe(1)
    expect(serversList.innerHTML).toContain('http://10.0.1.12:80')

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(2)
    expect(tableBody?.innerHTML).toContain('router-test1@docker')
    expect(tableBody?.innerHTML).toContain('router-test2@docker')

    expect(() => {
      getByTestId('health-check')
    }).toThrow('Unable to find an element by: [data-testid="health-check"]')

    expect(() => {
      getByTestId('mirror-services')
    }).toThrow('Unable to find an element by: [data-testid="mirror-services"]')
  })

  it('should render a service with health check', async () => {
    const mockData = {
      loadBalancer: {
        servers: [
          {
            url: 'http://10.0.1.12:81',
          },
        ],
        passHostHeader: true,
        healthCheck: {
          scheme: 'https',
          path: '/health',
          port: 80,
          interval: '5s',
          timeout: '10s',
          hostname: 'domain.com',
          headers: {
            'X-Custom-A': 'foobar,gi,ji;ji,ok',
            'X-Custom-B': 'foobar foobar foobar foobar foobar',
          },
        },
      },
      status: 'enabled',
      usedBy: [],
      serverStatus: {
        'http://10.0.1.12:81': 'UP',
      },
      name: 'service-test2',
      provider: 'docker',
      type: 'loadbalancer',
      routers: [],
    }

    const { getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpServiceRender name="mock-service" data={mockData as any} error={undefined} />,
    )

    const healthCheck = getByTestId('health-check')
    expect(healthCheck.innerHTML).toContain('Scheme')
    expect(healthCheck.innerHTML).toContain('https')
    expect(healthCheck.innerHTML).toContain('Interval')
    expect(healthCheck.innerHTML).toContain('5s')
    expect(healthCheck.innerHTML).toContain('Path')
    expect(healthCheck.innerHTML).toContain('/health')
    expect(healthCheck.innerHTML).toContain('Timeout')
    expect(healthCheck.innerHTML).toContain('10s')
    expect(healthCheck.innerHTML).toContain('Port')
    expect(healthCheck.innerHTML).toContain('80')
    expect(healthCheck.innerHTML).toContain('Hostname')
    expect(healthCheck.innerHTML).toContain('domain.com')
    expect(healthCheck.innerHTML).toContain('Headers')
    expect(healthCheck.innerHTML).toContain('X-Custom-A: foobar,gi,ji;ji,ok')
    expect(healthCheck.innerHTML).toContain('X-Custom-B: foobar foobar foobar foobar foobar')

    expect(() => {
      getByTestId('mirror-services')
    }).toThrow('Unable to find an element by: [data-testid="mirror-services"]')
  })

  it('should render a service with mirror services', async () => {
    const mockData = {
      mirroring: {
        service: 'one@docker',
        mirrors: [
          {
            name: 'two@docker',
            percent: 10,
          },
          {
            name: 'three@docker',
            percent: 15,
          },
          {
            name: 'four@docker',
            percent: 80,
          },
        ],
      },
      status: 'enabled',
      usedBy: [],
      name: 'service-test3',
      provider: 'docker',
      type: 'mirroring',
      routers: [],
    }

    const { getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpServiceRender name="mock-service" data={mockData as any} error={undefined} />,
    )

    const mirrorServices = getByTestId('mirror-services')
    const providers = Array.from(mirrorServices.querySelectorAll('svg[data-testid="docker"]'))
    expect(mirrorServices.childNodes.length).toBe(3)
    expect(mirrorServices.innerHTML).toContain('two@docker')
    expect(mirrorServices.innerHTML).toContain('three@docker')
    expect(mirrorServices.innerHTML).toContain('four@docker')
    expect(mirrorServices.innerHTML).toContain('10')
    expect(mirrorServices.innerHTML).toContain('15')
    expect(mirrorServices.innerHTML).toContain('80')
    expect(providers.length).toBe(3)

    expect(() => {
      getByTestId('health-check')
    }).toThrow('Unable to find an element by: [data-testid="health-check"]')

    expect(() => {
      getByTestId('servers-list')
    }).toThrow('Unable to find an element by: [data-testid="servers-list"]')
  })
})
