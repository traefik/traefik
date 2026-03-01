import Page from './Page'

import { renderWithProviders } from 'utils/test'

describe('<Page />', () => {
  it('should render an empty page', () => {
    const { getByTestId } = renderWithProviders(<Page />, { route: '/test' })
    expect(getByTestId('/test page')).toBeInTheDocument()
  })
})
