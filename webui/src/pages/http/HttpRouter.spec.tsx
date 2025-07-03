import { HttpRouterRender } from './HttpRouter'

import { ResourceDetailDataType } from 'hooks/use-resource-detail'
import apiEntrypoints from 'mocks/data/api-entrypoints.json'
import apiHttpMiddlewares from 'mocks/data/api-http_middlewares.json'
import apiHttpRouters from 'mocks/data/api-http_routers.json'
import { renderWithProviders } from 'utils/test'

describe('<HttpRouterPage />', () => {
  it('should render the error message', () => {
    const { getByTestId } = renderWithProviders(
      <HttpRouterRender name="mock-router" data={undefined} error={new Error('Test error')} />,
    )
    expect(getByTestId('error-text')).toBeInTheDocument()
  })

  it('should render the skeleton', () => {
    const { getByTestId } = renderWithProviders(
      <HttpRouterRender name="mock-router" data={undefined} error={undefined} />,
    )
    expect(getByTestId('skeleton')).toBeInTheDocument()
  })

  it('should render the not found page', () => {
    const { getByTestId } = renderWithProviders(
      <HttpRouterRender name="mock-router" data={{} as ResourceDetailDataType} error={undefined} />,
    )
    expect(getByTestId('Not found page')).toBeInTheDocument()
  })

  it('should render the router details', async () => {
    const router = apiHttpRouters.find((x) => x.name === 'orphan-router@file')
    const mockData = {
      ...router!,
      middlewares: apiHttpMiddlewares.filter((x) => router?.middlewares?.includes(x.name)),
      hasValidMiddlewares: true,
      entryPointsData: apiEntrypoints.filter((x) => router?.using?.includes(x.name)),
    }

    const { getByTestId } = renderWithProviders(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      <HttpRouterRender name="mock-router" data={mockData as any} error={undefined} />,
    )

    const routerStructure = getByTestId('router-structure')
    expect(routerStructure.innerHTML).toContain(':80')
    expect(routerStructure.innerHTML).toContain(':443')
    expect(routerStructure.innerHTML).toContain(':8080')
    expect(routerStructure.innerHTML).toContain(':8002')
    expect(routerStructure.innerHTML).toContain(':8003')
    expect(routerStructure.innerHTML).toContain('orphan-router@file')
    expect(routerStructure.innerHTML).toContain('middleware00')
    expect(routerStructure.innerHTML).toContain('middleware01')
    expect(routerStructure.innerHTML).toContain('middleware02')
    expect(routerStructure.innerHTML).toContain('middleware03')
    expect(routerStructure.innerHTML).toContain('middleware04')
    expect(routerStructure.innerHTML).toContain('middleware05')
    expect(routerStructure.innerHTML).toContain('middleware06')
    expect(routerStructure.innerHTML).toContain('middleware07')
    expect(routerStructure.innerHTML).toContain('middleware08')
    expect(routerStructure.innerHTML).toContain('middleware09')
    expect(routerStructure.innerHTML).toContain('middleware10')
    expect(routerStructure.innerHTML).toContain('middleware11')
    expect(routerStructure.innerHTML).toContain('middleware12')
    expect(routerStructure.innerHTML).toContain('middleware13')
    expect(routerStructure.innerHTML).toContain('middleware14')
    expect(routerStructure.innerHTML).toContain('middleware15')
    expect(routerStructure.innerHTML).toContain('middleware16')
    expect(routerStructure.innerHTML).toContain('middleware17')
    expect(routerStructure.innerHTML).toContain('middleware18')
    expect(routerStructure.innerHTML).toContain('middleware19')
    expect(routerStructure.innerHTML).toContain('middleware20')
    expect(routerStructure.innerHTML).toContain('unexistingservice')
    expect(routerStructure.innerHTML).toContain('HTTP Router')
    expect(routerStructure.innerHTML).not.toContain('TCP Router')

    const routerDetailsSection = getByTestId('router-detail')

    const routerDetailsPanel = routerDetailsSection.querySelector(':scope > div:nth-child(1)')
    expect(routerDetailsPanel?.innerHTML).toContain('orphan-router@file')
    expect(routerDetailsPanel?.innerHTML).toContain('Error')
    expect(routerDetailsPanel?.querySelector('svg[data-testid="file"]')).toBeTruthy()
    expect(routerDetailsPanel?.innerHTML).toContain(
      'Path(`somethingreallyunexpectedbutalsoverylongitgetsoutofthecontainermaybe`)',
    )
    expect(routerDetailsPanel?.innerHTML).toContain('unexistingservice')
    expect(routerDetailsPanel?.innerHTML).toContain('the service "unexistingservice@file" does not exist')

    const middlewaresPanel = routerDetailsSection.querySelector(':scope > div:nth-child(3)')
    const providers = Array.from(middlewaresPanel?.querySelectorAll('svg[data-testid="docker"]') || [])
    expect(middlewaresPanel?.innerHTML).toContain('middleware00')
    expect(middlewaresPanel?.innerHTML).toContain('middleware01')
    expect(middlewaresPanel?.innerHTML).toContain('middleware02')
    expect(middlewaresPanel?.innerHTML).toContain('middleware03')
    expect(middlewaresPanel?.innerHTML).toContain('middleware04')
    expect(middlewaresPanel?.innerHTML).toContain('middleware05')
    expect(middlewaresPanel?.innerHTML).toContain('middleware06')
    expect(middlewaresPanel?.innerHTML).toContain('middleware07')
    expect(middlewaresPanel?.innerHTML).toContain('middleware08')
    expect(middlewaresPanel?.innerHTML).toContain('middleware09')
    expect(middlewaresPanel?.innerHTML).toContain('middleware10')
    expect(middlewaresPanel?.innerHTML).toContain('middleware11')
    expect(middlewaresPanel?.innerHTML).toContain('middleware12')
    expect(middlewaresPanel?.innerHTML).toContain('middleware13')
    expect(middlewaresPanel?.innerHTML).toContain('middleware14')
    expect(middlewaresPanel?.innerHTML).toContain('middleware15')
    expect(middlewaresPanel?.innerHTML).toContain('middleware16')
    expect(middlewaresPanel?.innerHTML).toContain('middleware17')
    expect(middlewaresPanel?.innerHTML).toContain('middleware18')
    expect(middlewaresPanel?.innerHTML).toContain('middleware19')
    expect(middlewaresPanel?.innerHTML).toContain('middleware20')
    expect(middlewaresPanel?.innerHTML).toContain('Success')
    expect(providers.length).toBe(21)

    expect(getByTestId('/http/middlewares/middleware00@docker')).toBeInTheDocument()

    expect(getByTestId('/http/middlewares/middleware01@docker')).toBeInTheDocument()

    expect(getByTestId('/http/services/unexistingservice@file')).toBeInTheDocument()
  })
})
