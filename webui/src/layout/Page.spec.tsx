import { renderWithProviders } from 'utils/test'

import Page from './Page'

describe('<Page />', () => {
  it('should render an empty page', () => {
    const { getByTestId } = renderWithProviders(<Page title="Test" />)
    expect(getByTestId('Test page')).toBeInTheDocument()
  })
})
