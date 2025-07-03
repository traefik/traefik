import { TcpMiddlewareRender } from './TcpMiddleware'

import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import { renderWithProviders } from 'utils/test'

describe('<TcpMiddlewarePage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <TcpMiddlewareRender name="mock-middleware" data={undefined} error={new Error('Test error')} />,
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <TcpMiddlewareRender name="mock-middleware" data={undefined} error={undefined} />,
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <TcpMiddlewareRender name="mock-middleware" data={{} as ResourceDetailDataType} error={undefined} />,
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
      <TcpMiddlewareRender name="mock-middleware" data={mockData as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-simple')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.querySelector('svg[data-testid="docker"]')).toBeTruthy()
    expect(middlewareCard.innerHTML).toContain('Success')
    expect(middlewareCard.innerHTML).toContain('inFlightConn')
    expect(middlewareCard.innerHTML).toContain('amount')
    expect(middlewareCard.innerHTML).toContain('10')

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(tableBody?.innerHTML).toContain('router-test-simple@docker')
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
      <TcpMiddlewareRender name="mock-middleware" data={mockData as any} error={undefined} />,
    )

    const headings = Array.from(container.getElementsByTagName('h1'))
    const titleTags = headings.filter((h1) => h1.innerHTML === 'middleware-complex')
    expect(titleTags.length).toBe(1)

    const middlewareCard = getByTestId('middleware-card')
    expect(middlewareCard.innerHTML).toContain('Success')
    expect(middlewareCard.innerHTML).toContain('the-provider')
    expect(middlewareCard.innerHTML).toContain('inFlightConn')
    expect(middlewareCard.innerHTML).toContain('amount')
    expect(middlewareCard.innerHTML).toContain('10')
    expect(middlewareCard.innerHTML).toContain('ipWhiteList')
    expect(middlewareCard.innerHTML).toContain('source Range')
    expect(middlewareCard.innerHTML).toContain('125.0.0.1')
    expect(middlewareCard.innerHTML).toContain('125.0.0.4')

    const routersTable = getByTestId('routers-table')
    const tableBody = routersTable.querySelectorAll('div[role="rowgroup"]')[1]
    expect(tableBody?.querySelectorAll('a[role="row"]')).toHaveLength(1)
    expect(tableBody?.innerHTML).toContain('router-test-complex@docker')
  })
})
