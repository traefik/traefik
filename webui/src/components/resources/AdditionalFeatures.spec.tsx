import AdditionalFeatures from './AdditionalFeatures'

import { renderWithProviders } from 'utils/test'

describe('<AdditionalFeatures />', () => {
  it('should render the middleware info', () => {
    renderWithProviders(<AdditionalFeatures uid="test-key" />)
  })

  it('should render the middleware info with number', () => {
    const middlewares: Middleware.Props[] = [
      {
        retry: {
          attempts: 2,
        },
      },
    ]

    const { container } = renderWithProviders(<AdditionalFeatures uid="test-key" middlewares={middlewares} />)

    expect(container.innerHTML).toContain('Retry: Attempts=2')
  })

  it('should render the middleware info with string', () => {
    const middlewares: Middleware.Props[] = [
      {
        circuitBreaker: {
          expression: 'expression',
        },
      },
    ]

    const { container } = renderWithProviders(<AdditionalFeatures uid="test-key" middlewares={middlewares} />)

    expect(container.innerHTML).toContain('CircuitBreaker: Expression="expression"')
  })

  it('should render the middleware info with string', () => {
    const middlewares: Middleware.Props[] = [
      {
        rateLimit: {
          burst: 100,
          average: 100,
        },
      },
    ]

    const { container } = renderWithProviders(<AdditionalFeatures uid="test-key" middlewares={middlewares} />)

    expect(container.innerHTML).toContain('RateLimit: Burst=100, Average=100')
  })
})
