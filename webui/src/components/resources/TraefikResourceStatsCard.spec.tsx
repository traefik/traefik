import TraefikResourceStatsCard from './TraefikResourceStatsCard'

import { renderWithProviders } from 'utils/test'

describe('<TraefikResourceStatsCard />', () => {
  it('should render the component and show the expected data (success count is zero)', () => {
    const { getByTestId } = renderWithProviders(
      <TraefikResourceStatsCard title="test" errors={2} total={5} warnings={3} linkTo="" />,
    )
    expect(getByTestId('success-pc').innerHTML).toContain('0%')
    expect(getByTestId('success-count').innerHTML).toContain('0')
    expect(getByTestId('warnings-pc').innerHTML).toContain('60%')
    expect(getByTestId('warnings-count').innerHTML).toContain('3')
    expect(getByTestId('errors-pc').innerHTML).toContain('40%')
    expect(getByTestId('errors-count').innerHTML).toContain('2')
  })

  it('should render the component and show the expected data (success count is not zero)', async () => {
    const { getByTestId } = renderWithProviders(
      <TraefikResourceStatsCard title="test" errors={2} total={7} warnings={4} linkTo="" />,
    )
    expect(getByTestId('success-pc').innerHTML).toContain('14%')
    expect(getByTestId('success-count').innerHTML).toContain('1')
    expect(getByTestId('warnings-pc').innerHTML).toContain('57%')
    expect(getByTestId('warnings-count').innerHTML).toContain('4')
    expect(getByTestId('errors-pc').innerHTML).toContain('29%')
    expect(getByTestId('errors-count').innerHTML).toContain('2')
  })

  it('should not render the component when everything is zero', async () => {
    const { getByTestId } = renderWithProviders(
      <TraefikResourceStatsCard title="test" errors={0} total={0} warnings={0} linkTo="" />,
    )
    expect(() => {
      getByTestId('success-pc')
    }).toThrow('Unable to find an element by: [data-testid="success-pc"]')
    expect(() => {
      getByTestId('success-count')
    }).toThrow('Unable to find an element by: [data-testid="success-count"]')
    expect(() => {
      getByTestId('warnings-pc')
    }).toThrow('Unable to find an element by: [data-testid="warnings-pc"]')
    expect(() => {
      getByTestId('warnings-count')
    }).toThrow('Unable to find an element by: [data-testid="warnings-count"]')
    expect(() => {
      getByTestId('errors-pc')
    }).toThrow('Unable to find an element by: [data-testid="errors-pc"]')
    expect(() => {
      getByTestId('errors-count')
    }).toThrow('Unable to find an element by: [data-testid="errors-count"]')
  })
})
