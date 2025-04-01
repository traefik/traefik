import Page from './Page'

import { renderWithProviders } from 'utils/test'

describe('<Page />', () => {
  it('should render an empty page', () => {
    const { getByTestId } = renderWithProviders(<Page title="Test" />)
    expect(getByTestId('Test page')).toBeInTheDocument()
  })
})
