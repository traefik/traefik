import { MiddlewareDetail } from 'components/middlewares/MiddlewareDetail'
import { renderWithProviders } from 'utils/test'

describe('<TcpMiddlewarePage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <MiddlewareDetail name="mock-middleware" data={undefined} error={new Error('Test error')} protocol="tcp" />,
      { route: '/tcp/middlewares/mock-middleware', withPage: true },
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <MiddlewareDetail name="mock-middleware" data={undefined} error={undefined} protocol="tcp" />,
      { route: '/tcp/middlewares/mock-middleware', withPage: true },
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <MiddlewareDetail name="mock-middleware" data={{} as Resource.DetailsData} error={undefined} protocol="tcp" />,
      { route: '/tcp/middlewares/mock-middleware', withPage: true },
    )
    expect(getByTestId('Not found page')).toBeInTheDocument()
  })

  it('should render a simple middleware', async () => {
    const mockData = {
      inFlightConn: {
        amount: 10,
      },
      status: 'enabled',
      usedBy: ['router-test-simple@docker'],
      name: 'middleware-simple',
      provider: 'docker',
      type: 'addprefix',
      routers: [
        {
          entryPoints: ['web-redirect'],
          middlewares: ['middleware-simple'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-redirect'],
          name: 'router-test-simple@docker',
          provider: 'docker',
        },
      ],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <MiddlewareDetail name="mock-middleware" data={mockData as any} error={undefined} protocol="tcp" />,
      { route: '/tcp/middlewares/middleware-simple', withPage: true },
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-simple')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(middlewareCard.innerHTML).toContain('Success')
    expect(container.innerHTML).toContain('inFlightConn')
    expect(container.innerHTML).toContain('amount')
    expect(container.innerHTML).toContain('10')

    const routersTable = getByTestId('routers-table')
    expect(routersTable.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(routersTable.innerHTML).toContain('router-test-simple@docker')
  })

  it('should render a complex middleware', async () => {
    const mockData = {
      name: 'middleware-complex',
      type: 'sample-middleware',
      status: 'enabled',
      provider: 'the-provider',
      usedBy: ['router-test-complex@docker'],
      inFlightConn: {
        amount: 10,
      },
      ipWhiteList: {
        sourceRange: ['125.0.0.1', '125.0.0.4'],
      },
      routers: [
        {
          entryPoints: ['web-redirect'],
          middlewares: ['middleware-complex'],
          service: 'api2_v2-example-beta1',
          rule: 'Host(`server`)',
          tls: {},
          status: 'enabled',
          using: ['web-redirect'],
          name: 'router-test-complex@docker',
          provider: 'docker',
        },
      ],
    }

    const { container, getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <MiddlewareDetail name="mock-middleware" data={mockData as any} error={undefined} protocol="tcp" />,
      { route: '/tcp/middlewares/middleware-complex', withPage: true },
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-complex')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.innerHTML).toContain('Success')
    expect(middlewareCard.innerHTML).toContain('the-provider')
    expect(container.innerHTML).toContain('inFlightConn')
    expect(container.innerHTML).toContain('amount')
    expect(container.innerHTML).toContain('10')
    expect(container.innerHTML).toContain('ipWhiteList')
    expect(container.innerHTML).toContain('source Range')
    expect(container.innerHTML).toContain('125.0.0.1')
    expect(container.innerHTML).toContain('125.0.0.4')

    const routersTable = getByTestId('routers-table')
    expect(routersTable.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(routersTable.innerHTML).toContain('router-test-complex@docker')
  })
})
